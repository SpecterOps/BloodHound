# Data Quality OpenGraph Fixture

This fixture creates a synthetic OpenGraph source kind named `DQTBase`, one environment kind named `DQT_Environment`, and a few object kinds. It is intended for manually testing data-quality object counts.

1. Register the OpenGraph extension.

   Do not upload `extension.json` through the file-ingest upload flow. It is a schema-extension payload and must be sent directly to the OpenGraph extensions API:

   ```sh
   curl -X PUT "$BLOODHOUND_URL/api/v2/extensions" \
     -H "Authorization: Bearer $BLOODHOUND_TOKEN" \
     -H "Content-Type: application/json" \
     --data-binary @cmd/api/src/services/dataquality/fixtures/opengraph/extension.json
   ```

2. Upload `ingest.json` through the file-ingest upload flow or UI, then end the ingest job so the datapipe can analyze it.

3. If analysis does not run automatically, request it:

   ```sh
   curl -X PUT "$BLOODHOUND_URL/api/v2/analysis" \
     -H "Authorization: Bearer $BLOODHOUND_TOKEN"
   ```

4. After analysis/data-quality collection runs, query:

   ```sh
   curl "$BLOODHOUND_URL/api/v2/data-quality/source-object-counts?source_kind=DQTBase" \
     -H "Authorization: Bearer $BLOODHOUND_TOKEN"

   curl "$BLOODHOUND_URL/api/v2/data-quality/environment-object-counts?source_kind=DQTBase" \
     -H "Authorization: Bearer $BLOODHOUND_TOKEN"
   ```

Expected source-level counts include `DQT_Environment: 2`, `DQT_User: 3`, `DQT_Group: 2`, and `DQT_Application: 2`.

Expected environment-level counts use uppercased environment IDs. `DQT-ENV-ENGINEERING` should include `DQT_Environment: 1`, `DQT_User: 2`, `DQT_Group: 1`, and `DQT_Application: 1`. `DQT-ENV-SECURITY` should include `DQT_Environment: 1`, `DQT_User: 1`, and `DQT_Group: 1`. The global portal app is source-level only because it has no `environmentid`.
