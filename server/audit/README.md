# Audit logging

The `audit` module records the two rows that form an audit event:

1. An **intent** row before a mutating operation begins.
2. A **success** or **failure** outcome row after it finishes.

Both rows have the same generated commit ID. The module obtains request, source-IP,
and actor information from the request context. If no valid identity is present, it
records an anonymous actor, matching existing behavior.

## Lifecycle and transactions

The lifecycle intentionally matches `AuditableTransaction` in db layer per onion architecture:

```text
audit.Begin       -> inserts the intent row
service operation -> runs its own database transaction, when needed
audit.Complete    -> inserts the success or failure outcome row
```

The audit rows are not part of the service operation's transaction. This lets a
failed operation receive a durable failure audit record. An intent-write error must
prevent the operation from running. An outcome-write error is returned to the caller,
even when the operation itself has already completed; this preserves the legacy error
precedence.

## Wiring a module

`modules.Register` creates one `audit.Service` using the process's identity resolver
and supplies it to feature-module `Register` functions. A feature module should pass
that dependency to its service constructor:

```go
func Register(routerInst *router.Router, pool *pgxpool.Pool, auditor audit.Service) {
	var (
		store = appdb.NewStore(pool)
		svc   = services.NewService(store, auditor)
	)

	// Register routes with svc.
}
```

The feature service defines the narrow interface it consumes, rather than importing
the audit implementation internals:

```go
type Auditor interface {
	Begin(ctx context.Context, entry audit.Entry) (uuid.UUID, error)
	Complete(ctx context.Context, commitID uuid.UUID, entry audit.Entry, operationErr error) error
}
```

## Auditing a mutating service operation

Build the audit entry from the requested change, write the intent, run the database
operation, and always attempt to write its outcome after a successful `Begin`:

```go
func (s *Service) UpdateThing(ctx context.Context, thing Thing) error {
	var (
		entry = audit.Entry{
			Action: model.AuditLogActionUpdateThing,
			Fields: map[string]any{
				"thing_id": thing.ID,
				"enabled":  thing.Enabled,
			},
		}
		commitID     uuid.UUID
		operationErr error
		err          error
	)

	if commitID, err = s.auditor.Begin(ctx, entry); err != nil {
		return fmt.Errorf("writing update thing audit intent: %w", err)
	}

	operationErr = s.db.UpdateThing(ctx, thing)
	if err = s.auditor.Complete(ctx, commitID, entry, operationErr); err != nil {
		return fmt.Errorf("writing update thing audit outcome: %w", err)
	}

	return operationErr
}
```

The database adapter owns the transaction required by `UpdateThing`; it should not
insert audit rows itself. The service owns the audit lifecycle because it is the only
layer that can correctly describe the business operation and its outcome.
