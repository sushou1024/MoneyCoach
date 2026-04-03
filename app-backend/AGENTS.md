This directory contains the backend code for a cross-platform mobile app. Use Golang.

Rules in development:
1. Use the skill $effective-go when writing Golang code.
2. Use Postgress & GORM if you need a database. Don't write raw sql commands. Use Redis if necessary.
3. Use Resend to send Emails; Use CoinGecko to get token information (metadata, price); Use CMC for the Fear & Greed index.
4. When writing code that calls LLMs, always put the system prompt under the directory `./system-prompts`, one file for one system prompt, and use `//go:embed` to load the system prompt string from the file.
5. Use the skill $aws-ecs-postgres-cicd to implement CI/CD.

Documents you may refer to:
1. Resend docs are in `./docs/external/resend/*`.
2. Coin Market Cap (CMC) docs are in `./docs/external/coinmarketcap/*`. Some important docs are stored locally. Read `./docs/external/coinmarketcap/standards-and-conventions.md`, `./docs/external/coinmarketcap/erros-and-rate-limits.md`, and `./docs/external/coinmarketcap/standards-and-conventions.md` before reading other docs.
3. CoinGecko docs are in `./docs/external/coingecko/*`. Read `https://docs.coingecko.com/llms-full.txt` if you want to explore more endpoints and when you have Internet access.
4. MarketStack docs are in `./docs/external/marketstack/*`
5. OpenExchangeRates docs are in `./docs/external/openexchangerates/*`
6. Massive (for stock or FX logos) docs are in `./docs/external/massive/*`
