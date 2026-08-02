#!/usr/bin/env bash
set -euo pipefail

: "${OPENWRT_VERSION:?missing OPENWRT_VERSION}"
: "${CHANNEL:?missing CHANNEL}"
: "${SDK_URL:?missing SDK_URL}"
: "${SDK_SHA256:?missing SDK_SHA256}"
: "${OPENWRT_REF:?missing OPENWRT_REF}"
: "${FEEDS_BUILDINFO_URL:?missing FEEDS_BUILDINFO_URL}"
: "${PACKAGE_FORMAT:?missing PACKAGE_FORMAT}"

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
mapfile -t custom_package_dirs < <(
  find "$repo_root/packages" -mindepth 1 -maxdepth 1 -type d -print | sort
)
if (( ${#custom_package_dirs[@]} == 0 )); then
  echo "no custom package directories found" >&2
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

printf 'building custom targets:\n'
printf '  %s\n' "${custom_targets[@]}"
GOFLAGS=-buildvcs=false GOPROXY=https://proxy.golang.org,direct \
  make -j"${BUILD_JOBS:-2}" "${custom_targets[@]}" V=s
make package/index

arch="$(sed -n 's/^CONFIG_TARGET_ARCH_PACKAGES="\(.*\)"/\1/p' .config)"
if [[ -z "$arch" ]]; then
  arch="aarch64_generic"
fi

feed_dir="$sdk_root/bin/packages/$arch/souhana"
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

echo "published build output in $output_dir"
