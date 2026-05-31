#!/bin/bash
set -euo pipefail

# Usage: generate-changelog.sh <output_file> [tag_or_range]
# If tag_or_range is empty, generates from latest tag to HEAD

OUTPUT_FILE="$1"
TAG_OR_RANGE="${2:-}"

normalize_changelog_type() {
  case "${1,,}" in
    feat|feature)
      echo "feature"
      ;;
    fix)
      echo "fix"
      ;;
    misc)
      echo "misc"
      ;;
    *)
      echo ""
      ;;
  esac
}

extract_changelog_type() {
  local commit_text="$1"
  local trailer_value=""

  trailer_value=$(printf '%s\n' "$commit_text" |
    git interpret-trailers --parse |
    awk 'BEGIN{IGNORECASE=1} /^[[:space:]]*changelog[[:space:]]*:/ {sub(/^[^:]*:[[:space:]]*/, "", $0); print tolower($0); exit}')

  if [[ -n "$trailer_value" ]]; then
    normalize_changelog_type "$trailer_value"
    return
  fi

  if printf '%s\n' "$commit_text" | grep -iqE 'changelog[[:space:]]*\((feature|feat)\)'; then
    echo "feature"
  elif printf '%s\n' "$commit_text" | grep -iqE 'changelog[[:space:]]*\((fix)\)'; then
    echo "fix"
  elif printf '%s\n' "$commit_text" | grep -iqE 'changelog[[:space:]]*\((misc)\)'; then
    echo "misc"
  elif printf '%s\n' "$commit_text" | grep -iqE '^[[:space:]]*changelog[[:space:]]*:[[:space:]]*(feature|feat)[[:space:]]*$'; then
    echo "feature"
  elif printf '%s\n' "$commit_text" | grep -iqE '^[[:space:]]*changelog[[:space:]]*:[[:space:]]*(fix)[[:space:]]*$'; then
    echo "fix"
  elif printf '%s\n' "$commit_text" | grep -iqE '^[[:space:]]*changelog[[:space:]]*:[[:space:]]*(misc)[[:space:]]*$'; then
    echo "misc"
  else
    echo ""
  fi
}

extract_repo_slug() {
  if [[ -n "${GITHUB_REPOSITORY:-}" ]]; then
    echo "$GITHUB_REPOSITORY"
    return
  fi

  local remote_url=""
  remote_url=$(git config --get remote.origin.url 2>/dev/null || true)
  if [[ -z "$remote_url" ]]; then
    return
  fi

  if [[ "$remote_url" =~ ^git@github\.com:([^/]+/[^/.]+)(\.git)?$ ]]; then
    echo "${BASH_REMATCH[1]}"
  elif [[ "$remote_url" =~ ^https://github\.com/([^/]+/[^/.]+)/?(\.git)?$ ]]; then
    echo "${BASH_REMATCH[1]}"
  fi
}

resolve_github_login() {
  local commit_hash="$1"
  local repo_slug="$2"
  local author_name="$3"
  local api_url="${GITHUB_API_URL:-https://api.github.com}"
  local login=""

  if [[ -n "$repo_slug" ]] && command -v curl >/dev/null 2>&1; then
    local auth_args=()
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
      auth_args=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
    elif [[ -n "${GH_TOKEN:-}" ]]; then
      auth_args=(-H "Authorization: Bearer ${GH_TOKEN}")
    fi

    local response=""
    response=$(curl -fsSL "${auth_args[@]}" -H "Accept: application/vnd.github+json" "$api_url/repos/$repo_slug/commits/$commit_hash" 2>/dev/null || true)
    if [[ -n "$response" ]]; then
      local python_bin=""
      python_bin=$(command -v python3 2>/dev/null || command -v python 2>/dev/null || true)
      if [[ -n "$python_bin" ]]; then
        login=$(printf '%s' "$response" | "$python_bin" -c 'import json,sys; data=json.load(sys.stdin); print(((data.get("author") or {}).get("login")) or "")' 2>/dev/null || true)
      fi
    fi
  fi

  if [[ -z "$login" && "$author_name" =~ ^[A-Za-z0-9][A-Za-z0-9-]{0,38}$ ]]; then
    login="$author_name"
  fi

  echo "$login"
}

build_thanks_line() {
  local commit_hash="$1"
  local repo_slug="$2"
  local repo_owner="$3"
  local author_name="$4"
  local author_email="$5"
  local committer_name="$6"
  local committer_email="$7"

  local author_login=""
  author_login=$(resolve_github_login "$commit_hash" "$repo_slug" "$author_name")

  local is_third_party="false"
  if [[ -n "$author_login" && -n "$repo_owner" ]]; then
    if [[ "${author_login,,}" != "${repo_owner,,}" ]]; then
      is_third_party="true"
    fi
  elif [[ "${author_name,,}" != "${committer_name,,}" || "${author_email,,}" != "${committer_email,,}" ]]; then
    is_third_party="true"
  fi

  if [[ "$is_third_party" == "true" && -n "$author_login" ]]; then
    printf '    thanks to @%s' "$author_login"
  fi
}

mkdir -p "$(dirname "$OUTPUT_FILE")"

if [[ -z "$TAG_OR_RANGE" ]]; then
  # Dev version - from latest tag to HEAD
  LATEST_TAG=$(git describe --tags --abbrev=0 --match "v*.*.*" 2>/dev/null || echo "")
  if [[ -z "$LATEST_TAG" ]]; then
    LOG_RANGE=""
    CONTEXT_MSG="All unreleased changes:"
    VERSION_TITLE="Development Version"
  else
    LOG_RANGE="$LATEST_TAG..HEAD"
    CONTEXT_MSG="Changes since $LATEST_TAG:"
    VERSION_TITLE="Development Version"
  fi
else
  # Tagged version
  VERSION="$TAG_OR_RANGE"
  PREVIOUS_TAG=$(git describe --tags --abbrev=0 --match "v*.*.*" "$TAG_OR_RANGE^" 2>/dev/null || echo "")
  if [[ -z "$PREVIOUS_TAG" ]]; then
    LOG_RANGE="$TAG_OR_RANGE"
    CONTEXT_MSG="All changes in this release:"
  else
    LOG_RANGE="$PREVIOUS_TAG..$TAG_OR_RANGE"
    CONTEXT_MSG="Changes since $PREVIOUS_TAG:"
  fi
  VERSION_TITLE="Version ${VERSION#v}"
fi

# Collect commits
mapfile -t COMMITS < <(git log --pretty=format:'%H' $LOG_RANGE)
FEATURES=""
FIXES=""
MISC=""
REPO_SLUG=$(extract_repo_slug)
REPO_OWNER="${REPO_SLUG%%/*}"

for commit_hash in "${COMMITS[@]}"; do
  commit_msg=$(git log -1 --pretty=format:'%s' "$commit_hash")
  commit_body=$(git log -1 --pretty=format:'%b' "$commit_hash")
  commit_text=$(git log -1 --pretty=format:'%B' "$commit_hash")
  author_name=$(git log -1 --pretty=format:'%an' "$commit_hash")
  author_email=$(git log -1 --pretty=format:'%ae' "$commit_hash")
  committer_name=$(git log -1 --pretty=format:'%cn' "$commit_hash")
  committer_email=$(git log -1 --pretty=format:'%ce' "$commit_hash")
  changelog_type=$(extract_changelog_type "$commit_text")
  if [ -n "$changelog_type" ]; then
    body_content=$(echo "$commit_body" | awk 'BEGIN{IGNORECASE=1} !/changelog[(: ]/ && !/^[[:space:]]*co-authored-by:[[:space:]]*/ && NF')
    thanks_line=$(build_thanks_line "$commit_hash" "$REPO_SLUG" "$REPO_OWNER" "$author_name" "$author_email" "$committer_name" "$committer_email")
    entry="- $commit_msg"
    if [ -n "$body_content" ]; then
      entry=$(printf "%s\n%s" "$entry" "$(echo "$body_content" | sed 's/^/  /')")
    fi
    if [ -n "$thanks_line" ]; then
      entry=$(printf "%s  \n%s" "$entry" "$thanks_line")
    fi
    if [ "$changelog_type" = "feature" ]; then
      FEATURES=$(printf "%s\n%s" "$FEATURES" "$entry")
    elif [ "$changelog_type" = "fix" ]; then
      FIXES=$(printf "%s\n%s" "$FIXES" "$entry")
    else
      MISC=$(printf "%s\n%s" "$MISC" "$entry")
    fi
  fi
done

{
  echo ""
  echo "$CONTEXT_MSG"
  echo ""
  
  if [[ -n "$FEATURES" ]]; then
    echo "### ✨ New Features"
    echo ""
    echo "$FEATURES"
    echo ""
  fi
  
  if [[ -n "$FIXES" ]]; then
    echo "### 🐛 Fixes"
    echo ""
    echo "$FIXES"
    echo ""
  fi
  
  if [[ -n "$MISC" ]]; then
    echo "### 🔧 Miscellaneous"
    echo ""
    echo "$MISC"
    echo ""
  fi
  
  if [[ -z "$FEATURES" && -z "$FIXES" && -z "$MISC" ]]; then
    if [[ -z "$TAG_OR_RANGE" ]]; then
      echo "No changes yet."
    else
      echo "No changes."
    fi
  fi
} > "$OUTPUT_FILE"

echo "Changelog generated: $OUTPUT_FILE"
