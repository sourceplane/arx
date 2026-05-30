# Task 0004 (Verifier pass)

Agent: Verifier

## Current Repo Context

- Active spec: `specs/orun-state-redesign/` (Phase 1, local-only).
- Active milestone: **M2 — `internal/statestore`** (PR B in flight).
- Implementer Task 0004 (M2 PR B) shipped on branch
  `impl/task-0004-m2-statestore-prb` and opened PR **#155**:
  - Title: "Task 0004: M2 PR-B — statestore CompareAndSwap + List"
  - Head OID: `4875025` (`487502521d8d8013b08151b7c610399f2c75f42a`)
  - State: OPEN, MERGEABLE, mergeStateStatus CLEAN.
  - Required CI: `CI / Orun Plan` SUCCESS (run **26670829548**, completed
    2026-05-30T01:36:00Z) and `orun remote-state conformance / Harness
    dry-run guard` SUCCESS (run **26670829550**, completed
    2026-05-30T01:35:13Z). Matrix legs SKIPPED legitimately (still empty
    matrix at M2).
  - Implementer self-reported coverage on `internal/statestore` = **95.4%**
    via `make test-state-redesign`.
- Implementer report: `ai/reports/task-0004-implementer.md` (currently
  **untracked locally** — see "Implementer report not committed to PR"
  pitfall in `orun-saas-implementer` skill; you must commit + push it to
  the PR branch as part of verification, then re-check CI).
- Last completed merged: Task 0003 (PR #154 → main `9b0a39c` on
  2026-05-29). Repo health 🟢 green prior to this task.
- No open proposals under `ai/proposals/` for Task 0004.

## Objective

Validate Task 0004 against the Verifier Standard in `agents/orchestrator.md`
and the M2 PR-B "done when" criteria in
`specs/orun-state-redesign/implementation-plan.md`. Confirm that
`*LocalStore.CompareAndSwap` and `*LocalStore.List` match
`state-store.md` §3.3 / §3.4 verbatim; that the atomicity / exclusivity /
CAS / `pgregory.net/rapid` round-trip suite from `test-plan.md` §2 / §3 is
present, race-clean, and meaningful (not stubs); that no production-caller
wiring slipped in; and that `internal/statestore` is still leaf-clean
with coverage ≥ 95 %. On PASS, apply the Verifier Merge Protocol.

## PR Boundary

- Verification only. No production-code edits, no spec edits, no PR-C work.
- One verifier-only artifact may be added to the PR branch:
  `ai/reports/task-0004-verifier.md`. If `ai/reports/task-0004-implementer.md`
  is missing on the PR branch (it currently is, locally), commit and push
  it to the PR branch alongside your verifier report and wait for CI to
  re-pass before merging.
- Do NOT touch `internal/statestore/local.go`, `internal/statestore/store.go`,
  `internal/statestore/paths.go`, `Makefile`, or `go.mod`/`go.sum`.

## Read First

- `agents/orchestrator.md` — Verifier Standard + Verifier Merge Protocol.
- `ai/tasks/task-0004.md` — implementer prompt (acceptance criteria, scope).
- `ai/reports/task-0004-implementer.md` — implementer self-report.
- `specs/orun-state-redesign/README.md` — entry / read order.
- `specs/orun-state-redesign/state-store.md` — §1 (interface), §3.3
  (CompareAndSwap), §3.4 (Read / List / Delete), §4 (error taxonomy),
  §6 (atomicity contract).
- `specs/orun-state-redesign/test-plan.md` — §1 (≥ 95 % coverage gate),
  §2 (atomicity suite shape), §3 (rapid property tests).
- `specs/orun-state-redesign/implementation-plan.md` — Milestone M2 "done
  when" checklist (PR B segment).
- PR #155 diff: `gh pr diff 155` and `gh pr view 155 --json …`.
- Reference only: `ai/reports/task-0003-verifier.md` for the prior
  M2-PR-A verification shape.

## Required Outcomes

- [ ] PR **#155** corresponds exactly to Task 0004 — no scope creep.
- [ ] `*LocalStore.CompareAndSwap` matches `state-store.md` §3.3:
      `Read` → revision compare → `Write`; `ErrNotFound` on absent target;
      `ErrConflict` on revision mismatch; errors wrap sentinels via
      `fmt.Errorf("%w: …", ErrX, …)`. The PR-A stub error string
      (`"… not implemented in PR A …"`) is GONE from `local.go`.
- [ ] `*LocalStore.List` matches `state-store.md` §3.4: walks translated
      prefix; returns `[]ObjectInfo` with logical (forward-slash, root-relative,
      no leading slash) paths; skips symlinks; filters `.orun-tmp-*`
      orphan tempfiles; non-existent prefix → empty slice (no error);
      `ErrInvalid` on alphabet/escape violations.
- [ ] All four required tests exist in `internal/statestore/*_test.go`
      and pass under `-race`:
      - 100-goroutine `Write+Read` atomicity (decoders never see partial JSON).
      - 100-goroutine `CreateIfAbsent` exclusivity (exactly one wins).
      - Concurrent CAS conflict (one wins, the other returns
        `ErrConflict`).
      - `pgregory.net/rapid` round-trip on path-alphabet inputs through
        `Write` → `Read` with stable lowercase-hex sha256 `Revision`.
- [ ] Local quality gates green:
      `go build ./...`, `go vet ./...`,
      `go test -race -count=1 ./internal/statestore/...`,
      `make test-state-redesign` (with reported `internal/statestore`
      coverage ≥ 95 %, target ≥ 96 %).
- [ ] `kiox -- orun validate --intent intent.yaml` exits 0; record the
      `kiox -- orun plan --changed …` output verbatim if the known
      composition-cache failure reproduces (CI is authoritative — see
      Task 0001+ historical note in `ai/state.json`).
- [ ] PR CI logs **inspected at log level** (not just status badges) for
      `CI / Orun Plan` (run 26670829548) and `orun remote-state
      conformance / Harness dry-run guard` (run 26670829550):
      - Real `orun plan --from-ci github …` invocation observed.
      - Empty-matrix shape legitimate (M2 PR-B still has 0 components).
      - Dry-run guard `[guard] PASS:` assertions present.
- [ ] `internal/statestore` leaf-clean: `go list -deps
      ./internal/statestore/...` shows zero `…/orun/internal/*` imports
      besides itself (`pgregory.net/rapid` in test files is fine).
- [ ] No production-caller wiring snuck in:
      `git diff origin/main...impl/task-0004-m2-statestore-prb -- cmd/orun
      internal/state internal/runner internal/runbundle` is empty.
- [ ] No new dependencies in `go.mod` beyond what was already present
      after Task 0003.
- [ ] No secrets / credentials in any committed file or CI log.
- [ ] If `ai/reports/task-0004-implementer.md` is missing on the PR
      branch, commit + push it to the branch (this matches the recurring
      "implementer report not committed" pitfall) and wait for CI to
      re-pass before merging.
- [ ] Verifier report at `/ai/reports/task-0004-verifier.md` filed with
      the budget shape from `agents/orchestrator.md`.
- [ ] On PASS: PR squash-merged into `main`, branch deleted on origin,
      local checkout switched back to `main`, fast-forward pulled, working
      tree clean (`git status --short` empty).
- [ ] On FAIL: PR left OPEN with clear blockers; verifier report
      enumerates each failed criterion.

## Non-Goals

- No `refs.go` / `indexes.go` work (PR C / Task 0005).
- No production-caller migration.
- No empty-directory `Delete` semantics revisit (left as a non-blocking
  Minor in Task 0003 verifier report).
- No CLI surface changes.
- No spec edits (file `ai/proposals/task-0004-spec-update.md` if a real
  spec drift surfaces).

## Constraints

1. Verify code reality via direct file reads, not just diff summaries.
2. Detect spec drift through `errors.Is` / `errors.As` traversal — confirm
   every error path wraps one of the four sentinels.
3. CAS implementation may carry a per-path in-process mutex (additive,
   permitted by spec §6 "best-effort on local"); flag only if the mutex
   alters cross-process semantics.
4. List ordering is unspecified by §3.4 — do NOT require any particular
   order; only require completeness, tempfile filtering, and forward-slash
   logical paths.
5. Use `gh` for PR + CI inspection. Use `kiox -- orun …` for Orun
   validation. Prefer `/Users/irinelinson/.local/bin/kiox` if `kiox` is
   not on `PATH`.
6. Never merge a PR with unresolved verification blockers or red required
   CI checks.
7. After merge: `git checkout main && git pull --ff-only origin main`,
   then `git status --short` must be empty.

## Integration Notes

- The known persistent local `kiox -- orun plan --changed --intent
  examples/intent.yaml` composition-cache failure carried from Task 0001+
  is NOT a regression. Record verbatim if it reproduces; CI is
  authoritative.
- If you add the verifier report (and the missing implementer report) to
  the PR branch, expect required CI to re-run. Do not merge until
  `mergeStateStatus=CLEAN` again with all required checks SUCCESS.

## Acceptance Criteria

✅ PR #155 maps to Task 0004 exactly (no scope creep, no extra files
   beyond `internal/statestore/*.go`, `_test.go` siblings, and the
   `ai/` lineage docs the implementer was scoped to touch).

✅ Code-path inspection of `internal/statestore/local.go`:
   - `CompareAndSwap` body implements Read → revision-compare → Write
     with sentinel-wrapping and the `ErrNotFound` / `ErrConflict` paths
     visible.
   - `List` body implements the translated-prefix walk with symlink
     skip, `.orun-tmp-*` filter, forward-slash logical-path conversion,
     and empty-on-not-exist semantics.
   - The PR-A stub strings are gone.

✅ Test inspection in `internal/statestore/*_test.go`:
   - Atomicity test runs ≥ 100 concurrent writers/readers and asserts
     decoder success.
   - `CreateIfAbsent` exclusivity test runs ≥ 100 concurrent goroutines
     and asserts exactly one nil return.
   - CAS conflict test runs concurrent CAS with shared `oldRev` and
     asserts exactly one `ErrConflict`.
   - `pgregory.net/rapid` round-trip generates path-alphabet segments
     `[a-zA-Z0-9._-]{1,32}`, depth 1–5, and asserts byte-for-byte
     `Read` equality plus stable lowercase-hex sha256 `Revision`.

✅ Local quality gates (run from the PR branch):
```
go build ./...
go vet ./...
go test -race -count=1 ./internal/statestore/...
make test-state-redesign
```
all exit 0; `make test-state-redesign` reports `internal/statestore`
coverage ≥ 95 %.

✅ Orun verification (per Verifier Merge Protocol):
```
/Users/irinelinson/.local/bin/kiox -- orun validate --intent intent.yaml
/Users/irinelinson/.local/bin/kiox -- orun plan --changed --intent intent.yaml --output plan.json
/Users/irinelinson/.local/bin/kiox -- orun run --plan plan.json --dry-run --runner github-actions
```
Recorded verbatim. Known composition-cache failure on `--changed` is
acceptable (CI authoritative).

✅ PR CI log inspection (not status only):
```
gh run view 26670829548 --log | head -200      # CI / Orun Plan
gh run view 26670829550 --log | head -200      # Harness dry-run guard
```
Real plan invocation + dry-run guard `[guard] PASS:` assertions present.

✅ Leaf-clean:
```
go list -deps ./internal/statestore/... | grep '/orun/internal/'
```
prints only `…/orun/internal/statestore` itself.

✅ No production-caller wiring:
```
git diff origin/main...impl/task-0004-m2-statestore-prb -- \
  cmd/orun internal/state internal/runner internal/runbundle
```
is empty.

✅ Secret scan clean:
```
git diff origin/main...impl/task-0004-m2-statestore-prb | \
  grep -Ei 'api[_-]?key|secret|token|password' || true
```
returns no committed credential material.

✅ On PASS: PR squash-merged, branch deleted, `main` fast-forwarded,
   `git status --short` empty.

## Verification

Execute, in order:

1. **Repo & PR state** — `git fetch origin`, `gh pr view 155 --json …`,
   confirm OPEN/MERGEABLE/CLEAN and SUCCESS on required checks.
2. **Implementer-report check** — confirm
   `ai/reports/task-0004-implementer.md` is present on the PR branch
   (`git ls-tree origin/impl/task-0004-m2-statestore-prb --name-only
   ai/reports/task-0004-implementer.md`). If absent, commit + push it.
3. **Code-path inspection** — read `internal/statestore/local.go`
   `CompareAndSwap` and `List` bodies. Confirm error-sentinel wrapping.
4. **Test-suite inspection** — open the new `*_test.go` file(s) and
   verify the four required tests are present with the required shape.
5. **Local gates** — run the four commands listed above on the PR
   branch.
6. **Orun validation** — run the `kiox -- orun validate / plan / run
   --dry-run` triple. Record verbatim.
7. **CI log inspection** — `gh run view --log` for both required runs;
   confirm real plan invocation + dry-run guard PASS lines.
8. **Leaf-clean + caller-wiring + secrets scans** — as above.
9. **Decision** — PASS only if every Required Outcome above is met.
10. **Merge protocol on PASS** — squash-merge PR #155, capture merge
    commit SHA, `git checkout main && git pull --ff-only origin main`,
    `git branch -D impl/task-0004-m2-statestore-prb`, confirm
    `git status --short` empty.
11. **State updates** (after merge) — orchestrator owns the post-merge
    `ai/state.json` / `ai/context/*` advance; the verifier only files
    the verifier report and updates the ledger entry status.

## PR Creation Requirement

The Implementer has already created PR #155. Your job is to verify it.
You may add `ai/reports/task-0004-verifier.md` (and the missing
implementer report, if needed) as a verifier-only commit to the PR
branch before merging.

## When Done Report

Write `/ai/reports/task-0004-verifier.md` with:

- Result: PASS | FAIL
- Checks (numbered list of every verification step run + result)
- CI Log Review (run IDs, expected commands observed)
- Code Path Inspection (CAS + List body summaries vs §3.3 / §3.4)
- Test Suite Inspection (file paths + assertion shape per required test)
- Local Quality Gates (exact commands + outputs)
- Orun Validation (verbatim, including any known cache failure)
- Secret Handling Review
- Leaf-Clean Confirmation
- Issues (blocking / non-blocking, severity)
- Risk Notes (residual risk for PR C / M3)
- Spec Proposals (links + one-line reason; expected: none)
- Recommended Next Move (Task 0005 = M2 PR C scoping cue)
- Merge Outcome (commit SHA on `main`, branch deleted, repo clean)
