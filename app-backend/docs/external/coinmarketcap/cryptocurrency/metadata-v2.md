https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyInfo

# Metadata v2

GET https://pro-api.coinmarketcap.com/v2/cryptocurrency/info

Returns all static metadata available for one or more cryptocurrencies. This information includes details like logo, description, official website URL, social links, and links to a cryptocurrency's technical documentation.

Cache / Update frequency: Static data is updated only as needed, every 30 seconds.

## Query Parameters
 id	
string
One or more comma-separated CoinMarketCap cryptocurrency IDs. Example: "1,2"

 slug	
string
Alternatively pass a comma-separated list of cryptocurrency slugs. Example: "bitcoin,ethereum"

 symbol	
string
Alternatively pass one or more comma-separated cryptocurrency symbols. Example: "BTC,ETH". At least one "id" or "slug" or "symbol" is required for this request. Please note that starting in the v2 endpoint, due to the fact that a symbol is not unique, if you request by symbol each data response will contain an array of objects containing all of the coins that use each requested symbol. The v1 endpoint will still return a single object, the highest ranked coin using that symbol.

 address	
string
Alternatively pass in a contract address. Example: "0xc40af1e4fecfa05ce6bab79dcd8b373d2e436c4e"

 skip_invalid	
boolean
Default: false
Pass true to relax request validation rules. When requesting records on multiple cryptocurrencies an error is returned if any invalid cryptocurrencies are requested or a cryptocurrency does not have matching records in the requested timeframe. If set to true, invalid lookups will be skipped allowing valid cryptocurrencies to still be returned.

 aux	
string
Default: "urls,logo,description,tags,platform,date_added,notice"
Optionally specify a comma-separated list of supplemental data fields to return. Pass urls,logo,description,tags,platform,date_added,notice,status to include all auxiliary fields.

## Response

{
  "data": {
    "1": {
      "urls": {
        "website": [
          "https://bitcoin.org/"
        ],
        "technical_doc": [
          "https://bitcoin.org/bitcoin.pdf"
        ],
        "twitter": [],
        "reddit": [
          "https://reddit.com/r/bitcoin"
        ],
        "message_board": [
          "https://bitcointalk.org"
        ],
        "announcement": [],
        "chat": [],
        "explorer": [
          "https://blockchain.coinmarketcap.com/chain/bitcoin",
          "https://blockchain.info/",
          "https://live.blockcypher.com/btc/"
        ],
        "source_code": [
          "https://github.com/bitcoin/"
        ]
      },
      "logo": "https://s2.coinmarketcap.com/static/img/coins/64x64/1.png",
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "description": "Bitcoin (BTC) is a consensus network that enables a new payment system and a completely digital currency. Powered by its users, it is a peer to peer payment network that requires no central authority to operate",
      "date_added": "2013-04-28T00:00:00.000Z",
      "date_launched": "2013-04-28T00:00:00.000Z",
      "tags": [
        "mineable"
      ],
      "platform": null,
      "category": "coin"
    },
    "1027": {
      "urls": {
        "website": [
          "https://www.ethereum.org/"
        ],
        "technical_doc": [
          "https://github.com/ethereum/wiki/wiki/White-Paper"
        ],
        "twitter": [
          "https://twitter.com/ethereum"
        ],
        "reddit": [
          "https://reddit.com/r/ethereum"
        ],
        "message_board": [
          "https://forum.ethereum.org/"
        ],
        "announcement": [
          "https://bitcointalk.org/index.php?topic=428589.0"
        ],
        "chat": [
          "https://gitter.im/orgs/ethereum/rooms"
        ],
        "explorer": [
          "https://blockchain.coinmarketcap.com/chain/ethereum",
          "https://etherscan.io/",
          "https://ethplorer.io/"
        ],
        "source_code": [
          "https://github.com/ethereum"
        ]
      },
      "logo": "https://s2.coinmarketcap.com/static/img/coins/64x64/1027.png",
      "id": 1027,
      "name": "Ethereum",
      "symbol": "ETH",
      "slug": "ethereum",
      "description": "Ethereum (ETH) is a smart contract platform that enables developers to build decentralized applications (dapps) conceptualized by Vitalik Buterin in 2013. ETH is the native currency for the Ethereum platform and also works as the transaction fees to miners on the Ethereum network. Ethereum is the pioneer for blockchain based smart contracts. When running on the blockchain a smart contract becomes like a self-operating computer program that automatically executes when specific conditions are met. On the blockchain, smart contracts allow for code to be run exactly as programmed without any possibility of downtime, censorship, fraud or third-party interference. It can facilitate the exchange of money, content, property, shares, or anything of value. The Ethereum network went live on July 30th, 2015 with 72 million Ethereum premined.",
      "notice": null,
      "date_added": "2015-08-07T00:00:00.000Z",
      "date_launched": "2015-08-07T00:00:00.000Z",
      "tags": [
        "mineable"
      ],
      "platform": null,
      "category": "coin",
      "self_reported_circulating_supply": null,
      "self_reported_market_cap": null,
      "self_reported_tags": null,
      "infinite_supply": false
    }
  },
  "status": {
    "timestamp": "2026-01-01T19:49:23.887Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 10,
    "credit_count": 1,
    "notice": ""
  }
}