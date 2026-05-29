# task-0146 — Implementer Report

**Task**: Orun Cockpit TUI Phase 2 — Plan Studio
**Branch**: `impl/task-0146-plan-studio`
**PR**: https://github.com/sourceplane/orun/pull/145
**Base commit**: `c710bf5`

## What shipped

1. **`internal/tui/services/plan_service.go`** (new, ~340 LOC)
   - `LiveOrunService.GeneratePlan(ctx, PlanRequest) (*PlanResult, error)`
   - Mirrors `cmd/orun/main.go:generatePlan` exactly via internal packages:
     `loader.LoadResolvedIntent` → `LoadCompositionsForIntent` → preset
     resolution → `normalize.NormalizeIntent` →
     `composition.ValidateAllComponents` → `trigger.ValidateProfileRules` /
     `ValidateTriggerContext` / `ResolveActiveEnvironments` →
     `expand.NewExpander` → env/component/changed-only filtering →
     `planner.NewJobPlanner.PlanJobs` →
     `planner.ResolvePromotionDependencies` → `planner.NewJobGraph` topo
     sort → `render.NewRenderer.RenderPlanWithOrder` → optional
     `state.Store.SavePlan` when `NamedPlan` is set.
   - Honours `ctx.Err()` at every stage boundary.
   - **Never imports `cmd/orun`** — duplicates the needed helpers
     (`changedFilesFromGit`, `changedComponentsFromFiles`, path utils)
     inside `internal/tui/services`. Comment links back to the canonical
     CLI implementation; consolidation is a Phase 3 refactor.

2. **`internal/tui/views/plan_studio.go`** (rewritten, ~290 LOC)
   - State machine: `Idle → Configuring → Generating → Review → Saved → Error`.
   - Cursor navigation across the rendered DAG.
   - Local keymap: `g`=generate, `s`=save, `c`=clear, `j/k`=cursor.
   - Deterministic `View()`; jobs rendered with composition, env, deps.
   - View NEVER imports `LiveOrunService` directly — generate commands
     flow through `views.GeneratePlanCmd(svc, req)`, dispatched by the
     root model. Keeps the view trivially testable with mock messages.

3. **`internal/tui/model.go`** (extended ~60 LOC)
   - Routes `services.PlanGeneratedMsg` and
     `views.PlanStudioSaveRequestedMsg` to the Plan Studio view.
   - Seeds the staged `PlanRequest` from the workspace snapshot once it
     loads.
   - Mode-switch shortcuts: `p` (Plan Studio), `b` (Browse), `h` (History).
   - Save reuses `GeneratePlan` with `NamedPlan` set — guarantees the
     persisted plan is byte-identical to the reviewed checksum.

4. **`internal/tui/services/live_service.go`** — removed the Phase 1
   `errNotImplemented` stub for `GeneratePlan`.

## Tests added

- `internal/tui/services/plan_service_test.go` — 4 tests covering
  context cancellation, missing intent, malformed YAML, and
  request-vs-config precedence.
- `internal/tui/views/plan_studio_test.go` — 10 unit tests covering all
  state transitions, cursor clamping, save dispatch, and rendering,
  plus **two `pgregory.net/rapid` property tests**:
  1. `TestPlanStudioModel_PropertyValidTransitions` — random action
     sequences never escape the declared state set; cursor stays in
     `[0, len(jobs))`; `Review`/`Saved` states always have a non-nil
     Result.
  2. `TestPlanStudioModel_PropertyDeterministicView` — repeated
     `PlanGeneratedMsg` with the same input yields identical view output.

All pre-existing Phase 1 tests remain green.

## Verification

```
go build ./...                                    # ✓
go test ./...                                     # ✓ (25+ packages)
go test ./internal/tui/... -count=1               # ✓
go run ./cmd/orun validate -i examples/intent.yaml   # ✓
go run ./cmd/orun plan -i examples/intent.yaml       # ✓
   → 15 components × 5 envs → 38 jobs, plan ad8f82c030a5 (unchanged)
```

## go.mod changes

- Promoted to **direct** dependencies (were already in go.sum from
  Phase 1): `bubbletea v1.3.5`, `bubbles v0.21.0`, `lipgloss v1.1.0`,
  `pgregory.net/rapid v1.1.0`.
- **Did NOT** reintroduce stale `github.com/flyingmutant/rapid` (per
  task-0146 §43).

## Explicit scope gaps (documented, not bugs)

1. **ChangedOnly safe subset**: explicit Base/Head + component-path
   matching only. No CI auto-detect, no semantic intent diff, no
   `--uncommitted`/`--untracked`. Surfaced at runtime via
   `PlanResult.Warnings`. Phase 2.1 follow-up.
2. **`RunPlan`, `Describe`, follow-mode `TailLogs`, `RemoteState.ListRuns`**
   remain `errNotImplemented` (Phase 3 / task-0147+).
3. **Spec drift in `.kiro/specs/orun-tui-cockpit/tasks.md`**
   (lines 5/14/122/296 still reference `github.com/flyingmutant/rapid`)
   left untouched — task-0146 §43 explicitly forbids touching it; this
   was accepted by the task-0144 verifier.

## Files changed

```
ai/tasks/task-0146.md                       (new, brief)
go.mod                                      (deps promoted to direct)
go.sum                                      (unchanged content; tidy)
internal/tui/model.go                       (+60 LOC routing/seeding)
internal/tui/services/live_service.go      (-5 LOC stub removed)
internal/tui/services/plan_service.go      (new, ~340 LOC)
internal/tui/services/plan_service_test.go (new, 4 tests)
internal/tui/views/plan_studio.go          (rewritten, ~290 LOC)
internal/tui/views/plan_studio_test.go     (new, 10 tests + 2 rapid props)
```

## Handoff to verifier

PR #145 is ready for review. Key things to spot-check:
1. `internal/tui/services/plan_service.go` parity vs
   `cmd/orun/main.go:generatePlan` — particularly preset merging,
   trigger resolution, and the env-filter ordering.
2. ChangedOnly warning text — should be visible in any Plan Studio
   session that toggles it.
3. Property test convergence (`go test ./internal/tui/views/... -run Property -count=10`).
4. CI must pass `go vet`, `go test ./...`, and any TUI golden tests.
