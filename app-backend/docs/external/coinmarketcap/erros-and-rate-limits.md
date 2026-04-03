https://coinmarketcap.com/api/documentation/v1/#section/Errors-and-Rate-Limits

# Errors and Rate Limits

## API Request Throttling
Use of the CoinMarketCap API is subject to API call rate limiting or "request throttling". This is the number of HTTP calls that can be made simultaneously or within the same minute with your API Key before receiving an HTTP 429 "Too Many Requests" throttling error. This limit scales with the usage tier and resets every 60 seconds. Please review our Best Practices for implementation strategies that work well with rate limiting.

## HTTP Status Codes
The API uses standard HTTP status codes to indicate the success or failure of an API call.

- 400 (Bad Request) The server could not process the request, likely due to an invalid argument.
- 401 (Unauthorized) Your request lacks valid authentication credentials, likely an issue with your API Key.
- 402 (Payment Required) Your API request was rejected due to it being a paid subscription plan with an overdue balance. Pay the balance in the Developer Portal billing tab and it will be enabled.
- 403 (Forbidden) Your request was rejected due to a permission issue, likely a restriction on the API Key's associated service plan. Here is a convenient map of service plans to endpoints.
- 429 (Too Many Requests) The API Key's rate limit was exceeded; consider slowing down your API Request frequency if this is an HTTP request throttling error. Consider upgrading your service plan if you have reached your monthly API call credit limit for the day/month.
- 500 (Internal Server Error) An unexpected server issue was encountered.

## Error Response Codes
A Status object is always included in the JSON response payload for both successful calls and failures when possible. During error scenarios you may reference the error_code and error_message properties of the Status object. One of the API error codes below will be returned if applicable otherwise the HTTP status code for the general error type is returned.

| HTTP Status | Error Code                                     | Error Message                                                                           |
|-------------|------------------------------------------------|-----------------------------------------------------------------------------------------|
| 401         | 1001 [API_KEY_INVALID]                         | This API Key is invalid.                                                                |
| 401         | 1002 [API_KEY_MISSING]                         | API key missing.                                                                        |
| 402         | 1003 [API_KEY_PLAN_REQUIRES_PAYEMENT]          | Your API Key must be activated. Please go to pro.coinmarketcap.com/account/plan.        |
| 402         | 1004 [API_KEY_PLAN_PAYMENT_EXPIRED]            | Your API Key's subscription plan has expired.                                           |
| 403         | 1005 [API_KEY_REQUIRED]                        | An API Key is required for this call.                                                   |
| 403         | 1006 [API_KEY_PLAN_NOT_AUTHORIZED]             | Your API Key subscription plan doesn't support this endpoint.                           |
| 403         | 1007 [API_KEY_DISABLED]                        | This API Key has been disabled. Please contact support.                                 |
| 429         | 1008 [API_KEY_PLAN_MINUTE_RATE_LIMIT_REACHED]  | You've exceeded your API Key's HTTP request rate limit. Rate limits reset every minute. |
| 429         | 1009 [API_KEY_PLAN_DAILY_RATE_LIMIT_REACHED]   | You've exceeded your API Key's daily rate limit.                                        |
| 429         | 1010 [API_KEY_PLAN_MONTHLY_RATE_LIMIT_REACHED] | You've exceeded your API Key's monthly rate limit.                                      |
| 429         | 1011 [IP_RATE_LIMIT_REACHED]                   | You've hit an IP rate limit.                                                            |