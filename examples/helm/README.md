# BloodHound Community Edition Helm Chart Example

BloodHound Community Edition is composed of three distinct parts:

-   A PostgreSQL database used for application state storage
-   A Neo4J graph database used for storing all the graph data
-   A single binary containing the BloodHound API server and the UI assets

This Helm chart templates these three services and facilitates communication between them inside a Kubernetes cluster. This chart uses the Prebuilt docker images to run inside of Kubernetes. Helm will allow you to template important information about your deployment. Please use the included `values.yaml` to customize deployment configuration.

## Prerequisites

In order to run this chart you will need the following:
- A Kubernetes Cluster in which to deploy (any supported version should work with this chart).
- Admin access or the ability to create Helm charts and the associated Objects in this chart (validate RBAC Policies and Policy Engine configuration)
- Helm (v3+) installed on a local workstation or CI Pipeline.

## Quick start

To install this chart inside of a Kubernetes cluster **make sure helm is properly configured to point at the correct cluster within the kubeconfig**. Once You have validated your helm install, clone this repo and run the following command from your workstation or pipeline:

```
$ helm install bhce $BH_ROOT/examples/helm/.
```

## Accessing
By default, the ingress is enabled. This means the application will be available on the ingress endpoint with the host set to `bloodhound.example.com` (Each of these values can be configured in the values.yaml). As long as you have a properly configured ingress controller and valid DNS configuration the application will be available at `https://bloodhound.example.com`, else without DNS you can test with curl by passing the 'Host' Header: `curl -H 'Host: bloodhound.example.com' https://$endpointIP`.

 If you have a TLS cert you can enable `bloodhound.tls.customCert` in the values.yaml and provide the ingress secret in `bloodhound.tls.certSecret`. Else, the provided cert will be self signed from the ingress controller.

 ## Configuring BloodHound Community Edition

To configure the Helm Chart deployment of BloodHound Community Edition you can use the two files specified below:

-   `values.yaml` - A general Helm Chart configuration file - you can use this to configure aspects of the deployment of BloodHound and set Environment Variables. This file generally will be the single source of truth for application and deployment configuration.
-   `templates/cmbh.yaml` - This is the Kubernetes configmap used for the BloodHound Application - you can add your own custom configuration in the `bloodhound-config.json` section of this file. Some of it is populated with the `values.yaml` as well for portibility

If using a custom Certificate for TLS please terminate it on the ingress and not the application directly for easier Kubernetes native management. 

Changing the database credentials is not a requirement for development/testing but is **Highly** Recommended for production environments.

 ## Ports
The default ports are as follows:

-   8080 - BloodHound Web Port. This is the backend port of the service used by BloodHound. 
-   7474 - Neo4J Web Interface. Useful for when you need to run queries directly against the Neo4J database. (Note: this is not exposed externally by default)
-   7687 - Neo4J Database Port. This is provided in case you want to access the Neo4J database from some other application on your machine. (Note: this is not exposed externally by default)


## Choosing a BloodHound Version

BloodHound docker images are tagged for each release:

-   `latest` will give you the most recent stable release
-   `X` (e.g. `5`) will give you the latest stable release for that major version
-   `X.X` (e.g. `5.0`) will give you the latest stable release for that minor version
-   `X.X.X` (e.g. `5.0.0`) will give you the release for that specific patch version
-   `X.X.X-rcX` (e.g. `5.0.0-rc1`) will give you a specific release candidate for an upcoming release
-   `edge` will give you the most recent main commit (not guaranteed to be stable)

By Default the `latest` tag is set. You can change this by setting the `bloodhound.container.version` in the values.yaml

## Troubleshooting

Please assure any local host-based firewall frontends such as `firewalld` or `ufw` on the node is disabled. Please use Kubernetes NetworkPolicies and admission controllers instead.

Validate all 3 of the Deployments are healthy, Services are active and the ingress is properly configured  with `kubectl get -n bloodhoundad deployments,svc,ingress`

Assure the cluster DNS is resolving the dns for the 'app-db' service and the 'graph-db' service. An easy way to validate this is to spin up a busybox pod in the bloodhoundad namespace and use nslookup to check for resolution issues.

```
$ kubectl run -it --rm dns-test -n bloodhoundad --image busybox:latest
# for name in app-db graph-db; do nslookup $name; done
```
If you are having trouble with cluster DNS please refer to the upstream Kubernetes documentation:
https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/

Else, please refer to the upstream Kubernetes Documentation for General cluster troubleshooting: https://kubernetes.io/docs/tasks/debug/debug-cluster/.