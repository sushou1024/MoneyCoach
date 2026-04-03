# Getting Started
## API Access Key & Authentication
For every API request you make, you will need to make sure to be authenticated with the API by passing your API access key to the API's access_key parameter. You can find an example below.

### Example API Request

Sign Up to Run API Request
https://api.marketstack.com/v2/eod
    ? access_key = YOUR_ACCESS_KEY
    & symbols = AAPL
Important: Please make sure not to expose your API access key publicly. If you believe your API access key may be compromised, you can always reset in your account dashboard.

## 256-bit HTTPS Encryption 
If you're subscribed to either the free or any paid plans, you will be able to access the Marketstack API using industry-standard HTTPS. To do that, simply use the https protocol when making API requests.

## API Error Codes
API errors consist of error code and message response objects. If an error occurs, the marketstack will return HTTP status codes, such as 404 for "not found" errors. If your API request succeeds, a status code 200 will be sent.

For validation errors, the marketstack API will also provide a context response object returning additional information about the error that occurred in the form of one or multiple sub-objects, each equipped with the name of the affected parameter as well as key and message objects. You can find an example error below.

### Example Error
{
   "error": {
      "code": "validation_error",
      "message": "Request failed with validation error",
      "context": {
         "symbols": [
            {
               "key": "missing_symbols",
               "message": "You did not specify any symbols."
            }
         ]
      }
   }
}
 

### Common API Errors

| Code | Type                       | Description                                                              |
|------|----------------------------|--------------------------------------------------------------------------|
| 401  | unauthorized               | Authentication failed. Please verify your access key or account status.  |
| 403  | function_access_restricted | This API endpoint is not available under your current subscription plan. |
| 404  | invalid_api_function       | The specified API endpoint does not exist.                               |
| 404  | 404_not_found              | The requested resource could not be found.                               |
| 429  | too_many_requests          | The account has exceeded the allowed monthly request quota.              |
| 429  | rate_limit_reached         | The given user account has reached the rate limit.                       |
| 500  | internal_error             | An internal server error has occurred.                                   |
| 406  | data_not_available         | The requested data is currently unavailable.                             |

Note: The API is limited to 5 requests per second.

 