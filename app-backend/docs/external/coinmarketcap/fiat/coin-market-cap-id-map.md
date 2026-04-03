https://coinmarketcap.com/api/documentation/v1/#operation/getV1FiatMap

# CoinMarketCap ID Map

GET https://pro-api.coinmarketcap.com/v1/fiat/map

Returns a mapping of all supported fiat currencies to unique CoinMarketCap ids. Per our Best Practices we recommend utilizing CMC ID instead of currency symbols to securely identify assets with our other endpoints and in your own application logic.

Cache / Update frequency: Mapping data is updated only as needed, every 30 seconds.

## Query Parameters
 start	
integer >= 1
Default: 1
Optionally offset the start (1-based index) of the paginated list of items to return.

 limit	
integer [ 1 .. 5000 ]
Optionally specify the number of results to return. Use this parameter and the "start" parameter to determine your own pagination size.

 sort	
string
Default: "id"
Valid values: "name""id"
What field to sort the list by.

 include_metals	
boolean
Default: false
Pass true to include precious metals.

## Response

{
  "data": [
    {
      "id": 2781,
      "name": "United States Dollar",
      "sign": "$",
      "symbol": "USD"
    },
    {
      "id": 2787,
      "name": "Chinese Yuan",
      "sign": "¥",
      "symbol": "CNY"
    },
    {
      "id": 2781,
      "name": "South Korean Won",
      "sign": "₩",
      "symbol": "KRW"
    }
  ],
  "status": {
    "timestamp": "2020-01-07T22:51:28.209Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 3,
    "credit_count": 1
  }
}