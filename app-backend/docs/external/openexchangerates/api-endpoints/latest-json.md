# /latest.json

GET https://openexchangerates.org/api/latest.json

Get the latest exchange rates available from the Open Exchange Rates API.

The most simple route in our API, latest.json provides a standard response object containing all the conversion rates for all of the currently available symbols/currencies, labeled by their international-standard 3-letter ISO currency codes.

The latest rates will always be the most up-to-date data available on your plan.

The base property provides the 3-letter currency code to which all the delivered exchange rates are relative. This base currency is also given in the rates object by default (e.g. "USD": 1).

The rates property is an object (hash/dictionary/associative array) containing all the conversion or exchange rates for all of the available (or requested) currencies, labeled by their international-standard 3-letter currency codes. All the values are relative to 1 unit of the requested base currency.

The timestamp property indicates the time (UNIX) that the rates were published. (If you’re using the timestamp value in JavaScript, remember to multiply it by 1000, because JavaScript uses time in milliseconds instead of seconds.)

> Additional Parameters
Choosing specific symbols and fetching extra rates with show_alternative are available for all plans, including free. Changing the base currency is available for all clients of paid plans.

## Query Params

| Parameter name   | Description                                                                      | Type    | Default value | Required or Optional |
|------------------|----------------------------------------------------------------------------------|---------|---------------|----------------------|
| app_id           | Your unique App ID                                                               | string  |               | Required             |
| base             | Change base currency (3-letter code, default: USD)                               | string  | USD           | Optional             |
| symbols          | Limit results to specific currencies (comma-separated list of 3-letter codes)    | string  |               | Optional             |
| prettyprint      | Set to false to reduce response size (removes whitespace)                        | boolean | false         | Optional             |
| show_alternative | Extend returned values with alternative, black market and digital currency rates | boolean | false         | Optional             |

### Sample Request

```
curl --request GET \
     --url 'https://openexchangerates.org/api/latest.json?app_id=Required&base=Optional&symbols=Optional&prettyprint=false&show_alternative=false' \
     --header 'accept: application/json'
```