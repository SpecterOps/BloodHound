# LLM Instructions: RACF-Aware BloodHound Fork / RACFHound

## Project goal

Build a RACF-aware analysis layer and UI experience on top of BloodHound Community Edition.

The project should help security consultants and defenders analyze RACF authorization paths, privilege inheritance, sensitive dataset exposure, SURROGAT abuse paths, APF write risk, and other z/OS-specific escalation/control relationships.

The intent is not merely to display RACF data as a generic graph. The tool should explain why a RACF relationship matters, how a user/group/service identity reaches authority, and what the effective risk is.

## Repository strategy

Assume the BloodHound Community Edition codebase is maintained as a fork.

Use the following Git model:

```text
origin        = user's fork
upstream      = SpecterOps/BloodHound
main          = clean mirror of upstream/main
racf-main     = stable RACF integration branch
feature/*     = small focused implementation branches
```

Do not put experimental or unrelated changes directly into `main`.

`main` should remain a clean mirror of upstream BloodHound so that updates from SpecterOps can be merged or rebased cleanly.

All RACF-specific work should land in `racf-main` via focused feature branches, for example:

```text
feature/racf-schema
feature/racf-importer
feature/racf-group-detail-panel
feature/racf-user-detail-panel
feature/racf-dataset-detail-panel
feature/racf-dashboard-special
feature/racf-dashboard-apf-write
feature/racf-rule-engine
feature/racf-path-explanations
```

When changing the BloodHound fork, keep the diff against upstream as small and isolated as possible.

## Related repositories and ownership boundaries

The RACFHound system currently spans three repositories. Treat them as separate
release units with one-way dependencies:

```text
mfpandas
  -> mfpandas-racfhound
       -> racfhound
            -> BloodHound CE APIs
```

Repositories reviewed on 2026-06-27:

```text
JonathanPrince/mfpandas-racfhound
  reviewed revision: 7690aa663ac79c89af07acdfcb3c46fb1bc37f67

JonathanPrince/racfhound
  reviewed revision: 71cf0a6b85742bda947f84301c679e4313d214d7

JonathanPrince/BloodHound
  integration branch: racf-main
```

Repository responsibilities:

### `JonathanPrince/mfpandas-racfhound`

This is the transformation and graph-model package. It currently:

- Accepts a parsed `mfpandas.IRRDBU00` object.
- Builds RACF nodes and relationships.
- Resolves controlling generic dataset profiles for supplied concrete
  APF/PARMLIB/PROCLIB datasets.
- Emits BloodHound OpenGraph JSON.
- Owns deterministic graph-model, generic-profile, dataset-resolution, and
  export tests.

RACF semantic derivation belongs here unless it requires data that is outside
the parsed unload and collector inputs.

### `JonathanPrince/racfhound`

This is the user-facing collector and orchestration package. It currently:

- Collects RACF and system context from z/OS using SSH, JCL, and FTP.
- Calls `mfpandas-racfhound` for transformation.
- Authenticates to BloodHound.
- Provisions node presentation metadata.
- Installs bundled RACF saved queries.
- Uploads OpenGraph data.

Collector behavior, CLI/API compatibility adapters, provisioning, saved-query
distribution, and end-to-end workflow tests belong here.

### `JonathanPrince/BloodHound`

This fork should contain only functionality that cannot be delivered through
the public OpenGraph, upload, saved-query, or other supported BloodHound APIs.
Examples include richer RACF-specific entity panels, dashboards, navigation,
and presentation behavior that the generic extension UI cannot express.

Do not copy the Python parser, graph builder, collector, provisioning client, or
query bundle into the BloodHound fork. Do not add a dependency from either
Python repository back to fork-internal Go or TypeScript packages.

When a feature crosses repositories, change the lowest-level contract first:

1. Update and test `mfpandas-racfhound`.
2. Update `racfhound` to consume and provision the contract.
3. Update the BloodHound fork only if additional presentation is still needed.
4. Record the compatible versions or revisions in release notes.

## Licensing and attribution constraints

BloodHound Community Edition is Apache-2.0 licensed. When modifying or redistributing code:

- Preserve existing copyright notices.
- Preserve the Apache-2.0 license text.
- Mark modified files clearly where appropriate.
- Do not imply that RACFHound or the fork is an official SpecterOps product.
- Avoid reusing BloodHound/SpecterOps branding beyond accurate attribution such as “based on BloodHound Community Edition.”

Prefer neutral project naming such as:

```text
RACFHound
RACFGraph
MainframeHound
```

## Architecture principle

Separate RACF reasoning from BloodHound presentation.

BloodHound fork responsibilities:

- RACF-specific UI panels.
- Entity drill-down views.
- Dashboards.
- Clickable graph navigation.
- Finding and path visualization.
- Integration with BloodHound data/query APIs.

Independent RACF core responsibilities:

- IRRDBU00 parsing.
- DSMon parsing.
- SETROPTS parsing.
- STARTED class parsing.
- APF/LINKLIST/PARMLIB/PROCLIB enrichment.
- Generic dataset profile matching and precedence handling.
- Effective access calculation.
- Group nesting resolution.
- SURROGAT relationship derivation.
- Analysis rules.
- Finding generation.
- Export to BloodHound/OpenGraph-compatible JSON.

The RACF-specific analysis engine should be usable without the BloodHound UI. It should support headless execution for report generation, CI testing, and client deliverables.

Suggested logical layout:

```text
racfhound-core/
  parsers/
    irrdbu00/
    dsmon/
    setropts/
    started/
  model/
    nodes/
    edges/
    properties/
  analysis/
    rules/
    scoring/
    findings/
    paths/
  export/
    opengraph/
    json/
    markdown/
  tests/

bloodhound-racf-fork/
  racf/
    dashboards/
    entity-panels/
    path-explanations/
    finding-views/
    query-library/
```

Adapt the exact paths to the real BloodHound repository structure, but preserve the separation of concerns.

## Update-safe BloodHound integration

Use supported extension points before modifying upstream BloodHound code.
Preferred order:

1. OpenGraph schema extension for node kinds, relationship kinds,
   traversability, environments, icons, colors, and relationship findings.
2. OpenGraph data upload for nodes and edges.
3. Saved-query APIs for RACF query distribution.
4. A self-contained RACF feature module or UI package in this fork.
5. Small, explicitly tracked upstream patches only when no stable extension
   point exists.

This BloodHound checkout currently exposes the modern schema-extension API at
`PUT /api/v2/extensions` behind `opengraph_extension_management`. It also
contains OpenGraph search, findings, pathfinding support, and an
`opengraph_entity_panel` feature flag. Inspect the current implementation and
feature-flag state before adding replacement code.

The existing `racfhound provision` implementation targets the older
`POST /api/v2/custom-node-kinds` endpoint and the existing upload client targets
`POST /api/v2/graphs/custom-upload`. Treat both endpoints as version-dependent
compatibility behavior, not as the permanent integration contract.

Evolve `racfhound provision` into the customization/bootstrap command. It
should:

- Detect or accept the target BloodHound compatibility level.
- Install or update the RACF OpenGraph extension schema.
- Install saved queries idempotently.
- Validate required feature flags and permissions.
- Upload through the supported ingest workflow for the target release.
- Print actionable diagnostics when the target release is unsupported.
- Never edit a BloodHound source checkout.

If source-level customization is necessary, keep it under a clearly named RACF
boundary and maintain a machine-readable customization manifest. The manifest
should identify:

- The upstream base revision tested.
- Every upstream-owned file touched.
- The reason each patch is required.
- The test that proves the customization still applies.
- Whether a newer public extension point can replace it.

A helper script may verify and apply these tracked customizations, but it must
fail safely when expected anchors or upstream revisions do not match. Do not
use broad search-and-replace against upstream source.

Before merging an upstream update:

1. Update the clean `main` mirror from `upstream/main`.
2. Run converter contract fixtures against the updated BloodHound ingest path.
3. Rebase or merge `main` into a temporary RACF update branch.
4. Run the customization-manifest checks.
5. Run `just prepare-for-codereview`.
6. Perform a smoke test: provision, upload a fixture, search for a RACF node,
   run a saved query, open entity details, and find a path.
7. Merge the verified update into `racf-main`.

The current local checkout has only an `origin` remote configured. Add the
official SpecterOps repository as `upstream` before using the update workflow;
do not guess or silently rewrite an existing remote.

## OpenGraph contract and compatibility

Maintain one versioned, canonical RACF graph contract. The exporter,
provisioning schema, saved queries, fork UI, fixtures, and documentation must
agree on exact kind names, edge direction, required properties, and identifiers.

Modern BloodHound graph-extension kinds must be prefixed with the extension
namespace. Use namespace `racf` and migrate the existing kinds consistently,
for example:

```text
RACFUser       -> racf_RACFUser
RACFGroup      -> racf_RACFGroup
RACFMemberOf   -> racf_RACFMemberOf
RACFCanWrite   -> racf_RACFCanWrite
```

Keep display labels RACF-native even when stored kind names are namespaced.
Provide an explicit legacy-output mode only while older BloodHound releases
must remain supported. Do not emit a mixture of namespaced and legacy RACF
kinds in one dataset.

The modern data payload should include:

- A stable `metadata.source_kind` owned by the RACF extension.
- Stable node IDs that do not depend on array order.
- The display kind plus the RACF base/source kind in each node's `kinds`.
- `environmentid` and `collected` properties where required by the target
  OpenGraph environment model.
- Edge endpoints using the target release's supported `match_by` format.
- Provenance and explanation properties on derived relationships.

Keep a compatibility matrix in the `racfhound` documentation with:

```text
BloodHound version or revision
schema API and required feature flags
ingest API/workflow
RACF graph contract version
compatible mfpandas-racfhound version
compatible racfhound version
verified date
```

Contract changes require:

- A version increment.
- Golden schema and graph fixtures.
- Validation against the BloodHound extension and ingest handlers.
- Saved-query regression tests.
- A documented migration path for existing RACF data and queries.

## RACF graph model

Use explicit RACF-oriented node and edge types. Avoid forcing RACF concepts into Active Directory terminology unless there is a clear UX reason.

The names below are conceptual names used in discussion and UI copy. Serialize
them using the versioned, namespaced OpenGraph names defined by the canonical
contract.

Core node types:

```text
RACFUser
RACFGroup
RACFDataset
RACFResource
RACFClass
RACFPrivilege
RACFFinding
RACFPath
```

Useful privilege target nodes:

```text
SPECIAL
OPERATIONS
AUDITOR
ROAUDIT
BPX.SUPERUSER
BPX.FILEATTR.APF
BPX.FILEATTR.PROGCTL
APF Code Execution
PARMLIB Write
PROCLIB Write
LINKLIST Write
RACF DB Write
```

Core edge types:

```text
RACFMemberOf
RACFSubgroupOf
RACFOwns
RACFHasPrivilege
RACFClassAuth
RACFCanRead
RACFCanUpdate
RACFCanControl
RACFCanAlter
RACFCanWrite
RACFSurrogateFor
RACFCanSubmitAs
RACFStartedAs
RACFAffects
RACFEvidencedBy
```

Every edge should carry enough evidence and explanation to support drill-downs and reports.

Suggested edge properties:

```json
{
  "source_record": "0205",
  "source_file": "IRRDBU00",
  "access": "ALTER",
  "profile": "SYS1.PARMLIB",
  "profile_type": "GENERIC",
  "class": "DATASET",
  "effective_via": "GROUPA",
  "inherited": true,
  "risk_weight": 80,
  "explanation": "Group has ALTER access to a sensitive dataset profile"
}
```

Do not create edges without enough provenance to explain where the relationship came from.

## RACF source records to support

Prioritize these IRRDBU00 record types:

```text
0100 Group basic data
0102 Group member/subgroup data
0200 User basic data
0205 User connect data
0400 Dataset profile basic data
0404 Dataset profile access list
0500 General resource profile basic data
0505 General resource access list
```

Additional sources to support:

```text
DSMon APF list
DSMon LINKLIST
DSMon PARMLIB
DSMon PROCLIB
DSMon STARTED table
SETROPTS output
RLIST/LISTDSD evidence where available
SMF-derived evidence where available
```

## Important RACF semantics

Implement RACF semantics carefully. Do not treat all graph edges as equal.

### Group membership

Support:

- Direct user-to-group connections.
- Nested groups.
- Effective group expansion.
- Connect authority.
- Group-SPECIAL, group-OPERATIONS, group-AUDITOR flags from user connect data.
- Default group.
- Revoked connections where visible.

### Dataset access

Support:

- UACC.
- Access list entries.
- READ, UPDATE, CONTROL, ALTER.
- Generic and discrete profiles.
- Generic profile precedence.
- Enhanced generic considerations.
- Sensitive dataset classification.

Access should be interpreted contextually. For example, WRITE/UPDATE/ALTER to an APF-authorized library has much higher risk than READ to a normal application dataset.

### General resources

Support at least:

```text
FACILITY
SURROGAT
STARTED
OPERCMDS
UNIXPRIV
XFACILIT
JESSPOOL
TSOAUTH
SDSF
```

Use the class name and profile name to determine abuse meaning.

### SURROGAT

SURROGAT should be modeled as a control relationship, not just a resource permission.

For example:

```text
USERA --READ to SURROGAT profile--> TARGET.SUBMIT
USERA --RACFCanSubmitAs--> TARGET
```

Support chained analysis where submitting as a target ID leads to privileges, dataset access, or STC control.

### APF write

Model APF write as a high-impact path.

If a user or group can UPDATE/CONTROL/ALTER an APF-authorized library, create a meaningful path toward privileged code execution.

Example:

```text
USERA -> RACFMemberOf -> APPDEV -> RACFCanWrite -> SYS1.APP.LOADLIB -> APF Code Execution
```

### STARTED identities

Model started task identities and their privileges.

Important questions:

- Which user ID does this started task run as?
- What groups is that user connected to?
- Does the STC user have SPECIAL, OPERATIONS, BPX.SUPERUSER, UID(0), or sensitive dataset access?
- Can anyone modify the JCL, PROC, PARMLIB, or datasets that influence this STC?
- Can anyone submit as or otherwise control the STC identity?

## Entity drill-down requirements

RACF entities should have rich detail pages or side panels comparable to BloodHound's AD node experience.

### RACFUser panel

Include tabs or sections for:

- Overview.
- Direct groups.
- Effective groups.
- Default group.
- Direct privileges.
- Inherited privileges.
- Dataset access.
- Resource access.
- SURROGAT relationships.
- Can submit as.
- Can become / can influence.
- Inbound control relationships.
- Shortest paths to high-value targets.
- Raw evidence records.

### RACFGroup panel

Include:

- Overview.
- Direct users.
- Direct subgroups.
- Effective recursive members.
- Parent groups.
- Group privileges.
- Dataset access.
- Resource access.
- Owned objects.
- High-risk members.
- Risk summary.

The group page must support “show all effective members” and allow clicking from member lists into user pages.

### RACFDataset panel

Include:

- Overview.
- Profile name.
- Discrete vs generic.
- Owner.
- UACC.
- WARNING status.
- Access list.
- Effective users and groups.
- Write-capable principals.
- Sensitive classification.
- Matching logic explaining why a profile applies.
- Paths to write access.

### RACFResource panel

Include:

- Class.
- Profile.
- Owner.
- UACC where applicable.
- Access list.
- Effective users and groups.
- Abuse meaning.
- Related paths.
- Raw evidence.

### RACFFinding panel

Include:

- Title.
- Severity.
- Affected entities.
- Evidence.
- Paths.
- Explanation.
- Recommendation.
- False-positive notes.
- Exportable report text.

## Dashboards

Build dashboards around security questions, not just object counts.

Initial dashboards:

```text
Privileged Access Overview
SPECIAL Exposure
OPERATIONS Exposure
AUDITOR / ROAUDIT Exposure
BPX.SUPERUSER Exposure
APF Write Exposure
SURROGAT Abuse Paths
STARTED Identity Risk
Sensitive Dataset Access
PARMLIB / PROCLIB Write Exposure
RACF Hygiene
Attack Path Summary
```

Each dashboard should show:

- Total affected objects.
- Direct vs inherited exposure.
- Top risky users/groups.
- Shortest paths to dangerous targets.
- Recently changed evidence where timestamps exist.
- Links to affected entity panels.
- Links to findings.

Avoid dashboards that only show raw counts without interpretation.

## Automated analysis rules

Implement a rule engine that runs graph queries and emits findings.

Rules should be data-driven where possible, using YAML or JSON.

Example rule shape:

```yaml
id: RACF-APF-WRITE-GROUP
title: Group has write access to APF library
severity: high
category: APF
query: |
  MATCH p = (:RACFGroup)-[:RACFCanWrite]->(:RACFDataset {apf: true})
  RETURN p
explanation: |
  Members of this group may be able to modify APF-authorized code, which can lead to privileged execution on z/OS.
recommendation: |
  Restrict UPDATE, CONTROL, and ALTER access to APF libraries to tightly controlled system programming groups.
evidence_requirements:
  - Dataset profile
  - Access list entry
  - APF classification source
```

Findings should be first-class objects that can be clicked from dashboards and graph paths.

Suggested finding relationships:

```text
(:RACFFinding)-[:RACFAffects]->(:RACFUser)
(:RACFFinding)-[:RACFAffects]->(:RACFGroup)
(:RACFFinding)-[:RACFAffects]->(:RACFDataset)
(:RACFFinding)-[:RACFEvidencedBy]->(:RACFPath)
```

## Initial high-value analysis rules

Prioritize rules that support realistic consulting findings.

Start with:

```text
Direct SPECIAL users
Inherited SPECIAL through group connect
Group-SPECIAL exposure
OPERATIONS exposure
BPX.SUPERUSER exposure
UID(0) exposure
APF library write access
LINKLIST library write access
PARMLIB write access
PROCLIB write access
SURROGAT to privileged user
SURROGAT chain to privileged user
Broad group with sensitive dataset write
Broad group with BPX.SUPERUSER
STC user with excessive privilege
Writable JCL/PROC influencing privileged STC
RACF database dataset access
WARNING mode sensitive profiles
UACC access to sensitive resources
```

## Path explanation requirements

Every high-risk path should be explainable in plain English.

Example:

```text
USR123 can reach APF Code Execution because:
1. USR123 is connected to group APPDEV.
2. APPDEV has ALTER access to SYS1.APP.LOADLIB.
3. SYS1.APP.LOADLIB is APF-authorized.
4. Modifying APF-authorized code can allow privileged execution on z/OS.
```

Another example:

```text
USR456 can reach OPERATIONS because:
1. USR456 has READ access to SURROGAT profile BATCHADM.SUBMIT.
2. This allows USR456 to submit work as BATCHADM.
3. BATCHADM has OPERATIONS.
4. OPERATIONS can allow broad dataset access depending on policy and logging controls.
```

Path explanations should avoid overstating certainty. Use wording like “may allow,” “can provide a path to,” or “depending on local controls” where appropriate.

## UX principles

The user experience should support investigative workflows:

- From dashboard to finding.
- From finding to affected entities.
- From entity to graph path.
- From path to evidence.
- From group to all effective members.
- From member to user details.
- From dataset to write-capable principals.
- From resource to abuse explanation.

Always allow click-through between connected entities.

Lists should distinguish:

```text
Direct members
Effective inherited members
Direct privileges
Inherited privileges
Direct access
Effective access through group nesting
```

Do not hide how an effective result was calculated.

## Data quality and evidence

Prefer explainable, evidence-backed results over clever but opaque inference.

Each derived edge or finding should include:

- Source file or source type.
- Source record type where applicable.
- Profile name.
- Access level.
- Principal.
- Effective-via group where applicable.
- Timestamp where available.
- Confidence where inference is involved.

If required source data is missing, surface that clearly in the UI and findings.

Example:

```text
APF classification unavailable because no DSMon APF input was imported.
```

## Testing requirements

Create small deterministic test fixtures for RACF scenarios.

Test cases should include:

```text
Direct SPECIAL user
SPECIAL via group connect
Nested group membership
Group-SPECIAL connect
SURROGAT submit-as relationship
SURROGAT chain to SPECIAL
APF write via direct user ACL
APF write via group ACL
Generic dataset profile precedence
Discrete dataset profile override
UACC READ/UPDATE behavior
WARNING mode profile
STARTED identity mapping
PARMLIB write path
PROCLIB write path
```

Each test should validate both:

- Correct graph objects/edges are produced.
- Correct findings/path explanations are produced.

Do not rely only on visual testing.

## Coding style instructions for Codex

When modifying the codebase:

1. Inspect the existing project structure before creating new directories.
2. Prefer small, focused changes.
3. Avoid broad rewrites of upstream BloodHound code.
4. Add RACF-specific code behind clear boundaries.
5. Preserve existing behavior for AD/Azure/other BloodHound functionality.
6. Add tests for new parsing, graph modeling, and rule logic.
7. Include meaningful comments for RACF semantics, especially where they differ from AD assumptions.
8. Do not hard-code client-specific names, dataset names, group names, or user IDs except in test fixtures.
9. Keep UI labels precise and RACF-native.
10. Do not imply exploitability beyond what the evidence supports.

## Safety and wording constraints

This tool is for authorized security analysis and defensive review.

Avoid building features that directly automate exploitation. The tool may show risk paths and explain impact, but it should not generate exploit JCL, payloads, backdoors, credential theft routines, or destructive actions.

Use professional language suitable for client reports and conference demos.

Prefer:

```text
This path may allow privilege escalation if the user can modify executable code in an APF-authorized library.
```

Avoid:

```text
Click here to own the LPAR.
```

## First implementation milestone

The first milestone should prove the core workflow:

1. Import a small RACF fixture.
2. Create users, groups, datasets, resources, and privileges.
3. Resolve direct and nested group membership.
4. Show a group detail page with direct and effective members.
5. Show a user detail page with direct and inherited privileges.
6. Detect APF write exposure.
7. Detect SURROGAT to privileged user.
8. Display a dashboard with these findings.
9. Click from dashboard finding to affected entity.
10. Click from affected entity to path explanation and evidence.

Do not start with every RACF class. Start with enough end-to-end functionality to prove the model.

## Suggested demo fixture

Create synthetic data with:

```text
USR001: ordinary user
USR002: user connected to APPDEV
USR003: direct SPECIAL user
BATCHADM: privileged batch user with OPERATIONS
STCADM: started task user
APPDEV: application development group
SYSADM: system administration group
NESTED1: subgroup of APPDEV
SYS1.APP.LOADLIB: APF-authorized library
SYS1.PARMLIB: sensitive PARMLIB dataset
BATCHADM.SUBMIT: SURROGAT profile
```

Demo paths:

```text
USR002 -> APPDEV -> ALTER SYS1.APP.LOADLIB -> APF Code Execution
USR001 -> READ BATCHADM.SUBMIT -> BATCHADM -> OPERATIONS
USR002 -> APPDEV -> NESTED1 -> inherited access example
```

## Acceptance criteria for generated code

A change is acceptable only if:

- It builds successfully.
- Existing BloodHound behavior remains intact.
- RACF-specific code is isolated where practical.
- Tests or fixtures are added for new RACF behavior.
- UI changes support click-through investigation.
- Findings include evidence and explanation.
- No client-sensitive data is embedded.
- Documentation is updated when new RACF concepts or rules are introduced.

For cross-repository changes, acceptance also requires:

- The ownership boundary above is preserved.
- The OpenGraph contract version is explicit.
- Supported BloodHound versions are recorded.
- Provisioning and saved-query installation remain idempotent.
- A golden RACF fixture passes exporter-to-ingest compatibility testing.
- Any fork patch is listed in the customization manifest.

## Documentation and decision log

Keep durable documentation close to the component it describes:

```text
mfpandas-racfhound:
  graph contract, RACF mapping semantics, exporter fixtures

racfhound:
  collection, provisioning, compatibility matrix, operator workflow

BloodHound fork:
  fork-only UI/backend changes, customization manifest, upstream update runbook
```

For each material design decision, record:

```text
date
decision
reason
alternatives considered
affected repositories
contract or migration impact
verification performed
```

Update this file when repository boundaries, compatibility assumptions,
milestones, or non-obvious RACF semantics change. Do not use it as a substitute
for user-facing setup and operator documentation.

## Long-term direction

The long-term goal is a RACF reasoning layer that answers:

```text
Who has dangerous authority?
How did they get it?
Is it direct or inherited?
Which RACF mechanism grants it?
Which source record proves it?
What is the likely impact?
What should be remediated?
```

The UI should make RACF authorization review understandable to both mainframe specialists and non-mainframe security teams.
