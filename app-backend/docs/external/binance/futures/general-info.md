# General Info
## General API Information
- Some endpoints will require an API Key. Please refer to this page
- The base endpoint is: https://fapi.binance.com
- All endpoints return either a JSON object or array.
- Data is returned in ascending order. Oldest first, newest last.
- All time and timestamp related fields are in milliseconds.
- All data types adopt definition in JAVA.
### Testnet API Information
- Most of the endpoints can be used in the testnet platform.
- The REST base url for testnet is "https://demo-fapi.binance.com"
- The Websocket base url for testnet is "wss://fstream.binancefuture.com"
## General Information on Endpoints
- For GET endpoints, parameters must be sent as a query string.
- For POST, PUT, and DELETE endpoints, the parameters may be sent as a query string or in the request body with content type application/x-www-form-urlencoded. You may mix parameters between both the query string and request body if you wish to do so.
- Parameters may be sent in any order.
- If a parameter sent in both the query string and request body, the query string parameter will be used.
### HTTP Return Codes
- HTTP 4XX return codes are used for for malformed requests; the issue is on the sender's side.
- HTTP 403 return code is used when the WAF Limit (Web Application Firewall) has been violated.
- HTTP 408 return code is used when a timeout has occurred while waiting for a response from the backend server.
- HTTP 429 return code is used when breaking a request rate limit.
- HTTP 418 return code is used when an IP has been auto-banned for continuing to send requests after receiving 429 codes.
- HTTP 5XX return codes are used for internal errors; the issue is on Binance's side.
- If there is an error message "Request occur unknown error.", please retry later.
- HTTP 503 return code is used when:
    - If there is an error message "Unknown error, please check your request or try again later." returned in the response, the API successfully sent the request but not get a response within the timeout period. It is important to NOT treat this as a failure operation; the execution status is UNKNOWN and could have been a success;
    - If there is an error message "Service Unavailable." returned in the response, it means this is a failure API operation and the service might be unavailable at the moment, you need to retry later.
    - If there is an error message "Internal error; unable to process your request. Please try again." returned in the response, it means this is a failure API operation and you can resend your request if you need.
    - If the response contains the error message "Request throttled by system-level protection. Reduce-only/close-position orders are exempt. Please try again." (-1008), This indicates the node has exceeded its maximum concurrency and is temporarily throttled. Close-position, reduce-only, and cancel orders are exempt and will not receive this error.
### HTTP 503 Status: Message Variants & Handling
#### A. “Unknown error, please check your request or try again later.” (Execution status unknown)
- Meaning: Request accepted but no response before timeout; execution may have succeeded.
- Handling:
    - Do not treat as immediate failure; first verify via WebSocket updates or orderId queries to avoid duplicates.
    - During peaks, prefer single orders over batch to reduce uncertainty.
- Rate-limit counting: May or may not count, check header to verify rate limit info
#### B. “Service Unavailable.” (Failure)
- Meaning: Service temporarily unavailable; 100% failure.
- Handling: Retry with exponential backoff (e.g., 200ms → 400ms → 800ms, max 3–5 attempts).
- Rate-limit counting: not counted
#### C. “Request throttled by system-level protection. Reduce-only/close-position orders are exempt. Please try again.” (-1008, Failure)
- Meaning: System overload; 100% failure.
- Handling: Retry with backoff and reduce concurrency;
- Applicable endpoints:
    - POST /fapi/v1/order
    - POST /fapi/v1/batchOrders
    - POST /fapi/v1/order/test
- Rate-limit counting: Not counted (overload protection).
- Exception integrated here: When a request reduces exposure (Reduce-only / Close-position: closePosition = true, or positionSide = BOTH with reduceOnly = true, or LONG+SELL, or SHORT+BUY), it is not affected or prioritized under -1008 to ensure risk reduction.
    - Covered endpoints: POST /fapi/v1/order、POST /fapi/v1/batchOrders (when parameters satisfy the condition)
##### Error Codes and Messages
- Any endpoint can return an ERROR
> The error payload is as follows:

{
  "code": -1121,
  "msg": "Invalid symbol."
}

- Specific error codes and messages defined in Error Codes.