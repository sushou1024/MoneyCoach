# API Introduction
Open Exchange Rates provides a simple, lightweight and portable JSON API with live and historical foreign exchange (forex) rates, via a simple and easy-to-integrate API, in JSON format. Data are tracked and blended algorithmically from multiple reliable sources, ensuring fair and unbiased consistency.

Exchange rates published through the Open Exchange Rates API are collected from multiple reliable providers, blended together and served up in JSON format for everybody to use. There are no complex queries, confusing authentication methods or long-term contracts.

End-of-day rates are available historically for all days going back to 1st January, 1999.

## Common Use Cases
Data from the Open Exchange Rates API are suitable for use in every framework, language and application, and have been successfully integrated in:

Shopping carts from WooCommerce to Shopify, and thousands of individual web stores
Overseas campaigns from the smallest startups to Fortune 500 heavyweights
Accounting departments for multinational brands and shipping/logistics firms
Open source projects and charities
Enterprise-level analytics software
Hundreds of smartphone, tablet and desktop apps
School and university research projects across the world
Our clients range from freelancers and the smallest one-man development shops, to international sports networks and post-IPO startups.

## Connecting To The API
We serve our data in JSON format via a simple URL-based interface over HTTPS, which enables you to use the rates in whichever way you require.

This is the high-level introduction – for more in-depth guides, please see the relevant Documentation sections.

### Connection Types

Any language or software that can make HTTP requests or fetch web addresses can access our API (for example, you can visit any of the API routes in your browser to verify they’re working as expected).

For your integration, you can use whichever library you require. This will vary depending on your development environment. There are guides and a wide range of open source integrations available, also covered in our documentation.

URLs (routes) are requested once over HTTPS, and deliver all their data in one go, just like a normal web request.

We do not currently support websockets, webhooks or any other keep-alive or push-notification style connections – in other words, when you want fresh data, you simply request it from our server. We're considering these methods for a future version of our API, so please email us if interested.

### URL Format

The API base path is https://openexchangerates.org/api/.

API routes/endpoints are then appended to this base path, like so:

HTTP
```
https://openexchangerates.org/api/
                                  latest.json
                                  currencies.json
                                  historical/2013-02-16.json
```

Query parameters (such as your App ID, requested base currency, or JSONP callback) are appended as GET request parameters, for example:

HTTP
```
https://openexchangerates.org/api/latest.json
                                             ?app_id=YOUR_APP_ID
                                             &base=GBP
                                             &callback=someCallbackFunction
```

If your request is valid and permitted, you will receive a JSON-formatted response to work with. If something is wrong with the request, you will receive an error message.

## API Response Formats
Responses are delivered over HTTPS as plain-text JSON (JavaScript Object Notation) format, ready to be used however your integration requires.

This format doesn't limit how and where you can use the data in any way: JSON is simply a fast, simple and lightweight delivery mechanism, which is supported in every major language and framework.

We designed these responses to be simple to integrate into a variety of apps and software. If needed, you can also programmatically convert JSON data to CSV/spreadsheet format, or any other format.

There are several main response styles/formats: latest/historical rates, currencies list, time-series and currency conversion. These are detailed individually on each relevant documentation page, and you can see an example (for latest.json) below.

Here's an example basic API request for all the latest rates, relative to USD (default):

HTTP

https://openexchangerates.org/api/latest.json?app_id=YOUR_APP_ID
When requesting this URL (assuming your App ID is valid) you will receive a JSON object containing a UNIX timestamp (UTC seconds), base currency (3-letter ISO code), and a rates object with symbol:value pairs, relative to the requested base currency:

JSON - latest.json
```
{
    disclaimer: "https://openexchangerates.org/terms/",
    license: "https://openexchangerates.org/license/",
    timestamp: 1449877801,
    base: "USD",
    rates: {
        AED: 3.672538,
        AFN: 66.809999,
        ALL: 125.716501,
        AMD: 484.902502,
        ANG: 1.788575,
        AOA: 135.295998,
        ARS: 9.750101,
        AUD: 1.390866,
        /* ... */
    }
}
```

The response format is the same for Historical Data (historical/YYYY-MM-DD.json) requests.

Other API routes – i.e. currencies.json, time-series.json and convert/ – have a different request and response format. Please see their relevant pages for details and examples.