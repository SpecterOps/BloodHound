## Migrating Graph Data from Neo4j to Postgres 

### Endpoints
| Endpoint | HTTP Request | Usage | Expected Response |
| --- | --- | --- | --- |
| `/pg-migration/status/` | `GET` | Returns a status indicating whether the migrator is currently running. | **Status:** `200 OK`</br></br><pre>{</br>&nbsp;&nbsp;"state": "idle" \| "migrating" \| "canceling"</br>}</pre> |
| `/pg-migration/start/` | `PUT` | Kicks off the migration process from neo4j to postgres. | **Status:** `202 Accepted` |
| `/pg-migration/cancel/` | `PUT` | Cancels the currently running migration. | **Status:** `202 Accepted` |
| `/graph-db/switch/pg/` | `PUT` | Switches the current graph database driver to postgres. | **Status:** `200 OK` |
| `/graph-db/switch/ne04j/` | `PUT` | Switches the current graph database driver to ne04j. | **Status:** `200 OK` |

### Running a Migration
1. Confirm the migration status is currently "idle" before running a migration with the `/pg-migration/status/` endpoint. The migration will run in the same direction regardless of the currently selected graph driver.
2. Start the migration process using the `/pg-migration/start/` endpoint. Since the migration occurs asynchronously, you will want to monitor the API logs to see information regarding the currently running migration.
   - When the migration starts, there should be a log with the message `"Dispatching live migration from Neo4j to PostgreSQL"`
   - Upon completion, you should see the message `"Migration to PostgreSQL completed successfully"`
   - Any errors that occur during the migration process will also surface here
   - You can also poll the `/pg-migration/status/` endpoint and wait for an `"idle"` status to indicate the migration has completed
   - An in-progess migration can be cancelled with the `pg-migration/cancel/` endpoint and run again at any time
3. Once you are ready to switch over to the postgres graph driver, you can use the `/graph-db/switch/pg/` endpoint.