# App Store Screenshot Checks

Use this file as a quick pre-upload checklist. Always trust the latest sizes shown in App Store Connect because accepted dimensions can change.

## Current Common iPhone Sizes

Common accepted portrait/landscape pairs include:
- `1284 x 2778` or `2778 x 1284`
- `1242 x 2688` or `2688 x 1242`

## Quick Verification Commands

```bash
sips -g pixelWidth -g pixelHeight <image.png>
```

Batch check:

```bash
for f in *.png; do
  printf "%s: " "$f"
  sips -g pixelWidth -g pixelHeight "$f" 2>/dev/null | awk '/pixelWidth|pixelHeight/{printf $2" "} END{print ""}'
done
```

## Content Quality Checklist

- Remove debug overlays, dev menus, inspector panels, and permission popups.
- Keep status bar stable across the set (for example `9:41`).
- Keep typography and language consistent within the same locale set.
- Ensure screenshots represent real, currently shipped functionality.
- Do not expose private user data.
