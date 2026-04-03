# Route File Format

Use a UTF-8 TSV file with one capture per line:

```tsv
# filename<TAB>url<TAB>wait_seconds
01-home.png	myapp://home	2
02-portfolio.png	myapp://portfolio	3
03-settings.png	myapp://settings	2
```

Rules:
- Use TAB as the separator.
- Keep `filename` relative (no leading slash).
- Keep URL as a valid app deep link.
- Set `wait_seconds` high enough for data fetch and animation.
- Use `#` for comments.
- Omit `wait_seconds` to use default `2`.

Recommended naming:
- Prefix with ordered numbers for deterministic App Store upload order.
- Use lowercase and hyphens: `07-report-overview.png`.

Validation tips:
- Ensure each filename ends in `.png` or `.jpg`.
- Ensure each URL opens the intended screen when run alone:
  `xcrun simctl openurl <UDID> '<URL>'`
