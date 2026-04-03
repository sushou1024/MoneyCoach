# Implementation Gaps (Intentional Drifts)

This document records intentional deviations from the specs for the initial backend implementation.

## Payment & Entitlements
- None.

## Auth
- Google OAuth login (`/v1/auth/oauth`) is implemented; Apple OAuth login is still not implemented.

## Insights
- iOS/Android push delivery requires APNS/FCM credentials; without those, insights are generated but no push is sent.

## Precision
- Monetary fields are stored as float64 in GORM models rather than numeric(36,18)/numeric(36,6) with a decimal type.

## LLM Prompt Parsing
- LLM outputs are parsed into prompt-shaped structs; the backend injects computed fields before persisting and returning API payloads.
- The backend overwrites several prompt-output fields at save time. Preview reports always replace fixed_metrics, net_worth_display, base_currency, base_fx_rate_to_usd, asset_allocation, and meta_data/valuation/snapshot identifiers. Paid reports replace report_header values/statuses, charts.radar_chart, net_worth_display, base_currency/base_fx_rate_to_usd, asset_allocation, risk_summary, daily_alpha_signal, and meta_data/valuation/snapshot identifiers.
- Trade slip OCR parsing tolerates null amount/price/currency/fees fields. Missing amount causes the trade row to be skipped; missing price/currency fall back to market pricing logic.
