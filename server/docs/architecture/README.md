# BHCE Server Architecture Documentation

This directory contains LikeC4 architecture diagrams for the BHCE API Server feature module architecture.

## Quick Start

```bash
# View interactive diagrams
cd bhce/server/docs/architecture
npx -y likec4@latest serve

# Export to images
npx -y likec4@latest export png -o ./diagrams
```

## File Structure

- **`specification.c4`** – Defines element types (person, system, container, goPackage), relationship types (uses, calls, imports), and tags (external, example, test)
- **`model.c4`** – Declares all architectural elements and their relationships across all C4 levels
- **`views.c4`** – Defines specific views (projections) of the model for different audiences

## Available Views

### 1. System Context (`index`)
**Audience:** Any stakeholder
**Shows:** BHCE as a black box, its users (operators, analysts), and external systems (AD/Azure tenants)

### 2. Containers (`containers`)
**Audience:** Developers and architects
**Shows:** Deployable units inside BHCE: web UI, API server, PostgreSQL, Neo4j

### 3. API Server Components (`apiServerComponents`)
**Audience:** Go developers
**Shows:** Go packages inside the API server: wireup, modules, features
**Key insight:** Features are self-contained vertical slices registered through the module system

### 4. Analysis Internals (`analysisInternals`)
**Audience:** Feature developers
**Shows:** Four-layer architecture inside the analysis feature with call relationships
**Layers:** handlers → services → appdb → postgres
**Key insight:** Each layer defines the interface it needs from the layer below (Dependency Inversion)
**Focus:** Runtime call relationships between layers

### 5. Type Imports (`analysisTypeImports`)
**Audience:** Developers learning the layered pattern
**Shows:** ONLY Go package import relationships (not runtime calls)
**Key insights:**
- handlers imports `services.RequestedAnalysis` (domain type) and `services.ErrNoPendingRequest` (for error comparison)
- appdb imports `services.RequestedAnalysis` (return type) and `services.ErrNotFound` (sentinel error)
- services defines both interfaces and domain types but imports nothing from other layers
- This enables independent testing at each layer
**Focus:** Compile-time import dependencies showing dependency inversion

### 6. GET Request Flow (`analysisGetFlow`)
**Audience:** Developers debugging GET requests
**Shows:** Step-by-step execution of GET /api/v2/analysis
**Includes:**
- HTTP route: GET /api/v2/analysis
- Type transformations: PostgreSQL row → analysisRequest → services.RequestedAnalysis → handlers.RequestedAnalysisView → JSON
- Error mapping: pgx.ErrNoRows → services.ErrNotFound → services.ErrNoPendingRequest → HTTP 204
**Focus:** Read operations and error handling

### 7. PUT Request Flow (`analysisPutFlow`)
**Audience:** Developers debugging write operations
**Shows:** Step-by-step execution of PUT /api/v2/analysis
**Includes:**
- HTTP route: PUT /api/v2/analysis
- Authentication flow
- Concurrency handling via INSERT ... ON CONFLICT (singleton) DO NOTHING
- Type transformations: User ID → SQL → services.RequestedAnalysis → handlers.RequestedAnalysisView → JSON
- Status codes: 202 Accepted (created) vs 200 OK (already exists)
**Focus:** Write operations, idempotency, and concurrency

### 8. Shared Database Access (`sharedDatabaseAccess`)
**Audience:** Architects designing feature boundaries
**Shows:** How multiple features access the same database tables independently
**Key insights:**
- Table "ownership" is logical, not enforced
- Features share the database schema via SQL, not Go packages
- Each feature defines its own minimal Database interface
- Both features wrap the same pgxpool.Pool but remain decoupled

### 9. Module Registration (`moduleRegistrationFlow`)
**Audience:** Developers adding new features
**Shows:** Server startup sequence and module wireup
**Key insights:**
- modules.RegisterAll(deps) calls each Module function
- Each feature builds its own store → service → handler chain
- Adding a feature = append to modules.all slice

## Adding a New Feature

See the main [README.md](../../README.md#adding-a-new-feature-module) for step-by-step instructions.

To add the feature to these diagrams:

1. **Add to `model.c4`:**
   ```c4
   myFeature = goPackage 'myfeature' {
     description '...'
     technology 'server/myfeature'
     
     myFeatureHandlers = goPackage 'handlers' { ... }
     myFeatureServices = goPackage 'services' { ... }
     myFeatureAppdb = goPackage 'appdb' { ... }
   }
   ```

2. **Add relationships in `model.c4`:**
   ```c4
   bhe.apiServer.modulesPkg -[registers]-> bhe.apiServer.myFeature
   bhe.apiServer.myFeature.myFeatureHandlers -> bhe.apiServer.myFeature.myFeatureServices
   bhe.apiServer.myFeature.myFeatureServices -> bhe.apiServer.myFeature.myFeatureAppdb
   bhe.apiServer.myFeature.myFeatureAppdb -> bhe.postgres
   ```

3. **Views automatically include the feature** in apiServerComponents

No need to modify `specification.c4` unless you're adding new element types or relationship kinds.

## Relationship Types

- **uses** – General dependency
- **calls** – Function/method invocation
- **imports** – Go package import (shows type dependencies)
- **registers** – Module registration at startup
- **reads/writes** – Data access

## Tags

- **#external** – Systems outside BHE's boundary
- **#example** – Hypothetical elements for documentation
- **#generated** – Generated code (mocks, stubs)
- **#test** – Test-only elements

## C4 Model Levels

- **Level 1 (System Context):** People and systems
- **Level 2 (Containers):** Deployable units (processes, databases)
- **Level 3 (Components):** Go packages within containers
- **Level 4 (Internal):** Internal structure within a component

## Further Reading

- [C4 Model](https://c4model.com/)
- [LikeC4 Documentation](https://likec4.dev/)
- [Server README](../../README.md)
