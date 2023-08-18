<p align="center">
    <img src="cmd/ui/public/img/logo-white-full.svg" alt="BloodHound Community Edition" style="width: 400px;" />
</p>

<hr />

BloodHound is a monolithic web application composed of an embedded React frontend
with [Sigma.js](https://www.sigmajs.org/)
and a [Go](https://go.dev/) based REST API backend. It is deployed with a [Postgresql](https://www.postgresql.org/)
application
database and a [Neo4j](https://neo4j.com/) graph database, and is fed by
the [SharpHound](https://github.com/BloodHoundAD/SharpHound)
and [AzureHound](https://github.com/BloodHoundAD/AzureHound) data collectors.

BloodHound uses graph theory to reveal the hidden and often unintended relationships within an Active Directory or Azure
environment. Attackers can use BloodHound to easily identify highly complex attack paths that would otherwise be
impossible
to quickly identify. Defenders can use BloodHound to identify and eliminate those same attack paths. Both blue and red
teams
can use BloodHound to easily gain a deeper understanding of privilege relationships in an Active Directory or Azure
environment.

BloodHound CE is created and maintained by the [BloodHound Enterprise Team](https://bloodhoundenterprise.io). The
original
BloodHound was created by [@_wald0](https://www.twitter.com/_wald0), [@CptJesus](https://twitter.com/CptJesus), and
[@harmj0y](https://twitter.com/harmj0y).

## Quick Start

BloodHound CE is distributed as a Docker image available at https://hub.docker.com/r/specterops/bloodhound.
In order to get started, an example docker compose folder is provided
at [examples/docker-compose](examples/docker-compose/README.md).

### Prerequisites

Running the example Docker Compose project requires the following:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

The simplest way to get started is to install Docker Desktop, as it will provide both prerequisites and requires no
additional
configuration: [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Because BloodHound CE is distributed as a Docker image, there are many other ways to deploy, but this provides the
simplest setup.

### Running with Docker Compose

- Download our example [docker-compose.yml](./examples/docker-compose/docker-compose.yml)
- In whatever directory you copied the `docker-compose.yml` file to, run `docker compose up`
- The initial password will show up in the logs when the API server fully starts (there's a "server started" message
  that will appear)
- Go to `http://localhost:8080/ui/login` and use the username `admin` with your randomly generated password
- Congrats, you're now running BloodHound and can complete your application setup

### Helpful Tips

- You'll be asked to change your password on first login
- To get the latest compatible collectors, simply click the gear icon in the corner and select "Download Collectors"
- If you restart the service before copying your random password, it will not be regenerated. Simply
  run `docker compose down -v`
  and then `docker compose up` to reset your databases.
- More information, troubleshooting, and how to configure your deployments can be found
  in [Docker Compose Example README](./examples/docker-compose/README.md)

## Useful Links

- [Contributors](./CONTRIBUTORS.md)
- [Docker Compose Example](./examples/docker-compose/README.md)
- [BloodHound Docs](https://support.bloodhoundenterprise.io/)
- [Developer Quick Start Guide](./DEVREADME.md)
- [Contributing Guide](./CONTRIBUTING.md)

## Contact

### Join the BloodHound Gang Slack

[You may sign up for the BloodHound Slack workspace here.](https://join.slack.com/t/bloodhoundhq/shared_invite/zt-1tgq6ojd2-ixpx5nz9Wjtbhc3i8AVAWw)

### BloodHound Support

If you need to contact our team directly and do not wish to use Slack you may do so by sending an email
to `support [AT] specterops.io`.

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
