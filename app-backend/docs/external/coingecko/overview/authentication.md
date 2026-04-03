# Authentication (Pro API)

> Authentication method for CoinGecko Pro API (Paid plan subscribers with Pro-API keys)

<Note>
  ### **Notes**

  * Pro API Key is only available for [CoinGecko API paid plan](https://www.coingecko.com/en/api/pricing) subscribers, the root URL for CoinGecko Pro API must be `https://pro-api.coingecko.com/api/v3/`.
  * You are recommended to store the API key securely in your own backend and use a proxy to insert the key into the request URL.
  * It's highly recommended to use the Headers method when making API requests for better security. Using query string parameters can risk exposing your API key.
</Note>

## CoinGecko API Authentication Method

If this is your first time using the Pro API key, you can supply API Key to the root URL using one of these ways:

1. Header (Recommended): `x-cg-pro-api-key`
2. Query String Parameter: `x_cg_pro_api_key`

| Authentication Method  | Example using [Ping](/reference/ping-server) Endpoint                                         |
| ---------------------- | --------------------------------------------------------------------------------------------- |
| Header (cURL)          | `curl -X GET "https://pro-api.coingecko.com/api/v3/ping" -H "x-cg-pro-api-key: YOUR_API_KEY"` |
| Query String Parameter | `https://pro-api.coingecko.com/api/v3/ping?x_cg_pro_api_key=YOUR_API_KEY`                     |

## 🔥 Accessing Onchain DEX data

You can now use the Pro-API key (exclusive to any paid plan subscriber) to call onchain DEX data powered by [GeckoTerminal](https://www.geckoterminal.com/).

<Note>
  ### **Notes**

  * Authentication method for onchain endpoints is exactly same as other endpoints.
  * When using the CG Pro API to access onchain DEX data, include the `/onchain` endpoint path in the request.
</Note>

| Authentication Method  | Example using [Simple Token Price](/reference/onchain-simple-price) Endpoint                                                                                                  |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Header (cURL)          | `curl -X GET "<https://pro-api.coingecko.com/api/v3/onchain/simple/networks/eth/token_price/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2>" -H "x-cg-pro-api-key: YOUR_API_KEY"` |
| Query String Parameter | `https://pro-api.coingecko.com/api/v3/onchain/simple/networks/eth/token_price/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2?x_cg_pro_api_key=YOUR_API_KEY`                       |

## API Key Usage Credits

* Each request made to any endpoint counts as a single call (1 call = 1 credit).
* Each successful API request (Status 200) will deduct 1 credit from your monthly credit allowance.
* Unsuccessful Requests (Status 4xx, 5xx, etc) will not count towards credit deduction.
* Regardless of the HTTP status code returned (including 4xx and 5xx errors), all API requests will count towards your **minute rate limit**.
* Your monthly credit & rate limit are determined by the paid plan to which you subscribe. For more details, please refer to this [page](https://www.coingecko.com/en/api/pricing).
* To check the API usage, please go to the [developer dashboard](https://www.coingecko.com/en/developers/dashboard) or follow the guide [here](/reference/setting-up-your-api-key#4-api-usage-report)


---

> To find navigation and other pages in this documentation, fetch the llms.txt file at: https://docs.coingecko.com/llms.txt