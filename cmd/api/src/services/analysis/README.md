# Analysis Service

The Analysis Service owns the domain logic for the BloodHound analysis pipeline. It is the home for analysis orchestration — queueing runs, deciding when and how the pipeline should execute, and managing the state around it. Today the service handles request queueing and the precedence rules between competing requests; over time more orchestration logic will move out of the datapipe daemon and the database layer into this package.

## Queueing analysis requests

The `analysis_request` table holds a single row at a time, so each new request is reconciled against whatever is already queued. `reconcile()` in `reconcile.go` applies the following rules:

| Already queued   | Incoming      | Outcome                                                          |
| ---------------- | ------------- | ---------------------------------------------------------------- |
| nothing          | any           | Write the incoming request.                                      |
| deletion         | any           | Drop the incoming request; a queued deletion must run first.     |
| analysis         | deletion      | Replace the queued analysis with the deletion.                   |
| partial analysis | full analysis | Add the incoming steps to the queued request.                    |
| analysis         | analysis      | Drop the incoming request if the queued scope already covers it. |
