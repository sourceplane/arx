# Task 0025 (Verifier pass)

# Agent
Verifier

# Current Repo Context
- Phase 2 / Milestone **C2 PR-1** of `specs/orun-component-catalog/`
  (`internal/catalogresolve` — discover + load + inherit only) was
  implemented by Task 0025 on branch
  `impl/task-0025-c2-discover-load-inherit` and opened as **PR #170**.
  Implementer report: `ai/reports/task-0025-implementer.md`.
- Predecessors merged on `main`: C0 (PR #168 / `7f3f2bf`), C1 (PR #169 /
  `b50d799`). Phase 1 floors held: `internal/statestore` 95.7 %,
  `internal/revision` 90.3 %, `internal/executionstate` 90.0 %.
  Phase 2 floors held on `main`: `internal/catalogmodel` 91.1 %,
  `internal/sourcectx` 91.1 %, Sanitize* 100 %.
- **PR #170 status:** OPEN, MERGEABLE, mergeStateStatus **CLEAN**. All
  required checks green:
  - `state-redesign-tests / test` — SUCCESS (run `26705772895`)
  - `CI / Orun Plan` — SUCCESS (run `26705772887`)
  - `orun remote-state conformance / Harness dry-run guard` — SUCCESS
    (run `26705772906`)
  - Matrix legs (component env fanout, remote-state Run/Verify) SKIPPED
    legitimately on this branch.
- **Diff shape:** +1 413 / −0, 26 files; net new under
  `internal/catalogresolve/` (10 source/test files + 13 testdata
  fixtures), one additive file `internal/catalogmodel/schema_embed.go`
  (8 lines, `//go:embed`-only), `Makefile` +7 lines (catalogresolve
  coverage gate).
- **Two implementer call-outs to adjudicate:**
  1. Implementer added a new file to `internal/catalogmodel/`
     (`schema_embed.go`) so `catalogresolve` can compile the embedded
     schema without vendoring or absolute-path reads. The Task 0025
     prompt's PR Boundary said *"No edits to `internal/catalogmodel/`
     or `internal/sourcectx/`"*; implementer reads that as
     *"no edits to existing source files"* and proposes a wording
     tightening. **Verifier must adjudicate.**
  2. `internal/catalogresolve` coverage measured **exactly 90.0 %** —
     identical to the gate, zero headroom. Task 0024 hit the same
     pattern on `internal/catalogmodel` (90.2 % local, 87.9 % CI on
     some seeds) and required a deterministic backstop. Verifier must
     check whether `catalogresolve` has any rapid-driven /
     iteration-sensitive coverage and add headroom or a deterministic
     pin if so.

# Objective
Verify PR #170 against the Verifier Standard in `agents/orchestrator.md`
and the C2 PR-1 acceptance criteria in `ai/tasks/task-0025.md` +
`specs/orun-component-catalog/implementation-plan.md` §C2 +
`resolution-pipeline.md` stages 2 / 3 / 5. Decide PASS / FAIL and, on
PASS, merge PR #170, fast-forward `main`, and leave the local repo
clean. Default disposition is FAIL until every required check is green
on the merge commit candidate AND the two implementer call-outs are
adjudicated.

# PR Boundary
Verification only. Permitted edits to the PR branch (commit + push
before merging):

- `ai/reports/task-0025-verifier.md` (this verifier's report).
- A minimal, surgical fix to make the `internal/catalogresolve`
  coverage gate deterministic if reproducible runs show flake across
  the 90 % boundary — choose ONE of:
  1. Add 1–3 deterministic table-driven tests in
     `internal/catalogresolve/*_test.go` covering the exact branches
     that probabilistic / map-iteration-ordered tests sometimes miss.
     **Preferred** — same pattern Task 0024 verifier used for
     `catalogmodel`.
  2. Pin any `pgregory.net/rapid` `Check`/`Custom` seed and/or
     iteration count so the same branches are exercised every run.
  3. Bump the gate down to 88 % only if the package is provably hard to
     deterministically cover and the orchestrator approves the floor
     change in the verifier report (this is a last resort — Phase 2
     floors should not move down).

Out of scope for this verifier pass:

- Any inference / dependency / validation / `manifestHash` work — that
  is Task 0026 (C2 PR-2). Do NOT ask the implementer to bring
  Resolve()-level work into PR #170.
- Any C3+ functionality (graph, snapshot, store, CLI).
- Any Phase 1 source edit. Phase 1 invariants must hold byte-for-byte.

# Read First
- `ai/tasks/task-0025.md` — implementer prompt (acceptance criteria,
  PR Boundary, non-goals).
- `ai/reports/task-0025-implementer.md` — implementer's self-report
  (call-outs, assumptions, spec-proposal text).
- `agents/orchestrator.md` — Verifier Standard + Verifier Merge
  Protocol (sections "Verifier Standard" and "Verifier Merge Protocol").
- `specs/orun-component-catalog/implementation-plan.md` §C2 — milestone
  goal, suggested PR scope, "done when" (note: full C2 done-when spans
  two PRs; only the discover/load/inherit slice is gated here).
- `specs/orun-component-catalog/resolution-pipeline.md` §1–§3
  (workspace walk, manifest load, intent-defaults inheritance).
- `specs/orun-component-catalog/data-model.md` §6 + §7 (component
  manifest authored shape, provenance keys).
- `specs/orun-component-catalog/identity-and-keys.md` §10 (manifestHash
  context — confirm it is **not** computed in this PR; it lands in
  Task 0026).
- `specs/orun-component-catalog/test-plan.md` §1 + §3 (T-RES-1
  determinism mini-form, provenance shape).
- PR #170 diff (`gh pr diff 170`) and CI run logs
  (`gh run view 26705772895 --log` for the `state-redesign-tests`
  job specifically — confirm the catalogresolve gate output line is
  present).

# Required Outcomes
- PR #170 verified PASS or FAIL with reasoned adjudication of:
  - PR Boundary fidelity (only `internal/catalogresolve/`,
    `internal/catalogmodel/schema_embed.go`, `Makefile`).
  - The schema-embed call-out (Assumption #1 in the implementer
    report). Reasonable verifier outcome: ACCEPT with the implementer's
    proposed PR-Boundary wording tightening folded into a docs-only
    follow-up (do NOT block PR #170 on it). Document acceptance in the
    verifier report; if a spec proposal file is warranted, write
    `ai/proposals/task-0025-spec-update.md`.
  - The `catalogresolve` coverage headroom (Call-out #2). Reproduce
    `make test-state-redesign` locally three times back-to-back; if
    coverage is stable at ≥ 90 % every time, ACCEPT with risk note;
    if it drifts below 90 % on any run, attach a deterministic backstop
    per PR-Boundary option 1 above and re-push.
  - Phase 1 floors hold byte-for-byte (statestore 95.7 %, revision
    90.3 %, executionstate 90.0 %).
  - C0 + C1 floors hold (catalogmodel 91.1 %, sourcectx 91.1 %,
    Sanitize* 100 %).
- Verifier report committed to PR #170 branch under
  `ai/reports/task-0025-verifier.md` and CI re-run if any backstop
  commit was pushed.
- On PASS: PR #170 merged via `gh pr merge --squash --admin` (this
  repo's standard), local `main` fast-forwarded from `origin/main`,
  PR branch deleted, working tree clean (`git status --short` empty).
- On FAIL: PR #170 left OPEN with explicit blocker list in the verifier
  report, and the orchestrator pinged via `Recommended Next Move`.

# Non-Goals
- Do NOT bring forward C2 PR-2 work (infer / deps / validate /
  `manifestHash`).
- Do NOT touch Phase 1 packages.
- Do NOT touch `internal/sourcectx` source.
- Do NOT vendor a duplicate copy of the JSON Schema anywhere.
- Do NOT lower any Phase 1 or Phase 2 coverage floor without explicit
  orchestrator written approval in the verifier report's `Spec Proposals`
  section.

# Constraints
- Use `/Users/irinelinson/.local/bin/kiox` if `kiox` is not on `PATH`.
- Commit verifier-attached fixes (if any) on the PR branch, push, wait
  for CI green, only then merge.
- Never merge with any required CI check non-green on the head commit.
- Never include secrets, tokens, or absolute home paths in committed
  test fixtures or reports.

# Integration Notes
- The implementer's "schema embed" pattern (additive `schema_embed.go`
  exposing `ComponentYAMLSchema []byte`) is the same pattern any future
  package needing the schema (Task 0026 validate, C5 CLI, C8 catalogdiff)
  will reuse. If you accept it, note explicitly in the verifier report
  that the convention is **one additive file in `catalogmodel/` per
  cross-package contract surface, no edits to existing files** — this
  becomes a load-bearing convention for the rest of Phase 2.
- The `inherit.go` "explicit empty list = preserve" semantic mirrors the
  Phase 1 manifest behavior; verify that the explicit_empty_list fixture
  asserts an *empty* list survived (not a nil-merged result).
- Provenance keys are RFC 6901-escaped JSON pointers per
  `data-model.md` §7. Spot-check at least one nested field in the
  `canonical/` fixture (e.g. a key under `runtime.env`).

# Acceptance Criteria
- [ ] PR #170 corresponds exactly to Task 0025 scope (discover + load +
      inherit; no inference, no deps, no validation, no `manifestHash`).
- [ ] `git diff origin/main...HEAD --stat` shows ONLY paths inside
      `internal/catalogresolve/`,
      `internal/catalogmodel/schema_embed.go`, `Makefile`,
      `ai/reports/task-0025-implementer.md`, and (after this verifier
      pass) `ai/reports/task-0025-verifier.md`.
- [ ] No edits to existing source files in `internal/catalogmodel/` or
      `internal/sourcectx/` (additive `schema_embed.go` accepted with
      documented justification).
- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` all green.
- [ ] `make test-state-redesign` green three runs in a row, with
      `internal/catalogresolve/` coverage ≥ 90 % on every run; Phase 1
      and C0/C1 floors held byte-for-byte.
- [ ] `make verify-generated` green (no schema drift).
- [ ] `kiox -- orun validate --intent intent.yaml` green when run on
      `main` post-merge.
- [ ] Determinism: two consecutive `DiscoverAndLoad` calls produce
      byte-identical `[]AuthoredManifest` (mini-T-RES-1 asserted in
      `resolve_test.go`); spot-check by running
      `go test -count=10 ./internal/catalogresolve/...` and confirm no
      flake.
- [ ] PR #170 CI rollup is `CLEAN` on the merge candidate head; required
      checks SUCCESS; remaining checks SKIPPED legitimately (no FAILURES,
      no CANCELLED, no in-progress at merge time).
- [ ] Verifier report `ai/reports/task-0025-verifier.md` written with
      Result, Checks, Issues, CI Log Review, Risk Notes, Spec Proposals,
      Recommended Next Move.
- [ ] On PASS: PR squash-merged; `main` fast-forwarded; PR branch
      deleted; `git status --short` empty.

# Verification
1. **Repo + PR audit.** `git fetch origin`,
   `git checkout impl/task-0025-c2-discover-load-inherit`,
   `git pull --ff-only`. Confirm head matches `gh pr view 170 --json
   headRefOid`. Run `gh pr view 170 --json
   state,mergeable,mergeStateStatus,statusCheckRollup` and assert
   `OPEN/MERGEABLE/CLEAN` plus all required checks SUCCESS.
2. **Diff scope check.** `gh pr diff 170 --name-only` — must equal the
   set under `Acceptance Criteria` second bullet.
3. **PR Boundary fidelity.**
   - `git diff origin/main...HEAD -- internal/catalogmodel/` should
     contain ONLY the new `schema_embed.go` file (no edits to existing
     `.go` files in `catalogmodel/`).
   - `git diff origin/main...HEAD -- internal/sourcectx/` should be
     empty.
   - `git diff origin/main...HEAD -- internal/statestore/
     internal/revision/ internal/executionstate/ internal/triggerctx/`
     should be empty (Phase 1 untouched).
4. **Local gates.**
   ```
   go build ./...
   go vet ./...
   go test -race ./...
   make test-state-redesign
   make verify-generated
   ```
   Repeat `make test-state-redesign` three times to confirm coverage
   stability.
5. **Determinism stress.** `go test -count=10 -race
   ./internal/catalogresolve/...` — must be 0 failures.
6. **Spec drift check.** Read `inherit.go` and confirm the
   precedence ladder matches `resolution-pipeline.md` §3
   (authored ≻ intent-defaults ≻ inferred-future). Read `discover.go`
   and confirm default excludes (`.git .orun build dist node_modules
   vendor`) match the spec; intent excludes append, never replace.
7. **Provenance shape.** Read at least one fixture-driven test in
   `resolve_test.go` and confirm the asserted JSON-pointer keys are
   RFC 6901-escaped (`/foo~1bar` for slashes, `/a~0b` for tildes).
   This is the contract C2 PR-2 + C8 catalogdiff will rely on.
8. **Schema-embed adjudication.** Read
   `internal/catalogmodel/schema_embed.go`. Confirm:
   - 8 lines, `//go:embed`-only, no logic, no exported type other
     than `var ComponentYAMLSchema []byte`.
   - No existing `catalogmodel` source file was edited (use
     `git diff origin/main...HEAD -- internal/catalogmodel/*.go`
     filtered to non-new files).
   - Decide ACCEPT (preferred) or REQUEST-CHANGE. If ACCEPT, write
     a single-paragraph "convention adopted" note in the verifier
     report so future tasks reuse the pattern without re-litigating.
9. **CI log inspection.**
   ```
   gh run view 26705772895 --log | grep -E "catalogresolve|coverage"
   ```
   Confirm the catalogresolve gate measured ≥ 90 % on the CI run.
   Compare against three local `make test-state-redesign` runs.
10. **Backfill + commit (only if needed).** If any deterministic
    backstop is added: commit with message
    `test(catalogresolve): deterministic coverage backstop for C2 PR-1
    (task-0025 verifier)`, push, wait for `gh pr checks 170 --watch`.
11. **Report.** Write `ai/reports/task-0025-verifier.md` with the
    standard sections.
12. **Merge protocol (PASS path).**
    - `gh pr merge 170 --squash --admin --delete-branch` (matches
      Task 0024 / Task 0023 merge shape).
    - `git checkout main && git pull --ff-only origin main`.
    - `git status --short` must be empty before reporting complete.
    - Note the squash commit SHA in the verifier report.

# PR Creation Requirement
The Implementer has already created PR #170. Your job is to verify it.
You do NOT open a new PR. If a deterministic backstop is required, push
to the existing PR branch.

# When Done Report
File: `ai/reports/task-0025-verifier.md`

Sections:

- **Result**: PASS or FAIL.
- **Checks**: every command run + outcome (build, vet, test -race,
  test-state-redesign x3, verify-generated, kiox orun validate,
  determinism stress, CI log inspection).
- **Issues**: blockers (FAIL only) or accepted-with-note items (PASS).
  Specifically adjudicate the schema-embed call-out and the coverage
  headroom call-out.
- **CI Log Review**: PR #170 head SHA, run IDs verified, link to the
  catalogresolve gate line in CI logs.
- **Risk Notes**: residual risks (e.g. no headroom on
  `catalogresolve` floor; convention "one additive file per
  cross-package contract surface in `catalogmodel/`" now load-bearing
  for Phase 2).
- **Spec Proposals**: if you accept the implementer's PR-Boundary
  wording tightening, write `ai/proposals/task-0025-spec-update.md`
  per the Proposal template in `agents/orchestrator.md` and reference
  it here.
- **Recommended Next Move**: orchestrator should advance `current_task`
  to `0026` (C2 PR-2: infer + deps + validate + `manifestHash`) and
  bump notes; `active_milestone` stays at `C2` until both PRs land.
- **Squash Commit SHA** (PASS path) and **Merged At** timestamp.
