<p align="center">
    <picture>
        <source media="(prefers-color-scheme: dark)" srcset="cmd/ui/public/img/logo-secondary-transparent-full.svg">
        <img src="cmd/ui/public/img/logo-transparent-full.svg" alt="BloodHound Community Edition" width='400' />
    </picture>
</p>

<hr />
# THIS IS A TEST!!!

BloodHound is a monolithic web application composed of an embedded React frontend with [Sigma.js](https://www.sigmajs.org/) and a [Go](https://go.dev/) based REST API backend. It is deployed with a [Postgresql](https://www.postgresql.org/) application database and a [Neo4j](https://neo4j.com/) graph database, and is fed by the [SharpHound](https://github.com/SpecterOps/SharpHound) and [AzureHound](https://github.com/SpecterOps/AzureHound) data collectors.

BloodHound uses graph theory to reveal the hidden and often unintended relationships within an Active Directory or Azure environment. Attackers can use BloodHound to quickly identify highly complex attack paths that would otherwise be impossible to find. Defenders can use BloodHound to identify and eliminate those same attack paths. Both red and blue teams can use BloodHound to better understand privileged relationships in an Active Directory or Azure environment.

BloodHound CE is created and maintained by the [BloodHound Enterprise Team](https://bloodhoundenterprise.io). The original BloodHound was created by [@\_wald0](https://www.twitter.com/_wald0), [@CptJesus](https://twitter.com/CptJesus), and [@harmj0y](https://twitter.com/harmj0y).

## Running BloodHound Community Edition
Docker Compose is the easiest way to get up and running with BloodHound CE. Instructions below describe how to install and upgrade your deployment.

### System Requirements
BloodHound CE deploys in a traditional multi-tier container architecture consisting of databases, application, and UI layers. 

**Minimum specifications:**

- 4GB of RAM
- 4 processor cores
- 10GB hard disk space

**For large environments (>50,000 users):**

- 96GB of RAM
- 12 processor cores
- 50GB hard disk space

### Deploy BloodHound CE
Deploying BloodHound CE quickly with the following steps:

1. Install [Docker Desktop](https://www.docker.com/products/docker-desktop/). Docker Desktop includes Docker Compose as part of the installation.
2. Download the [Docker Compose YAML file](examples/docker-compose/docker-compose.yml) and save it to a directory where you'd like to run BloodHound. You can do this from a terminal application with `curl -L https://ghst.ly/getbhce`.
   > On Windows: Execute the command in CMD, or use `curl.exe` instead of `curl` in PowerShell.
3. Navigate to the folder with the saved `docker-compose.yml` file and run `docker compose pull && docker compose up`.
4. Locate the randomly generated password in the terminal output of Docker Compose.
5. In a browser, navigate to `http://localhost:8080/ui/login`. Login with a username of `admin` and the randomly generated password from the logs.

*NOTE: The default `docker-compose.yml` example binds only to localhost (127.0.0.1). If you want to access BloodHound outside of localhost, you'll need to follow the instructions in [examples/docker-compose/README.md](examples/docker-compose/README.md) to configure the host binding for the container.*

### Upgrade BloodHound CE
Once installed, upgrade BloodHound CE to the latest version with the following steps:

1. Navigate to the folder with the saved `docker-compose.yml` file and run `docker compose pull && docker compose up`.
2. In a browser, navigate to `http://localhost:8080/ui/login` and log in with your previously configured username and password.

### Importing sample data

The BloodHound team has provided some sample data for testing BloodHound without performing a SharpHound or AzureHound collection. That data may be found [here](https://github.com/SpecterOps/BloodHound/wiki/Example-Data).

## Installation Error Handling

- If you encounter a "failed to get console mode for stdin: The handle is invalid." ensure Docker Desktop (and associated Engine is running). Docker Desktop does not automatically register as a startup entry.

<p align="center">
    <img width="302" alt="Docker Engine Running" src="cmd/ui/public/img/Docker-Engine-Running.png">
</p>

- If you encounter an "Error response from daemon: Ports are not available: exposing port TCP 127.0.0.1:7474 -> 0.0.0.0:0: listen tcp 127.0.0.1:7474: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted." this is normally attributed to the "Neo4J Graph Database - neo4j" service already running on your local system. Please stop or delete the service to continue.

```
# Verify if Docker Engine is Running
docker info

# Attempt to stop Neo4j Service if running (on Windows)
Stop-Service "Neo4j" -ErrorAction SilentlyContinue
```

- A successful installation of BloodHound CE would look like the below:

https://github.com/SpecterOps/BloodHound/assets/12970156/ea9dc042-1866-4ccb-9839-933140cc38b9

## Useful Links

- [BloodHound Slack](https://ghst.ly/BHSlack)
- [Wiki](https://github.com/SpecterOps/BloodHound/wiki)
- [Contributors](./CONTRIBUTORS.md)
- [Docker Compose Example](./examples/docker-compose/README.md)
- [BloodHound Docs](https://support.bloodhoundenterprise.io/)
- [Developer Quick Start Guide](https://github.com/SpecterOps/BloodHound/wiki/Development)
- [Contributing Guide](https://github.com/SpecterOps/BloodHound/wiki/Contributing)

## Contact

Please check out the [Contact page](https://github.com/SpecterOps/BloodHound/wiki/Contact) in our wiki for details on how to reach out with questions and suggestions.

## Licensing

```
Copyright 2023 Specter Ops, Inc.

Licensed under the Apache License, Version 2.0
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

Unless otherwise annotated by a lower-level LICENSE file or license header, all files in this repository are released
under the `Apache-2.0` license. A full copy of the license may be found in the top-level [LICENSE](LICENSE) file.
