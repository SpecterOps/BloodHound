# Diff Engine POC

The Diff Engine is a utility POC for comparing two JSON schema definition files (an old schema vs a new schema) to determine the changes required to bring a database into alignment.

It inspects schemas, fields, metadata, and other attributes to produce a structured report of creates, updates, and deletes which it then prints as a result.

## Run the app

1. Run the database: `just bhe-testing up`
2. Run the app with the two schemas that you want to compare `CONNECTION_STRING="user=xyz password=xyz dbname=bloodhound host=localhost port=5432" OG_SCHEMA_PATH="./bhce/packages/go/diffengine/data/v1.json" NEW_SCHEMA_PATH="./bhce/packages/go/diffengine/data/v2.json" SCHEMA_NAME="CatProfile" go run ./bhce/packages/go/diffengine/cmd/.`

### Environment Variables

| Variable              | Description                                                                   |
| --------------------- | ----------------------------------------------------------------------------- |
| `CONNECTION_STRING`   | PostgreSQL connection string                                                  |
| `OLD_SCHEMA`          | Path to the JSON file containing the previous version of one or more schemas. |
| `NEW_SCHEMA`          | Path to the JSON file containing the updated version of one or more schemas.  |
| `SCHEMA_NAME`         | The exact name of the schema to compare (e.g. `"CatProfile"`).                |
