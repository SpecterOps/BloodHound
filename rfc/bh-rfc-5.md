---
bh-rfc: 5
title: Layered Architecture for API Separation of Concerns
authors: |
    BloodHound Engineering Team
status: DRAFT
created: 2025-12-15
audiences: |
    Backend Engineers
    API Developers
---

# Layered Architecture for API Separation of Concerns

```mermaid
flowchart LR
    subgraph View["View Layer"]
        H[Handlers]
        V[Views]
        MW[Middleware]
    end

    subgraph Service["Service Layer"]
        BL[Business Logic]
        INT[Interfaces]
    end

    subgraph Data["Data Layer"]
        DB[(Database)]
        Graph[(Graph DB)]
        FS[(Filesystem)]
    end

    subgraph Models["Shared Models"]
        AM[Application Models]
        SE[Sentinel Errors]
    end

    View -->|App Models| Service
    Service -->|App Models| Data
    Data -->|App Models| Service
    Service -->|App Models| View

    View -.->|uses| Models
    Service -.->|uses| Models
    Data -.->|uses| Models
```

## 1. Overview

This RFC defines a layered architecture for BloodHound APIs that establishes clear separation of concerns between the data layer, service layer, and view layer. It provides guidelines for model usage, error handling, and dependency management to reduce architectural debt and improve maintainability.

## 2. Motivation & Goals

Our current API layers require updating to establish better-defined boundaries and appropriate separation of concerns. There is significant architectural debt that has been tracked for years, and this document provides guidance on addressing that debt in our new APIs.

- **Separation of Concerns** - Establish clear boundaries between data, service, and view layers to prevent implementation details from leaking across layers.
- **Model Independence** - Move away from a single shared model to layer-specific models that protect API contracts from unintended drift.
- **Error Handling** - Define consistent error handling patterns that prevent external dependency error types from crossing layer boundaries.
- **Dependency Isolation** - Enable incremental refactoring by ensuring layers depend on interfaces rather than concrete implementations.
- **Maintainability** - Create an architecture that engineers want to maintain, following the "no broken windows" philosophy.

## 3. Considerations

### 3.1 Guiding Philosophy

No broken windows. If something bothers us about the current way things are done (e.g., API filtering logic), we should address it. Part of defining a clearer separation of concerns is to bring it in line with what engineers would actually want to maintain. YAGNI still applies—start with the simplest solution and let it grow organically, abstracting when necessary rather than before the lack of abstraction causes actual pain.

### 3.2 Impact on Existing Systems

This proposal affects the existing database package, graph querying packages (including dawgs), API handlers, and shared model definitions. Migration should be incremental, implementing one service's interface at a time rather than requiring a complete refactor in one step.

### 3.3 Implementation Plan

Refactors should be doable without changing services beyond which data layer is injected to a specific service. By interfacing dependencies, underlying data layers or libraries can be swapped out incrementally as long as they provide the same interface that the service needs.

## 4. Models

### 4.1 Layer-Specific Models

Each layer should own its own model types:

- **Database Layer** - Owns structs that represent the current state of database entities.
- **View Layer** - Owns view structs that represent the API contract, ensuring stability despite changes to the database or underlying services.
- **Application-wide Models** - Shared models used for transferring data in a standard way between layer boundaries, devoid of database or JSON tags.

### 4.2 Model Translation Flow

```mermaid
flowchart LR
    HTTP["HTTP"]
    ViewLayer["View Layer"]
    ServiceLayer["Service Layer"]
    DataLayer["Data Layer"]
    
    HTTP -->|"requests"| ViewLayer
    ViewLayer -->|"translate and call"| ServiceLayer
    ServiceLayer -->|"validate, operate, and call"| DataLayer
    DataLayer -->|"translate and return"| ServiceLayer
    ServiceLayer -->|"handle and return"| ViewLayer
    ViewLayer -->|"translate and return"| HTTP
```

### 4.3 Benefits

This approach allows each layer to be updated without significant changes in other layers. It prevents unintended contract drift by requiring explicit translation between models at layer boundaries.

## 5. Error Types

### 5.1 Error Boundary Rules

Public error types must not cross boundaries. Each layer should:

1. Convert errors to strings at the public boundary.
2. Wrap error strings with layer-specific sentinel errors before returning.
3. Avoid leaking error types of dependencies of the layer to prevent implicit contracts with higher layers.

### 5.2 Example

```go
var ErrNotFound = errors.New("not found")
...
err := gorm.ThingThatErrors() // This returns a gorm.ErrNotFound
if err != nil {
  // Note that the error is converted to a string, then wrapped with an appropriate sentinel
  return fmt.Errorf("%w: %s", ErrNotFound, err)
}
...
```

```mermaid
flowchart LR
    A["subdependency"]-->|"returns subdependency sentinel"| B["private method"]
    B-->|"runs some error handling, then returns, wrapping as normal"| C["public method"]
    C-->|"formats error as string, wraps with package level sentinel"| D["external layer"]
```

### 5.3 Sentinel Error Registry

Error types are part of a package's implicit API contract. Layers should either define their own sentinels or use a shared application registry of sentinel errors (a separate errors package that all layers import).

## 6. Data Layer

### 6.1 Responsibilities

The data layer abstracts access to data away from the service layer. It focuses on data access with limited to no business logic present. Current components include the database package, graph querying packages (including dawgs), and filesystem access.

### 6.2 Design Principles

- Packages should provide methods, not their own interfaces.
- Can be monolithic like the existing DB struct, or broken up along useful boundaries.
- Methods should always take and return application-wide types (typically declared in models).
- Internal conversion between application types and data layer-specific types should occur within the layer.

### 6.3 Transaction Handling

Services may need transaction objects that can be injected into additional database calls. The underlying transaction should remain private to the data layer. This allows the underlying database implementation to change without affecting the service layer.

## 7. Service Layer

### 7.1 Responsibilities

The service layer is where all business logic should be defined and wire-up for data access occurs. Services communicate with other layers using shared models and shared errors.

### 7.2 Dependency Injection

Services must be wired up during application bootstrap by injecting dependencies. Dependencies should be accepted using interfaces that the service controls. These interfaces should only cover the methods the service actually uses, not all methods the dependency provides.

```mermaid
flowchart TB
    subgraph Bootstrap["Application Bootstrap"]
        DL[Data Layer Implementation]
    end

    subgraph Service["Service Package"]
        SI[Service Interface<br/>Only required methods]
        SVC[Service]
    end

    DL -->|"inject"| SI
    SI -->|"uses"| SVC

    subgraph Testing["Testing"]
        MOCK[Generated Mock<br/>Per-service]
    end

    SI -.->|"mockgen"| MOCK
```

### 7.3 Benefits

- Services keep smaller interfaces on the consumer side where they belong.
- Mocks will be per-service instead of global.
- Services no longer need to import multiple mock packages.
- Service packages will need to set up mockgen for their interfaces.

## 8. View Layer

### 8.1 Responsibilities

The view layer interfaces with the outside world. Currently, the primary view layer consists of versioned API handlers, views, and middleware packages.

### 8.2 Handler Behavior

Handlers should:

1. Have one or more services injected for business logic.
2. Take HTTP request information and perform basic validations.
3. Negotiate HTTP errors as needed.
4. Convert HTTP request data into application types.
5. Call service methods to process data.
6. Convert resulting application data into view forms.

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Data

    Client->>Handler: HTTP Request
    Handler->>Handler: Basic Validation
    Handler->>Handler: Convert to App Model
    Handler->>Service: App Model
    Service->>Service: Business Logic
    Service->>Data: App Model
    Data->>Data: Convert to DB Model
    Data->>Data: Query/Store
    Data->>Data: Convert to App Model
    Data->>Service: App Model
    Service->>Handler: App Model
    Handler->>Handler: Convert to View Model
    Handler->>Client: HTTP Response (JSON)
```

### 8.3 Views

Views are types that map application models into externally formatted types. This typically means structs with JSON tags, but could include methods for CSV views or other transformations. This ensures most changes to application models will not leak into API contracts.

### 8.4 Future Extensibility

The separated view layer supports multiple API versions and could accommodate additional views such as HTMX or Wails in the future.
