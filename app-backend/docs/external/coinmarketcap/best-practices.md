# Best Practices
This section contains a few recommendations on how to efficiently utilize the CoinMarketCap API for your enterprise application, particularly if you already have a large base of users for your application.

## Use CoinMarketCap ID Instead of Cryptocurrency Symbol
Utilizing common cryptocurrency symbols to reference cryptocurrencies on the API is easy and convenient but brittle. You should know that many cryptocurrencies have the same symbol, for example, there are currently three cryptocurrencies that commonly refer to themselves by the symbol HOT. Cryptocurrency symbols also often change with cryptocurrency rebrands. When fetching cryptocurrency by a symbol that matches several active cryptocurrencies we return the one with the highest market cap at the time of the query. To ensure you always target the cryptocurrency you expect, use our permanent CoinMarketCap IDs. These IDs are used reliably by numerous mission critical platforms and never change.

We make fetching a map of all active cryptocurrencies' CoinMarketCap IDs very easy. Just call our /cryptocurrency/map endpoint to receive a list of all active currencies mapped to the unique id property. This map also includes other typical identifiying properties like name, symbol and platform token_address that can be cross referenced. In cryptocurrency calls you would then send, for example id=1027, instead of symbol=ETH. It's strongly recommended that any production code utilize these IDs for cryptocurrencies, exchanges, and markets to future-proof your code.

## Use the Right Endpoints for the Job
You may have noticed that /cryptocurrency/listings/latest and /cryptocurrency/quotes/latest return the same crypto data but in different formats. This is because the former is for requesting paginated and ordered lists of all cryptocurrencies while the latter is for selectively requesting only the specific cryptocurrencies you require. Many endpoints follow this pattern, allow the design of these endpoints to work for you!

## Implement a Caching Strategy If Needed
There are standard legal data safeguards built into the Commercial User Terms that application developers should keep in mind. These Terms help prevent unauthorized scraping and redistributing of CMC data but are intentionally worded to allow legitimate local caching of market data to support the operation of your application. If your application has a significant user base and you are concerned with staying within the call credit and API throttling limits of your subscription plan consider implementing a data caching strategy.

For example instead of making a /cryptocurrency/quotes/latest call every time one of your application's users needs to fetch market rates for specific cryptocurrencies, you could pre-fetch and cache the latest market data for every cryptocurrency in your application's local database every 60 seconds. This would only require 1 API call, /cryptocurrency/listings/latest?limit=5000, every 60 seconds. Then, anytime one of your application's users need to load a custom list of cryptocurrencies you could simply pull this latest market data from your local cache without the overhead of additional calls. This kind of optimization is practical for customers with large, demanding user bases.

## Code Defensively to Ensure a Robust REST API Integration
Whenever implementing any high availability REST API service for mission critical operations it's recommended to code defensively. Since the API is versioned, any breaking request or response format change would only be introduced through new versions of each endpoint, however existing endpoints may still introduce new convenience properties over time.

We suggest these best practices:

- You should parse the API response JSON as JSON and not through a regular expression or other means to avoid brittle parsing logic.
- Your parsing code should explicitly parse only the response properties you require to guarantee new fields that may be returned in the future are ignored.
- You should add robust field validation to your response parsing logic. You can wrap complex field parsing, like dates, in try/catch statements to minimize the impact of unexpected parsing issues (like the unlikely return of a null value).
- Implement a "Retry with exponential backoff" coding pattern for your REST API call logic. This means if your HTTP request happens to get rate limited (HTTP 429) or encounters an unexpected server-side condition (HTTP 5xx) your code would automatically recover and try again using an intelligent recovery scheme. You may use one of the many libraries available; for example, this one for Node or this one for Python.