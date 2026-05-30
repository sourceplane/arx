# Task 0009 — Verifier (M3 PR-A)

Agent: Verifier

## Current Repo Context

- Active spec: `specs/orun-state-redesign/` (Phase 1, local-only). Active milestone: **M3 — `internal/revision`**. M2 closed at PR #156 (`cd8b3e8`, verified PASS 2026-05-30).
- Implementer Task 0007 (M3 PR-A — `internal/revision` model + keys + writer skeleton) was implemented locally on branch `impl/task-0007-m3-revision-pra`, then delivered into a real PR by the corrective Task 0008 chore. Both implementer reports name PR **#157**.
- PR **#157** state at orchestrator emission time:
  - `state=OPEN`, `mergeable=MERGEABLE`, `mergeStateStatus=CLEAN`
  - head SHA `500218c0bcbb0a671d528453b24043ebc8da4d53` (commit `500218c` "Task 0008: backfill PR #157 number into 0007 report + file 0008 chore report"; parent `96621ed` "Task 0007: M3 PR-A — internal/revision model + keys + writer skeleton")
  - Required CI both SUCCESS:
    - `CI / Orun Plan` — run `26672937657`, job `78619489862`
    - `orun remote-state conformance / Harness dry-run guard` — run `26672937641`, job `78619489786`
  - Five matrix legs SKIPPED legitimately (empty matrix at M3 PR-A — same shape as #152/#155/#156).
  - Files changed (13): `Makefile`, `internal/revision/{model,keys,writer,version}.go`, `internal/revision/{keys_test,writer_test,coverage_test}.go`, `ai/reports/task-{0007,0008}-implementer.md`, `ai/tasks/task-{0006-verifier,0007,0008}.md`. Total `+2292 / -0`.
- Implementer reports:
  - `ai/reports/task-0007-implementer.md` — claims local quality gates green: `go build ./...`, `go vet ./...`, `go test -race -count=1` on revision/statestore/triggerctx, `make test-state-redesign` (revision pkg coverage **93.3 %**, gate ≥ 90 %; statestore stays at 96.1 %).
  - `ai/reports/task-0008-implementer.md` — corrective delivery chore; no production-code changes beyond Task 0007's tree.
- Outstanding implementer flag for the verifier: **claim-first ordering deviation from `cli-surface.md` §1.2** — index slot reserved via `CreateIfAbsent` BEFORE body writes (original spec lists indexes last). Rationale in `ai/reports/task-0007-implementer.md` § "Step-Order Deviation From cli-surface.md §1.2 (claim-first)". Verifier MUST adjudicate (accept-and-document OR file proposal at `ai/proposals/task-0007-spec-update.md`).
- After PR #157 merges and is verified, M3 PR-A closes and Task 0010 (M3 PR-B — `manifest.go` + `resolver.go` + legacy `.orun/plans/<checksum>.json` mirror body behind `Config.CompatibilityWrites`) is the next implementer task.

## Objective

Validate Task 0007 / PR #157 against the Verifier Standard in `agents/orchestrator.md` and the **M3 "Done when"** criteria in `specs/orun-state-redesign/implementation-plan.md`. Confirm the model, keys, and writer skeleton match `data-model.md` §3 / §6 / §7 and `design.md` §5–§6 within the PR-A scope, that no overreach exists (no `manifest.go`, no `resolver.go`, no production-caller wiring), and that the claim-first deviation is either acceptable as-implemented or formally captured as a spec-change proposal. On PASS, merge PR #157, sync local `main`, and leave the working tree clean.

## PR Boundary

- Verification only. Verifier may commit to the PR branch:
  - `ai/reports/task-0009-verifier.md` (this report).
  - Optionally `ai/proposals/task-0007-spec-update.md` if the claim-first ordering decision requires a spec amendment.
  - Any tiny verification-only fix that is strictly necessary for mergeability (typo, stray TODO removal). Anything beyond that → FAIL with explicit blockers; do not edit production code.
- Out of scope: M3 PR-B work (`manifest.go`, `resolver.go`, legacy mirror body), production-caller wiring, M4+ work, refactor of M0/M1/M2 surface.

## Read First

1. `agents/orchestrator.md` — Verifier Standard + Verifier Merge Protocol.
2. `specs/orun-state-redesign/README.md` — index + read order.
3. `specs/orun-state-redesign/implementation-plan.md` — Milestone **M3** goal, suggested PR scope, "Done when" checklist.
4. `specs/orun-state-redesign/design.md` — §5 (architecture / writer ordering) and §6 (compatibility writes / `stateCompatibilityWrites`).
5. `specs/orun-state-redesign/data-model.md` — §3 (`PlanRevision`), §4 (manifest — for non-goal awareness), §6 (refs), §7 (indexes).
6. `specs/orun-state-redesign/state-store.md` — §1 (frozen interface), §3 (CreateIfAbsent / Write atomicity), §6 (caller-owns-retry).
7. `specs/orun-state-redesign/cli-surface.md` — §1.2 step ordering (the contested spec text).
8. `specs/orun-state-redesign/test-plan.md` — coverage targets and JSON-byte-stability requirement.
9. `ai/tasks/task-0007.md`, `ai/tasks/task-0008.md` — implementer prompts.
10. `ai/reports/task-0007-implementer.md`, `ai/reports/task-0008-implementer.md` — implementer reports (especially Task 0007's "Step-Order Deviation" section).
11. PR **#157** diff and commits on branch `impl/task-0007-m3-revision-pra`.

## Required Outcomes

- [ ] Verifier report at `ai/reports/task-0009-verifier.md` with sections: Result, Checks, CI Log Review, Issues, Risk Notes, Spec Proposals, Recommended Next Move, PR Number + merge SHA.
- [ ] Explicit adjudication of the claim-first ordering deviation: either (a) accept-and-document inline in the verifier report's Risk Notes, OR (b) file `ai/proposals/task-0007-spec-update.md` per the proposal template in `agents/orchestrator.md`.
- [ ] PR #157 either squash-merged into `main` (PASS) or left OPEN with explicit blockers (FAIL).
- [ ] If PASS: local `main` fast-forwarded to `origin/main`; PR branch deleted; `git status --short` clean (resolve any verifier-created scratchpad edits before ending the task).
- [ ] `/ai/state.json` `task_agent` updated to `/ai/tasks/task-0009-verifier.md` while this verifier task is in flight (orchestrator sets it on emission; verifier flips it to its own report path on completion if that is the most recently produced file).

## Constraints

1. **No production-code edits** beyond a verifier-only typo/TODO fix strictly required for mergeability. The M3 PR-A surface (`internal/revision/{model,keys,writer,version}.go`) stays as authored.
2. **Spec edits only via proposal.** If the claim-first deviation needs spec amendment, write `ai/proposals/task-0007-spec-update.md` with the required Proposal sections; do NOT edit `cli-surface.md` directly.
3. **No M3 PR-B scope.** `manifest.go`, `resolver.go`, the legacy `.orun/plans/<checksum>.json` mirror body, and any wiring into `cmd/orun` / `internal/runner` / `internal/runbundle` MUST be absent from the PR. The `// TODO(m5)` compatibility-mirror stub in `writer.go` may exist (gated by `Config.CompatibilityWrites`); confirm it is a stub and not a real body write.
4. **Leaf-clean.** `internal/revision` may import `internal/triggerctx` and `internal/statestore` (M3 depends on M1+M2) but MUST NOT import `cmd/`, `internal/state`, `internal/runner`, `internal/runbundle`, or anything outside the documented dependency set in `design.md` §13. Verify with `go list -deps ./internal/revision | rg "/orun/internal/"`.
5. **Coverage gate ≥ 90 %** on `internal/revision` per M3 "Done when". `internal/statestore` must remain ≥ 95 % (regression check).
6. **Existing milestones unchanged.** `internal/triggerctx`, `internal/statestore`, and `internal/testfx/statefs` files MUST be byte-identical to `origin/main` (the diff in step 2 must show zero lines touching them).
7. **CI is authoritative for `orun plan --changed`.** The local composition-cache quirk (`stack.yaml at ~/.orun/cache/compositions/c41fc08… has no spec.compositions`) is a known environment artifact carried since Task 0001. Reproducing it is not a blocker; failing CI is.
8. **Merge gate**: never merge unless BOTH local quality gates AND the two required CI checks (`CI / Orun Plan`, `Harness dry-run guard`) are SUCCESS at log level on the PR head SHA.

## Verification Steps

Run, in order:

### 1. Repo State

```bash
git fetch origin
git status --short
git log --oneline -5 origin/main
gh pr view 157 --json number,state,mergeable,mergeStateStatus,headRefName,headRefOid,statusCheckRollup
```

Confirm: PR is OPEN, MERGEABLE, CLEAN, head SHA = `500218c0bcbb0a671d528453b24043ebc8da4d53`. If a new commit has been pushed since orchestrator emission, re-evaluate from the new head and inspect any new diff.

### 2. Diff Audit (overreach detection)

```bash
git fetch origin pull/157/head:pr-157
git diff --stat origin/main...pr-157
# Surfaces that MUST be empty:
git diff origin/main...pr-157 -- cmd/orun internal/state internal/runner internal/runbundle
git diff origin/main...pr-157 -- internal/triggerctx internal/statestore internal/testfx/statefs
# Surfaces that MUST NOT exist (PR-B / M4 leakage):
git diff origin/main...pr-157 -- internal/revision/manifest.go internal/revision/resolver.go
ls pr-157 2>/dev/null  # ignore — sanity only
git ls-tree -r --name-only origin/main..pr-157 -- internal/executionstate || true
```

The "MUST be empty" calls have to return zero lines. The "MUST NOT exist" files must not be present in the PR tree. If any of these fail, FAIL with the exact lines.

### 3. Spec Conformance — Code-Path Inspection

Open these in parallel with the spec sections:

- `internal/revision/model.go` ↔ `data-model.md` §3 (PlanRevision) + §3 sibling (RevSummary if present).
- `internal/revision/keys.go` ↔ `design.md` §5 (revision key shape) + collision-suffix logic in M3 "Done when".
- `internal/revision/writer.go` ↔ `design.md` §5.1 (writer order) + `cli-surface.md` §1.2 + `state-store.md` §3 (`CreateIfAbsent` / `Write`) + `design.md` §6 (`stateCompatibilityWrites`).
- `internal/revision/version.go` ↔ `data-model.md` §1 (the `version.json` envelope).

Check:

- `PlanRevision` and any sibling structs: every field name + JSON tag matches `data-model.md` §3 byte-for-byte. Deterministic JSON marshalling: `SetIndent("", "  ")`, `SetEscapeHTML(false)`, trailing `\n`.
- `RevisionKey(trig, planHash)` regex validator + collision-suffix logic exist; collision suffix path is exercised by a property test (M3 "Done when" calls this out explicitly).
- `WriteRevision` body order is one of:
  - `cli-surface.md` §1.2 verbatim — bodies → refs → indexes (canonical), OR
  - claim-first — index slot reserved via `CreateIfAbsent` BEFORE bodies, then bodies, then refs (the implementer's chosen ordering).
  - **If claim-first**: confirm the rationale in `ai/reports/task-0007-implementer.md` § "Step-Order Deviation From cli-surface.md §1.2" is sound (`CreateIfAbsent` is exclusive per `state-store.md` §3 and proven by Task 0004's 100-goroutine atomicity property test; refs still land after bodies preserving `state-store.md` §6 crash-recovery invariants). Adjudicate:
    - Accept → record in Risk Notes with one paragraph of justification.
    - Reject → file `ai/proposals/task-0007-spec-update.md` requesting `cli-surface.md` §1.2 amendment, and decide whether to FAIL or continue (continue if the PR otherwise meets every "Done when" and the proposal supersedes the spec text non-controversially).
- `stateCompatibilityWrites` flag is plumbed via `Config.CompatibilityWrites` (default true) and routes to a `writeCompatibilityMirror` (or equivalent) seam that is currently a `// TODO(m5)` stub — i.e. NOT writing real bodies in this PR.
- `EnsureStateStoreVersion` (or equivalent) writes to logical path `"version.json"` (`.orun/version.json` on disk) per `data-model.md` §1. The implementer flagged this in Task 0007 report § "Open Items For Verifier" as a possible follow-up to relocate to a `statestore.StateStoreVersionPath()` helper — decide whether to require that now (FAIL/proposal) or defer with a Risk Note.
- Errors wrap existing sentinels (`ErrInvalid`, `ErrNotFound`, `ErrExists`, `ErrConflict`) via `fmt.Errorf("%w: …", ErrX, …)`. No new sentinels.
- All exported symbols carry doc comments (M2 / M3 cross-cutting standard).
- No string concatenation for paths anywhere in `keys.go` / `writer.go` — all paths come through `internal/statestore/paths.go` helpers (or local helpers that themselves use `paths.go`).

### 4. Local Quality Gates

```bash
go build ./...
go vet ./...
go test -race -count=1 ./internal/revision/...
go test -race -count=1 ./internal/statestore/...
go test -race -count=1 ./internal/triggerctx/...
make test-state-redesign        # confirm coverage gate prints "measured: <≥90.0>%" for revision and "<≥95.0>%" for statestore
go test -cover ./internal/revision/... ./internal/statestore/...   # second-source coverage measurement
go list -deps ./internal/revision | rg "/orun/internal/" || echo "no internal imports"
/Users/irinelinson/.local/bin/kiox -- orun validate --intent examples/intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent examples/intent.yaml --output /tmp/plan-0009.json || \
  echo "EXPECTED: composition-cache quirk on local; CI is authoritative"
/Users/irinelinson/.local/bin/kiox -- orun run --plan /tmp/plan-0009.json --dry-run --runner github-actions || \
  echo "skipped because plan not produced; record no-op"
```

All non-quirk steps must exit 0. Coverage thresholds:
- `internal/revision` ≥ 90 % (M3 gate; implementer reported 93.3 %)
- `internal/statestore` ≥ 95 % (M2 floor; should still be ~96.1 %)

The `internal/revision` package is allowed to import `internal/triggerctx` and `internal/statestore`; `go list -deps` should show exactly those plus stdlib + `oklog/ulid/v2`. Anything else is overreach.

### 5. CI Log Review

Inspect both required CI runs at log level (not just summary):

```bash
gh run view 26672937657 --log-failed | head -200    # CI / Orun Plan
gh run view 26672937657 --log | rg -n "orun plan" | head -20
gh run view 26672937641 --log | rg -n "guard\] PASS:" | head -50
```

Confirm:

- `CI / Orun Plan` (run `26672937657`) actually invoked `orun plan --from-ci github …` against `examples/intent.yaml`, recorded the legitimate empty-matrix shape (`0 components × 3 envs → 0 jobs`), and uploaded the plan artifact.
- `Harness dry-run guard` (run `26672937641`) emitted the full `[guard] PASS:` battery (bash syntax, command-count thresholds, duplicate-claim helper PASS+FAIL, status helper PASS+FAIL, exported env asserts).
- The five SKIPPED matrix legs are legitimate empty-matrix skips (not silently-failed required checks).

### 6. Secret Hygiene & Production-Grade Basics

```bash
rg -n -i "(token|password|secret|key=)" -- internal/revision/ || echo "clean"
```

Confirm: no plaintext tokens, no logging of sensitive material, deterministic JSON, all exported symbols have doc comments.

### 7. Decision and Merge

If every step is green and the claim-first deviation is adjudicated (accepted-with-Risk-Note OR proposal filed):

```bash
# (Optional) commit the verifier report and proposal to the PR branch first
git checkout impl/task-0007-m3-revision-pra
git pull --ff-only origin impl/task-0007-m3-revision-pra
# write ai/reports/task-0009-verifier.md (and ai/proposals/task-0007-spec-update.md if needed)
git add ai/reports/task-0009-verifier.md ai/proposals/task-0007-spec-update.md 2>/dev/null || true
git commit -m "Task 0009: M3 PR-A verifier report (PASS)"
git push origin impl/task-0007-m3-revision-pra
# wait for CI to re-run on the new commit and confirm both required checks SUCCESS at log level
gh pr checks 157 --watch --interval 30
gh pr review 157 --approve --body "Verified PASS — Task 0009 (M3 PR-A). See ai/reports/task-0009-verifier.md."
gh pr merge 157 --squash --delete-branch
git checkout main
git pull --ff-only origin main
git status --short
```

If any blocker exists, leave PR OPEN and write the verifier report with `Result: FAIL` and explicit blockers. Do NOT merge.

## Acceptance Criteria

- ✅ PR #157 corresponds 1:1 to Task 0007 / M3 PR-A (model + keys + writer skeleton + version helper); only `internal/revision/**`, `Makefile`, and `ai/**` task/report files changed. No `manifest.go`, no `resolver.go`, no production-caller wiring.
- ✅ Both required CI checks on PR #157 are SUCCESS at log level (`CI / Orun Plan` run `26672937657`, `Harness dry-run guard` run `26672937641`) on the head SHA — re-checked after any verifier-side commit.
- ✅ `internal/revision` coverage ≥ 90 % locally (target ≥ 93 %).
- ✅ `internal/statestore` coverage ≥ 95 % (no regression).
- ✅ `internal/revision` imports stay within `oklog/ulid/v2` + `internal/triggerctx` + `internal/statestore` + stdlib.
- ✅ `PlanRevision` struct and `version.json` envelope byte-match `data-model.md` §3 / §1.
- ✅ Revision-key regex + collision-suffix logic + 100-iteration uniqueness/collision property test all present.
- ✅ Writer-order matches either `cli-surface.md` §1.2 verbatim OR claim-first with verifier acceptance and either an inline Risk Note or a filed proposal at `ai/proposals/task-0007-spec-update.md`.
- ✅ `Config.CompatibilityWrites` plumbed through to a `// TODO(m5)` stub mirror seam — no real legacy body writes in PR-A.
- ✅ All exported symbols carry doc comments.
- ✅ No new error sentinels; no string concatenation for paths; all paths via `internal/statestore` helpers.
- ✅ `internal/triggerctx`, `internal/statestore`, and `internal/testfx/statefs` byte-identical to `origin/main`.
- ✅ On PASS: PR #157 squash-merged, branch deleted, local `main` fast-forwarded, `git status --short` clean.
- ✅ Verifier report `ai/reports/task-0009-verifier.md` filed; `ai/proposals/task-0007-spec-update.md` filed if claim-first deviation is rejected.

## When Done Report

Save to: `/ai/reports/task-0009-verifier.md`

Sections:

- Result: PASS | FAIL
- Checks (every command from steps 1–6 with exit code / outcome)
- CI Log Review (run IDs, expected commands actually executed, evidence quoted)
- Claim-First Adjudication (accept-and-document OR proposal filed at `ai/proposals/task-0007-spec-update.md`, with the reasoning)
- Issues (blocking or non-blocking, ranked)
- Risk Notes (residual risk after merge — e.g. `version.json` helper location, `// TODO(m5)` mirror stub, M3 PR-B scope still owed)
- Spec Proposals (links + one-line reason; expected to be **None** OR a single `ai/proposals/task-0007-spec-update.md`)
- Recommended Next Move (Task 0010 — M3 PR-B implementer scoping if PASS; remediation steps if FAIL)
- PR Number (157) and merge commit SHA (post-merge) or "OPEN — blocked" (FAIL)
