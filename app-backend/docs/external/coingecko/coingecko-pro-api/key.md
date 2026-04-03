# 💼 API Usage

> This endpoint allows you to **monitor your account's API usage, including rate limits, monthly total credits, remaining credits, and more**

<Note>
  ### Note

  For a more comprehensive overview of your API usage, please log in to [https://www.coingecko.com/en/developers/dashboard](https://www.coingecko.com/en/developers/dashboard).
</Note>


## OpenAPI

````yaml reference/api-reference/coingecko-pro.json get /key
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
  /key:
    get:
      tags:
        - Key
      summary: 💼 API Usage
      description: >-
        This endpoint allows you to **monitor your account's API usage,
        including rate limits, monthly total credits, remaining credits, and
        more**
      operationId: api-usage
      responses:
        '200':
          description: API Usage
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Key'
components:
  schemas:
    Key:
      type: object
      properties:
        plan:
          type: string
        rate_limit_request_per_minute:
          type: number
        monthly_call_credit:
          type: number
        current_total_monthly_calls:
          type: number
        current_remaining_monthly_calls:
          type: number
      example:
        plan: Other
        rate_limit_request_per_minute: 1000
        monthly_call_credit: 1000000
        current_total_monthly_calls: 104
        current_remaining_monthly_calls: 999896
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