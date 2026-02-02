---
bh-rfc: 6
title: Dogtags - SKU-Based Feature Entitlements
authors: |
    [Pomeroy, Kaleb](kpomeroy@specterops.io)
status: DRAFT
created: 2025-01-07
---

# Dogtags - SKU-Based Feature Entitlements

## 1. Overview

Dogtags is a provider-based system for managing feature entitlements. It abstracts where SKU level configuration comes from, allowing the same application code to work whether entitlements are sourced from a license file, a database, or some future backend we haven't thought of yet.

## 2. Motivation & Goals

BloodHound supports multiple deployment models with different configuration needs. Some deployments read entitlements from signed license files, others from a database. The application shouldn't care where the values come from.

-   **Abstraction** - Decouple feature checks from configuration source
-   **Extensibility** - Add new providers without touching application code
-   **Simplicity** - One interface for all entitlement lookups

## 3. Considerations

### 3.1 Provider Injection

Providers are injected at runtime. The application receives a configured provider at startup and uses it for all entitlement lookups. Provider implementation details are intentionally hidden from the application. The FOSS application uses a no-op provider, meaning default values in the SKU lookup field will apply.

### 3.2 Startup-Time Loading

Entitlements are loaded once at startup and cached. Changes require a restart. This is intentional; Runtime updates add complexity and create less predictable behavior.

### 3.3 Fail-Fast

If a provider can't load its configuration, the application fails to start. Silent fallbacks mask misconfigurations and make debugging painful.

### 3.4 Public Interfaces, Private Implementations

The dogtags interfaces and service layer live in this public repository, but the feature itself is enterprise-only. This is a deliberate pattern: CE ships with a `NoopProvider` that returns defaults, effectively disabling SKU-gated functionality. Enterprise builds inject real providers that source actual entitlements.

This lets us maintain a single codebase where enterprise features are structurally present but functionally inert in CE. Application code doesn't branch on "is this enterprise?" - it just asks dogtags for values and gets sensible defaults in CE.

## 4. Architecture

### 4.1 Provider Interface

Providers implement a simple interface:

```go
type Provider interface {
    GetFlagAsBool(key string) (bool, error)
    GetFlagAsString(key string) (string, error)
    GetFlagAsInt(key string) (int64, error)
}
```

Providers return errors when a key isn't found. This lets the service layer decide whether to use defaults.

### 4.2 Service Layer

The service wraps providers and handles defaults. If a provider doesn't have a flag, the service falls back to the default defined in the SKU spec.

```go
type Service interface {
    GetFlagAsBool(key BoolDogTag) bool
    GetFlagAsString(key StringDogTag) string
    GetFlagAsInt(key IntDogTag) int64
    GetAllDogTags() map[string]any
}
```

Note the typed keys (`BoolDogTag`, `IntDogTag`, etc.) - this prevents mixing up flag types at compile time.

### 4.3 API Endpoint

Dogtags are exposed via `GET /api/v2/dog-tags`. The response is always a flat map (no nested objects). Keys use dot notation for namespacing (`namespace.key`), but this is purely a naming convention, not structure.

```json
{
    "data": {
        "auth.environment_targeted_access_control": false,
        "privilege_zones.tier_limit": 3,
        "privilege_zones.label_limit": 10,
        "privilege_zones.multi_tier_analysis": false
    }
}
```

## 5. Adding a New SKU Flag

### 5.1 Define the Flag

Add to `bhce/cmd/api/src/services/dogtags/sku_flags.go`:

```go
const (
    MY_NEW_FEATURE BoolDogTag = "my_feature.enabled"
)

var AllBoolDogTags = map[BoolDogTag]BoolDogTagSpec{
    MY_NEW_FEATURE: {Description: "My New Feature", Default: false},
}
```

Pick a sensible default. This is what the service returns if a provider doesn't have the flag.

### 5.2 Configure Providers

Every provider must be updated to source the new flag's value. Consult provider-specific documentation for configuration details. Missing this step means some deployments will silently fall back to defaults.

### 5.3 Use It

```go
if s.DogTags.GetFlagAsBool(dogtags.MY_NEW_FEATURE) {
    // feature enabled
}
```

The service handles defaults, so you never need to check errors.

### 5.4 Migrating Existing Flags

When moving a configuration value to dogtags, all previous references must be deprecated and removed. The old configuration endpoint, any backend code reading from the parameters table, and any UI components consuming the old values should be updated to use the dogtags service instead. Leaving stale references creates a confusing split where the displayed value may not match what the application actually uses.

## 6. Testing

For tests that don't care about specific flag values, use `NewDefaultService()` which returns defaults for everything:

```go
dogtagsService := dogtags.NewDefaultService()
```

For tests that need specific flag values, use `NewTestService` with `TestOverrides`. Values not in the overrides map fall back to defaults:

```go
dogtagsService := dogtags.NewTestService(dogtags.TestOverrides{
    Bools: map[dogtags.BoolDogTag]bool{dogtags.PZ_MULTI_TIER_ANALYSIS: true},
    Ints:  map[dogtags.IntDogTag]int64{dogtags.PZ_TIER_LIMIT: 5, dogtags.PZ_LABEL_LIMIT: 3},
})
```

## 7. File Locations

| File                                             | Purpose                              |
| ------------------------------------------------ | ------------------------------------ |
| `bhce/cmd/api/src/services/dogtags/sku_flags.go` | Flag definitions and defaults        |
| `bhce/cmd/api/src/services/dogtags/service.go`   | Service interface and implementation |
| `bhce/cmd/api/src/services/dogtags/provider.go`  | Provider interface and NoopProvider  |
