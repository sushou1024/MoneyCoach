# Get Funding Rate Info
## API Description
Query funding rate info for symbols that had FundingRateCap/ FundingRateFloor / fundingIntervalHours adjustment

## HTTP Request
GET /fapi/v1/fundingInfo

## Request Weight
0 share 500/5min/IP rate limit with GET /fapi/v1/fundingInfo

## Request Parameters

None

## Response Example

```
[
    {
        "symbol": "BLZUSDT",
        "adjustedFundingRateCap": "0.02500000",
        "adjustedFundingRateFloor": "-0.02500000",
        "fundingIntervalHours": 8,
        "disclaimer": false   // ingore
    }
]
```

Note: This endpoint responses with the information of every symbol, and the information doesn't seem to be updated very frequently.