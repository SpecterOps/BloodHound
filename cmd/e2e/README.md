# BDD E2E BHCE Tests

BHCE BDD E2E tests using [Cucumber] & opensource [PlayWright](https://playwright.dev) also utilizing playwright to write API tests. Initial phase is to write BDD tests using Cucumber since their test runner is much mature and stable while leveraging playwright API library to run tests on modern browsers, the playwright project is backed by Microsoft. Once the playwright test runner provides support for running [BDD](https://github.com/microsoft/playwright/issues/11975) natively, then we can configure to run BDD tests using playwright runners instead of Cucumber if needed.

## Install Dependencies 

- Node `v22.11.0`
- Yarn `v3.6.1`

```
yarn install 
```

## Run Tests Locally

- `just bh-dev` standup BHCE instance 

End2End [just](https://github.com/casey/just) recipes are also avaible at the root level of BHCE

Setup Prisma Schema and Client: 
- `cp .env.example .env` & update DB URL
- `just prisma-generate`

Update Cross-Enviroment variables
- `cp .env-cmdrc.example .env-cmdrc` 

Runs Cucumber BDD tests with `@e2e` tag
- `just run-e2e-tests` 

