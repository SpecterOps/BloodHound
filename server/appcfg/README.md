# Appcfg

The `appcfg` module has two parts: datapipe and appconfigs.

## Datapipe

Datapipe is the background graph processing pipeline, run by the [datapipe daemon](/cmd/api/src/daemons/datapipe/datapipe.go), which transitions the datapipe through a series of statuses as it executes processing. The `appcfg` module just handles the GET API for datapipe, allowing API access to fetch the current datapipe status.

### Interface

This module handles the [GET api/v2/datapipe/status endpoint](https://bloodhound.specterops.io/api-reference/datapipe/get-datapipe-status#get-datapipe-status).

### Implementation notes

The status is stored as a singleton in the enterprise database `datapipe_status` table.

## Appconfig

Application configurations are key values set on a particular instance of the Bloodhound application which control specific configurable aspects of Bloodhound's behavior, such as password expiration or scheduled analysis rules. These application configuration parameters have a unique key identifier and a custom data format to represent the value of a given parameter.

### Interface

This module handles the `api/v2/config` [GET](https://bloodhound.specterops.io/reference/config/list-application-config-parameters) and [PUT](https://bloodhound.specterops.io/reference/config/write-application-configuration-parameters) (TODO BED-9765) API endpoints, supporting reading and writing an allowed subset of configuration parameters.

`GetConfig` is also used internally by the Bloodhound application, allowing relevant modules to read configuration values to control their config-dependant behavior.

### Implementation notes

Configs are stored in the `parameters` table of the enterprise database.

## Working in this module

### Creating a new application configuration parameter

New parameters must be defined in [param_types.go](./internal/services/param_types.go).

(1) Add your parameter's key under `# Param keys` with an appropriate name.

(2) Under `## Param types list`, create the struct definition for your params fields.
* Specifiy `json` tags to match the stored value. 
* Note that for any ISO duration values, you can use the ISODuration type to automatically unmarshal the value into a Go duration.
* Specify `validation` tags if you want validation enforced on http API parameter update requests.

(3) Add a value to `paramTypeDefinitions` to define your parameter behavior.
* Create a new entry with your parameter key as the key
* Specify whether this key should be API-updatable
* Specify a default and any data transformations using your parameter struct type

You should now be able to fetch this parameter using the `GetConfig` function, any value under that key should be returned via API (if allowed), and be updatable via API (if allowed).