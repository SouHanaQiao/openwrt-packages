#!/usr/bin/env bash
set -euo pipefail

: "${OPENWRT_VERSION:?missing OPENWRT_VERSION}"
: "${CHANNEL:?missing CHANNEL}"
: "${SDK_URL:?missing SDK_URL}"
: "${SDK_SHA256:?missing SDK_SHA256}"
: "${OPENWRT_REF:?missing OPENWRT_REF}"
: "${FEEDS_BUILDINFO_URL:?missing FEEDS_BUILDINFO_URL}"
: "${PACKAGE_FORMAT:?missing PACKAGE_FORMAT}"

shard_index="${SHARD_INDEX:-0}"
shard_count="${SHARD_COUNT:-1}"
if ! [[ "$shard_index" =~ ^[0-9]+$ && "$shard_count" =~ ^[1-9][0-9]*$ ]] ||
   (( shard_index >= shard_count )); then
  echo "invalid shard selection: index=$shard_index count=$shard_count" >&2
  exit 1
fi

repo_root="${GITHUB_WORKSPACE:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
runner_temp="${RUNNER_TEMP:-/tmp}"
work_root="$runner_temp/souhana-openwrt-$OPENWRT_VERSION"
archive="$work_root/sdk.tar.zst"
extract_root="$work_root/extracted"

mkdir -p "$work_root" "$extract_root"
curl --fail --location --retry 5 --retry-delay 3 "$SDK_URL" -o "$archive"
printf '%s  %s\n' "$SDK_SHA256" "$archive" | sha256sum --check -
tar --zstd -xf "$archive" -C "$extract_root"

sdk_root="$(find "$extract_root" -mindepth 1 -maxdepth 1 -type d -name 'openwrt-sdk-*' -print -quit)"
if [[ -z "$sdk_root" ]]; then
  echo "OpenWrt SDK directory not found" >&2
  exit 1
fi

cd "$sdk_root"
feeds_buildinfo="$work_root/feeds.buildinfo"
curl --fail --location --retry 5 --retry-delay 3 \
  "$FEEDS_BUILDINFO_URL" -o "$feeds_buildinfo"
grep -Eq '^src-git packages https://git\.openwrt\.org/feed/packages\.git\^[0-9a-f]{40}$' \
  "$feeds_buildinfo"
{
  printf 'src-link souhana %s/packages\n' "$repo_root"
  printf 'src-git base https://git.openwrt.org/openwrt/openwrt.git^%s\n' "$OPENWRT_REF"
  cat "$feeds_buildinfo"
} > feeds.conf.default
./scripts/feeds update -a
./scripts/feeds install -a -p souhana -d m

go_root="$(go env GOROOT)"
printf 'CONFIG_GOLANG_EXTERNAL_BOOTSTRAP_ROOT="%s"\n' "$go_root" >> .config
printf 'CONFIG_SIGNED_PACKAGES=y\n' >> .config

# Ruby enables YJIT by default on aarch64 in OpenWrt 24.10. That optional
# feature pulls in rust/host, whose pinned CI LLVM bootstrap is no longer
# available upstream. OpenClash only needs Ruby itself, not YJIT.
if [[ "$OPENWRT_VERSION" == "24.10.1" ]]; then
  printf '# CONFIG_RUBY_ENABLE_YJIT is not set\n' >> .config
  # The historical SmartDNS source also defines an optional Rust dashboard.
  # LuCI uses the normal smartdns package and does not require this dashboard.
  printf '# CONFIG_PACKAGE_smartdns-ui is not set\n' >> .config
fi

if [[ "$PACKAGE_FORMAT" == "apk" ]]; then
  : "${APK_PRIVATE_KEY:?missing APK_PRIVATE_KEY secret}"
  umask 077
  printf '%s' "$APK_PRIVATE_KEY" > private-key.pem
  cp "$repo_root/keys/25.12.pem" public-key.pem
else
  : "${OPKG_PRIVATE_KEY:?missing OPKG_PRIVATE_KEY secret}"
  umask 077
  printf '%s' "$OPKG_PRIVATE_KEY" > key-build
  cp "$repo_root/keys/24.10.pub" key-build.pub
fi

make defconfig

if [[ "$OPENWRT_VERSION" == "24.10.1" ]] && grep -q '^CONFIG_RUBY_ENABLE_YJIT=y' .config; then
  echo "RUBY_ENABLE_YJIT must remain disabled for the OpenWrt 24.10 build" >&2
  exit 1
fi
if [[ "$OPENWRT_VERSION" == "24.10.1" ]] && grep -q '^CONFIG_PACKAGE_smartdns-ui=' .config; then
  echo "The optional SmartDNS Rust dashboard must remain disabled for OpenWrt 24.10" >&2
  exit 1
fi
mapfile -t all_custom_package_dirs < <(
  find "$repo_root/packages" -mindepth 1 -maxdepth 1 -type d -print | sort
)
if (( ${#all_custom_package_dirs[@]} == 0 )); then
  echo "no custom package directories found" >&2
  exit 1
fi

custom_package_dirs=()
if [[ -n "${PACKAGE_DIRS:-}" ]]; then
  read -r -a requested_package_names <<< "$PACKAGE_DIRS"
  for package_name in "${requested_package_names[@]}"; do
    if ! [[ "$package_name" =~ ^[A-Za-z0-9._+-]+$ ]] ||
       [[ ! -d "$repo_root/packages/$package_name" ]]; then
      echo "invalid requested package directory: $package_name" >&2
      exit 1
    fi
    custom_package_dirs+=("$repo_root/packages/$package_name")
  done
else
  for package_index in "${!all_custom_package_dirs[@]}"; do
    if (( package_index % shard_count == shard_index )); then
      custom_package_dirs+=("${all_custom_package_dirs[$package_index]}")
    fi
  done
fi
if (( ${#custom_package_dirs[@]} == 0 )); then
  echo "shard contains no custom package directories" >&2
  exit 1
fi

custom_targets=()
for package_dir in "${custom_package_dirs[@]}"; do
  package_name="$(basename "$package_dir")"
  if [[ ! -e "package/feeds/souhana/$package_name" ]]; then
    echo "custom package was not installed from souhana feed: $package_name" >&2
    exit 1
  fi
  custom_targets+=("package/feeds/souhana/$package_name/compile")
done

printf 'building shard %s/%s with %s custom targets:\n' \
  "$shard_index" "$shard_count" "${#custom_targets[@]}"
printf '  %s\n' "${custom_targets[@]}"

# Keep compiling independent packages when one historical package is not yet
# compatible with the selected SDK. This lets the signed feed publish every
# successful package while the build log and report retain the partial status.
set +e
GOFLAGS=-buildvcs=false GOPROXY=https://proxy.golang.org,direct \
  make -k -j"${BUILD_JOBS:-2}" "${custom_targets[@]}" V=s
build_status=$?
set -e

if (( build_status != 0 )); then
  echo "::warning::Some historical package targets failed; publishing all successful packages"
fi

arch="$(sed -n 's/^CONFIG_TARGET_ARCH_PACKAGES="\(.*\)"/\1/p' .config)"
if [[ -z "$arch" ]]; then
  arch="aarch64_generic"
fi

feed_dir="$sdk_root/bin/packages/$arch/souhana"

if (( shard_count > 1 )); then
  output_dir="$repo_root/shard-output/$CHANNEL/$arch"
  package_output_dir="$output_dir/packages"
  report_dir="$output_dir/reports"
  mkdir -p "$package_output_dir" "$report_dir"

  if [[ -d "$feed_dir" ]]; then
    find "$feed_dir" -maxdepth 1 -type f \( -name '*.apk' -o -name '*.ipk' \) \
      -exec cp -a {} "$package_output_dir/" \;
  fi

  report_name="${REPORT_NAME:-shard-$shard_index}"
  if ! [[ "$report_name" =~ ^[A-Za-z0-9._-]+$ ]]; then
    echo "invalid report name: $report_name" >&2
    exit 1
  fi

  {
    printf 'OpenWrt: %s\n' "$OPENWRT_VERSION"
    printf 'Channel: %s\n' "$CHANNEL"
    printf 'Shard: %s/%s\n' "$shard_index" "$shard_count"
    if (( build_status == 0 )); then
      printf 'Build status: complete\n'
    else
      printf 'Build status: partial (see GitHub Actions log)\n'
    fi
    printf 'Total source directories: %d\n' "${#all_custom_package_dirs[@]}"
    printf 'Shard source directories: %d\n' "${#custom_package_dirs[@]}"
    printf 'Selected source directories:\n'
    printf '  %s\n' "${custom_package_dirs[@]##*/}"
    printf 'Produced package files:\n'
    find "$package_output_dir" -maxdepth 1 -type f \
      -printf '  %f\n' | sort
  } > "$report_dir/$report_name.txt"

  echo "published shard output in $output_dir"
  exit 0
fi

make package/index

if [[ ! -d "$feed_dir" ]]; then
  echo "custom feed output not found: $feed_dir" >&2
  find "$sdk_root/bin/packages" -maxdepth 3 -type d -print >&2 || true
  exit 1
fi

output_dir="$repo_root/site/$CHANNEL/$arch"
mkdir -p "$output_dir"
cp -a "$feed_dir/." "$output_dir/"

if [[ "$PACKAGE_FORMAT" == "apk" ]]; then
  cp "$repo_root/keys/25.12.pem" "$output_dir/public-key.pem"
  test -s "$output_dir/packages.adb"
else
  cp "$repo_root/keys/24.10.pub" "$output_dir/key-build.pub"
  test -s "$output_dir/Packages.gz"
  test -s "$output_dir/Packages.sig"
fi

{
  printf 'OpenWrt: %s\n' "$OPENWRT_VERSION"
  printf 'Channel: %s\n' "$CHANNEL"
  if (( build_status == 0 )); then
    printf 'Build status: complete\n'
  else
    printf 'Build status: partial (see GitHub Actions log)\n'
  fi
  printf 'Source directories: %d\n' "${#custom_package_dirs[@]}"
  printf 'Published package files:\n'
  find "$output_dir" -maxdepth 1 -type f \( -name '*.apk' -o -name '*.ipk' \) \
    -printf '  %f\n' | sort
} > "$output_dir/build-report.txt"

echo "published build output in $output_dir"
