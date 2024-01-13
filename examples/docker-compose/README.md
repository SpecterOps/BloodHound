# BloodHound Community Edition Docker Compose Example

BloodHound Community Edition is composed of three distinct parts:

-   A PostgreSQL database used for application state storage
-   A Neo4J graph database used for storing all the graph data
-   A single binary containing the BloodHound API server and the UI assets

While these can all be built and run locally, we provide an official Docker image for running the BloodHound binary.
The databases need to be provided separately to provide more modular options for users. As such, this example folder
contains an example `docker-compose.yml` file and some supporting configuration files to help you get started with just
a single command.

## Prerequisites

Using this `docker-compose` configuration requires:

-   A Docker compatible container runtime. Either [Docker](https://www.docker.com/) or
    [Podman (with Docker compatibility enabled)](https://www.redhat.com/sysadmin/podman-docker-compose) will work great
-   [Docker Compose](https://docs.docker.com/compose/install/), which is automatically included with Docker Desktop if you\
    choose to go that route

## Running BloodHound Community Edition

If you're just looking to run the application as quickly as possible locally and don't care about configuration, you can
simply copy the `docker-compose.yml` file from this directory to a location on disk that you want to run it from. Then simply
run `docker compose up` from that directory to start the application. To stop the application, simply run `docker compose down`.

The default ports are as follows:

-   8080 - BloodHound Web Port. You'll access the UI by going to `http://localhost:8080/ui/login` when the server is running
-   7474 - Neo4J Web Interface. Useful for when you need to run queries directly against the Neo4J database
-   7687 - Neo4J Database Port. This is provided in case you want to access the Neo4J database from some other application on your machine

## Configuring BloodHound Community Edition

There are additional files included in this directory to help you configure the application to your needs (such as modifying
what ports different services run on and what credentials should be used):

-   `bloodhound.config.json`: Configuration file used by the BloodHound API server. This example is a direct copy of the one included
    in the official `bloodhound` docker image to be used as a starting point for modifying the configuration. If you want to change
    database credentials, you'll need to update them here as well so `bloodhound` will know how to connect to them.
-   `.env.example`: Copying this file to `.env` in the same directory as your `docker-compose.yml` will allow you to change
    the environment variables easily. Changes to database credentials here will need to be reflected in the `bloodhound.config.json`.

Changing database credentials isn't necessary when running locally, but it is encouraged. It is _highly recommended_ that
if you're going to make any of these ports available outside of localhost that you change the credentials to something secure.
These configuration files are provided to make this process as easy as possible.

## Accessing BloodHound Community Edition

Once the `bloodhound` server shows the message "Server started successfully", you'll be able to access the UI at
`http://localhost:8080/ui/login`. The default port is `8080` but can be configured as mentioned above. In order to login,
you will need to locate the message `# Initial Password Set To:    <password-here>    #`, which is conveniently located
in a decorated block. The default login will be `admin:<password-here>`, and by default the password is randomized at creation.
You will then be asked to choose a new secure password for your admin account. Keep this handy as all future logins will
require it. Afterward, you'll be greeted with the BloodHound Community Edition interface.

## Choosing a BloodHound Version

BloodHound docker images are tagged for each release:

-   `latest` will give you the most recent stable release
-   `X` (e.g. `5`) will give you the latest stable release for that major version
-   `X.X` (e.g. `5.0`) will give you the latest stable release for that minor version
-   `X.X.X` (e.g. `5.0.0`) will give you the release for that specific patch version
-   `X.X.X-rcX` (e.g. `5.0.0-rc1`) will give you a specific release candidate for an upcoming release
-   `edge` will give you the most recent main commit (not guaranteed to be stable)

You can either modify your `docker-compose.yml` configuration to change tags, or if using the example, you can change tags
in your `.env` file under `BLOODHOUND_TAG`.

## Configuration with Environment Variables

See the [wiki](https://github.com/SpecterOps/BloodHound/wiki/Using-Environment-Variables-For-Sensitive-Configuration) for
more information on using environment variables for sensitive configuration options.

## FAQ

### Q: "I'm wanting to connect to BloodHound on a host other than localhost, how do I allow it?"

A: With the default Docker Compose configuration, we bind to `localhost` rather than `0.0.0.0`. This is to ensure that by
default users won't be publishing to their local/external network interfaces. This is great if you're just running on localhost
(even if using a VM with port forwarding, like Docker Desktop), since it ensures outside entities can't connect to your local
instance, but also means that by default you can't intentionally run this configuration on a separate machine or non-port forwarded
virtual machine.

To help with this, the host is set using an environment variable calle `BLOODHOUND_HOST`. This defaults to `127.0.0.1` when
not set elsewhere, but can be altered in one of two ways:

-   Set in your `.env` file. In this directory is a file called `.env.example`. You can copy it to `.env` to have your own
local changes to the environment in file form, or you can create a new `.env` file and simply save it with the variables
you wish to add. For setting the host, simply save the following line: `BLOODHOUND_HOST=0.0.0.0`. When running `docker compose up`
afterwards, the new host should be set (keep in mind, it's better to use the adapter IP you wish to bind to in most scenarios,
`0.0.0.0` is used here for demonstration purposes)
-   Set an environment variable on your Docker host. Depends on your exact host configuration, so please see the documentation
for the host OS you run Docker in.

### Q: "How do I reset my Neo4J database without affecting my application state?"

A: You'll need to find the full name of your Neo4J volume and then run `docker rm <volume-name>`. The following command examples
will help do it all in one step:

-   For Bash compatible shells: `docker volume rm $(docker volume ls -q | grep neo4j-data)`
-   For PowerShell: `docker volume rm @(docker volume ls -q | Select-String neo4j-data)`

### Q: "One of the ports used by default conflicts with a port running on my computer. How can I change it?"

A: Ports can be configured by changing the port in your `.env` file. Simply find the port you want to change and change
it there before restarting the docker compose services.

### Q: "This is great, but Neo4j needs more juice. How do I configure Neo4j (the easy way)?"

A: The Neo4j Docker image reads configuration parameters as environment variables with the prefix `NEO4J_`. A handy conversion
table for Neo4j parameters to environment variables can be found here: https://neo4j.com/docs/operations-manual/4.4/docker/ref-settings/.
These environment variables can be added to the `docker-compose.yml` file in this example directory under the `graph-db`
service as additional list items for the `environment` parameter (follow the way the NEO4J auth is passed in for an example)

### Q: "How can I access the databases directly (especially Neo4j's browser)?"

A: Port forwarding is commented out by default for the databases due to a default password being used. If you change your
database passwords, you can easily uncomment the lines in `docker-compose.yml` to provide port forwarded access.

### Q: "Can I run these services in the background?"

A: Absolutely, simply run `docker compose up -d` to start the services up in the background. To access the logs, you can
use `docker compose logs -f` to open the logs in follow mode (ignore the -f if you just want the logs as of that moment in time).
To stop the services, use `docker compose down`.

### Q: "My databases persist between runs, how can I fully reset them?"

A: You can clear out all volumes by using `docker compose down --volumes`. This will lead to both databases being reset
the next time you run `docker compose up`. Docker Compose does not expose an easy way to reset only one volume, so deleting
a single volume is left as an exercise for the reader (you'll need to look at removing directly through Docker with
`docker volume rm` or using the Docker Desktop GUI)

### Q: "Restarting the application requires logging in again, why?"

A: By default, we generate a secure random 256-bit key for JWT signing. Because this happens on every server restart,
any existing sessions will be invalidated. If you need sessions to survive a server restart, there is a configuration
value available that will allow you to specify your own `base64` encoded 256-bit key. It is recommended that you configure
this when running Bloodhound on a standalone server, alongside other security configurations.
