# Coins List (ID Map)

> This endpoint allows you to **query all the supported coins on CoinGecko with coins ID, name and symbol**

<Tip>
  ### Tips

  * You may use this endpoint to query the list of coins with coin id for other endpoints that contain params like `id` or `ids` (coin ID).
  * By default, this endpoint returns full list of active coins that are currently listed on CoinGecko.com , you can also flag `status=inactive` to retrieve coins that are no longer available on CoinGecko.com . The inactive coin IDs can also be used with [selected historical data](/changelog#april-2024) endpoints.
</Tip>

<Note>
  ### Note

  * There is no pagination required for this endpoint.
  * Cache/Update Frequency: every 5 minutes for Pro API (Analyst, Lite, Pro, Enterprise).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /coins/list
openapi: 3.0.0
info:
  title: CoinGecko Pro API
  version: 3.0.0
servers:
  - url: https://pro-api.coingecko.com/api/v3
security:
  - apiKeyAuth: []
  - apiKeyQueryParam: []
paths:
  /coins/list:
    get:
      tags:
        - Coins
      summary: Coins List (ID Map)
      description: >-
        This endpoint allows you to **query all the supported coins on CoinGecko
        with coins ID, name and symbol**
      operationId: coins-list
      parameters:
        - name: include_platform
          in: query
          description: 'include platform and token''s contract addresses, default: false'
          required: false
          schema:
            type: boolean
        - name: status
          in: query
          description: 'filter by status of coins, default: active'
          required: false
          schema:
            type: string
            enum:
              - active
              - inactive
      responses:
        '200':
          description: List all coins with ID, name, and symbol
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CoinsList'
components:
  schemas:
    CoinsList:
      type: array
      items:
        type: object
        properties:
          id:
            type: string
            description: coin ID
          symbol:
            type: string
            description: coin symbol
          name:
            type: string
            description: coin name
          platforms:
            type: object
            description: coin asset platform and contract address
            additionalProperties:
              type: string
      example:
        - id: 0chain
          symbol: zcn
          name: Zus
          platforms:
            ethereum: '0xb9ef770b6a5e12e45983c5d80545258aa38f3b78'
            polygon-pos: '0x8bb30e0e67b11b978a5040144c410e1ccddcba30'
        - id: 01coin
          symbol: zoc
          name: 01coin
          platforms: {}
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: x-cg-pro-api-key
    apiKeyQueryParam:
      type: apiKey
      in: query
      name: x_cg_pro_api_key

````

---

> To find navigation and other pages in this documentation, fetch the llms.txt file at: https://docs.coingecko.com/llms.txt