https://coinmarketcap.com/api/documentation/v1/#operation/getV3FearandgreedHistorical

# CMC Crypto Fear and Greed Historical

GET https://pro-api.coinmarketcap.com/v3/fear-and-greed/historical

Returns a paginated list of all CMC Crypto Fear and Greed values at 12am UTC time.

Cache / Update frequency: Every 15 seconds.

Query Parameters
 start	
integer >= 1
Default: 1
Optionally offset the start (1-based index) of the paginated list of items to return.

 limit	
integer [ 1 .. 500 ]
Default: 50
Optionally specify the number of results to return. Use this parameter and the "start" parameter to determine your own pagination size.

## Response

{
  "data": [
    {
      "timestamp": "1726617600",
      "value": 38,
      "value_classification": "Fear"
    },
    {
      "timestamp": "1726531200",
      "value": 34,
      "value_classification": "Fear"
    },
    {
      "timestamp": "1726444800",
      "value": 36,
      "value_classification": "Fear"
    },
    {
      "timestamp": "1726358400",
      "value": 38,
      "value_classification": "Fear"
    },
    {
      "timestamp": "1726272000",
      "value": 38,
      "value_classification": "Fear"
    }
  ],
  "status": {
    "timestamp": "2026-01-01T19:49:23.887Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 10,
    "credit_count": 1,
    "notice": ""
  }
}