This directory contains the code for the project `moneycoach` (previously it was `quanta`, so there are some legacy namings, but we are still far from prod).
- The project PRD documentation is specified in `./specs/PRD.md`.
- The mobile app code is in `./mobile-app`.
- The backend for the mobile app code is in `./app-backend`.

# Development Rules

- When installing dependencies, always prefer the latest version. Check if you're using the latest version online before you proceed with a certain version.
- Never use fallbacks. Never keep legacy code. We'd prefer drastic refactor right now rather than pile up spaghetti code and feel the pain in the future.
- Always think of best practices and modularity. Keep the code maintainable.
- When fixing a problem, don't blindly resolve to patches or fallbacks. You should look deep into it and fix the root cause.
- When using LLMs, always force the output to be JSON, and meticulously check that each field is intended, and there should be no unused or redundant fields. Design the data flow with wisdom. And to remind you, I repeat: Never use fallbacks. Never keep legacy code. We'd prefer drastic refactor right now rather than pile up spaghetti code and feel the pain in the future.
- It's forbidden to write "keyword enumeration router" styled code. Use LLMs as the intelligence layer instead.
- In the development process, if you find certain environment variables are missing, don't silently go ahead and fallback to mocks. Speak out what environment variables are missing and the developer will configure them.
- Always update the docs accordingly after you make code changes.
- For both iOS and Android, all release/publish actions must go through `npx eas-cli ...`; do not use GitHub workflows for mobile release/publish.

Follow `app-backend/AGENTS.md` and `mobile-app/AGENTS.md`.

# Available Tools

- You can use the `gh` command to interact with Github, say, when you want to push or view the status or logs of the workflows.
- You can use the `aws` command to interact with AWS, say, when you want to check the status of a Cloudformation deployment.
