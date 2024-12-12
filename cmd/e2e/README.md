# BDD E2E BHCE Tests

BHCE BDD E2E tests using [Cucumber] & opensource [PlayWright](https://playwright.dev) also utilizing playwright to write API tests. Initial phase is to write BDD tests using Cucumber since their test runner is much mature and stable while leveraging playwright API library to run tests on modern browsers, the playwright project is backed by Microsoft. Once the playwright test runner provides support for running [BDD](https://github.com/microsoft/playwright/issues/11975) natively, then we can configure to run BDD tests using playwright runners instead of Cucumber if needed.

## Install Dependencies 

- Node `v22.11.0`
- Yarn `v3.6.1`

```
yarn install 
```

## Run Tests Locally

Setup Prisma Schema and Client 
- `cp .env.example .env` & update DB URL
- `just prisma-generate`

Update Cross-Enviroment variables
- `cp .env-cmdrc.example .env-cmdrc` 

Scripts configured in `package.json` to support running different type of tests.

- `just run-e2e-tests` runs all cucumber BDD tests using [just](https://github.com/casey/just) 
