#!/usr/bin/env bash
set -euo pipefail

: "${OPENWRT_VERSION:?missing OPENWRT_VERSION}"
: "${CHANNEL:?missing CHANNEL}"
: "${SDK_URL:?missing SDK_URL}"
: "${SDK_SHA256:?missing SDK_SHA256}"
: "${PACKAGE_FORMAT:?missing PACKAGE_FORMAT}"

arch="${PACKAGE_ARCH:-aarch64_generic}"
repo_root="${GITHUB_WORKSPACE:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
runner_temp="${RUNNER_TEMP:-/tmp}"
shard_root="${SHARD_ROOT:-$repo_root/shard-output}"
work_root="$runner_temp/souhana-finalize-$OPENWRT_VERSION"
archive="$work_root/sdk.tar.zst"
extract_root="$work_root/extracted"

mkdir -p "$work_root" "$extract_root"
curl --fail --location --retry 5 --retry-delay 3 "$SDK_URL" -o "$archive"
printf '%s  %s\n' "$SDK_SHA256" "$archive" | sha256sum --check -
tar --zstd -xf "$archive" -C "$extract_root"

sdk_root="$(find "$extract_root" -mindepth 1 -maxdepth 1 -type d \
  -name 'openwrt-sdk-*' -print -quit)"
if [[ -z "$sdk_root" ]]; then
  echo "OpenWrt SDK directory not found" >&2
  exit 1
fi

cd "$sdk_root"
printf 'CONFIG_SIGNED_PACKAGES=y\n' >> .config
printf 'src-link souhana %s/packages\n' "$repo_root" > feeds.conf

if [[ "$PACKAGE_FORMAT" == "apk" ]]; then
  : "${APK_PRIVATE_KEY:?missing APK_PRIVATE_KEY secret}"
  umask 077
  printf '%s' "$APK_PRIVATE_KEY" > private-key.pem
  cp "$repo_root/keys/25.12.pem" public-key.pem
  package_pattern='*.apk'
else
  : "${OPKG_PRIVATE_KEY:?missing OPKG_PRIVATE_KEY secret}"
  umask 077
  printf '%s' "$OPKG_PRIVATE_KEY" > key-build
  cp "$repo_root/keys/24.10.pub" key-build.pub
  package_pattern='*.ipk'
fi

# OpenWrt rejects package/index operations under a restrictive umask because
# the resulting repository metadata would not be readable by all users.
umask 022
make defconfig

source_package_dir="$shard_root/$CHANNEL/$arch/packages"
if [[ ! -d "$source_package_dir" ]]; then
  echo "merged shard package directory not found: $source_package_dir" >&2
  exit 1
fi

feed_dir="$sdk_root/bin/packages/$arch/souhana"
mkdir -p "$feed_dir"
find "$source_package_dir" -maxdepth 1 -type f -name "$package_pattern" \
  -exec cp -a {} "$feed_dir/" \;

package_count="$(find "$feed_dir" -maxdepth 1 -type f -name "$package_pattern" | wc -l)"
if (( package_count == 0 )); then
  echo "no package files were produced by the build shards" >&2
  exit 1
fi

make package/index V=s

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

report_dir="$shard_root/$CHANNEL/$arch/reports"
partial_shards=0
if [[ -d "$report_dir" ]]; then
  partial_shards="$(awk '
    /Build status: partial/ { partial[FILENAME] = 1 }
    END { print length(partial) }
  ' "$report_dir"/*.txt)"
fi

{
  printf 'OpenWrt: %s\n' "$OPENWRT_VERSION"
  printf 'Channel: %s\n' "$CHANNEL"
  printf 'Published package files: %s\n' "$package_count"
  printf 'Partial shards: %s\n' "$partial_shards"
  printf 'Package list:\n'
  find "$feed_dir" -maxdepth 1 -type f -name "$package_pattern" \
    -printf '  %f\n' | sort
  if [[ -d "$report_dir" ]]; then
    printf '\nShard reports:\n'
    for report in "$report_dir"/*.txt; do
      printf '\n--- %s ---\n' "$(basename "$report")"
      cat "$report"
    done
  fi
} > "$output_dir/build-report.txt"

echo "published $package_count package files in $output_dir"
