https://coinmarketcap.com/api/documentation/v1/#section/Standards-and-Conventions

# Standards and Conventions

Each HTTP request must contain the header Accept: application/json. You should also send an Accept-Encoding: deflate, gzip header to receive data fast and efficiently.

## Endpoint Response Payload Format

All endpoints return data in JSON format with the results of your query under data if the call is successful.

A Status object is always included for both successful calls and failures when possible. The Status object always includes the current time on the server when the call was executed as timestamp, the number of API call credits this call utilized as credit_count, and the number of milliseconds it took to process the request as elapsed. Any details about errors encountered can be found under the error_code and error_message. See Errors and Rate Limits for details on errors.

```
{
  "data" : {
    ...
  },
  "status": {
    "timestamp": "2018-06-06T07:52:27.273Z",
    "error_code": 400,
    "error_message": "Invalid value for \"id\"",
    "elapsed": 0,
    "credit_count": 0
  }
}
```

## Cryptocurrency, Exchange, and Fiat currency identifiers

Cryptocurrencies may be identified in endpoints using either the cryptocurrency's unique CoinMarketCap ID as id (eg. id=1 for Bitcoin) or the cryptocurrency's symbol (eg. symbol=BTC for Bitcoin). For a current list of supported cryptocurrencies use our /cryptocurrency/map call.
Exchanges may be identified in endpoints using either the exchange's unique CoinMarketCap ID as id (eg. id=270 for Binance) or the exchange's web slug (eg. slug=binance for Binance). For a current list of supported exchanges use our /exchange/map call.
All fiat currency options use the standard ISO 8601 currency code (eg. USD for the US Dollar). For a current list of supported fiat currencies use our /fiat/map endpoint. Unless otherwise stated, endpoints with fiat currency options like our convert parameter support these 93 major currency codes:

| Currency                                 | Currency Code | CoinMarketCap ID |
|------------------------------------------|---------------|------------------|
| United States Dollar ($)                 | USD           | 2781             |
| Albanian Lek (L)                         | ALL           | 3526             |
| Algerian Dinar (د.ج)                     | DZD           | 3537             |
| Argentine Peso ($)                       | ARS           | 2821             |
| Armenian Dram (֏)                        | AMD           | 3527             |
| Australian Dollar ($)                    | AUD           | 2782             |
| Azerbaijani Manat (₼)                    | AZN           | 3528             |
| Bahraini Dinar (.د.ب)                    | BHD           | 3531             |
| Bangladeshi Taka (৳)                     | BDT           | 3530             |
| Belarusian Ruble (Br)                    | BYN           | 3533             |
| Bermudan Dollar ($)                      | BMD           | 3532             |
| Bolivian Boliviano (Bs.)                 | BOB           | 2832             |
| Bosnia-Herzegovina Convertible Mark (KM) | BAM           | 3529             |
| Brazilian Real (R$)                      | BRL           | 2783             |
| Bulgarian Lev (лв)                       | BGN           | 2814             |
| Cambodian Riel (៛)                       | KHR           | 3549             |
| Canadian Dollar ($)                      | CAD           | 2784             |
| Chilean Peso ($)                         | CLP           | 2786             |
| Chinese Yuan (¥)                         | CNY           | 2787             |
| Colombian Peso ($)                       | COP           | 2820             |
| Costa Rican Colón (₡)                    | CRC           | 3534             |
| Croatian Kuna (kn)                       | HRK           | 2815             |
| Cuban Peso ($)                           | CUP           | 3535             |
| Czech Koruna (Kč)                        | CZK           | 2788             |
| Danish Krone (kr)                        | DKK           | 2789             |
| Dominican Peso ($)                       | DOP           | 3536             |
| Egyptian Pound (£)                       | EGP           | 3538             |
| Euro (€)                                 | EUR           | 2790             |
| Georgian Lari (₾)                        | GEL           | 3539             |
| Ghanaian Cedi (₵)                        | GHS           | 3540             |
| Guatemalan Quetzal (Q)                   | GTQ           | 3541             |
| Honduran Lempira (L)                     | HNL           | 3542             |
| Hong Kong Dollar ($)                     | HKD           | 2792             |
| Hungarian Forint (Ft)                    | HUF           | 2793             |
| Icelandic Króna (kr)                     | ISK           | 2818             |
| Indian Rupee (₹)                         | INR           | 2796             |
| Indonesian Rupiah (Rp)                   | IDR           | 2794             |
| Iranian Rial (﷼)                         | IRR           | 3544             |
| Iraqi Dinar (ع.د)                        | IQD           | 3543             |
| Israeli New Shekel (₪)                   | ILS           | 2795             |
| Jamaican Dollar ($)                      | JMD           | 3545             |
| Japanese Yen (¥)                         | JPY           | 2797             |
| Jordanian Dinar (د.ا)                    | JOD           | 3546             |
| Kazakhstani Tenge (₸)                    | KZT           | 3551             |
| Kenyan Shilling (Sh)                     | KES           | 3547             |
| Kuwaiti Dinar (د.ك)                      | KWD           | 3550             |
| Kyrgystani Som (с)                       | KGS           | 3548             |
| Lebanese Pound (ل.ل)                     | LBP           | 3552             |
| Macedonian Denar (ден)                   | MKD           | 3556             |
| Malaysian Ringgit (RM)                   | MYR           | 2800             |
| Mauritian Rupee (₨)                      | MUR           | 2816             |
| Mexican Peso ($)                         | MXN           | 2799             |
| Moldovan Leu (L)                         | MDL           | 3555             |
| Mongolian Tugrik (₮)                     | MNT           | 3558             |
| Moroccan Dirham (د.م.)                   | MAD           | 3554             |
| Myanma Kyat (Ks)                         | MMK           | 3557             |
| Namibian Dollar ($)                      | NAD           | 3559             |
| Nepalese Rupee (₨)                       | NPR           | 3561             |
| New Taiwan Dollar (NT$)                  | TWD           | 2811             |
| New Zealand Dollar ($)                   | NZD           | 2802             |
| Nicaraguan Córdoba (C$)                  | NIO           | 3560             |
| Nigerian Naira (₦)                       | NGN           | 2819             |
| Norwegian Krone (kr)                     | NOK           | 2801             |
| Omani Rial (ر.ع.)                        | OMR           | 3562             |
| Pakistani Rupee (₨)                      | PKR           | 2804             |
| Panamanian Balboa (B/.)                  | PAB           | 3563             |
| Peruvian Sol (S/.)                       | PEN           | 2822             |
| Philippine Peso (₱)                      | PHP           | 2803             |
| Polish Złoty (zł)                        | PLN           | 2805             |
| Pound Sterling (£)                       | GBP           | 2791             |
| Qatari Rial (ر.ق)                        | QAR           | 3564             |
| Romanian Leu (lei)                       | RON           | 2817             |
| Russian Ruble (₽)                        | RUB           | 2806             |
| Saudi Riyal (ر.س)                        | SAR           | 3566             |
| Serbian Dinar (дин.)                     | RSD           | 3565             |
| Singapore Dollar (S$)                    | SGD           | 2808             |
| South African Rand (R)                   | ZAR           | 2812             |
| South Korean Won (₩)                     | KRW           | 2798             |
| South Sudanese Pound (£)                 | SSP           | 3567             |
| Sovereign Bolivar (Bs.)                  | VES           | 3573             |
| Sri Lankan Rupee (Rs)                    | LKR           | 3553             |
| Swedish Krona ( kr)                      | SEK           | 2807             |
| Swiss Franc (Fr)                         | CHF           | 2785             |
| Thai Baht (฿)                            | THB           | 2809             |
| Trinidad and Tobago Dollar ($)           | TTD           | 3569             |
| Tunisian Dinar (د.ت)                     | TND           | 3568             |
| Turkish Lira (₺)                         | TRY           | 2810             |
| Ugandan Shilling (Sh)                    | UGX           | 3570             |
| Ukrainian Hryvnia (₴)                    | UAH           | 2824             |
| United Arab Emirates Dirham (د.إ)        | AED           | 2813             |
| Uruguayan Peso ($)                       | UYU           | 3571             |
| Uzbekistan Som (so'm)                    | UZS           | 3572             |
| Vietnamese Dong (₫)                      | VND           | 2823             |

Along with these four precious metals:

| Precious Metal    | Currency Code | CoinMarketCap ID |
|-------------------|---------------|------------------|
| Gold Troy Ounce   | XAU           | 3575             |
| Silver Troy Ounce | XAG           | 3574             |
| Platinum Ounce    | XPT           | 3577             |
| Palladium Ounce   | XPD           | 3576             |

Warning: Using CoinMarketCap IDs is always recommended as not all cryptocurrency symbols are unique. They can also change with a cryptocurrency rebrand. If a symbol is used the API will always default to the cryptocurrency with the highest market cap if there are multiple matches. Our convert parameter also defaults to fiat if a cryptocurrency symbol also matches a supported fiat currency. You may use the convenient /map endpoints to quickly find the corresponding CoinMarketCap ID for a cryptocurrency or exchange.

## Bundling API Calls
- Many endpoints support ID and crypto/fiat currency conversion bundling. This means you can pass multiple comma-separated values to an endpoint to query or convert several items at once. Check the id, symbol, slug, and convert query parameter descriptions in the endpoint documentation to see if this is supported for an endpoint.
- Endpoints that support bundling return data as an object map instead of an array. Each key-value pair will use the identifier you passed in as the key.
For example, if you passed symbol=BTC,ETH to /v1/cryptocurrency/quotes/latest you would receive:

```
"data" : {
    "BTC" : {
      ...
    },
    "ETH" : {
      ...
    }
}
```

Or if you passed id=1,1027 you would receive:

```
"data" : {
    "1" : {
      ...
    },
    "1027" : {
      ...
    }
}
```

Price conversions that are returned inside endpoint responses behave in the same fashion. These are enclosed in a quote object.

## Date and Time Formats
- All endpoints that require date/time parameters allow timestamps to be passed in either ISO 8601 format (eg. 2018-06-06T01:46:40Z) or in Unix time (eg. 1528249600). Timestamps that are passed in ISO 8601 format support basic and extended notations; if a timezone is not included, UTC will be the default.
- All timestamps returned in JSON payloads are returned in UTC time using human-readable ISO 8601 format which follows this pattern: yyyy-mm-ddThh:mm:ss.mmmZ. The final .mmm designates milliseconds. Per the ISO 8601 spec the final Z is a constant that represents UTC time.
- Data is collected, recorded, and reported in UTC time unless otherwise specified.