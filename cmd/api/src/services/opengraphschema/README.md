# Open Graph Schema Service

The Open Graph Schema Service acts as an intermediary between the API and our database layer, ensuring that all graph schema extensions are valid and comply with internal business logic.

## Validations

### Basic Required Fields

- Extension name is required
- Extension version is required
- Extension namespace is required
- At least one node kind is required

### Node Kinds Validations

- Each node kind name must have the extension namespace as a prefix (e.g., ns_User)
- No duplicate node kind names

### Relationship Kinds Validations

- Each relationship kind name must have the extension namespace as a prefix
- No duplicate relationship kind names
- Relationship kind names cannot collide with node kind names (no name can be both)

### Properties Validations

- No duplicate property names

### Environments Validations

- Environment kind name must have the extension namespace as a prefix
- Environment kind name must reference a declared node kind
- Environment source kind name cannot be empty
- Environment source kind name must NOT be a declared node kind
- Environment source kind name must NOT be a declared relationship kind
- Each principal kind name must have the extension namespace as a prefix
- Each principal kind name must reference a declared node kind

### Relationship Findings Validations

- Relationship finding name must have the extension namespace as a prefix
- Relationship finding environment kind name must have the extension namespace as a prefix
- Relationship finding relationship kind name must have the extension namespace as a prefix
- Relationship finding environment kind name must reference a declared node kind
- Relationship finding relationship kind name must reference a declared relationship kind
- Relationship finding source kind name cannot be empty
- Relationship finding source kind name must NOT be a declared node kind
- Relationship finding source kind name must NOT be a declared relationship kind
