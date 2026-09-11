# Paths
Paths correspond to the endpoints available.
See https://spec.openapis.org/oas/v3.0.3#path-item-object for more info.

---

## Groups
Each path file is prepended with a group name to help organize endpoints.
When moving/creating paths, make sure the operation tag matches other
tags within the same prepended value.

### Example:
The `domains.*.yaml` prepend has operations defined that all have the
`Domains API` tag to group them all together.

---

## File Names
File names roughly match the URI they correspond to. This should
help people identify paths when browsing the directories as well
as help spot errors in path `$refs` in the main `openapi.yaml` doc.

### Example:
The `/api/v2/collectors/{collector_type}/{release_tag}/checksum`
endpoint has the filename `collectors.collectors.type.tag.checksum.yaml`.

---

## Operation Definitions

### Operation ID
Each operation should have an `operationId` defined. This ID *must*
be unique across *all* operations on all endpoints. Think of it
as a unique name, since most code generation tools will use this
value for things like method/action names.

### Parameters
Parameters can be path level or operation level (per verb). Things
like global headers and path variables should be defined at the path
level, query params will often be at the operation level.

#### Common Parameters
```yaml
# For your copy-paste convenience:

# Path level
- $ref: './../parameters/header.prefer.yaml'
- $ref: './../parameters/path.object-id.yaml'

# Operation level
- $ref: './../parameters/query.created-at.yaml'
- $ref: './../parameters/query.updated-at.yaml'
- $ref: './../parameters/query.deleted-at.yaml'
- $ref: './../parameters/query.skip.yaml'
- $ref: './../parameters/query.limit.yaml'
  
# Filter params with predicates
- name: sort_by
  in: query
  description: Sortable columns are [list_columns_here].
  schema:
      $ref: './../schemas/api.params.query.sort-by.yaml'
- name: [string_param_name]
  in: query
  schema:
    $ref: './../schemas/api.params.predicate.filter.string.yaml'
- name: [int_param_name]
  in: query
  schema:
    $ref: './../schemas/api.params.predicate.filter.integer.yaml'
```

### Responses
We should make sure to define all expected responses an endpoint
may return. In most cases, this will be one success case, and one
or more error cases (different status codes).

Make sure to consider more than just the possible status' that the
handlers return. Most endpoints also have auth middleware which can
return `401` and `403` before ever making it to the handler.

#### Common Error Responses
```yaml
# For your copy-paste convenience:
400:
  $ref: './../responses/bad-request.yaml'
401:
  $ref: './../responses/unauthorized.yaml'
403:
  $ref: './../responses/forbidden.yaml'
404:
  $ref: './../responses/not-found.yaml'
429:
  $ref: './../responses/too-many-requests.yaml'
500:
  $ref: './../responses/internal-server-error.yaml'
```
