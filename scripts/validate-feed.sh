#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
package_root="$repo_root/packages"

if [[ ! -d "$package_root" ]]; then
  echo "missing packages directory" >&2
  exit 1
fi

package_count=0
while IFS= read -r -d '' package_dir; do
  package_count=$((package_count + 1))
  if [[ ! -f "$package_dir/Makefile" ]]; then
    echo "missing Makefile: $package_dir" >&2
    exit 1
  fi
done < <(find "$package_root" -mindepth 1 -maxdepth 1 -type d -print0)

if [[ "$package_count" -eq 0 ]]; then
  echo "no package directories found" >&2
  exit 1
fi

if find "$repo_root" -type f -size +50M -print -quit | grep -q .; then
  echo "file larger than 50 MiB found" >&2
  find "$repo_root" -type f -size +50M -print >&2
  exit 1
fi

if find "$repo_root" -mindepth 2 -name .git -print -quit | grep -q .; then
  echo "nested .git directory found" >&2
  exit 1
fi

if grep -RIlE 'ghp_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]+' "$repo_root" \
  --exclude-dir=.git | grep -q .; then
  echo "possible GitHub credential found" >&2
  exit 1
fi

echo "validated $package_count package directories"

