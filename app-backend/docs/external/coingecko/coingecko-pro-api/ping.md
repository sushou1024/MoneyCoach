# Check API server status

> This endpoint allows you to **check the API server status**

<Note>
  ### Note

  * You can also go to [status.coingecko.com](https://status.coingecko.com/) to check the API server status and further maintenance notices.
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /ping
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
  /ping:
    get:
      tags:
        - Ping
      summary: Check API server status
      description: This endpoint allows you to **check the API server status**
      operationId: ping-server
      responses:
        '200':
          description: Status OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Ping'
components:
  schemas:
    Ping:
      type: object
      properties:
        gecko_says:
          type: string
      example:
        gecko_says: (V3) To the Moon!
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