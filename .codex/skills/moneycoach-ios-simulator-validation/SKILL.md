---
name: moneycoach-ios-simulator-validation
description: Validate Money Coach mobile-app flows on iOS Simulator against the deployed backend. Use when deploying app-backend changes and smoke-testing authenticated intelligence/report/auth flows in the Expo iOS dev build, including Hermes inspector session injection, deep links, screenshots, and production-endpoint verification.
---

# Money Coach iOS Simulator Validation

Use this skill for repo-local validation of `mobile-app` against the deployed backend at `https://api.moneycoach.cc`.

Use `ios-screenshot-capture` if the goal is App Store grade screenshots. This skill is for smoke testing, state injection, and deploy verification.

## Preconditions

- Work from repo root: `/Users/jack/shawn/quanta`
- `mobile-app/.env` should point iOS to deployed backend via `EXPO_PUBLIC_API_BASE_URL`
- Review credentials live in `DEVELOPER.md`; do not hardcode them into the skill
- Metro / Expo iOS dev build is acceptable for simulator validation

## Workflow

### 1. Validate and deploy backend changes

If the change touches `app-backend`, run:

```bash
cd app-backend
docker compose up -d postgres redis
TEST_DATABASE_URL='postgres://quanta:quanta@localhost:5432/quanta?sslmode=disable' go test ./internal/app/...
```

Then push to `main`.

Do not rely solely on `gh run list` for deployment confirmation. In this repo, the GitHub Actions API can return `404` even when push succeeded.

Verify deployment by probing the real production endpoint instead:

```bash
ACCESS=$(curl -sS -X POST https://api.moneycoach.cc/v1/auth/email/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"<REVIEW_EMAIL>","password":"<REVIEW_PASSWORD>"}' | jq -r '.access_token')

curl -i -sS "https://api.moneycoach.cc/v1/intelligence/assets/stock%3Amic%3AXNAS%3ATSLA" \
  -H "Authorization: Bearer $ACCESS"
```

Wait until the production response reflects the new behavior.

### 2. Launch iOS simulator app

```bash
cd mobile-app
npm run ios -- -d "iPhone 16 Pro"
```

Useful device id from prior runs:

```bash
xcrun simctl list devices | rg 'iPhone 16 Pro'
```

### 3. Find the Hermes inspector websocket

```bash
curl -sS http://127.0.0.1:8081/json/list | jq .
```

Use the `webSocketDebuggerUrl` for the running app.

### 4. Inspect runtime module ids

Module ids are not stable. Discover them from `__r.getModules()` instead of hardcoding old ids.

Useful evaluator pattern:

```js
(() => {
  const mods = Array.from(__r.getModules().entries()).map(([id, m]) => ({
    id,
    verboseName: m.verboseName || '',
    output: m.output?.[0] || '',
  }))
  const pick = (needle) => mods.find((m) => m.verboseName.includes(needle) || m.output.includes(needle))?.id ?? null
  return JSON.stringify({
    router: pick('expo-router/build/index.js'),
    secureStore: pick('expo-secure-store/build/SecureStore.js'),
    intelligence: pick('src/services/intelligence.ts'),
  })
})()
```

### 5. Inject authenticated session when the UI path is noisy

If onboarding, dev menu, or modal sequencing makes manual sign-in slow, inject the review session into SecureStore through Hermes.

1. Get a fresh login token from production backend.
2. In Hermes, write the values with `ExpoSecureStore`.

Important pitfall: `getValueWithKeySync` and `setValueWithKeySync` require the second argument to be an object, not a string.

Correct form:

```js
expo.modules.ExpoSecureStore.setValueWithKeySync('mc_access_token', accessToken, {})
expo.modules.ExpoSecureStore.setValueWithKeySync('mc_refresh_token', refreshToken, {})
expo.modules.ExpoSecureStore.setValueWithKeySync('mc_user_id', userId, {})
```

Readback check:

```js
JSON.stringify({
  access: expo.modules.ExpoSecureStore.getValueWithKeySync('mc_access_token', {}),
  refresh: expo.modules.ExpoSecureStore.getValueWithKeySync('mc_refresh_token', {}),
  user: expo.modules.ExpoSecureStore.getValueWithKeySync('mc_user_id', {}),
})
```

Then terminate and relaunch the app:

```bash
xcrun simctl terminate <SIM_UDID> cc.moneycoach.app.ios || true
xcrun simctl launch <SIM_UDID> cc.moneycoach.app.ios
```

### 6. Open target flows via deep link

Examples:

```bash
xcrun simctl openurl <SIM_UDID> 'moneycoach://market-regime'
xcrun simctl openurl <SIM_UDID> 'moneycoach://asset-brief?asset_key=stock%3Amic%3AXNAS%3ATSLA'
```

For route debugging in Hermes:

```js
__r(<expo-router-store-module-id>).store.getRouteInfo()
```

### 7. Capture validation screenshots

```bash
mkdir -p .tmp/ios-test
xcrun simctl io <SIM_UDID> screenshot .tmp/ios-test/current.png
```

Keep screenshots in `.tmp/ios-test/` unless they are intended for docs or store assets.

## Known pitfalls in this repo

### Encoded asset keys in path params

`asset_key` values such as `stock:mic:XNAS:TSLA` are routed as path segments. Clients correctly percent-encode them. If the backend forgets to decode the chi URL param, production returns `404` for encoded values even though raw paths may work.

Symptom:
- `GET /v1/intelligence/assets/stock:mic:XNAS:TSLA` returns `200`
- `GET /v1/intelligence/assets/stock%3Amic%3AXNAS%3ATSLA` returns `404`

Fix belongs in backend path decoding, not by removing encoding in the client.

### Cold-start deep links to authenticated modals

If the app cold-starts directly into an authenticated modal such as `asset-brief`, auth bootstrap can finish after the modal mounts.

Correct UX rule:
- show loading until auth bootstrap and the first fetch settle
- do not flash the empty-state card before the first fetch completes

If you see an empty card immediately on cold start, wait for bootstrap or verify the screen after the fix above. Do not mistake this for a backend outage without checking the production endpoint.

### Expo developer menu overlay

The Expo dev menu or first-run overlay can sit above the real UI and ruin screenshots. Dismiss it before concluding the app is on the wrong screen.

### GitHub Actions visibility

This repo can push successfully while `gh run list` still returns `404`. Treat production endpoint behavior as the source of truth for deploy verification.

## Minimal end-to-end check for the intelligence flow

1. Launch the iOS dev build.
2. Ensure the app has a valid authenticated session.
3. Confirm `Market Pulse` renders with the `Market Regime` card.
4. Open `moneycoach://market-regime` and verify the modal content.
5. Open `moneycoach://asset-brief?asset_key=stock%3Amic%3AXNAS%3ATSLA` and verify:
   - header asset metadata
   - action badge
   - price cards
   - chart / entry plan / technical setup
6. If something looks wrong, first verify the corresponding production API endpoint with the same user token.
