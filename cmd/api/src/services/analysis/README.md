# Analysis Service

The single front door for everything analysis-related in BloodHound:
queueing analysis runs, reading what's currently queued, cancelling
pending runs, and — over time — deciding when and how analysis runs.

## What "analysis" means here

When BloodHound has new graph data (a fresh AD ingest, a user editing an
asset group tag selector, an admin deleting a tenant), it doesn't yet know
what's exposed, who has Tier Zero access, or which findings are stale.
**Analysis** is the pipeline that figures that out: AD/Azure
post-processing, tagging nodes against asset-group rules, and generating
the findings the UI shows.

The pipeline is expensive, so it doesn't run on every change. Instead,
callers **queue an analysis request** — a row in the `analysis_request`
table — and the datapipe daemon picks it up on its next tick and runs the
pipeline.

This service owns the "queue an analysis request" flow and, over time,
will own more of the orchestration around it.

## Why a service exists at all

Before this service, every part of the codebase that wanted to trigger
analysis called the database directly (`db.RequestAnalysis(...)`). That
worked, but had two real problems:

1. **Precedence rules were scattered.** If a partial analysis was already
   queued and a full one came in, what should happen? The right answer
   ("the full one absorbs the partial") needs to live in *one* place —
   not get redecided at every call site, where someone will eventually
   get it wrong.

2. **Handler tests had to mock the whole database.** A handler whose only
   job is "queue an analysis when X happens" shouldn't need a
   200-method database mock to test that one behavior.

The service fixes both. Callers say `s.Analysis.RequestFullAnalysis(ctx, userId)`.
The service worries about *how* that's stored, *what* precedence rules
apply, and *how* it interacts with anything already queued. Tests mock
one small interface instead of the whole database.

**Mental model:** think of a restaurant. A customer says "I'd like the
salmon." The waiter doesn't hand them a knife and point at the kitchen.
The waiter takes the order, checks whether the kitchen already has a
salmon order in progress, and handles the rest. The service is the
waiter.

## How a request flows through

```text
Caller (handler, daemon, etc.)
     │
     │  s.Analysis.RequestFullAnalysis(ctx, userId)
     ▼
┌──────────────────────────────────────────┐
│ Service method picks the right step:           │
│   RequestFullAnalysis    → AnalysisStepAll     │
│   RequestAnalysisForAGT  → partial or full     │
│                             (depends on flag)  │
│   RequestGraphDataDeletion → deletion request  │
└────────────────────────────────────────────────┘
     │
     ▼
┌──────────────────────────────────────────┐
│ apply() reads any request already        │
│ queued, calls merge() to reconcile,      │
│ writes the result.                       │
└──────────────────────────────────────────┘
     │
     ▼
Row sits in analysis_request table.
Datapipe daemon ticks, picks it up, runs the pipeline.
```

## Precedence rules

Many things can queue a request at the same time — the scheduler, an
admin clicking "Run Analysis," an asset-group-tag mutation, a graph data
deletion. Only one row exists at a time, so a new request has to be
reconciled with whatever's there.

`merge()` in `merge.go` is the pure function that encodes those rules.
No side effects, no dependencies, trivial to unit test:

| Already queued | Incoming     | Outcome                                          |
| -------------- | ------------ | ------------------------------------------------ |
| nothing        | analysis     | write incoming                                   |
| nothing        | deletion     | write incoming                                   |
| deletion       | *anything*   | drop incoming — deletion is sticky               |
| analysis       | deletion     | replace with deletion                            |
| analysis (A)   | analysis (B) | OR the step bitmasks; write only if widened      |

**The bitmask-OR rule is the interesting one.** Every analysis request
carries a `model.AnalysisStep` bitmask telling the pipeline which stages
to run (post-processing, tagging, findings generation). When a queued
*partial* request collides with an incoming *full* one, the queued row's
bitmask gets widened so the full pipeline still runs. Without this rule,
an asset-group-tag change immediately followed by a user clicking "Run
Analysis" would silently skip post-processing. That's the bug class this
service exists to prevent.

## Feature gating

`RequestAnalysisForAGT` is the only method whose behavior depends on
configuration. When the `agt_partial_analysis` feature flag is **on**,
AGT mutations queue a *partial* analysis (tagging through findings,
skipping AD/Azure post-processing). When it's **off**, AGT mutations
queue a full analysis like everything else.

The flag is `user_updatable=false` — an internal kill-switch, not a
customer-facing toggle. If partial analysis misbehaves in production,
an operator flips the flag and the system reverts to full analysis on
every AGT change without a code deploy.

## What lives here today

```text
services/analysis/
├── README.md       ← you are here
├── request.go      ← Service interface, struct, all request-queueing methods
├── merge.go        ← pure precedence rules (the table above)
└── mocks/          ← generated mock for handler tests
```

## What's planned to land here

This package is named `analysis` (not `analysisrequest`) because it's
intended as the home for analysis-domain orchestration as it gets pulled
out of the datapipe daemon and various scattered helpers. The plan, in
rough priority order:

### 1. `flags.go` — flag passthrough helpers

Today every caller that needs to know "is partial analysis enabled?"
imports `appcfg` and calls a free function. That couples every caller to
the flag system. Methods like `s.Analysis.IsPartialEnabled(ctx)`
consolidate flag access here. Thin today; the right home if a flag's
semantics ever get richer (e.g., "partial enabled, but only when
scheduled analysis is off").

### 2. `dispatcher.go` — "should I run, and at what step?"

The datapipe daemon currently makes this decision inline, mixing several
concerns: filter out deletion rows (they're not analysis), check whether
ingest jobs are still pending, decide whether to upgrade a partial to
full. That logic belongs here, not in the daemon. The daemon should ask
the service "what should I do?" and the service should give a clear
answer.

This is where the open bug fixes go:

- **Filter deletion rows** so they don't accidentally trigger analysis
  code.
- **Upgrade partial → full** when ingest jobs are still waiting, since
  partial analysis on incomplete data is misleading.

### 3. `bookkeeping.go` — post-run state

After analysis completes, several things need to happen: update the
last-analysis timestamp, refresh data quality metrics, mark ingest jobs
complete. Today this lives in `daemons/datapipe/pipeline.go`. Moving it
here keeps the datapipe daemon focused on *running* the pipeline and the
service focused on *managing the state around* the pipeline.

### 4. `status.go` — read-side queries

Things like "is analysis running right now?", "when did it last
complete?", "what's the current step?" — for API endpoints and the UI.
Today these are scattered free functions and ad-hoc database queries.
Centralizing them gives the UI a stable contract and lets us cache
cheaply if it ever becomes hot.

## Architectural principles for future contributors

A few rules to follow as this package grows:

**Keep the package flat.** Add `dispatcher.go`, `bookkeeping.go`, etc.
as siblings to `request.go`. Don't create sub-packages like
`analysis/request/` or `analysis/dispatcher/`. In Go, nested packages
add real cost (more interfaces, more mocks, harder navigation) and pay
off only when a sub-concern has independent dependencies, a different
lifecycle, or could plausibly become its own service someday. Analysis
sub-concerns share the database, the same request model, and the same
test fixtures — so they belong together.

**Promote to `internal/` only when complexity demands it.** If `merge`
ever grows from 30 lines into a 300-line precedence engine with its own
concepts, *then* it earns a home like `analysis/internal/precedence/`.
Until then it's a file. The trigger is "this concept has gotten complex
enough that I want to test it without standing up the whole service" —
not "this concept is topically distinct."

**Methods can be passthroughs.** It's fine to add `IsPartialEnabled(ctx)`
as a one-line wrapper around `appcfg.GetAGTPartialAnalysisEnabled`. The
value isn't in the wrapper — it's in giving future logic *one place to
land*. The day someone needs that decision to depend on three flags, a
config value, and the current time, they'll be glad the entry point
already exists.

**The public surface is the `Service` interface; the struct is
unexported.** Callers depend on the interface, not the concrete type.
This is what makes the generated mock work and what lets you swap
implementations (e.g., in-memory for tests) without consumers caring.

## Quick reference

```go
// Queue a full analysis (PUT /api/v2/analysis, scheduler, startup init).
err := s.Analysis.RequestFullAnalysis(ctx, userId)

// Queue an AGT-triggered analysis — partial or full depending on the flag.
err := s.Analysis.RequestAnalysisForAGT(ctx, userId)

// Queue a graph data deletion (overrides any pending analysis).
err := s.Analysis.RequestGraphDataDeletion(ctx, request)

// Read whatever's currently queued (returns ErrNotFound if nothing).
req, err := s.Analysis.GetAnalysisRequest(ctx)

// Clear the queue (e.g. user-initiated cancel).
err := s.Analysis.DeleteAnalysisRequest(ctx)
```

Construction:

```go
analysisService := analysis.NewAnalysisService(db)
```

The service is wired into `v2.Resources` and `ResourcesDeprecated` as
the `Analysis` field, so handlers reach it as `s.Analysis.*` and never
need to import the package directly.
