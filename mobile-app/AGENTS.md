This directory contains the code for a cross-platform mobile app.

Rules in development:
1. Use the skill $react-native-architecture.
2. Before writing/updating code, always write/update the UI/UX prototypes in `../prototypes`, so the code and the prototypes match. They should not only contain detailed UI information, but also include user interaction logic and data contracts. Every screen should have an ID for easy tracking, and there should be integration descriptions to specify how various screens are orchistrated and the data contracts. There should not be inconsistency, ambiguity, or under-specification in the prototype docs.
3. The app should support iOS and android.
4. Run `npm run lint` and resolve all errors/warnings (some warnings can be resolved with `npm run lint -- --fix`).
5. Run `npm run test` and make sure all tests pass.
6. Run `npx tsc --noEmit -p tsconfig.json` and resolve all errors/warnings.
7. For both iOS and Android, all release/publish actions must go through `npx eas-cli ...`; do not use GitHub workflows for mobile release/publish.
