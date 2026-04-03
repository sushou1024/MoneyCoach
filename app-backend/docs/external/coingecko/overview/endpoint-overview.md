# Endpoint Overview

<Note>
  ### Notes

  In the API reference pages, the plan-specific endpoint access will be marked as below:

  * 💼 — exclusive for [Analyst Plan & above](https://www.coingecko.com/en/api/pricing) subscribers only (excluding Basic plan).
  * 👑 — exclusive for [Enterprise Plan](https://www.coingecko.com/en/api/enterprise) subscribers only.

  Some endpoints may have parameters or data access that are exclusive to different plan subscribers, please refer to the endpoint reference page for details.
</Note>

## CoinGecko Endpoints: Coins

| Endpoint                                                                                               | Description                                                                                                                                                                            |
| ------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [/ping](/reference/ping-server)                                                                        | Check the API server status                                                                                                                                                            |
| 💼 [/key](/reference/api-usage)                                                                        | Check account's API usage                                                                                                                                                              |
| [/simple/price](/reference/simple-price)                                                               | Query the prices of one or more coins by using their unique Coin API IDs                                                                                                               |
| [/simple/token\_price/\{id}](/reference/simple-token-price)                                            | Query the prices of one or more coins by using their unique Coin API IDs                                                                                                               |
| [/simple/supported\_vs\_currencies](/reference/simple-supported-currencies)                            | Query all the supported currencies on CoinGecko                                                                                                                                        |
| [/coins/list](/reference/coins-list)                                                                   | Query all the supported coins on CoinGecko with coins ID, name and symbol                                                                                                              |
| 💼 [/coins/top\_gainers\_losers](/reference/coins-top-gainers-losers)                                  | Query the top 30 coins with largest price gain and loss by a specific time duration                                                                                                    |
| 💼 [/coins/list/new](/reference/coins-list-new)                                                        | Query the latest 200 coins that recently listed on CoinGecko                                                                                                                           |
| [/coins/markets](/reference/coins-markets)                                                             | Query all the supported coins with price, market cap, volume and market related data                                                                                                   |
| [/coins/\{id}](/reference/coins-id)                                                                    | Query all the metadata (image, websites, socials, description, contract address, etc.) from the CoinGecko coin page based on a particular coin ID                                      |
| [/coins/\{id}/tickers](/reference/coins-id-tickers)                                                    | Query the coin tickers on both centralized exchange (CEX) and decentralized exchange (DEX) based on a particular coin ID                                                               |
| [/coins/\{id}/history](/reference/coins-id-history)                                                    | Query the historical data (price, market cap, 24hr volume, ...) at a given date for a coin based on a particular coin ID                                                               |
| [/coins/\{id}/market\_chart](/reference/coins-id-market-chart)                                         | Get the historical chart data of a coin including time in UNIX, price, market cap and 24hr volume based on particular coin ID                                                          |
| [/coins/\{id}/market\_chart/range](/reference/coins-id-market-chart-range)                             | Get the historical chart data of a coin within certain time range in UNIX along with price, market cap and 24hr volume based on particular coin ID                                     |
| [/coins-id-ohlc](/reference/coins-id-ohlc)                                                             | Get the OHLC chart (Open, High, Low, Close) of a coin based on particular coin ID                                                                                                      |
| 💼 [/coins/\{id}/ohlc/range](/reference/coins-id-ohlc-range)                                           | Get the OHLC chart (Open, High, Low, Close) of a coin within a range of timestamp based on particular coin ID                                                                          |
| 👑 [/coins/\{id}/circulating\_supply\_chart](/reference/coins-id-circulating-supply-chart)             | Query historical circulating supply of a coin by number of days away from now based on provided coin ID                                                                                |
| 👑 [/coins/\{id}/circulating\_supply\_chart/range](/reference/coins-id-circulating-supply-chart-range) | Query historical circulating supply of a coin, within a range of timestamp based on the provided coin ID                                                                               |
| 👑 [/coins/\{id}/total\_supply\_chart](/reference/coins-id-total-supply-chart)                         | Query historical total supply of a coin by number of days away from now based on provided coin ID                                                                                      |
| 👑 [/coins/\{id}/total\_supply\_chart/range](/reference/coins-id-total-supply-chart-range)             | Query historical total supply of a coin, within a range of timestamp based on the provided coin ID                                                                                     |
| [/coins/../contract/..](/reference/coins-contract-address)                                             | Query all the metadata (image, websites, socials, description, contract address, etc.) from the CoinGecko coin page based on an asset platform and a particular token contract address |
| [/coins/../contract/../market\_chart](/reference/contract-address-market-chart)                        | Get the historical chart data including time in UNIX, price, market cap and 24hr volume based on asset platform and particular token contract address                                  |
| [/coins/../contract/../market\_chart/range](/reference/contract-address-market-chart-range)            | Get the historical chart data within certain time range in UNIX along with price, market cap and 24hr volume based on asset platform and particular token contract address             |
| [/coins/categories/list](/reference/coins-categories-list)                                             | Query all the coins categories on CoinGecko                                                                                                                                            |
| [/coins/categories](/reference/coins-categories)                                                       | Query all the coins categories with market data (market cap, volume, ...) on CoinGecko                                                                                                 |

## CoinGecko Endpoints: NFT

| Endpoint                                                                               | Description                                                                                                                                                                  |
| -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [/nfts/list](/reference/nfts-list)                                                     | Query all supported NFTs with ID, contract address, name, asset platform ID and symbol on CoinGecko                                                                          |
| [/nfts/..](/reference/nfts-id)                                                         | Query all the NFT data (name, floor price, 24hr volume, ...) based on the NFT collection ID                                                                                  |
| [/nfts/../contract/..](/reference/nfts-contract-address)                               | Query all the NFT data (name, floor price, 24hr volume, ...) based on the NFT collection contract address and respective asset platform                                      |
| 💼 [/nfts/markets](/reference/nfts-markets)                                            | Query all the supported NFT collections with floor price, market cap, volume and market related data on CoinGecko                                                            |
| 💼 [/nfts/../market\_chart](/reference/nfts-id-market-chart)                           | Query historical market data of a NFT collection, including floor price, market cap, and 24hr volume, by number of days away from now                                        |
| 💼 [/nfts/../contract/../market\_chart](/reference/nfts-contract-address-market-chart) | Query historical market data of a NFT collection, including floor price, market cap, and 24hr volume, by number of days away from now based on the provided contract address |
| 💼 [/nfts/../tickers](/reference/nfts-id-tickers)                                      | Query the latest floor price and 24hr volume of a NFT collection, on each NFT marketplace, e.g. OpenSea and LooksRare                                                        |

## CoinGecko Endpoints: Exchanges & Derivatives

| Endpoint                                                                              | Description                                                                                                                   |
| ------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| [/exchanges](/reference/exchanges)                                                    | Query all the supported exchanges with exchanges' data (ID, name, country, ...) that have active trading volumes on CoinGecko |
| [/exchanges/list](/reference/exchanges-list)                                          | Query all the exchanges with ID and name                                                                                      |
| [/exchanges/\{id}](/reference/exchanges-id)                                           | Query exchange's data (name, year established, country, ...), exchange volume in BTC and tickers based on exchange's ID       |
| [/exchanges/\{id}/tickers](/reference/exchanges-id-tickers)                           | Query exchange's tickers based on exchange's ID                                                                               |
| [/exchanges/\{id}/volume\_chart](/reference/exchanges-id-volume-chart)                | Query the historical volume chart data with time in UNIX and trading volume data in BTC based on exchange's ID                |
| 💼 [/exchanges/\{id}/volume\_chart/range](/reference/exchanges-id-volume-chart-range) | Query the historical volume chart data in BTC by specifying date range in UNIX based on exchange's ID                         |
| [/derivatives](/reference/derivatives-tickers)                                        | Query all the tickers from derivatives exchanges on CoinGecko                                                                 |
| [/derivatives/exchanges](/reference/derivatives-exchanges)                            | Query all the derivatives exchanges with related data (ID, name, open interest, ...) on CoinGecko                             |
| [/derivatives/exchanges/\{id}](/reference/derivatives-exchanges-id)                   | Query the derivatives exchange's related data (ID, name, open interest, ...) based on the exchanges' ID                       |
| [/derivatives/exchanges/list](/reference/derivatives-exchanges-list)                  | Query all the derivatives exchanges with ID and name on CoinGecko                                                             |

## CoinGecko Endpoints: Public Treasuries

| Endpoint                                                                                               | Description                                                                                       |
| ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| [/entities/list](/reference/entities-list)                                                             | Query all the supported entities on CoinGecko with entities ID, name, symbol, and country         |
| [/\{entity}/public\_treasury/\{coin\_id}](/reference/companies-public-treasury)                        | Query public companies & governments' cryptocurrency holdings by coin ID                          |
| [/public\_treasury/\{entity\_id}](/reference/public-treasury-entity)                                   | Query public companies & governments' cryptocurrency holdings by entity ID                        |
| [/public\_treasury/\{entity\_id}/.../holding\_chart](/reference/public-treasury-entity-chart)          | Query public companies & governments' cryptocurrency historical holdings by entity ID and coin ID |
| [/public\_treasury/\{entity\_id}/transaction\_history](/reference/public-treasury-transaction-history) | Query public companies & governments' cryptocurrency transaction history by entity ID             |

## CoinGecko Endpoints: General

| Endpoint                                                                | Description                                                                                                        |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| [/asset\_platforms](/reference/asset-platforms-list)                    | Query all the asset platforms (blockchain networks) on CoinGecko                                                   |
| [/token\_lists/\{asset\_platform\_id}/all.json](/reference/token-lists) | Get full list of tokens of a blockchain network (asset platform) that is supported by Ethereum token list standard |
| [/exchange\_rates](/reference/exchange-rates)                           | Query BTC exchange rates with other currencies                                                                     |
| [/search](/reference/search-data)                                       | Search for coins, categories and markets listed on CoinGecko                                                       |
| [/search/trending](/reference/trending-search)                          | Query trending search coins, NFTs and categories on CoinGecko in the last 24 hours                                 |
| [/global](/reference/crypto-global)                                     | Query cryptocurrency global data including active cryptocurrencies, markets, total crypto market cap and etc.      |
| [/global/decentralized\_finance\_defi](/reference/global-defi)          | Query cryptocurrency global decentralized finance (DeFi) data including DeFi market cap, trading volume            |
| 💼 [/global/market\_cap\_chart](/reference/global-market-cap-chart)     | Query historical global market cap and volume data by number of days away from now                                 |

## Onchain DEX Endpoints (GeckoTerminal)

| Endpoint                                                                                         | Description                                                                                                                                                              |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [/onchain/simple/networks/../token\_price/..](/reference/onchain-simple-price)                   | Get token price based on the provided token contract address on a network                                                                                                |
| [/onchain/networks](/reference/networks-list)                                                    | Query all the supported networks on GeckoTerminal                                                                                                                        |
| [/onchain/networks/../dexes](/reference/dexes-list)                                              | Query all the supported decentralized exchanges (DEXs) based on the provided network on GeckoTerminal                                                                    |
| [/onchain/networks/../pools/..](/reference/pool-address)                                         | Query the specific pool based on the provided network and pool address                                                                                                   |
| [/onchain/networks/../pools/multi/..](/reference/pools-addresses)                                | Query multiple pools based on the provided network and pool address                                                                                                      |
| [/onchain/networks/trending\_pools](/reference/trending-pools-list)                              | Query all the trending pools across all networks on GeckoTerminal                                                                                                        |
| [/onchain/networks/../trending\_pools](/reference/trending-pools-network)                        | Query the trending pools based on the provided network                                                                                                                   |
| [/onchain/networks/../pools](/reference/top-pools-network)                                       | Query all the top pools based on the provided network                                                                                                                    |
| [/onchain/networks/../dexes/../pools](/reference/top-pools-dex)                                  | Query all the top pools based on the provided network and decentralized exchange (DEX)                                                                                   |
| [/onchain/networks/new\_pools](/reference/latest-pools-list)                                     | Query all the latest pools across all networks on GeckoTerminal                                                                                                          |
| [/onchain/networks/../new\_pools](/reference/latest-pools-network)                               | Query all the latest pools based on provided network                                                                                                                     |
| 🔥 💼 [/onchain/pools/megafilter](/reference/pools-megafilter)                                   | Query pools based on various filters across all networks on GeckoTerminal                                                                                                |
| [/onchain/search/pools](/reference/search-pools)                                                 | Search for pools on a network                                                                                                                                            |
| 💼 [/onchain/pools/trending\_search](/reference/trending-search-pools)                           | Query all the trending search pools across all networks on GeckoTerminal                                                                                                 |
| [/onchain/networks/../tokens/../pools](/reference/top-pools-contract-address)                    | Query top pools based on the provided token contract address on a network                                                                                                |
| [/onchain/networks/../tokens/..](/reference/token-data-contract-address)                         | Query specific token data based on the provided token contract address on a network                                                                                      |
| [/onchain/networks/../tokens/multi/..](/reference/tokens-data-contract-addresses)                | Query multiple tokens data based on the provided token contract addresses on a network                                                                                   |
| [/onchain/networks/../tokens/../info](/reference/token-info-contract-address)                    | Query token metadata (name, symbol, CoinGecko ID, image, socials, websites, description, etc.) based on a provided token contract address on a network                   |
| [/onchain/networks/../pools/../info](/reference/pool-token-info-contract-address)                | Query pool metadata (base and quote token details, image, socials, websites, description, contract address, etc.) based on a provided pool contract address on a network |
| [/onchain/tokens/info\_recently\_updated](/reference/tokens-info-recent-updated)                 | Query 100 most recently updated tokens info across all networks on GeckoTerminal                                                                                         |
| 💼 [/onchain/networks/../tokens/../top\_traders](/reference/top-token-traders-token-address)     | Query top token traders based on the provided token contract address on a network                                                                                        |
| 💼 [/onchain/networks/../tokens/../top\_holders](/reference/top-token-holders-token-address)     | Query top token holders based on the provided token contract address on a network                                                                                        |
| 💼 [/onchain/networks/../tokens/../holders\_chart](/reference/token-holders-chart-token-address) | Get the historical token holders chart based on the provided token contract address on a network                                                                         |
| [/onchain/networks/../pools/../ohlcv/..](/reference/pool-ohlcv-contract-address)                 | Get the OHLCV chart (Open, High, Low, Close, Volume) of a pool based on the provided pool address on a network                                                           |
| 💼 [/onchain/networks/../tokens/../ohlcv/..](/reference/token-ohlcv-token-address)               | Get the OHLCV chart (Open, High, Low, Close, Volume) of a token based on the provided token address on a network                                                         |
| [/onchain/networks/../pools/../trades](/reference/pool-trades-contract-address)                  | Query the last 300 trades in the past 24 hours based on the provided pool address                                                                                        |
| 💼 [/onchain/networks/../tokens/../trades](/reference/token-trades-contract-address)             | Query the last 300 trades in the past 24 hours across all pools, based on the provided token contract address on a network                                               |
| 💼 [/onchain/categories](/reference/categories-list)                                             | Query all the supported categories on GeckoTerminal                                                                                                                      |
| 💼 [/onchain/categories/../pools](/reference/pools-category)                                     | Query all the pools based on the provided category ID                                                                                                                    |

⚡️ Need Real-time Data Streams? Try [WebSocket API](https://docs.coingecko.com/websocket)

<a href="/websocket">
  <Frame>
    <img src="https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=2c88f667113256b6285720c468fb53a1" noZoom data-og-width="2400" width="2400" data-og-height="470" height="470" data-path="images/wss-banner-2.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=280&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=d2eafb93fcd670d5df221d617fd6f6a7 280w, https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=560&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=24f635622a42c0ae03695cc940112699 560w, https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=840&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=82ef1c05b6f45d6d8ec0bcef0f19d49a 840w, https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=1100&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=b119e8746bb1a78b759e6d94d96b7c8b 1100w, https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=1650&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=95797e7366c7f280e3e4b570b6db2b49 1650w, https://mintcdn.com/coingecko/VlaOc2UnIs8mj72v/images/wss-banner-2.png?w=2500&fit=max&auto=format&n=VlaOc2UnIs8mj72v&q=85&s=2f120e8a31b5793213494d4ae2d46fb3 2500w" />
  </Frame>
</a>

With WebSocket, you can now stream ultra-low latency, real-time prices, trades, and OHLCV chart data. <br />
Subscribe to our [paid API plan](https://www.coingecko.com/en/api/pricing) (Analyst plan & above) to access WebSocket and REST API data delivery methods.


---

> To find navigation and other pages in this documentation, fetch the llms.txt file at: https://docs.coingecko.com/llms.txt