---
name: ios-screenshot-capture
description: Capture clean and reproducible iOS app screenshots in Xcode Simulator for App Store submission, QA evidence, and release documentation. Use when Codex needs to take iOS screenshots, batch-capture screens from deep links/routes, normalize status bar and appearance, validate final dimensions, or troubleshoot simulator screenshot issues for any iOS project.
---

# iOS Screenshot Capture

Capture deterministic iOS screenshots from Simulator with repeatable commands and a reusable batch script. Apply this skill to any iOS app stack (SwiftUI, UIKit, React Native, Flutter, or Capacitor).

## Workflow

1. Discover target simulator and app bundle id.
   ```bash
   xcrun simctl list devices
   xcrun simctl listapps <UDID> | rg -n "CFBundleIdentifier|CFBundleDisplayName|<app keyword>"
   ```

2. Prepare a clean capture state.
   ```bash
   xcrun simctl boot "<UDID>"
   xcrun simctl bootstatus "<UDID>" -b
   xcrun simctl ui "<UDID>" appearance light
   xcrun simctl launch --terminate-running-process "<UDID>" "<BUNDLE_ID>"
   ```

3. Capture a single screenshot.
   ```bash
   xcrun simctl openurl "<UDID>" "<APP_SCHEME_URL>"
   sleep 2
   xcrun simctl io "<UDID>" screenshot "<OUTPUT_PATH>.png"
   ```

4. Capture a batch using `scripts/capture_routes.sh`.
   - Create a TSV file in the format `filename<TAB>url<TAB>wait_seconds`.
   - Run:
     ```bash
     bash scripts/capture_routes.sh \
       --device "<UDID-or-Device-Name>" \
       --bundle-id "<BUNDLE_ID>" \
       --routes "<routes.tsv>" \
       --out-dir "<output-dir>" \
       --status-bar \
       --appearance light
     ```

5. Validate each output image dimension.
   ```bash
   sips -g pixelWidth -g pixelHeight "<output-file>.png"
   ```
   - If targeting App Store submission, verify accepted sizes in App Store Connect and use `references/app-store-specs.md` as a quick checklist.

## Route File Format

Read `references/routes-format.md` before first run. Keep one route per line and use comments (`#`) for notes.

## Quality Rules

- Capture app UI only. Avoid simulator chrome, dev menus, debug overlays, keyboard popups, and permission alerts unless intentionally documenting them.
- Keep appearance, status bar, locale, and font size consistent across one screenshot set.
- Use realistic in-app data and states for release screenshots.
- Keep one canonical output directory per run and remove temporary captures.

## Troubleshooting

- If `openurl` fails, verify app scheme and test one route manually.
- If target content is still loading, increase `wait_seconds` for that route.
- If status bar overrides are stale:
  ```bash
  xcrun simctl status_bar "<UDID>" clear
  ```
- If simulator state is broken:
  ```bash
  xcrun simctl shutdown "<UDID>"
  xcrun simctl erase "<UDID>"
  ```
