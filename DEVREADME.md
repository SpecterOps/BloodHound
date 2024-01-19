# BloodHound Community Edition Development

Welcome to BloodHound Community Edition! This README should help point you in the right direction if you're looking to begin development with BloodHound. 
It is intended to be a simple place to look to get your dev environment set up. 

More detailed information regarding [contributing](https://github.com/SpecterOps/BloodHound/wiki/Contributing), [code structure](https://github.com/SpecterOps/BloodHound/wiki/Code), and [development](https://github.com/SpecterOps/BloodHound/wiki/Development) can be found in our [GitHub wiki](https://github.com/SpecterOps/BloodHound/wiki).

# Dev Environment Prerequisites

-   [Just](https://github.com/casey/just)
-   [Python 3.10](https://www.python.org/downloads/)
-   [Go 1.21](https://go.dev/dl/)
-   [Node 18](https://nodejs.dev/en/download/)
-   [Yarn 3.6](https://yarnpkg.com/getting-started/install)
-   [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker/Docker Compose compatible runtime)

# Quick Start

* Run `just init`
* Run `just bh-dev`
* To access the UI, navigate to `http://bloodhound.localhost`. Login with the user `admin` and password that can be found in the log output of the app.
