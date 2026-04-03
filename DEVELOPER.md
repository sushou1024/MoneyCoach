# Developer Guide

## Ship Readiness Audit (February 20, 2026, 10:30 Local)

### Verdict
- **No-go for App Store / Play Store submission yet.**
- Core content/auth/subscription code paths are now in good shape, and iOS release build is no longer blocked.
- The remaining blockers are operational: deploy legal pages to production and provide the remaining release secrets.

### Fixed in This Pass
1. Dangerous reset endpoints are now properly guarded:
   - `GET /v1/reset-app-state` and `POST /v1/debug/reset-user` only register when `ENABLE_DANGEROUS_RESET_ROUTES=true`.
   - Reset secret is now env-driven via `RESET_APP_SECRET`.
2. In-app account deletion is implemented end-to-end:
   - Backend: `DELETE /v1/users/me` with `{"confirm_text":"DELETE"}`.
   - Mobile: Settings > Account & Data > Delete Account with explicit confirmation and sign-out.
3. Subscription compliance text/links are implemented in-app:
   - Auto-renew disclosure shown on paywall.
   - Terms/Privacy links shown on paywall (`EXPO_PUBLIC_TERMS_URL`, `EXPO_PUBLIC_PRIVACY_URL`).
4. Legal pages were added to web frontend:
   - `web-frontend/app/terms/page.tsx`
   - `web-frontend/app/privacy/page.tsx`
   - Footer links added on homepage.
5. Workflows now wire legal URLs:
   - `.github/workflows/mobile-app-android-release.yml`
   - `.github/workflows/mobile-app-pwa-deploy.yml`
6. Missing GitHub **variables** are now populated in `jackcpku/money-coach` (non-secret env coverage is complete).
7. iOS Release simulator build is now passing:
   - `cd mobile-app/ios && xcodebuild -workspace MoneyCoach.xcworkspace -scheme MoneyCoach -configuration Release -destination 'platform=iOS Simulator,OS=18.6,name=iPhone 16' CODE_SIGNING_ALLOWED=NO build` ✅

### Evidence Snapshot
- Mobile checks pass:
  - `cd mobile-app && npm run lint` ✅
  - `cd mobile-app && npm run test -- --runInBand` ✅
  - `cd mobile-app && npx tsc --noEmit -p tsconfig.json` ✅
- Web frontend checks pass:
  - `cd web-frontend && npm run lint` ✅
  - `cd web-frontend && npm run build` ✅
  - Build routes include `/terms` and `/privacy` ✅
- Backend compile check passes:
  - `cd app-backend && go test ./... -run '^$'` ✅
- Dockerized backend + targeted e2e with extended timeout passes:
  - `cd app-backend && docker compose down -v && docker compose up --build -d` ✅
  - `E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 go test ./e2e -run TestDeployedPortfolioCurrencySwitch -timeout 15m` ✅
- GitHub workflow env audit:
  - Missing vars: none ✅
  - Remaining missing secrets: `FCM_SERVER_KEY`, `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64` ⚠️
- Live legal URL check against production domain:
  - `https://moneycoach.cc/` -> `200`
  - `https://moneycoach.cc/terms` -> `404`
  - `https://moneycoach.cc/privacy` -> `404`

### Remaining Blockers
1. Legal pages are coded but not deployed to production yet (`/terms` and `/privacy` still return `404`).
2. Missing release secrets:
   - `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64` (required for Play upload and backend Google purchase verification).
   - `FCM_SERVER_KEY` (required only if Android push notifications are needed).
3. Store operations checklist is still pending (App Store Connect + Play Console metadata/review prep).

### Step-by-Step Next Actions

#### Phase 1: Deploy Legal URLs (Required for Review)
1. Push current `web-frontend` changes to `main`.
2. Run web deploy workflow:
   - `gh workflow run web-frontend-deploy.yml --repo jackcpku/money-coach`
3. Verify URLs are live:
   - `curl -I https://moneycoach.cc/terms`
   - `curl -I https://moneycoach.cc/privacy`
4. Confirm both return `200` before App Review submission.

#### Phase 2: Complete Release Secrets
1. Add Play service account secret:
   - `gh secret set GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64 --repo jackcpku/money-coach`
2. Add FCM secret if Android push is required:
   - `gh secret set FCM_SERVER_KEY --repo jackcpku/money-coach`

#### Phase 3: Keep Production Safety Locked
1. Ensure backend runtime uses:
   - `BILLING_DEV_MODE=false`
   - `ENABLE_DANGEROUS_RESET_ROUTES=false`
2. Keep `RESET_APP_SECRET` unset in production unless reset routes are intentionally enabled in non-prod only.

#### Phase 4: Pre-Submit Validation
1. Backend e2e on Docker:
   - Targeted: `E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 go test ./e2e -run TestDeployedPortfolioCurrencySwitch -timeout 15m`
   - Full suite: `E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 go test ./e2e -timeout 45m`
2. iOS release build:
   - `cd mobile-app/ios && xcodebuild -workspace MoneyCoach.xcworkspace -scheme MoneyCoach -configuration Release -destination 'platform=iOS Simulator,OS=18.6,name=iPhone 16' CODE_SIGNING_ALLOWED=NO build`
3. Android release pipeline:
   - Internal smoke test:
     ```bash
     cd mobile-app
     npx eas-cli build --platform android --profile production
     npx eas-cli submit --platform android --profile android-internal --latest
     ```
   - Production release:
     ```bash
     cd mobile-app
     npx eas-cli build --platform android --profile production
     npx eas-cli submit --platform android --profile production --latest
     ```

#### Phase 5: Store Submission Checklist
1. App Store Connect:
   - Verify Sign in with Apple + IAP product wiring.
   - Provide Privacy URL + Terms URL + support URL + review notes.
2. Play Console:
   - Verify subscription product IDs, internal track release, and service account permissions.
3. Final manual QA:
   - auth, purchase, restore, manage subscription, delete account.

### Final Go/No-Go Rule
- **Go** only when Phase 1 to Phase 5 are complete.
- Current state: **No-Go** (operational blockers remain, code blockers largely resolved).

## Company Apple Account Cutover Runbook (Personal -> Company)

This section is a click-by-click procedure for migrating from personal Apple credentials to company-owned Apple credentials.  
UI labels are written as of February 20, 2026.

### Inputs You Need Before Starting
1. Company Apple team access:
   - Apple Developer: https://developer.apple.com/account/
   - App Store Connect: https://appstoreconnect.apple.com/
2. Existing project identifiers:
   - iOS bundle ID (currently expected: `cc.moneycoach.app.ios`)
   - desired Apple IAP product IDs (currently expected by this repo: `weekly`, `yearly`)
3. Repo and CLI access:
   - GitHub repo: `jackcpku/money-coach`
   - `gh` CLI authenticated as a repo admin.

### Phase A: Freeze Personal Account and Record Old Values
1. Open https://appstoreconnect.apple.com/ and sign in to the **personal** team.
2. Go to `Apps` and open your app.
3. Go to `TestFlight` and stop distributing new builds from personal team.
4. Go to app `General` -> `App Information` and record:
   - Bundle ID
   - Apple ID / SKU (for audit reference only)
5. Go to app `Monetization` -> `Subscriptions` and record old subscription product IDs.
   - If you only see a `+` dialog with `Type = Consumable / Non-Consumable`, you are in the wrong section (`In-App Purchases` for one-time products). Go back and open `Monetization` -> `Subscriptions`.
6. Go to https://developer.apple.com/account/resources/authkeys/list and record old APNs key metadata (`Key ID`, `Team ID`).
7. Store these values in an internal migration note; do not reuse them in production.

### Phase B: Create/Configure Company Apple App Identity
1. Open https://developer.apple.com/account/resources/identifiers/list and sign in to the **company** team.
2. Click `+` (Create Identifier).
3. Select `App IDs`, click `Continue`.
4. Select `App`, click `Continue`.
5. Fill:
   - `Description`: `Money Coach iOS`
   - `Bundle ID`:
     - If reusing existing: enter existing bundle ID.
     - If fresh: enter your new reverse-domain ID.
6. In `Capabilities`, check:
   - `Sign In with Apple`
   - `In-App Purchase`
7. Click `Continue`, then `Register`.

### Phase C: Create Company App Record in App Store Connect
1. Open https://appstoreconnect.apple.com/apps and sign in to the **company** team.
2. Click `+` -> `New App`.
3. Fill fields:
   - `Platforms`: `iOS`
   - `Name`: `Money Coach` (or your final public name)
   - `Primary language`: choose your production language
   - `Bundle ID`: select the bundle ID from Phase B
   - `SKU`: e.g. `moneycoach-ios-prod`
   - `User Access`: `Full Access`
4. Click `Create`.

### Phase D: Create Company Subscriptions (Weekly + Yearly)
1. In company App Store Connect app, open `Monetization` -> `Subscriptions`.
   - If clicking `+` asks for `Type (Consumable / Non-Consumable)`, switch to `Monetization` -> `Subscriptions` (not `In-App Purchases`).
2. If no subscription group exists:
   - Click `+` -> `Create Subscription Group`.
   - Group `Reference Name`: `Money Coach Pro`.
   - Click `Create`.
3. Create weekly product:
   - Open the `Money Coach Pro` subscription group.
   - If this is the first product in the group, click `Create Subscription`.
   - If there is already a product, click `+` -> `Create Subscription`.
   - `Reference Name`: `Pro Weekly`.
   - `Product ID`: `weekly` (or your chosen ID).
   - Click `Create`.
4. Create yearly product:
   - In the same `Money Coach Pro` group, click `+` -> `Create Subscription`.
   - `Reference Name`: `Pro Yearly`.
   - `Product ID`: `yearly` (or your chosen ID).
   - Click `Create`.
5. For each subscription product:
   - Open product.
   - Configure pricing and availability.
   - Click `Add Localization` and create all 5 app locales (`English (U.S.)`, `Chinese (Simplified)`, `Chinese (Traditional)`, `Japanese`, `Korean`) with these exact values:
     - Weekly subscription (`Product ID = weekly`):
       - `Language`: `English (U.S.)` | `Display Name`: `Money Coach Pro Weekly` | `Description`: `Weekly access to all Money Coach Pro features.`
       - `Language`: `Chinese (Simplified)` | `Display Name`: `Money Coach Pro 周订阅` | `Description`: `解锁 Money Coach Pro 全部功能，按周自动续订。`
       - `Language`: `Chinese (Traditional)` | `Display Name`: `Money Coach Pro 週訂閱` | `Description`: `解鎖 Money Coach Pro 全部功能，按週自動續訂。`
       - `Language`: `Japanese` | `Display Name`: `Money Coach Pro 週間プラン` | `Description`: `Money Coach Pro のすべての機能を利用できる週間自動更新サブスクリプションです。`
       - `Language`: `Korean` | `Display Name`: `Money Coach Pro 주간 플랜` | `Description`: `Money Coach Pro의 모든 기능을 이용할 수 있는 주간 자동 갱신 구독입니다.`
     - Yearly subscription (`Product ID = yearly`):
       - `Language`: `English (U.S.)` | `Display Name`: `Money Coach Pro Yearly` | `Description`: `Yearly access to all Money Coach Pro features.`
       - `Language`: `Chinese (Simplified)` | `Display Name`: `Money Coach Pro 年订阅` | `Description`: `解锁 Money Coach Pro 全部功能，按年自动续订。`
       - `Language`: `Chinese (Traditional)` | `Display Name`: `Money Coach Pro 年訂閱` | `Description`: `解鎖 Money Coach Pro 全部功能，按年自動續訂。`
       - `Language`: `Japanese` | `Display Name`: `Money Coach Pro 年間プラン` | `Description`: `Money Coach Pro のすべての機能を利用できる年間自動更新サブスクリプションです。`
       - `Language`: `Korean` | `Display Name`: `Money Coach Pro 연간 플랜` | `Description`: `Money Coach Pro의 모든 기능을 이용할 수 있는 연간 자동 갱신 구독입니다.`
   - In section `Review Information`, complete both fields:
     - `Screenshot for review`:
       - Use a real iPhone screenshot of the in-app paywall modal (not a design mock).
       - The screenshot should show:
         - paywall title (`Stop Trading Blindly` in English build),
         - plan cards with prices,
         - subscription CTA button (for example `Unlock Weekly` / `Unlock Annual`),
         - legal links (`Terms of Use`, `Privacy Policy`).
       - Do not include personal email, real portfolio values, or debug UI.
       - The screenshot must not show any internal development controls (for example `Dev Tools` / `Simulate Pro entitlement`).
       - Use a valid App Store screenshot size for iPhone. Recommended: 6.9-inch portrait (`1320 x 2868`).
       - Apple only uses this image for review; it is not shown on the public App Store.
       - After upload, Apple does not allow deleting this screenshot; only replacement. So upload a clean final image.
     - `Review Notes (Optional)`:
       - Recommended: always fill this field to reduce back-and-forth during review.
       - Paste this exact note for **weekly** (`Product ID = weekly`):
         ```text
         This auto-renewable subscription unlocks Money Coach Pro features (full diagnostics, strategy execution guidance, and portfolio insights).

         How to reach this purchase screen:
         1) Install and open the app.
         2) Complete onboarding and sign in by tapping "Continue with Email" and using the credentials provided in App Review Information.
         3) Go to Assets tab and tap "Unlock Full Report" (or open Insights and tap the locked upgrade CTA).
         4) The paywall modal appears with Weekly and Annual plans.

         Product under review on this page: weekly (Money Coach Pro Weekly).
         No special hardware is required.
         A pre-created review account is provided in App Review Information.
         Terms of Use: https://moneycoach.cc/terms
         Privacy Policy: https://moneycoach.cc/privacy
         ```
       - Paste this exact note for **yearly** (`Product ID = yearly`):
         ```text
         This auto-renewable subscription unlocks Money Coach Pro features (full diagnostics, strategy execution guidance, and portfolio insights).

         How to reach this purchase screen:
         1) Install and open the app.
         2) Complete onboarding and sign in by tapping "Continue with Email" and using the credentials provided in App Review Information.
         3) Go to Assets tab and tap "Unlock Full Report" (or open Insights and tap the locked upgrade CTA).
         4) The paywall modal appears with Weekly and Annual plans.

         Product under review on this page: yearly (Money Coach Pro Yearly).
         No special hardware is required.
         A pre-created review account is provided in App Review Information.
         Terms of Use: https://moneycoach.cc/terms
         Privacy Policy: https://moneycoach.cc/privacy
         ```
   - Ensure status is ready for submission/testing.

### Phase E: Configure App Store Server Notifications (Company)
1. Open https://appstoreconnect.apple.com/apps and sign in to the **company** team.
2. Click your app (`Money Coach`).
3. In the left sidebar, click `General` -> `App Information`.
4. Scroll down to section `App Store Server Notifications`.
   - If you cannot find this section, your account role is likely missing permission. Use `Account Holder`, `Admin`, `App Manager`, or `Marketing`.
5. Under `Production Server URL`, click `Set Up URL` (or `Edit` if already set).
6. In dialog `Production Server URL`, fill:
   - `URL`: `https://api.moneycoach.cc/v1/webhooks/apple`
   - If a `Version` selector is shown, choose `Version 2`.
   - If no `Version` selector is shown, that is valid (some `Edit URL` dialogs only expose the URL field).
7. Click `Save`.
8. Under `Sandbox Server URL`, click `Set Up URL` (recommended for TestFlight validation), and fill the same values:
   - `URL`: `https://api.moneycoach.cc/v1/webhooks/apple`
   - If a `Version` selector is shown, choose `Version 2`.
   - If no `Version` selector is shown, that is valid (some `Edit URL` dialogs only expose the URL field).
9. Click `Save`.
10. Important behavior:
   - If only `Production Server URL` is set, Apple sends both production and sandbox notifications to production URL.
   - If only `Sandbox Server URL` is set, production notifications are **not** sent.

### Phase F: Get Company Shared Secret
1. In company app, open `General` -> `App Information`.
2. Find `App-Specific Shared Secret`.
3. Click `Generate` (or `Manage` -> `Generate New` if needed).
4. Copy the secret value immediately.
5. This value maps to GitHub Secret:
   - `APPSTORE_SHARED_SECRET`

### Phase G: Create Company APNs Key
1. Open https://developer.apple.com/account/resources/authkeys/list on company team.
2. Click `+`.
3. In `Key Name`, enter: `moneycoachapnsprod`.
   - Use letters/numbers only (no `-`, `.`, `@`, `&`, `*`, `'`, `"`).
4. Check `Apple Push Notifications service (APNs)`.
5. Click `Continue` -> `Register`.
6. Download `.p8` file once (cannot be downloaded again).
7. Record:
   - `Key ID` -> `APNS_KEY_ID`
   - `Team ID` -> `APNS_TEAM_ID`
     - Open `https://developer.apple.com/account`.
     - In the top-right team picker, make sure your **company** team is selected.
     - Click `Membership details`.
     - Copy the 10-character `Team ID` value.
   - `.p8` file contents -> `APNS_PRIVATE_KEY`
   - Bundle ID from Phase B -> `APNS_BUNDLE_ID`

### Phase H: Update Repo Configuration (Exact Names)
Run these commands from your terminal:

```bash
# Apple core identifiers (GitHub Variables)
gh variable set APPLE_IOS_BUNDLE_ID --repo jackcpku/money-coach --body 'cc.moneycoach.app.ios'
gh variable set APPLE_IAP_PRODUCT_ID_WEEKLY --repo jackcpku/money-coach --body 'weekly'
gh variable set APPLE_IAP_PRODUCT_ID_YEARLY --repo jackcpku/money-coach --body 'yearly'

# APNs metadata (GitHub Variables)
gh variable set APNS_KEY_ID --repo jackcpku/money-coach --body '<COMPANY_APNS_KEY_ID>'
gh variable set APNS_TEAM_ID --repo jackcpku/money-coach --body '<COMPANY_TEAM_ID>'
gh variable set APNS_BUNDLE_ID --repo jackcpku/money-coach --body 'cc.moneycoach.app.ios'

# Mobile build-time Apple IDs (GitHub Variables)
gh variable set EXPO_PUBLIC_APPLE_IAP_PRODUCT_ID_WEEKLY --repo jackcpku/money-coach --body 'weekly'
gh variable set EXPO_PUBLIC_APPLE_IAP_PRODUCT_ID_YEARLY --repo jackcpku/money-coach --body 'yearly'

# Apple secrets (GitHub Secrets; command prompts for value)
gh secret set APPSTORE_SHARED_SECRET --repo jackcpku/money-coach
gh secret set APNS_PRIVATE_KEY --repo jackcpku/money-coach
```

Also update local env files with the same company values:
1. `app-backend/.env`
2. `mobile-app/.env`

### Phase I: Google Adjustments for Apple Cutover
Only required if your iOS bundle ID changed.

1. Open Google Cloud Console Credentials:
   - https://console.cloud.google.com/apis/credentials
2. Select project used for Money Coach auth.
3. Click `+ CREATE CREDENTIALS` -> `OAuth client ID`.
4. `Application type`: `iOS`.
5. Fill:
   - `Name`: `MoneyCoach iOS (Company)`
   - `Bundle ID`: new company bundle ID
6. Click `Create`, copy client ID.
7. Rotate GitHub variables:
   - `EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID`
   - `GOOGLE_ALLOWED_CLIENT_IDS` (must include this new iOS client ID plus Android/Web/Expo IDs).

### Phase J: Google Play + RTDN (Required for Android Production)
Use the detailed section `Google Play Production Release (Exact Fill Guide)` below and complete every step.

Critical Phase J checks:
1. `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64` must exist in GitHub Secrets for:
   - Backend Google subscription verification.
   - Backend deployment / verification code that calls Google Play APIs.
2. `GOOGLE_PUBSUB_AUDIENCE` must be `https://api.moneycoach.cc/v1/webhooks/google`.
3. `GOOGLE_PLAY_PRODUCT_WEEKLY` and `GOOGLE_PLAY_PRODUCT_ANNUAL` must match Play subscription Product IDs exactly (`weekly`, `yearly`).
4. Run one internal-track Android release end-to-end before production rollout.

### Phase K: Fresh-Start Data Cleanup (Production Backend)
1. Take a production DB snapshot/backup.
2. Remove legacy billing lineage from personal Apple account:
   - `entitlements`
   - `external_subscriptions`
   - `payments`
3. Ensure production safety flags:
   - `BILLING_DEV_MODE=false`
   - `ENABLE_DANGEROUS_RESET_ROUTES=false`

### Phase L: Release Validation (Company-Only)
1. Check plans API returns company product IDs:
   - `GET https://api.moneycoach.cc/v1/billing/plans`
2. Run iOS release build and test on company setup:
   - Apple sign-in
   - purchase
   - restore
   - account deletion
3. Confirm Apple webhook traffic is accepted at:
   - `POST /v1/webhooks/apple`
4. Confirm Android purchase verification works with new Play service account.
5. Confirm legal URLs are live:
   - `https://moneycoach.cc/terms`
   - `https://moneycoach.cc/privacy`

### Final Cutover Exit Criteria
1. No personal Apple credentials remain in:
   - GitHub Variables
   - GitHub Secrets
   - backend env
   - mobile env
2. Company app + subscriptions + webhooks are live and tested.
3. Google Play service account secret is configured.
4. End-to-end auth + billing + restore + delete-account pass in release builds.

## iOS `1.0 Prepare for Submission` (Exact Fill Guide)

This section is for App Store Connect path:
- https://appstoreconnect.apple.com/apps
- `Your App` -> `Distribution` -> `App Store` -> `iOS App 1.0` -> `Prepare for Submission`

### 0. General -> App Information (Localizable Information)
1. Open https://appstoreconnect.apple.com/apps.
2. Click your app.
3. Left sidebar: `General` -> `App Information`.
4. In `Localizable Information`, set:
   - `Name`: `Money Coach Pro`
   - `Subtitle`: `AI Portfolio Coach`
5. Click `Save` (top-right).

### A. Previews and Screenshots
1. In section `Previews and Screenshots`, locate the iPhone slot that accepts:
   - `1242 × 2688`, `2688 × 1242`, `1284 × 2778`, or `2778 × 1284`.
2. Upload these files from this repo to the public App Store screenshot gallery:
   - `artifacts/app-store/ios-67-upload/01-portfolio.png`
   - `artifacts/app-store/ios-67-upload/02-alerts.png`
   - `artifacts/app-store/ios-67-upload/03-plan.png`
   - `artifacts/app-store/ios-67-upload/04-signin-methods.png`
   - `artifacts/app-store/ios-67-upload/05-email-password-signin.png`
   - `artifacts/app-store/ios-67-upload/07-insights-loggedin.png`
   - `artifacts/app-store/ios-67-upload/08-assets-loggedin.png`
   - `artifacts/app-store/ios-67-upload/09-me-loggedin.png`
   - `artifacts/app-store/ios-67-upload/10-report-overview.png`
   - `artifacts/app-store/ios-67-upload/11-report-strategy-detail.png`
3. Keep this order (left to right) so the story is coherent:
   - Onboarding value -> sign-in method -> signed-in product value -> full report depth.
4. Do not put paywall in the public screenshot gallery. Keep paywall only for subscription review evidence:
   - `artifacts/app-store/ios-67-upload/06-upgrade-paywall.png` is for `Monetization -> Subscriptions -> [Product] -> Review Information -> Screenshot for review`.
5. App preview videos are optional for this release; leave empty unless you have polished capture videos.
6. Latest capture baseline (executed on February 20, 2026):
   - Production fixture replay with your review account and `portfolio4`:
     - `cd app-backend && go run ./cmd/e2e_local --base-url https://api.moneycoach.cc --portfolio portfolio4 --email <REVIEW_EMAIL> --password '<REVIEW_PASSWORD>' --timeout 15m`
   - Generated upload batch and calculation:
     - `upload_batch_id=ub_01KHXNEJ841DKZ26H8H2ZEHHP2`
     - `calculation_id=calc_01KHXNFJBE4E8AZVTVPJK3FWCB`
     - `plan_id=plan_03`
   - Simulator recapture commands for the App Store gallery (`1284 x 2778`):
     ```bash
     DEVICE="ASC iPhone 13 Pro Max"
     xcrun simctl openurl "$DEVICE" 'moneycoach://sc00' && sleep 1 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/01-portfolio.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://sc01' && sleep 1 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/02-alerts.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://sc01b' && sleep 1 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/03-plan.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://sc07' && sleep 1 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/04-signin-methods.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://sc07a' && sleep 1 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/05-email-password-signin.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://insights' && sleep 2 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/07-insights-loggedin.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://assets' && sleep 5 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/08-assets-loggedin.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://me' && sleep 2 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/09-me-loggedin.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://report?id=calc_01KHXNFJBE4E8AZVTVPJK3FWCB' && sleep 5 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/10-report-overview.png
     xcrun simctl openurl "$DEVICE" 'moneycoach://strategy?calculation_id=calc_01KHXNFJBE4E8AZVTVPJK3FWCB&plan_id=plan_03' && sleep 5 && xcrun simctl io "$DEVICE" screenshot artifacts/app-store/ios-67-upload/11-report-strategy-detail.png
     ```
   - Paywall recapture is optional; keep `06-upgrade-paywall.png` as-is unless pricing text changes.

### B. Promotional Text
Paste this exact text:

```text
Turn scattered holdings into one clear portfolio. Money Coach Pro gives you AI-guided insights and daily action plans to invest with discipline.
```

### C. Description
Paste this exact text:

```text
Money Coach Pro helps you understand your entire portfolio and decide your next move with clarity.

Track everything in one place:
• Combine screenshots from exchanges, brokerages, and wallets
• See total allocation, concentration, and position mix in seconds

Get actionable guidance:
• Receive AI-generated insights tied to your real holdings
• Follow structured execution plans instead of ad-hoc decisions
• Stay consistent with daily signals and disciplined routines

Built for real investors:
• Supports both crypto and equities workflows
• Localized experience in English, Simplified Chinese, Traditional Chinese, Japanese, and Korean
• Private, account-based access with secure sign-in

Subscription:
Money Coach Pro is an auto-renewable subscription. Payment is charged to your Apple ID at confirmation. Subscription renews automatically unless canceled at least 24 hours before the end of the current period. Manage or cancel anytime in your App Store account settings.
```

### D. Keywords
Paste this exact text:

```text
portfolio,investing,stocks,crypto,tracker,insights,allocation,analysis,trading,finance,wealth
```

### E. URLs
1. `Support URL`:
   - `https://moneycoach.cc/support`
2. `Marketing URL`:
   - `https://moneycoach.cc/`

### F. Version / Copyright
1. `Version`:
   - Keep `1.0` for this submission.
2. `Copyright`:
   - Use your legal entity name exactly:
   - `2026 <Your Company Legal Name>`
   - Example format only: `2026 Money Coach Inc.`

### G. Routing App Coverage File
- Optional for map-routing apps only.
- Leave empty for Money Coach.

### H. Build
1. Upload build via EAS / Xcode / Transporter.
2. Select the uploaded build under `Build`.
3. If export compliance prompts appear, complete the encryption questionnaire before submission.

### I. App Review Information

#### 1) Sign-In Information (Important)
1. Check `Sign-in required`.
2. Use a dedicated **email/password** review account (no personal Google account needed), then fill:
   - `Username`: `nigaji3479@iaciu.com`
   - `Password`: `nigaji3479@iaciu.com`
3. In `Notes`, paste this exact text:

```text
Login is required to access the full app.

Primary review login:
- Method: Continue with Email
- Username: nigaji3479@iaciu.com
- Password: nigaji3479@iaciu.com

No secondary review account is currently configured.

How to test:
1) Open app.
2) Complete onboarding screens.
3) On "Save your profile", tap "Continue with Email".
4) On "Sign in with email", tap the `Sign in` tab (right side). `Create account` is the default tab.
5) Use the credentials above.
6) After login, open Assets and Insights tabs.
7) To test subscription purchase flow, open paywall from locked upgrade CTA.
8) Terms: https://moneycoach.cc/terms
9) Privacy: https://moneycoach.cc/privacy
10) Support: https://moneycoach.cc/support

Signup verification note:
- A one-time email verification code is required only when creating a new account.
- The review account above is pre-created, so App Review should not need a verification code.
```

Recommended operational setup:
1. Create two dedicated review accounts:
   - `review_free`: no active subscription (to test purchase flow).
   - `review_pro`: active subscription (to test unlocked features quickly).
2. Put `review_free` in Sign-In fields, and include `review_pro` in Notes as optional secondary account.
3. Keep these as persistent accounts; avoid deleting them between review resubmissions.

Review account creation (run after backend deploy that includes `/v1/auth/email/register/start`):

```bash
API_BASE_URL="https://api.moneycoach.cc"
REVIEW_FREE_EMAIL="nigaji3479@iaciu.com"
REVIEW_FREE_PASSWORD="nigaji3479@iaciu.com"

# 1) Send signup code
curl -X POST "$API_BASE_URL/v1/auth/email/register/start" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$REVIEW_FREE_EMAIL\"}"

# 2) Read the 6-digit code from mailbox, then complete signup
SIGNUP_CODE="<code_from_email>"
curl -X POST "$API_BASE_URL/v1/auth/email/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$REVIEW_FREE_EMAIL\",\"password\":\"$REVIEW_FREE_PASSWORD\",\"code\":\"$SIGNUP_CODE\"}"
```

Repeat the same for `review_pro`, then purchase/activate subscription on `review_pro` only.

#### 2) Contact Information
Fill with your real release owner details:
- `First name`: `<release_owner_first_name>`
- `Last name`: `<release_owner_last_name>`
- `Phone number`: `<release_owner_phone>`
- `Email`: `<release_owner_email>`

#### 3) Notes
Use the exact text from `#### 1) Sign-In Information`, step 3 (copy-paste without edits).

#### 4) Attachment
- Optional.
- Add only if reviewer needs additional walkthrough video or PDF.

### J. App Store Version Release
For production control, choose:
1. `Manually release this version`

Use automatic release only if you want immediate go-live right after approval.

## Backend (Docker)

The backend reads config from `app-backend/.env`.

Clean restart (resets Postgres + Redis volumes):

```bash
cd app-backend
docker compose down -v
docker compose up --build -d
```

Optional health check:

```bash
curl http://localhost:8080/healthz
```

Dangerous reset routes are disabled by default.

Enable only for local/staging:

```bash
export ENABLE_DANGEROUS_RESET_ROUTES=true
export RESET_APP_SECRET='replace-with-long-random-secret'
```

Dangerous reset (drops schema + flushes Redis, **never enable in production**):

```bash
curl -H "X-Reset-Secret: $RESET_APP_SECRET" http://localhost:8080/v1/reset-app-state
```

Debug reset user by email (deletes the account and all user data):

```bash
curl -X POST \
  -H "X-Reset-Secret: $RESET_APP_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}' \
  http://localhost:8080/v1/debug/reset-user
```

Delete current authenticated user (used by mobile Settings > Delete Account):

```bash
curl -X DELETE \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"confirm_text":"DELETE"}' \
  http://localhost:8080/v1/users/me
```

## Production DB Access via SSM Bastion (AWS)

Use this when you need to directly inspect or patch production rows (for example, setting a review account entitlement).

### 1) Ensure bastion is provisioned
1. Check existing bastion output:
   - `aws cloudformation describe-stacks --region ap-southeast-1 --stack-name app-backend --query "Stacks[0].Outputs[?OutputKey=='BastionInstanceId'].OutputValue" --output text`
2. If empty, deploy with bastion enabled:
   - `aws cloudformation deploy --region ap-southeast-1 --stack-name app-backend --template-file app-backend/cloudformation/backend.yaml --capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM --parameter-overrides EnableBastion=true BastionInstanceType=t3.micro --no-fail-on-empty-changeset`
3. Wait until stack status is `UPDATE_COMPLETE`:
   - `aws cloudformation describe-stacks --region ap-southeast-1 --stack-name app-backend --query "Stacks[0].StackStatus" --output text`

### 2) Start port forwarding from local machine to RDS through bastion
1. Print the exact command from stack output:
   - `aws cloudformation describe-stacks --region ap-southeast-1 --stack-name app-backend --query "Stacks[0].Outputs[?OutputKey=='BastionPortForwardCommand'].OutputValue" --output text`
2. Run that command in terminal and keep it running.
3. Confirm SSM sees bastion online (optional check):
   - `BASTION_ID=$(aws cloudformation describe-stacks --region ap-southeast-1 --stack-name app-backend --query "Stacks[0].Outputs[?OutputKey=='BastionInstanceId'].OutputValue" --output text) && aws ssm describe-instance-information --region ap-southeast-1 --query "InstanceInformationList[?InstanceId=='${BASTION_ID}'].PingStatus" --output text`

### 3) Read DB username/password from Secrets Manager
1. Get secret ARN:
   - `aws cloudformation describe-stacks --region ap-southeast-1 --stack-name app-backend --query "Stacks[0].Outputs[?OutputKey=='DatabaseCredentialsArn'].OutputValue" --output text`
2. Read secret JSON:
   - `aws secretsmanager get-secret-value --region ap-southeast-1 --secret-id <DatabaseCredentialsArn> --query SecretString --output text`
3. Extract `username` and `password` from the JSON.

### 4) Connect with `psql` through forwarded port
1. In another terminal:
   - `PGPASSWORD='<password>' psql "host=127.0.0.1 port=15432 user=<username> dbname=appdb sslmode=require"`

### 5) Promote a review account to Pro (idempotent SQL)
1. Run this SQL in `psql`:
   ```sql
   WITH target_user AS (
     SELECT id
     FROM users
     WHERE lower(email) = lower('<REVIEW_EMAIL>')
   )
   INSERT INTO entitlements (user_id, status, provider, plan_id, current_period_end, last_verified_at)
   SELECT id, 'active', 'manual', 'yearly', (now() AT TIME ZONE 'UTC') + interval '365 days', (now() AT TIME ZONE 'UTC')
   FROM target_user
   ON CONFLICT (user_id) DO UPDATE
   SET status = EXCLUDED.status,
       provider = EXCLUDED.provider,
       plan_id = EXCLUDED.plan_id,
       current_period_end = EXCLUDED.current_period_end,
       last_verified_at = EXCLUDED.last_verified_at;
   ```
2. Verify:
   ```sql
   SELECT u.id AS user_id, u.email, e.status, e.provider, e.plan_id, e.current_period_end, e.last_verified_at
   FROM users u
   JOIN entitlements e ON e.user_id = u.id
   WHERE lower(u.email) = lower('<REVIEW_EMAIL>');
   ```

### 5A) Simulate paywall, then restore Pro (end-to-end review check)
1. Temporarily expire entitlement (forces locked state / paywall in app):
   ```sql
   WITH target_user AS (
     SELECT id
     FROM users
     WHERE lower(email) = lower('<REVIEW_EMAIL>')
   )
   INSERT INTO entitlements (user_id, status, provider, plan_id, current_period_end, last_verified_at)
   SELECT id, 'expired', 'manual', 'yearly', (now() AT TIME ZONE 'UTC') - interval '1 day', (now() AT TIME ZONE 'UTC')
   FROM target_user
   ON CONFLICT (user_id) DO UPDATE
   SET status = EXCLUDED.status,
       provider = EXCLUDED.provider,
       plan_id = EXCLUDED.plan_id,
       current_period_end = EXCLUDED.current_period_end,
       last_verified_at = EXCLUDED.last_verified_at;
   ```
2. Verify lock state via API (`403 ENTITLEMENT_REQUIRED` expected):
   ```bash
   API='https://api.moneycoach.cc'
   TOKEN=$(curl -sS -X POST "$API/v1/auth/email/login" -H 'Content-Type: application/json' -d '{"email":"<REVIEW_EMAIL>","password":"<REVIEW_PASSWORD>"}' | jq -r '.access_token')
   curl -i -H "Authorization: Bearer $TOKEN" "$API/v1/insights"
   ```
3. Re-run step `5)` SQL to restore `active` Pro entitlement.
4. Verify unlocked state via API (`200` expected):
   ```bash
   API='https://api.moneycoach.cc'
   TOKEN=$(curl -sS -X POST "$API/v1/auth/email/login" -H 'Content-Type: application/json' -d '{"email":"<REVIEW_EMAIL>","password":"<REVIEW_PASSWORD>"}' | jq -r '.access_token')
   curl -i -H "Authorization: Bearer $TOKEN" "$API/v1/insights"
   ```

### 6) Non-interactive fallback (run SQL directly on bastion with SSM Run Command)
1. Use `aws ssm send-command` with `AWS-RunShellScript` to execute a shell script that calls `psql` on bastion.
2. This path avoids local Session Manager plugin setup and was validated during this release prep.

## Backend Environment Variables (app-backend/.env)

### Core runtime (required)
- `DATABASE_URL`: Postgres connection string (Docker uses `postgres://quanta:quanta@host.docker.internal:5432/quanta?sslmode=disable`).
- `REDIS_URL`: Redis connection string (Docker uses `redis://host.docker.internal:6379/0`).
- `JWT_SIGNING_SECRET`: generate a long random secret (32+ bytes).
- `OBJECT_STORAGE_MODE`: `local` or `s3`.
  - `local`: set `OBJECT_STORAGE_LOCAL_DIR` and `LOCAL_STORAGE_BASE_URL`.
  - `s3`: set `OBJECT_STORAGE_BUCKET`, `OBJECT_STORAGE_REGION`, and ensure AWS credentials or IAM role are available.
- `LOGOS_XHKG_BASE_URL`, `LOGOS_NASDAQ_BASE_URL`, `LOGOS_NYSE_BASE_URL`: base URLs that host the logo assets + mapping JSON (e.g., S3/CloudFront hosting `app-backend/data/logos/**`).
- `BINANCE_API_BASE_URL`: Binance REST base URL (defaults to `https://api.binance.com` in Docker/CI).
- `OPEN_EXCHANGE_APP_ID`: OpenExchangeRates app ID (create an app in the OpenExchangeRates dashboard).
- `MARKETSTACK_ACCESS_KEY`: Marketstack API key (from the Marketstack dashboard).
- `COINGECKO_PRO_API_KEY`: CoinGecko Pro API key (from CoinGecko Pro → API Keys).
- `GEMINI_API_KEY`: Gemini API key (Google AI Studio → API keys).
- `RESEND_API_KEY` + `RESEND_FROM_EMAIL`: Resend API key + verified sender email used for signup verification code emails (Resend dashboard → API Keys + verified sender).

Optional external data:
- `CMC_PRO_API_KEY`: CoinMarketCap key (Fear & Greed + optional API tests; create in the CoinMarketCap developer portal).

Optional runtime overrides:
- `PORT`: HTTP port for the API server (defaults to `8080`).
- `BINANCE_FUTURES_BASE_URL`: Binance futures REST base URL (defaults to `https://fapi.binance.com`).
- `EMAIL_OTP_MODE`: set to `debug`, `local`, or `test` to return signup verification codes in `/v1/auth/email/register/start` responses.
- `API_CORS_ALLOWED_ORIGINS`: comma-separated allowed origins for API CORS.
- `ENABLE_DANGEROUS_RESET_ROUTES`: `true`/`false`; enables `/v1/reset-app-state` and `/v1/debug/reset-user` routes (default `false`).
- `RESET_APP_SECRET`: required when `ENABLE_DANGEROUS_RESET_ROUTES=true`; secret for `X-Reset-Secret` header.

Testing:
- `TEST_DATABASE_URL`: optional Postgres connection string for running backend tests.
- `RUN_OPTIONAL_API_TESTS=1`: enables optional external API tests in `app-backend/external_apis_test.go` (CoinMarketCap, CoinGecko, Marketstack, Binance funding rate). Requires the corresponding API keys above.

E2E tests (optional):
- `E2E_BASE_URL`: API base URL (defaults to `http://localhost:8080`).
- `E2E_ENABLE_DEV_ENTITLEMENT`: `true`/`false` to toggle dev entitlement activation (defaults to `true`).
- `E2E_DEVICE_TIMEZONE`: IANA TZ (e.g. `America/Los_Angeles`, defaults to `UTC`).
- `E2E_POLL_INTERVAL_MS`: polling interval in ms (defaults to `1500`).
- `E2E_STAGE_TIMEOUT_MS`: per-stage timeout in ms (defaults to `600000`; use `900000` for a 15-minute stage timeout on slower runs).
- `E2E_RUN_ID`: optional run identifier (auto-generated if unset).

Recommended e2e commands (Docker backend):

```bash
# Targeted smoke (15-minute timeout)
E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 \
  go test ./e2e -run TestDeployedPortfolioCurrencySwitch -timeout 15m

# Full matrix (requires larger total timeout than 15m)
E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 \
  go test ./e2e -timeout 45m
```

### Auth (mobile)
- `GOOGLE_ALLOWED_CLIENT_IDS`: comma-separated OAuth client IDs (iOS, Android, Android debug, Expo Go, Web).
  - Create them in Google Cloud Console → APIs & Services → Credentials.
  - Android requires package name + SHA-1; iOS requires bundle ID; Web uses an OAuth Web client.
- `APPLE_IOS_BUNDLE_ID`: iOS bundle ID used as Apple Sign-In audience (`cc.moneycoach.app.ios`).

### Payments (set only if you enable a provider)
- Stripe (web/PWA checkout): `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_ID_WEEKLY`, `STRIPE_PRICE_ID_YEARLY`.
  - Get keys + price IDs from the Stripe Dashboard → Developers / Products.
- Apple IAP: `APPSTORE_SHARED_SECRET`, `APPLE_IAP_PRODUCT_ID_WEEKLY`, `APPLE_IAP_PRODUCT_ID_YEARLY`.
  - Shared secret: App Store Connect → your app → General → App Information → App-Specific Shared Secret.
- Google Play Billing:
  - `GOOGLE_PLAY_PACKAGE_NAME`: Android package name (`cc.moneycoach.app`).
  - `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON` (raw JSON) or `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64` (base64).
  - `GOOGLE_PLAY_PRODUCT_WEEKLY` + `GOOGLE_PLAY_PRODUCT_ANNUAL` (subscription product IDs used by the app).
  - `GOOGLE_PUBSUB_AUDIENCE` (required for RTDN webhook verification).
- `BILLING_DEV_MODE=true` enables dev entitlement endpoints (set `false` in production).
- Apple/Google subscriptions are bound to a single user account using the original transaction ID (Apple) or purchase token (Google). Reusing a receipt across accounts returns `409 BILLING_CONFLICT`.

### Push (optional)
- APNs: `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_PRIVATE_KEY`, `APNS_BUNDLE_ID`.
- FCM: `FCM_SERVER_KEY`.

## Web Frontend (Next.js)

Install dependencies:

```bash
cd web-frontend
npm install
```

Run locally:

```bash
npm run dev
```

Production build:

```bash
npm run build
npm run start
```

Notes:
- Default port is `3000` and health check is `GET /healthz`.
- The frontend is a public marketing site with no required environment variables.
- Requires Node `20.9+` for Next.js 16.

## Mobile App (Simulators)

Install dependencies:

```bash
cd mobile-app
npm install
```

iOS simulator (uses localhost):

```bash
EXPO_PUBLIC_API_BASE_URL=http://localhost:8080 npm run ios
```

iOS build troubleshooting (`xcodebuild` exit 65, Hermes/rsync code 23):

```bash
cd mobile-app/ios
pod install
rm -rf ~/Library/Developer/Xcode/DerivedData/MoneyCoach-*
cd ..
npm run ios
```

Android emulator (use host loopback):

```bash
EXPO_PUBLIC_API_BASE_URL_ANDROID=http://10.0.2.2:8080 npm run android
```

## Brand Icon Source

- Source of truth: repository root `logo.svg`.
- The SVG must keep an opaque square background (no transparent outer corners) for app-store-safe icon generation.
- Keep the generic brand mark inside a visible safe zone (roughly the inner 80% of the canvas), otherwise some launcher masks will crop the corners.
- For Android launcher icons, keep a dedicated source asset at `mobile-app/assets/adaptive-icon-source.svg`, but render it as the final dark square icon art instead of a nested badge inside a larger transparent plate. The launcher already applies its own mask; a second inner badge creates obvious extra whitespace.
- Set `expo.android.adaptiveIcon.backgroundColor` to the same dark tone as the icon art so the generated adaptive layers stay visually consistent.
- Keep the generated `android:roundIcon` in sync with the adaptive art. The real failure mode was a stale `ic_launcher_foreground` resource, not the presence of `roundIcon` itself.
- Do not add inner scanner-bracket corners to the launcher icon; they make the icon feel cramped once Android applies its own rounded mask.

Regenerate app + web icons after updating `logo.svg`:

```bash
cd /path/to/repo
cp logo.svg web-frontend/public/favicon.svg
TMP_DIR=$(mktemp -d)
qlmanage -t -s 1024 -o "$TMP_DIR" logo.svg >/dev/null 2>&1
SRC="$TMP_DIR/logo.svg.png"
cp "$SRC" mobile-app/assets/icon.png
sips -z 256 256 "$SRC" --out mobile-app/assets/favicon.png >/dev/null

npx -y -p @resvg/resvg-js@2.6.2 node - <<'NODE'
const fs = require('fs');
const { Resvg } = require('@resvg/resvg-js');
const svg = fs.readFileSync('mobile-app/assets/adaptive-icon-source.svg', 'utf8');
const pngData = new Resvg(svg, {
  fitTo: { mode: 'width', value: 1024 },
  background: 'rgba(0,0,0,0)',
}).render().asPng();
fs.writeFileSync('mobile-app/assets/adaptive-icon.png', pngData);
NODE

npx expo prebuild --platform android --no-install
```

Note:
- Do not leave `width="100%" height="100%"` on the root of `mobile-app/assets/adaptive-icon-source.svg`; it can bake an unexpected white `ic_launcher_foreground` even when the SVG previews correctly.
- Android adaptive icon assets must be regenerated with `npx expo prebuild --platform android --no-install` before the next `npx eas-cli build ...`, otherwise EAS may package stale launcher resources.
- In-app purchases require a dev client or EAS build (Expo Go does not support IAP).
- Android IAP builds must include the `react-native-iap` config plugin (adds Billing permission + OpenIAP dependency during prebuild/EAS).
- The PWA build is for internal testing only and is not a production surface.

## Mobile App Environment Variables (Expo)

Required for local/dev:
- `EXPO_PUBLIC_API_BASE_URL`: API base URL (iOS + web).
- `EXPO_PUBLIC_API_BASE_URL_ANDROID`: Android emulator base URL (typically `http://10.0.2.2:8080`).

Google Sign-In:
- `EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID`
- `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID` (required for web/PWA and Android native Google Sign-In)
- `EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID` (Expo Go only)
- Android OAuth client IDs are still required in Google Cloud Console for package/SHA registration, but current app builds do not read `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID` or `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID_DEBUG` at runtime.
- For EAS release builds, set these in EAS project environments (not only local `.env`), otherwise TestFlight/Play builds can miss them at runtime:
  ```bash
  cd mobile-app
  npx eas-cli env:create --environment production --name EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID --value '<your-ios-client-id>' --visibility plaintext --force
  npx eas-cli env:create --environment production --name EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID --value '<your-web-client-id>' --visibility plaintext --force
  npx eas-cli env:list --environment production
  ```
- `mobile-app/eas.json` is configured so profile->environment mapping is explicit:
  - `development` -> `development`
  - `preview` -> `preview`
  - `production` -> `production`
- If these are missing in a release build, Google auth is unavailable and the app shows a configuration alert on the Google button. Email sign-in remains available.
- Regression note (fixed): older builds could crash on SC07 ("Save your profile") if the required Google runtime client ID for that platform was missing at build time. Current code no longer crashes, but you should still set the required Google env vars before TestFlight/Play builds.
- Regression note (fixed): native Google OAuth may return `code` first and populate `id_token` after auth-session exchange; SC07 now waits for that exchange result instead of failing immediately.
- Google OAuth `Error 400: invalid_request` on iOS/TestFlight troubleshooting:
  - In Google Cloud Console -> `APIs & Services` -> `Credentials` -> your iOS OAuth client, verify `Bundle ID` is exactly `cc.moneycoach.app.ios`.
  - Do not use `moneycoach://...` as Google OAuth redirect URI in native builds.
  - Native redirect must resolve to `cc.moneycoach.app.ios:/oauthredirect` (generated by `expo-auth-session/providers/google`).
  - If you tested a build created before this fix, create a new build and retest:
    ```bash
    cd mobile-app
    npx eas-cli build --platform ios --profile production
    ```
- Google OAuth `Error 400: invalid_request` on Android/Google Play troubleshooting:
  - Money Coach Android release builds use package name `cc.moneycoach.app`.
  - Older Android builds (up to versionCode `5`) used `expo-auth-session/providers/google`, which generated the native redirect as `cc.moneycoach.app:/oauthredirect`.
  - A real Google Play-distributed install of Money Coach (versionCode `5`) was retested on Android emulator with Play Store login, and Google returned `Error 400: invalid_request` for the OAuth URL:
    - `client_id=472068794594-npvboq28aloi66gfq7ohgcpqgd8641vo.apps.googleusercontent.com`
    - `redirect_uri=cc.moneycoach.app:/oauthredirect`
  - Root cause: Google no longer supports **custom URI scheme redirects** for Android / Chrome app OAuth browser flows.
  - Fixed implementation (builds generated after this migration, starting with versionCode `6`):
    - Android native builds use `@react-native-google-signin/google-signin`.
    - Android native builds request an ID token using `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`.
    - Expo Go on Android keeps the Expo proxy browser flow because the native module is not available there.
  - Play signing certificate is still required for the Android OAuth client configured in Google Cloud:
  - Open Google Play Console App Integrity for Money Coach:
    - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/app-integrity`
  - In section `App signing key certificate`, copy the `SHA-1` value.
  - Do **not** copy `Upload key certificate` for production Google sign-in.
  - Open Google Cloud Console Credentials:
    - `https://console.cloud.google.com/apis/credentials`
  - Select the same Google Cloud project that owns the existing Money Coach OAuth clients.
  - Click `+ CREATE CREDENTIALS` -> `OAuth client ID`.
  - In `Application type`, choose `Android`.
  - Fill exactly:
    - `Name`: `Money Coach Android (Play Production)`
    - `Package name`: `cc.moneycoach.app`
    - `SHA-1 certificate fingerprint`: paste the `App signing key certificate` SHA-1 copied from Play Console.
  - Click `Create` or `Save`.
  - The Android OAuth client is required in Google Cloud, but the app runtime does **not** read that Android client ID from Expo env anymore.
  - Update the EAS production environment so Android native builds can mint the Google ID token with the web client:
    ```bash
    cd mobile-app
    npx eas-cli env:create --environment production --name EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID --value '<your-web-client-id>' --visibility plaintext --force
    npx eas-cli env:list --environment production | rg EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID
    ```
  - Update backend allowed Google audiences so the backend accepts ID tokens minted for the web client used by Android native sign-in:
    ```bash
    CURRENT="$(gh variable get GOOGLE_ALLOWED_CLIENT_IDS --repo jackcpku/money-coach)"
    gh variable set GOOGLE_ALLOWED_CLIENT_IDS --repo jackcpku/money-coach --body "${CURRENT},<your-web-client-id>"
    ```
  - Redeploy `app-backend` after updating `GOOGLE_ALLOWED_CLIENT_IDS`; the running backend container only reads that value at startup.
  - Rebuild the Android production app after the EAS environment is updated:
    ```bash
    cd mobile-app
    npx eas-cli build --platform android --profile production
    ```
  - Submit the rebuilt AAB to Google Play after the Android implementation has been migrated to native Google sign-in.
  - If a Play-installed Android build still opens a browser OAuth page with `cc.moneycoach.app:/oauthredirect`, you are testing an older build that predates this migration.
  - Local `expo run:android` / debug-build troubleshooting:
    - Native Google Sign-In uses the **currently installed app signature**, so local debug builds need their own Android OAuth client in Google Cloud.
    - If local debug testing shows:
      - `This android application is not registered to use OAuth2.0`
    - then register the debug keystore SHA-1 in Google Cloud:
      1. In terminal:
         ```bash
         cd mobile-app/android
         ./gradlew signingReport
         ```
      2. Under variant `debug`, copy the `SHA1` value.
      3. Open Google Cloud Console -> `APIs & Services` -> `Credentials`.
      4. Click `+ CREATE CREDENTIALS` -> `OAuth client ID`.
      5. Set:
         - `Application type`: `Android`
         - `Name`: `Money Coach Android (Debug)`
         - `Package name`: `cc.moneycoach.app`
         - `SHA-1 certificate fingerprint`: paste the debug SHA-1 from `signingReport`
      6. Save.
    - This debug OAuth client is only for local native testing. Public Play builds still depend on the **Play App Signing** SHA-1.
  - If Google sign-in works on emulator/dev build but fails on the public Play build, check both:
    - the Play App Signing SHA-1 on the Android OAuth client
    - whether the installed build is still using the browser-based `expo-auth-session` custom-scheme flow on Android

Optional:
- `EXPO_PUBLIC_PWA_URL`: used for Stripe checkout return URLs on web/PWA.
- `EXPO_PUBLIC_SHOW_DEV_OTP=1`: shows signup verification codes in dev when using email account creation.
- `EXPO_PUBLIC_TERMS_URL`: Terms of Use URL shown on paywall (recommended; required for App Store submission readiness).
- `EXPO_PUBLIC_PRIVACY_URL`: Privacy Policy URL shown on paywall (recommended; required for App Store submission readiness).

## Waitlist Flow

- Backend: `app-backend/internal/app/handlers_waitlist.go` uses helpers in `app-backend/internal/app/waitlist_store.go` to reuse a static per-user rank and create per-strategy entries.
- Mobile: `mobile-app/src/hooks/useWaitlistEntry.ts` owns the waitlist request lifecycle, and `mobile-app/src/components/WaitlistTicket.tsx` renders the queue ticket UI.

## Report Screen Architecture

- Backend: report plan chart helpers live in `app-backend/internal/app/report_plan_chart.go`.
- Mobile: plan action UI is encapsulated in `mobile-app/src/components/ReportPlanActions.tsx`.
- Mobile: notification permission/CTA state lives in `mobile-app/src/hooks/useReportNotifications.ts`.

## Report Analysis Helpers

- Backend: alpha/radar calculations and plan sorting live in `app-backend/internal/app/analysis_alpha.go`, `app-backend/internal/app/analysis_radar.go`, and `app-backend/internal/app/analysis_plan_sort.go`.
- Backend: report payload/build helpers live in `app-backend/internal/app/analysis_report_types.go` and `app-backend/internal/app/analysis_report_builders.go`.
- Backend: shared JSON marshaling helper lives in `app-backend/internal/app/json_helpers.go`.
- Backend: market series fetch + OHLC normalization lives in `app-backend/internal/app/analysis_series.go` (metrics math stays in `app-backend/internal/app/analysis_metrics.go`).
- Backend: plan parameter helpers (layers, safety orders, additions, typed getters) live in `app-backend/internal/app/plan_params.go`.
- Backend: shared stats helpers (returns, correlation, mean/stddev) live in `app-backend/internal/app/analysis_stats.go`.
- Backend: return series helpers (intersection, price series, drawdown) live in `app-backend/internal/app/analysis_return_series.go`.
- Backend: pairwise correlation helper lives in `app-backend/internal/app/analysis_correlation.go`.
- Backend: plan selection context + candidate logic live in `app-backend/internal/app/analysis_plan_context.go` and `app-backend/internal/app/analysis_plan_select.go`.
- Backend: S01 (smart stop-loss) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s01.go`.
- Backend: S02 (DCA ladder) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s02.go`.
- Backend: S03 (trailing stop) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s03.go`.
- Backend: S04 (take-profit layers) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s04.go`.
- Backend: S05 (steady DCA) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s05.go`.
- Backend: S09 (pyramiding adds) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s09.go`.
- Backend: S16 (cash-and-carry) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s16.go`.
- Backend: S18 (trend regime) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s18.go`.
- Backend: S22 (volatility-weighted rebalance) plan builder + helpers live in `app-backend/internal/app/analysis_plan_s22.go`.
- Backend: indicator math (SMA, RSI, Bollinger, volatility, log returns, close extraction, annualization) lives in `app-backend/internal/app/analysis_indicators.go`.
- Backend: trend classification helpers live in `app-backend/internal/app/analysis_trend.go`.
- Backend: support-level helpers live in `app-backend/internal/app/analysis_support.go`.
- Backend: risk linking helpers live in `app-backend/internal/app/analysis_risk_linking.go`.
- Backend: shared analysis constants live in `app-backend/internal/app/analysis_constants.go`.
- Backend: portfolio helpers (cash classification, crypto weight, annualization) live in `app-backend/internal/app/analysis_portfolio_helpers.go`.
- Backend: market metrics helpers (volatility, drawdown, correlation) live in `app-backend/internal/app/analysis_market_metrics.go`.
- Backend: health score baseline helper lives in `app-backend/internal/app/analysis_health_score.go`.

## Insights Signals

- Backend: insight orchestration (buildInsights) + shared constants live in `app-backend/internal/app/analysis_insights.go`.
- Backend: portfolio watch signal builders live in `app-backend/internal/app/analysis_insights_portfolio_watch.go`.
- Backend: portfolio watch strategy helpers live in `app-backend/internal/app/analysis_insights_portfolio_watch_helpers.go`.
- Backend: action alert signal builders live in `app-backend/internal/app/analysis_insights_action_alert.go`.
- Backend: action alert strategy helpers live in `app-backend/internal/app/analysis_insights_action_alert_helpers.go`.
- Backend: action alert portfolio helpers live in `app-backend/internal/app/analysis_insights_action_alert_portfolio.go`.
- Backend: market alpha signal builders live in `app-backend/internal/app/analysis_insights_market_alpha.go`.
- Backend: market alpha universe selection lives in `app-backend/internal/app/analysis_insights_market_alpha_universe.go`.
- Backend: return/beta helpers live in `app-backend/internal/app/analysis_returns.go`.
- Backend: rebalance helpers live in `app-backend/internal/app/analysis_rebalance_helpers.go`.
- Backend: shared insight helpers (trigger keys, formatting, USD suggestions) live in `app-backend/internal/app/analysis_insights_helpers.go`.

## Holdings Resolution

- Backend: asset resolution logic lives in `app-backend/internal/app/analysis_holdings.go`.
- Backend: CoinGecko ID + balance-type resolution helpers live in `app-backend/internal/app/analysis_holdings_resolution.go`.
- Backend: symbol/platform resolution + stock ticker cache helpers live in `app-backend/internal/app/analysis_holdings_resolution_helpers.go`.
- Backend: per-asset resolution workflow lives in `app-backend/internal/app/analysis_holdings_resolver.go`.
- Backend: aggregation helpers live in `app-backend/internal/app/analysis_holdings_aggregate.go`.
- Backend: net-worth helpers live in `app-backend/internal/app/analysis_holdings_net_worth.go`.
- Backend: pricing + cost basis helpers (avg price normalization + pricing orchestration) live in `app-backend/internal/app/analysis_holdings_pricing.go`.
- Backend: pricing timestamp helper lives in `app-backend/internal/app/analysis_pricing_time.go`.
- Backend: pricing fetch helpers live in `app-backend/internal/app/analysis_holdings_pricing_fetch.go`.
- Backend: asset pricing helpers + shared USD valuation setter live in `app-backend/internal/app/analysis_holdings_pricing_assets.go`.
- Backend: user-provided value helpers live in `app-backend/internal/app/analysis_holdings_user_values.go`.

## Delta Application

- Backend: portfolio delta application lives in `app-backend/internal/app/analysis_delta.go`.
- Backend: asset delta helpers live in `app-backend/internal/app/analysis_delta_asset.go`.
- Backend: cash adjustment helpers live in `app-backend/internal/app/analysis_delta_cash.go`.

## Market Data

- Backend: Binance OHLC parsing + symbol candidate helpers live in `app-backend/internal/app/market_binance_helpers.go`.
- Backend: indicator series selection helpers live in `app-backend/internal/app/analysis_indicator_series.go`.
- Backend: market snapshot defaults live in `app-backend/internal/app/analysis_market_snapshot.go`.

## Currency Helpers

- Backend: base currency conversions + FX rate helpers live in `app-backend/internal/app/currency.go`.
- Backend: display currency parsing + screenshot conversions live in `app-backend/internal/app/currency_display.go`.

## Strategy Detail Architecture

- Mobile: Strategy Detail view-model helpers (rules, intro copy, chart candles, schedule) live in `mobile-app/src/utils/strategyDetail.ts`.
- Backend: report plan loading/decoding is centralized in `app-backend/internal/app/report_store.go`.

## Apple Sign-In + IAP Setup

### Apple Sign-In (iOS-only)
1. Apple Developer → Certificates, Identifiers & Profiles → Identifiers.
2. Create or edit your App ID (bundle ID must match `ios.bundleIdentifier` in `mobile-app/app.json`).
3. Enable the **Sign in with Apple** capability.
4. Set `APPLE_IOS_BUNDLE_ID` to the same bundle ID (GitHub Variable and `app-backend/.env` for local).

### Apple IAP (subscriptions)
1. App Store Connect → your app → **Monetization** → **Subscriptions**.
2. Create a **Subscription Group** (example: `Money Coach Pro`).
3. Open that group and create two subscriptions (weekly + yearly).
4. If `+` asks for `Type (Consumable / Non-Consumable)`, you are in **In-App Purchases** (wrong page for subscriptions). Return to **Monetization → Subscriptions**.
5. Copy each **Product ID**:
   - `APPLE_IAP_PRODUCT_ID_WEEKLY`
   - `APPLE_IAP_PRODUCT_ID_YEARLY`
6. App Store Connect → your app → **General** → **App Information** → **App-Specific Shared Secret** → copy it to `APPSTORE_SHARED_SECRET`.
7. App Store Connect → your app → **General** → **App Information** → **App Store Server Notifications**:
   - Set **Production Server URL** to `https://api.moneycoach.cc/v1/webhooks/apple`.
   - Set **Sandbox Server URL** to `https://api.moneycoach.cc/v1/webhooks/apple`.
   - If a **Version** selector appears, choose **Version 2**.

## Google Play Production Release (Exact Fill Guide)

Use this section for Android ship. It is intentionally click-by-click, and it matches the current Play Console app that is already live.

Current fixed identifiers for this repo:
- Play Console app base URL: `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470`
- Android package name: `cc.moneycoach.app`
- Expo project owner/slug: `jackcpku/money-coach`
- EAS build profile: `production`
- EAS submit profiles:
  - `android-internal`
  - `production`
- Review account:
  - Email: `nigaji3479@iaciu.com`
  - Password: `nigaji3479@iaciu.com`

### 0) Hard rule: Android release/publish uses `npx eas-cli` only
Do not use GitHub Actions for Android release/publish.

Normal commands:

```bash
cd mobile-app
npx eas-cli build --platform android --profile production
npx eas-cli submit --platform android --profile android-internal --latest
npx eas-cli submit --platform android --profile production --latest
```

### 1) Pre-flight checks before every Android release
1. Confirm repo identifiers:
   - `mobile-app/app.json` -> `expo.android.package`: `cc.moneycoach.app`
   - `mobile-app/eas.json` -> submit profile `android-internal` maps to Play track `internal`
   - `mobile-app/eas.json` -> submit profile `production` maps to Play track `production`
2. Confirm production legal URLs are live:
   - `https://moneycoach.cc/terms`
   - `https://moneycoach.cc/privacy`
   - `https://moneycoach.cc/support`
3. Confirm Google Play product IDs:
   - `weekly`
   - `yearly`
4. Run mobile checks:
   ```bash
   cd mobile-app
   npm run lint
   npm run test -- --runInBand
   npx tsc --noEmit -p tsconfig.json
   npx expo-doctor
   ```
5. Run backend e2e against local Docker backend before a production push:
   ```bash
   cd app-backend
   docker compose up -d postgres redis
   docker compose up -d backend
   E2E_BASE_URL=http://localhost:8080 E2E_STAGE_TIMEOUT_MS=900000 go test ./e2e -timeout 45m
   ```
6. Confirm current Android release behavior on a Play-installed build before promoting:
   - Google sign-in
   - Email sign-in
   - Report open
   - Strategy detail open
   - Subscription purchase / restore

### 2) One-time Google OAuth setup for Android release builds
This is already configured for the current live app. Re-check it only if Google sign-in breaks on a Play-installed build.

1. Open Play Console app signing page:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/keymanagement`
2. In section `App signing key certificate`, copy `SHA-1 certificate fingerprint`.
   - Current value for this app: `0C:60:FB:28:F5:07:C9:41:A0:DE:9F:12:96:B0:B6:F9:E1:D2:5B:6C`
3. Do **not** use the `Upload key certificate` SHA-1 for Google OAuth.
   - Current upload-key SHA-1 is `04:95:84:99:44:AD:4C:84:F7:10:35:CD:AB:B8:FD:47:30:A7:F9:82`
4. Open Google Cloud OAuth credentials:
   - `https://console.cloud.google.com/apis/credentials`
5. Open the Android OAuth client for Money Coach release builds.
   - Current client ID: `472068794594-npvboq28aloi66gfq7ohgcpqgd8641vo.apps.googleusercontent.com`
6. Verify these exact fields:
   - `Application type`: `Android`
   - `Package name`: `cc.moneycoach.app`
   - `SHA-1 certificate fingerprint`: `0C:60:FB:28:F5:07:C9:41:A0:DE:9F:12:96:B0:B6:F9:E1:D2:5B:6C`
7. Verify the EAS production environment contains the Google **Web** client ID used by the native Android sign-in implementation:
   - `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`
8. Verify backend audience allow-list includes that same web client ID:
   - `GOOGLE_ALLOWED_CLIENT_IDS`
9. Important implementation note:
   - Android release builds now use native `@react-native-google-signin/google-signin`.
   - They do **not** use the old browser OAuth redirect `cc.moneycoach.app:/oauthredirect`.
   - If a Play-installed build still opens a browser Google OAuth page and fails with `Error 400: invalid_request`, you are testing an old build or the OAuth client is misconfigured.

### 3) One-time Play API / service-account setup
This is required for:
- EAS submit to Google Play
- backend Google Play billing verification

1. Open Play Console API access page:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/dev-account-details/api-access`
2. If Play Console asks you to link a Google Cloud project, click `Link project` and finish that first.
3. In `Service accounts`, click `View Play Console permissions`.
4. If `moneycoach-play-api` already exists, click it. Otherwise create it:
   - Open Google Cloud service accounts:
     - `https://console.cloud.google.com/iam-admin/serviceaccounts`
   - Click `+ CREATE SERVICE ACCOUNT`
   - `Service account name`: `moneycoach-play-api`
   - Click `Create and continue`
   - Do not add IAM roles here unless your org requires them
   - Click `Done`
5. Create a JSON key for that service account.
   - If Google Cloud blocks key creation with `iam.disableServiceAccountKeyCreation`, ask the org policy admin to allow service-account key creation for this project, then retry.
6. Back in Play Console -> `API access`, click `Grant access` for `moneycoach-play-api`.
7. In tab `App permissions`:
   - Click `Add app`
   - Choose `Money Coach`
   - Click `Apply`
8. Turn on these app permissions:
   - `View app information (read-only)`
   - `Release apps to testing tracks`
   - `Manage testing tracks and edit tester lists`
   - `Manage store presence`
   - `Manage policy declarations`
   - `View financial data`
   - `Manage orders and subscriptions`
   - For direct production submit, also enable:
     - `Release to production, exclude devices, and use Play App Signing`
9. Click `Apply` -> `Save`.
10. Save the JSON key file locally. Example file name used on this machine:
    - `gen-lang-client-0470835121-bb94926ef367.json`
11. Put the same JSON into backend deployment secrets, because backend billing verification reads `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64`:
    ```bash
    base64 < /absolute/path/to/play-service-account.json | tr -d '\n' | gh secret set GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64 --repo jackcpku/money-coach
    ```
12. Verify the secret exists:
    ```bash
    gh secret list --repo jackcpku/money-coach | rg GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64
    ```
13. First time you run `npx eas-cli submit` on a new machine:
    - run it **without** `--non-interactive`
    - when EAS asks for the Google service account JSON path, enter the absolute path to the downloaded JSON file
    - allow EAS to save that credential for future submissions
14. After EAS has the credential saved, later submissions can use:
    ```bash
    npx eas-cli submit --platform android --profile android-internal --latest
    npx eas-cli submit --platform android --profile production --latest
    ```

### 4) One-time Play Console listing / settings values
This app already exists. For future releases, open each page, verify the values below, and change only if you intentionally want a public-facing listing update.

#### A. Main store listing
1. Open:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/main-store-listing`
2. Fill / verify:
   - `App name`: keep the current published name. At the moment this is `Money Coach`.
   - `Short description`:
     ```text
     AI portfolio coach for market signals, portfolio health, and clear trade plans.
     ```
   - `Full description`:
     ```text
     Money Coach helps you turn portfolio data into clear, actionable decisions.

     WHAT YOU CAN DO
     • See market signals tailored to your holdings
     • Track portfolio health and key valuation metrics
     • Get AI-generated report summaries and strategy details
     • Follow concrete execution rules instead of guessing entries and exits
     • Keep your portfolio view, insights, and profile in one app

     WHY USERS CHOOSE MONEY COACH
     • Structured guidance over noisy feeds
     • Clear next actions, not generic commentary
     • A consistent workflow from insight to execution

     ACCOUNT & ACCESS
     • Sign in with Apple, Google, or Email
     • Email sign-in supports account creation and secure login
     • Subscription unlocks full premium report and strategy features

     Money Coach is designed for serious investors who want disciplined, repeatable portfolio decisions.
     ```
3. Upload / verify graphics:
   - `App icon`: `artifacts/android/store-listing/google-play-icon-512.png`
   - `Feature graphic`: `artifacts/android/store-listing/google-play-feature-graphic-1024x500.png`
   - These files are local release-workstation assets. If they do not exist on the current machine, regenerate Android store-listing assets before opening Play Console.
4. Upload / verify phone screenshots in this order:
   - `artifacts/android/store-listing/phone-1080x1920-ordered/01-insights.png`
   - `artifacts/android/store-listing/phone-1080x1920-ordered/02-assets.png`
   - `artifacts/android/store-listing/phone-1080x1920-ordered/03-report-overview.png`
   - `artifacts/android/store-listing/phone-1080x1920-ordered/04-strategy-detail.png`
   - `artifacts/android/store-listing/phone-1080x1920-ordered/05-profile.png`
   - `artifacts/android/store-listing/phone-1080x1920-ordered/06-onboarding.png`
5. Click `Save`.

#### B. Store settings
1. Open:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/store-settings`
2. Fill / verify:
   - `App or game`: `App`
   - `Category`: `Finance`
   - `Email address`: `support@moneycoach.cc`
   - `Website`: `https://moneycoach.cc/`
   - `Phone number`: leave blank unless you want a public phone number on the listing
3. `External marketing`:
   - leave on or off according to your distribution preference; it does not block release
4. Click `Save`.

### 5) App content pages (exact pages and exact text)
#### A. Privacy policy
1. Open:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/app-content/privacy-policy`
2. If the card is incomplete, click `Start` or `Manage`.
3. Paste:
   - `https://moneycoach.cc/privacy`
4. Click `Save` and `Submit`.

#### B. App access / testing credentials
1. Open:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/app-content/testing-credentials?source=dashboard`
2. Select:
   - `All or some functionality in my app is restricted`
3. Click `Add instructions`.
4. Fill exactly:
   - `Instruction name *`: `Main review access (email + Pro)`
   - `Username, email address, or phone number`: `nigaji3479@iaciu.com`
   - `Password`: `nigaji3479@iaciu.com`
   - `Any other information required to access your app`:
     ```text
     Use Email login for review (do not use Google/Apple login).
     Steps:
     1) Open app and complete onboarding until "Save your profile".
     2) Tap "Continue with Email".
     3) On "Sign in with email", select the "Sign in" tab.
     4) Enter the credentials above and sign in.
     This account is preconfigured as Pro Member, so premium report/strategy features are accessible without making a purchase.
     No 2FA/OTP is required.
     ```
5. Turn on:
   - `Allow Android to use the credentials you provide for performance and app compatibility testing`
6. Click `Save`.

#### C. Photo and video permissions
1. Current Android manifest is policy-safe:
   - allowed permissions:
     - `android.permission.VIBRATE`
     - `com.android.vending.BILLING`
   - blocked permissions include:
     - `READ_MEDIA_IMAGES`
     - `READ_MEDIA_VIDEO`
     - `READ_MEDIA_VISUAL_USER_SELECTED`
2. If Play still shows:
   - `Invalid use of the photo and video permissions`
3. Then one of your active tracks still contains an old build. Fix it like this:
   - open `Publishing overview`:
     - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/publishing`
   - replace old builds on **every** active track (`Internal testing`, `Closed testing`, `Production`) with a build created after this permission removal
   - for this repo, any build based on the current app.json permission blocklist is safe; the tested public fix path started at versionCode `5`
   - after each track is updated, refresh `Publishing overview`

#### D. Other app-content cards
For this already-live app, keep the currently accepted answers unless product behavior changes:
- `Content Rating`
- `Target audience and content`
- `Ads declaration`
- `Data safety`
- `Financial features`
- `Health`
- `Government apps`
- `Advertising ID`

Rule:
- If a card is already approved and the product behavior has not changed, do not rewrite it during a normal release.
- If Play shows one of these cards as incomplete, complete it before submission, then return to `Publishing overview`.

### 6) Play subscription setup (verify only unless you are changing pricing)
1. Open subscriptions page:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/monetization/subscriptions`
2. Verify two subscription products exist and are active:
   - `weekly`
   - `yearly`
3. Verify product details:
   - `weekly`
     - `Subscription name`: `Money Coach Pro Weekly`
     - base plan: `weekly-auto`
     - price: `USD 9.99`
   - `yearly`
     - `Subscription name`: `Money Coach Pro Annual`
     - base plan: `yearly-auto`
     - price: `USD 99.90`
4. If you need to create them from scratch:
   - click `Create subscription`
   - `Product ID`: `weekly` or `yearly`
   - `Subscription name`: use the names above
   - `Description`: `Access to full risk analysis, strategy details, and alerts.`
   - click `Save`
   - click `Add base plan`
   - use the base-plan IDs above
   - set `Auto-renewing`
   - set price
   - activate the base plan

### 7) Build the Android AAB with EAS
1. Build:
   ```bash
   cd mobile-app
   npx eas-cli build --platform android --profile production
   ```
2. Important behavior:
   - `mobile-app/eas.json` has `build.production.autoIncrement=true`
   - every production build increments `versionCode` remotely
3. If Play says `Version code X has already been used`:
   - do **not** re-upload the same `.aab`
   - run another production build to get a new version code
4. If you want the `.aab` file locally:
   - download it from the EAS build page after the build finishes
   - or use the direct artifact URL shown by EAS

### 8) Internal smoke-test release via EAS submit
This is the required gate before production.

1. Submit the latest finished production build to the internal track:
   ```bash
   cd mobile-app
   npx eas-cli submit --platform android --profile android-internal --latest
   ```
2. Open the internal track page:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/tracks/4700741181755087135`
3. If Play opens `Create internal testing release`, fill:
   - `Release name *`:
     ```text
     0.1.0 (<versionCode>) - Internal test update
     ```
   - `Release notes`:
     ```text
     Internal testing update.
     - Android native Google sign-in flow
     - System photo picker trade-slip upload flow
     - Sign-in stability and onboarding reliability improvements
     - General bug fixes and polish
     ```
4. Click:
   - `Next`
   - `Start rollout to internal testing`
5. Install the Play-delivered build from the internal tester account and verify:
   - Google sign-in returns to app successfully
   - Email sign-in works
   - Assets / Insights / Me tabs work
   - Full Report opens
   - Strategy Detail opens
   - Purchase / restore works in Play billing

### 9) Production release via EAS submit
1. Submit the latest finished production build to the production track:
   ```bash
   cd mobile-app
   npx eas-cli submit --platform android --profile production --latest
   ```
2. Open production-track release page:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/tracks/4697379121899730493`
3. If Play opens `Create production release`, fill:
   - `Release name *`:
     ```text
     0.1.0 (<versionCode>) - Production update
     ```
   - `Release notes`:
     ```text
     Production update.
     - Portfolio insights, reports, and strategy detail improvements
     - Android native Google sign-in flow
     - System photo picker trade-slip upload flow
     - Stability improvements and bug fixes
     ```
4. Click `Next` to reach `Preview and confirm`.
5. If Play shows the blocking error:
   - `No countries or regions have been selected for this track`
6. Fix it like this:
   - open `Publishing overview`:
     - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/publishing`
   - in the `Production` row, click `Countries / regions`
   - click `Add countries / regions`
   - select the countries you want live now
   - at minimum, select `United States` if you need to unblock submission quickly
   - click `Apply`
   - click `Save`
7. Return to the release page.
8. Ignore the warning:
   - `There is no deobfuscation file associated with this App Bundle`
   - it does **not** block submission
9. Go back to `Publishing overview`:
   - `https://play.google.com/console/u/0/developers/4814175513902304532/app/4974678330853071470/publishing`
10. Confirm the page shows:
    - Production row with the new version code
    - no blocking app-content issue
    - store listing complete
11. Click `Send changes for review`.
12. If `Managed publishing` is off, Google publishes automatically after approval.
13. If you want manual control after approval:
    - turn `Managed publishing` on **before** clicking `Send changes for review`

### 10) RTDN / backend webhook verification
1. In Play Console, open RTDN settings:
   - `Monetize with Play` -> `Monetization setup` -> `Real-time developer notifications`
2. Ensure the Pub/Sub topic is linked.
3. In Google Cloud Pub/Sub, ensure the push subscription points to:
   - `https://api.moneycoach.cc/v1/webhooks/google`
4. Ensure push subscription OIDC audience is:
   - `https://api.moneycoach.cc/v1/webhooks/google`
5. Ensure GitHub variable is set:
   ```bash
   gh variable set GOOGLE_PUBSUB_AUDIENCE --repo jackcpku/money-coach --body 'https://api.moneycoach.cc/v1/webhooks/google'
   ```

### 11) Common blockers and exact fixes
1. `Error 400: invalid_request` after Google login on public Android build:
   - verify Play App Signing SHA-1, not upload-key SHA-1
   - verify Android OAuth client uses package `cc.moneycoach.app`
   - verify you are testing a build that already migrated to native Android Google sign-in
2. `Version code X has already been used`:
   - run a new `npx eas-cli build --platform android --profile production`
3. `Invalid use of the photo and video permissions`:
   - remove older active builds from all tracks
   - keep all active tracks on builds created after the permission cleanup
4. `No countries or regions have been selected for this track`:
   - add countries on `Publishing overview` -> `Production` -> `Countries / regions`
5. `Your Android App Bundle is signed with the wrong key`:
   - your upload key does not match the one registered with Play
   - fix Play upload-key registration first, then rebuild/upload
6. EAS submit cannot talk to Play:
   - service account credential not configured in EAS yet
   - or Play Console permissions on `moneycoach-play-api` are missing

### 12) Final Android ship checklist
Release only when all boxes are true:
- `npx eas-cli build --platform android --profile production` succeeded
- `npx eas-cli submit --platform android --profile android-internal --latest` succeeded
- internal Play-installed build passed auth / report / purchase QA
- `npx eas-cli submit --platform android --profile production --latest` succeeded
- `Publishing overview` shows no blocking issue
- `App access` review credentials are up to date
- `Privacy policy` is `https://moneycoach.cc/privacy`
- countries / regions are selected
- Google sign-in works on the Play-installed production build

When all of the above are true, click `Send changes for review`.

## Push Notifications (Optional)

If you want to send push notifications, configure the APNs and FCM credentials below. The backend skips push delivery when these are unset.

### iOS (APNs)
1. Apple Developer -> **Certificates, Identifiers & Profiles** -> **Keys**.
2. Create a new key with **Apple Push Notification service (APNs)** enabled.
3. Download the `.p8` key and note:
   - `APNS_KEY_ID` (Key ID shown in the portal)
   - `APNS_TEAM_ID` (your Apple Developer Team ID)
4. Set `APNS_PRIVATE_KEY` to the **contents** of the `.p8` file (PEM) or base64-encode the PEM.
5. Set `APNS_BUNDLE_ID` to your iOS bundle ID (`cc.moneycoach.app.ios`).
6. In CI: store `APNS_PRIVATE_KEY` as a GitHub Secret and `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID` as GitHub Variables.

### Android (FCM)
1. Firebase Console -> your project -> **Project settings** -> **Cloud Messaging**.
2. Copy the **Server key** (Legacy) and set `FCM_SERVER_KEY`.
3. In CI: store `FCM_SERVER_KEY` as a GitHub Secret.

## GitHub CI/CD Configuration

The workflows read from GitHub **Secrets** and **Variables**. Secrets are always required (no defaults). Variables are split below into ones **without defaults** and ones **with defaults**.

Important: workflows reference `vars.*`. If you store a value as a **Secret** instead of a **Variable**, the workflow will not see it unless the workflow is updated.

### GitHub Secrets (No Defaults)

Backend deploy (`.github/workflows/ci.yml`):
- `AWS_ROLE_ARN`
- `JWT_SIGNING_SECRET`
- `GEMINI_API_KEY`
- `RESEND_API_KEY`
- `CMC_PRO_API_KEY`
- `COINGECKO_PRO_API_KEY`
- `MARKETSTACK_ACCESS_KEY`
- `OPEN_EXCHANGE_APP_ID`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `APPSTORE_SHARED_SECRET` (only if using Apple IAP)
- `APNS_PRIVATE_KEY` (only if sending iOS push notifications)
- `FCM_SERVER_KEY` (only if sending Android push notifications)
- `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64` (only if using Google Play Billing)

Legacy Android release workflow (`.github/workflows/mobile-app-android-release.yml`) — do not use for mobile release/publish:
- `ANDROID_UPLOAD_KEYSTORE_BASE64`
- `ANDROID_UPLOAD_STORE_PASSWORD`
- `ANDROID_UPLOAD_KEY_ALIAS`
- `ANDROID_UPLOAD_KEY_PASSWORD`
- `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON_BASE64`

PWA deploy (`.github/workflows/mobile-app-pwa-deploy.yml`):
- `AWS_ROLE_ARN`

Web frontend deploy (`.github/workflows/web-frontend-deploy.yml`):
- `AWS_ROLE_ARN`

### GitHub Secrets (Defaults)

None.

### GitHub Variables (No Defaults)

Backend deploy (`.github/workflows/ci.yml`):
- `RESEND_FROM_EMAIL`
- `STRIPE_PRICE_ID_WEEKLY`
- `STRIPE_PRICE_ID_YEARLY`
- `HOSTED_ZONE_ID`
- `APPLE_IOS_BUNDLE_ID` (required for iOS Sign in with Apple or Apple IAP)
- `APPLE_IAP_PRODUCT_ID_WEEKLY` (only if using Apple IAP)
- `APPLE_IAP_PRODUCT_ID_YEARLY` (only if using Apple IAP)
- `APNS_KEY_ID` (only if sending iOS push notifications)
- `APNS_TEAM_ID` (only if sending iOS push notifications)
- `APNS_BUNDLE_ID` (only if sending iOS push notifications)
- `GOOGLE_PLAY_PACKAGE_NAME` (only if using Google Play Billing)
- `GOOGLE_PLAY_PRODUCT_WEEKLY` (required for Expo IAP)
- `GOOGLE_PLAY_PRODUCT_ANNUAL` (required for Expo IAP)
- `GOOGLE_PUBSUB_AUDIENCE` (required for Google RTDN webhook verification)

Legacy Android release workflow (`.github/workflows/mobile-app-android-release.yml`) — do not use for mobile release/publish:
- `GOOGLE_PLAY_PACKAGE_NAME`
- `EXPO_PUBLIC_API_BASE_URL`
- `EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID`
- `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID`
- `EXPO_PUBLIC_STRIPE_PUBLISHABLE_KEY`
- `EXPO_PUBLIC_API_BASE_URL_ANDROID` (recommended)
- `EXPO_PUBLIC_PWA_URL` (recommended)
- `EXPO_PUBLIC_TERMS_URL` (recommended)
- `EXPO_PUBLIC_PRIVACY_URL` (recommended)
- `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID_DEBUG` (recommended)
- `EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID` (recommended)
- `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID` (recommended)
- `EXPO_PUBLIC_APPLE_IAP_PRODUCT_ID_WEEKLY` (recommended)
- `EXPO_PUBLIC_APPLE_IAP_PRODUCT_ID_YEARLY` (recommended)

PWA deploy (`.github/workflows/mobile-app-pwa-deploy.yml`):
- `HOSTED_ZONE_ID` (same value as backend)
- `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`
- `EXPO_PUBLIC_PWA_URL`

Web frontend deploy (`.github/workflows/web-frontend-deploy.yml`):
- `HOSTED_ZONE_ID` (same value as backend)

Mobile app build-time (Expo env vars; set as GitHub Variables if you want CI to inject them):
- `EXPO_PUBLIC_API_BASE_URL`
- `EXPO_PUBLIC_API_BASE_URL_ANDROID`
- `EXPO_PUBLIC_PWA_URL`
- `EXPO_PUBLIC_TERMS_URL`
- `EXPO_PUBLIC_PRIVACY_URL`
- `EXPO_PUBLIC_SHOW_DEV_OTP`
- `EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID`
- `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID`
- `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID_DEBUG`
- `EXPO_PUBLIC_GOOGLE_EXPO_CLIENT_ID`
- `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`

Note: the app reads plan product IDs from `GET /v1/billing/plans`, so you typically don't need extra Expo IAP env vars beyond the Google OAuth client IDs listed above.

### GitHub Variables (Defaults)

Backend deploy (`.github/workflows/ci.yml`):
- `AWS_REGION` (default `ap-southeast-1`)
- `STACK_NAME` (default `app-backend`)
- `SERVICE_NAME` (default `app-backend`)
- `ECR_REPO` (default `app-backend`)
- `CFN_TEMPLATE` (default `app-backend/cloudformation/backend.yaml`)
- `DOCKER_BUILD_CONTEXT` (default `app-backend`)
- `DOCKERFILE_PATH` (default `app-backend/Dockerfile`)
- `CPU` (default `1024`)
- `MEMORY` (default `2048`)
- `DESIRED_COUNT` (default `1`)
- `DB_INSTANCE_CLASS` (default `db.t4g.micro`)
- `DB_ALLOCATED_STORAGE` (default `20`)
- `DB_NAME` (default `appdb`)
- `REDIS_NODE_TYPE` (default `cache.t4g.micro`)
- `UPLOADS_CORS_ALLOWED_ORIGINS` (default `*`)
- `API_CORS_ALLOWED_ORIGINS` (default `*`)
- `BINANCE_API_BASE_URL` (default `https://api.binance.com`)
- `GOOGLE_ALLOWED_CLIENT_IDS` (default empty; include iOS, Android, Android debug, Expo Go, and Web client IDs)
- `HOSTED_ZONE_NAME` (default `moneycoach.cc.`)
- `API_DOMAIN_NAME` (default `api.moneycoach.cc`)

PWA deploy (`.github/workflows/mobile-app-pwa-deploy.yml`):
- `AWS_REGION` (default `ap-southeast-1`)
- `AWS_CERT_REGION` (default `us-east-1`)
- `PWA_STACK_NAME` (default `moneycoach-pwa`)
- `PWA_CERT_STACK_NAME` (default `moneycoach-pwa-cert`)
- `PWA_DOMAIN_NAME` (default `demo.moneycoach.cc`)
- `HOSTED_ZONE_NAME` (default `moneycoach.cc.`)
- `PWA_BUCKET_NAME` (default `demo.moneycoach.cc`)
- `PWA_API_BASE_URL` (default `https://api.moneycoach.cc`)

Web frontend deploy (`.github/workflows/web-frontend-deploy.yml`):
- `AWS_REGION` (default `ap-southeast-1`)
- `WEB_STACK_NAME` (default `moneycoach-web`)
- `WEB_SERVICE_NAME` (default `moneycoach-web`)
- `WEB_ECR_REPO` (default `moneycoach-web`)
- `WEB_CFN_TEMPLATE` (default `web-frontend/cloudformation/frontend.yaml`)
- `WEB_DOCKER_BUILD_CONTEXT` (default `web-frontend`)
- `WEB_DOCKERFILE_PATH` (default `web-frontend/Dockerfile`)
- `WEB_CPU` (default `512`)
- `WEB_MEMORY` (default `1024`)
- `WEB_DESIRED_COUNT` (default `1`)
- `WEB_CONTAINER_PORT` (default `3000`)
- `WEB_ROOT_DOMAIN_NAME` (default `moneycoach.cc`)
- `WEB_WWW_DOMAIN_NAME` (default `www.moneycoach.cc`)
- `HOSTED_ZONE_NAME` (default `moneycoach.cc.`)
- `WEB_HEALTH_CHECK_PATH` (default `/healthz`)
