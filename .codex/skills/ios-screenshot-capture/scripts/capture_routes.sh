#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Capture deterministic iOS Simulator screenshots from a route list.

Usage:
  capture_routes.sh --device <name|UDID|booted> --routes <file.tsv> --out-dir <dir> [options]

Required:
  --device <name|UDID|booted>   Simulator device name, UDID, or "booted".
  --routes <file.tsv>           TSV file: filename<TAB>url<TAB>wait_seconds
  --out-dir <dir>               Output directory for screenshots.

Options:
  --bundle-id <id>              Launch app before capture.
  --appearance <light|dark>     Simulator appearance (default: light).
  --status-bar                  Override status bar for deterministic capture.
  --time <HH:MM>                Status bar time when --status-bar is set (default: 9:41).
  --keep-status-bar             Keep status bar override after run.
  --boot-only                   Prepare simulator and exit without captures.
  --help                        Show this help.

Routes file format:
  - One row per screenshot.
  - Column 1: output filename (for example: 01-home.png)
  - Column 2: deep-link URL (for example: myapp://home)
  - Column 3: wait seconds before capture (optional, default: 2)
  - Lines starting with # are ignored.
EOF
}

require_bin() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

is_udid() {
  [[ "$1" =~ ^[0-9A-Fa-f-]{36}$ ]]
}

resolve_udid() {
  local input="$1"
  local list_json
  list_json="$(xcrun simctl list devices -j)"

  if [[ "$input" == "booted" ]]; then
    jq -r '[.devices[][] | select(.state == "Booted") | .udid] | first // empty' <<<"$list_json"
    return
  fi

  if is_udid "$input"; then
    echo "$input"
    return
  fi

  jq -r --arg name "$input" \
    '[.devices[][] | select(.name == $name and ((.isAvailable // true) == true)) | .udid] | first // empty' \
    <<<"$list_json"
}

device_state() {
  local udid="$1"
  local list_json
  list_json="$(xcrun simctl list devices -j)"
  jq -r --arg udid "$udid" '[.devices[][] | select(.udid == $udid) | .state] | first // empty' <<<"$list_json"
}

trim_trailing_cr() {
  local value="$1"
  printf '%s' "${value%$'\r'}"
}

DEVICE_INPUT=""
ROUTES_FILE=""
OUT_DIR=""
BUNDLE_ID=""
APPEARANCE="light"
STATUS_BAR=0
STATUS_BAR_TIME="9:41"
KEEP_STATUS_BAR=0
BOOT_ONLY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --device)
      DEVICE_INPUT="${2:-}"
      shift 2
      ;;
    --routes)
      ROUTES_FILE="${2:-}"
      shift 2
      ;;
    --out-dir)
      OUT_DIR="${2:-}"
      shift 2
      ;;
    --bundle-id)
      BUNDLE_ID="${2:-}"
      shift 2
      ;;
    --appearance)
      APPEARANCE="${2:-}"
      shift 2
      ;;
    --status-bar)
      STATUS_BAR=1
      shift
      ;;
    --time)
      STATUS_BAR_TIME="${2:-}"
      shift 2
      ;;
    --keep-status-bar)
      KEEP_STATUS_BAR=1
      shift
      ;;
    --boot-only)
      BOOT_ONLY=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$DEVICE_INPUT" ]]; then
  echo "--device is required" >&2
  usage
  exit 1
fi

if [[ -z "$OUT_DIR" ]]; then
  echo "--out-dir is required" >&2
  usage
  exit 1
fi

if [[ "$BOOT_ONLY" -eq 0 && -z "$ROUTES_FILE" ]]; then
  echo "--routes is required unless --boot-only is set" >&2
  usage
  exit 1
fi

if [[ "$APPEARANCE" != "light" && "$APPEARANCE" != "dark" ]]; then
  echo "--appearance must be 'light' or 'dark'" >&2
  exit 1
fi

require_bin xcrun
require_bin jq

if [[ "$BOOT_ONLY" -eq 0 && ! -f "$ROUTES_FILE" ]]; then
  echo "Routes file not found: $ROUTES_FILE" >&2
  exit 1
fi

UDID="$(resolve_udid "$DEVICE_INPUT")"
if [[ -z "$UDID" ]]; then
  echo "Unable to resolve simulator: $DEVICE_INPUT" >&2
  exit 1
fi

cleanup() {
  if [[ "$STATUS_BAR" -eq 1 && "$KEEP_STATUS_BAR" -eq 0 ]]; then
    set +e
    xcrun simctl status_bar "$UDID" clear >/dev/null 2>&1
  fi
}
trap cleanup EXIT

STATE="$(device_state "$UDID")"
if [[ "$STATE" != "Booted" ]]; then
  xcrun simctl boot "$UDID"
fi
xcrun simctl bootstatus "$UDID" -b
xcrun simctl ui "$UDID" appearance "$APPEARANCE"

if [[ "$STATUS_BAR" -eq 1 ]]; then
  xcrun simctl status_bar "$UDID" clear
  xcrun simctl status_bar "$UDID" override \
    --time "$STATUS_BAR_TIME" \
    --dataNetwork wifi \
    --wifiMode active \
    --wifiBars 3 \
    --cellularMode active \
    --cellularBars 4 \
    --batteryState charged \
    --batteryLevel 100
fi

if [[ -n "$BUNDLE_ID" ]]; then
  xcrun simctl launch --terminate-running-process "$UDID" "$BUNDLE_ID" >/dev/null
fi

if [[ "$BOOT_ONLY" -eq 1 ]]; then
  echo "Simulator prepared: $UDID"
  exit 0
fi

mkdir -p "$OUT_DIR"

captured=0
while IFS=$'\t' read -r raw_filename raw_url raw_wait || [[ -n "${raw_filename:-}" || -n "${raw_url:-}" || -n "${raw_wait:-}" ]]; do
  filename="$(trim_trailing_cr "${raw_filename:-}")"
  url="$(trim_trailing_cr "${raw_url:-}")"
  wait_secs="$(trim_trailing_cr "${raw_wait:-}")"

  if [[ -z "${filename// }" ]]; then
    continue
  fi
  if [[ "${filename:0:1}" == "#" ]]; then
    continue
  fi
  if [[ -z "${url// }" ]]; then
    echo "Missing URL for filename: $filename" >&2
    exit 1
  fi
  if [[ -z "${wait_secs// }" ]]; then
    wait_secs="2"
  fi

  xcrun simctl openurl "$UDID" "$url"
  sleep "$wait_secs"

  target="$OUT_DIR/$filename"
  mkdir -p "$(dirname "$target")"
  xcrun simctl io "$UDID" screenshot "$target" >/dev/null
  echo "Captured: $target"
  captured=$((captured + 1))
done < "$ROUTES_FILE"

echo "Completed. Captured $captured screenshot(s)."
