# Task 0022 — Verifier

Agent: Verifier

## Current Repo Context

- Phase 2 (`specs/orun-component-catalog/`) rollover. Task 0022 is the
  **docs-and-bookkeeping** half of milestone C0. The C0 code half ships
  separately as Task 0023.
- Implementer PR: **#167** on branch
  `impl/task-0022-phase2-rollover`, head SHA
  `1016a2b6dbfdcfecc79293c84dd5db78ecdabb18`.
  `mergeable=MERGEABLE`, `mergeStateStatus=CLEAN`.
- PR diff stat: 21 files changed, +5 803 / −287. File list (per
  `gh pr view 167 --json files`):
  - `agents/orchestrator.md` — rewritten Phase 2 sections; adds the
    "Deferred Decision Protocol" section (per implementer report,
    proactive addition not present in the prior file).
  - `ai/state.json` — Phase 2 rotation: `current_task=0022`,
    `active_spec=specs/orun-component-catalog`, `active_milestone=C0`,
    `phase_history.phase_1_orun_state_redesign` block recording M0–M6
    COMPLETE, final PR #165, coverage floors
    (statestore 95.7 %, revision 90.3 %, executionstate 90.0 %,
    triggerctx passes).
  - `ai/context/current.md` — refreshed for Phase 2 frame; Phase 1
    closure summary moved into a "Past Phase" section.
  - `ai/context/task-ledger.md` — appended Phase 2 header + Task 0022
    entry; Phase 1 ledger preserved verbatim.
  - `ai/waiting_for_input.md` — short "no input" note.
  - `ai/reports/task-0022-implementer.md` — implementer report.
  - `ai/tasks/task-0022.md` — implementer prompt committed for
    provenance.
  - `specs/orun-component-catalog/{README,design,data-model,identity-and-keys,resolution-pipeline,catalog-store,compatibility-and-migration,cli-surface,sync-model,implementation-plan,test-plan,risks-and-open-questions}.md`
    — the 12-doc authoritative spec pack.
  - `specs/orun-component-catalog/_archive/full-design-monolith.md` —
    the root-level `orun-catalog-full-design.md` moved here under
    `_archive/` (preserves provenance per user preference for archive-
    not-delete).
  - `specs/orun-component-catalog/_archive/README.md` — declares
    `_archive/` non-authoritative.
- Required CI on PR head SHA (per `gh pr view 167 --json statusCheckRollup`):
  `CI / Orun Plan` run `26703506317` SUCCESS (53 s);
  matrix leg `${{ matrix.component }}/${{ matrix.env }}` SKIPPED
  (legitimate empty-matrix shape for a docs-only PR — same shape
  as every prior state-redesign-era spec/test-only PR).
- Local repo on this machine: branch `impl/task-0022-phase2-rollover`
  checked out, 4 commits ahead of `main` (`b4178dd`):
  - `7503297` [1/3] land orun-component-catalog spec pack
  - `c3743a9` [2/3] rotate orchestrator + state.json
  - `455d635` [3/3] archive monolith, refresh context, land task prompt
  - `1016a2b` Task 0022: implementer report + flip state.json task_agent
- Implementer report: `ai/reports/task-0022-implementer.md`. Reports
  `go build ./...` and `go test ./...` PASS with all Phase 1
  coverage-floored packages green; spec content shipped as-is (zero
  `/ai/proposals/` entries filed); zero Go code in the diff.

## Objective

Verify PR #167 against Task 0022's docs-and-bookkeeping acceptance
criteria, the Phase 2 rollover protocol established in the rewritten
`agents/orchestrator.md`, and the Verifier Standard from
`agents/orchestrator.md` (sections 459–500). Confirm the diff is
**zero Go**, the spec pack contains all 12 named documents, the
orchestrator rewrite is internally consistent, and `ai/state.json` /
`ai/context/*` reflect Phase 2 cleanly. If PASS, merge PR #167,
fast-forward `main`, and leave the local repo clean. If FAIL, leave
PR #167 open with explicit blockers.

## PR Boundary (must match Task 0022 implementer)

- **Docs and bookkeeping ONLY.** Allowed paths:
  - `specs/orun-component-catalog/**` (12 named docs +
    `_archive/full-design-monolith.md` + `_archive/README.md`)
  - `agents/orchestrator.md`
  - `ai/state.json`
  - `ai/context/current.md`, `ai/context/task-ledger.md`
  - `ai/waiting_for_input.md`
  - `ai/tasks/task-0022.md`
  - `ai/reports/task-0022-implementer.md`
  - The repo-root `orun-catalog-full-design.md` MUST be deleted
    (the file moved to `_archive/full-design-monolith.md`).
- **Out of scope (FAIL on overreach):**
  - Any change under `internal/`, `cmd/`, `pkg/`, `Makefile`,
    `go.mod`, `go.sum`, `examples/`, `intent.yaml`, GitHub workflow
    files, or any `_test.go`.
  - Any edit to `specs/orun-state-redesign/**` (Phase 1 must remain
    intact).
  - Any edit to existing reports under `ai/reports/task-00{01..21}-*.md`
    or earlier `ai/tasks/task-00{01..21}*.md`.

## Read First

- `agents/orchestrator.md` — Verifier Standard (§459–500) and
  Verifier Merge Protocol (§485–500). On this branch you are also
  verifying the rewritten Phase 2 sections of this same file.
- `ai/tasks/task-0022.md` — implementer prompt (acceptance list, PR
  boundary, non-goals).
- `ai/reports/task-0022-implementer.md` — implementer claims and
  assumptions.
- `specs/orun-component-catalog/README.md` — entry point, doc index,
  read order, agent-role usage.
- `specs/orun-component-catalog/implementation-plan.md` — confirm C0
  goal / "done when" wording matches what `agents/orchestrator.md`
  and `ai/context/current.md` say about C0.
- `ai/state.json` — `phase_history.phase_1_orun_state_redesign`
  block must record M0–M6 COMPLETE, closure 2026-05-30, final PR
  #165, and the four coverage floors verbatim.
- `ai/context/current.md` — Phase 2 frame; Phase 1 closure preserved
  under "Past Phase".
- `ai/context/task-ledger.md` — Phase 2 header + Task 0022 entry
  appended; Phase 1 ledger above it byte-identical to `main`.

## Required Outcomes

- [ ] PR #167 maps exactly to Task 0022; no overreach (no Go,
      no Phase 1 spec edits, no prior-task report edits).
- [ ] All 12 named spec docs present under
      `specs/orun-component-catalog/`:
      `README.md`, `design.md`, `data-model.md`, `identity-and-keys.md`,
      `resolution-pipeline.md`, `catalog-store.md`,
      `compatibility-and-migration.md`, `cli-surface.md`,
      `sync-model.md`, `implementation-plan.md`, `test-plan.md`,
      `risks-and-open-questions.md`.
- [ ] `_archive/full-design-monolith.md` and `_archive/README.md`
      present under `specs/orun-component-catalog/_archive/`.
      `_archive/README.md` declares the archive non-authoritative.
- [ ] Repo root no longer contains `orun-catalog-full-design.md`.
- [ ] `agents/orchestrator.md` cites `specs/orun-component-catalog/`
      as the active spec, names milestones C0–C9, demotes
      `specs/orun-state-redesign/` to "predecessor (Phase 1, COMPLETE
      — M0–M6 merged)". The "Deferred Decision Protocol" section is
      present and consistent (orchestrator loop must keep producing
      PR-sized work whenever any human-independent candidate exists).
- [ ] `ai/state.json` shape matches: `current_task=0022`,
      `active_spec=specs/orun-component-catalog`,
      `active_milestone=C0`,
      `phase_history.phase_1_orun_state_redesign` block records
      M0–M6 COMPLETE / 2026-05-30 / PR #165 / coverage floors
      (statestore 95.7 %, revision 90.3 %, executionstate 90.0 %,
      triggerctx passes). Phase 1 floors must NOT be lowered or
      removed.
- [ ] `ai/context/current.md` describes the Phase 2 frame and moves
      Phase 1 closure into a "Past Phase" section. The C0 split
      (Task 0022 = docs/rollover, Task 0023 = code half) is
      explicitly named.
- [ ] `ai/context/task-ledger.md` Phase 1 entries are byte-identical
      to `main` for tasks 0001–0021; a new `## Task 0022` entry is
      appended under a Phase 2 header.
- [ ] `ai/waiting_for_input.md` is a short "no input currently
      requested" note (Phase 2 has no synchronous human-input
      blockers right now).
- [ ] Implementer report present at
      `ai/reports/task-0022-implementer.md` with PR Number `#167`.
- [ ] Local quality gates green on PR head: `go build ./...`,
      `go vet ./...`, `go test ./...`, `make test-state-redesign`.
      Phase 1 coverage floors preserved (statestore ≥ 95 %,
      revision ≥ 90 %, executionstate ≥ 90 %, triggerctx passes).
- [ ] PR CI checks green at log level on final head SHA per
      `gh run view <id> --log` (NOT just `--json conclusion`):
      `CI / Orun Plan` PASS; matrix leg SKIPPED is legitimate for
      a docs-only diff.
- [ ] No secrets / tokens / user emails / cloud credentials anywhere
      in the diff.
- [ ] No internal-link rot inside the new spec pack
      (`README.md` references that don't resolve to a sibling file
      = FAIL or proposal).

## Non-Goals

- No verifier-side feature work. Verifier may add a verifier report
  and, if essential, one tiny verification-only fix; otherwise the
  PR branch should not be modified.
- Do not edit specs unless filing `ai/proposals/task-0022-spec-update.md`
  for genuine drift.
- Do not start C0 code half (Task 0023).
- Do not introduce Go code, CLI changes, or `internal/catalogmodel`.

## Verification Steps

1. **Read implementer report and diff.**
   ```
   gh pr view 167 --json title,headRefOid,mergeable,mergeStateStatus,statusCheckRollup,files
   gh pr diff 167 | head -200
   ```
   Confirm the file list matches the boundary above.

2. **Scope-discipline scan.** From the local checkout:
   ```
   git fetch origin pull/167/head:verify/task-0022
   git checkout verify/task-0022
   git diff --stat main...verify/task-0022
   git diff --name-only main...verify/task-0022 \
     | grep -E '\.(go|mod|sum)$|^Makefile$|^cmd/|^internal/|^pkg/|^\.github/|^examples/|^intent\.yaml$' \
     && echo "OVERREACH: code/build files in diff" && exit 1 \
     || echo "scope clean"
   git diff --name-only main...verify/task-0022 \
     | grep -E '^specs/orun-state-redesign/' \
     && echo "OVERREACH: Phase 1 spec edits" && exit 1 \
     || echo "Phase 1 specs untouched"
   ```

3. **Spec pack inventory.**
   ```
   ls -1 specs/orun-component-catalog/*.md | sort
   ls -1 specs/orun-component-catalog/_archive/
   ```
   Expect 12 top-level docs + the `_archive/` directory with the
   monolith + README. Confirm root-level `orun-catalog-full-design.md`
   is gone (`ls orun-catalog-full-design.md` → not found).

4. **Internal link check.** From the spec pack:
   ```
   rg -n '\]\([^)]*\.md[^)]*\)' specs/orun-component-catalog/ \
     | awk -F: '{print $1, $3}' \
     | while read src link; do
         tgt=$(echo "$link" | sed 's/.*(\([^)]*\)).*/\1/')
         case "$tgt" in
           http*|'#'*|/*) ;;
           *) [ -f "$(dirname "$src")/$tgt" ] || echo "broken: $src -> $tgt" ;;
         esac
       done
   ```
   Any unresolved sibling reference is a blocker (or a proposal,
   verifier's call based on materiality).

5. **`agents/orchestrator.md` consistency.** Confirm:
   - `specs/orun-component-catalog/` named as the active spec.
   - Milestones C0–C9 referenced (search for `C0` … `C9`).
   - `specs/orun-state-redesign/` demoted to "predecessor (Phase 1,
     COMPLETE — M0–M6 merged)" everywhere.
   - "Deferred Decision Protocol" section exists, with the canonical
     wording: "Deferred is not blocked. The loop must keep producing
     PR-sized work whenever any human-independent candidate exists,
     even if multiple candidates are deferred awaiting input."
   - Embedded `state.json` template includes `active_spec`,
     `active_milestone`, `phase_history`.

6. **`ai/state.json` shape.** `jq .` on the file:
   - `current_task == "0022"`
   - `active_spec == "specs/orun-component-catalog"`
   - `active_milestone == "C0"`
   - `phase_history.phase_1_orun_state_redesign.status == "COMPLETE"`
   - `phase_history.phase_1_orun_state_redesign.milestones == "M0–M6"`
   - `phase_history.phase_1_orun_state_redesign.coverage_floors`
     contains `internal/statestore: "95.7%"`,
     `internal/revision: "90.3%"`,
     `internal/executionstate: "90.0%"`,
     `internal/triggerctx: "passes"`.
   - `waiting_for_input == "false"`.
   - `task_agent` is a path (not a role string) — likely
     `ai/reports/task-0022-implementer.md` post-flip.

7. **Context-file diff vs. main.**
   ```
   git diff main...verify/task-0022 -- ai/context/task-ledger.md \
     | grep -E '^-[^-]' | head
   ```
   Removed Phase 1 ledger lines should be ZERO (Phase 1 entries
   must be preserved verbatim; only additions allowed). If any
   `^-` lines beyond standard re-flow show up under tasks 0001–0021,
   that's a blocker.

8. **Local quality gates.** On `verify/task-0022`:
   ```
   go build ./...
   go vet ./...
   go test ./... -count=1 -timeout 600s
   make test-state-redesign
   ```
   Sanity check — these should be green because no Go changed.
   Coverage floors must hold (statestore ≥ 95 %, revision ≥ 90 %,
   executionstate ≥ 90 %, triggerctx passes). A regression here
   would imply spec/docs CI gating affected Go (extremely unlikely
   on this PR shape — investigate if observed).

9. **PR CI log review.** `gh run view 26703506317 --log | tail -200`
   for `CI / Orun Plan` — confirm the Orun walk actually ran and
   exited 0; no logged secrets. The matrix leg SKIPPED is expected
   for a docs-only diff (same shape as M5.a–c, M6 PRs); confirm
   the skip reason is the empty-matrix path, not a workflow error.

10. **Secret-handling sweep.**
    ```
    git diff main...verify/task-0022 \
      | rg -i '(token|secret|api[_-]?key|password|bearer)\s*[:=]\s*[\"a-z0-9_-]{12,}'
    ```
    Expect zero hits.

11. **Spec drift check.** If any of the 12 docs reference a flag,
    package, or behavior that contradicts another doc in the same
    pack (e.g. `cli-surface.md` vs `compatibility-and-migration.md`
    on `--catalog-strict` semantics), file
    `ai/proposals/task-0022-spec-update.md` and decide PASS-with-
    followup vs FAIL based on materiality.

12. **`kiox -- orun validate / plan / run --dry-run`.** The persistent
    composition-cache quirk on this developer machine is documented
    (see `ai/context/current.md` → Known Spec Drift). CI is
    authoritative. Note observation in the verifier report's "Local
    Resource Evidence" section if the quirk reproduces.

## Acceptance Criteria

✅ Diff stays within PR Boundary above (docs + bookkeeping only;
   zero Go; Phase 1 specs untouched; prior reports/tasks untouched).
✅ Spec pack contains all 12 named documents and the `_archive/`
   pair; root-level monolith removed.
✅ `agents/orchestrator.md` Phase 2 rewrite is internally consistent
   (active spec, milestones, demoted Phase 1 reference, Deferred
   Decision Protocol present).
✅ `ai/state.json` carries the Phase 2 rotation with the
   `phase_history.phase_1_orun_state_redesign` block recording
   coverage floors verbatim.
✅ `ai/context/current.md` and `ai/context/task-ledger.md` reflect
   Phase 2 cleanly without rewriting Phase 1 history.
✅ `go build / vet / test / make test-state-redesign` green on PR
   head; Phase 1 coverage floors preserved.
✅ Required PR CI check (`CI / Orun Plan`) PASS at log level on
   final head SHA `1016a2b`. Matrix-leg SKIP legitimate.
✅ No secrets / tokens in the diff.
✅ MergeStateStatus stays `CLEAN` through final push.

If any check fails, FAIL the verification and leave PR #167 open
with explicit blockers in the verifier report.

## Verifier Merge Protocol (per `agents/orchestrator.md` §485–500)

- If PASS:
  1. Optionally commit the verifier report to the PR branch:
     ```
     git checkout impl/task-0022-phase2-rollover
     git pull --ff-only origin impl/task-0022-phase2-rollover
     git add ai/reports/task-0022-verifier.md
     git commit -m "Task 0022 verifier report"
     git push
     ```
     Wait for CI to re-run; confirm SUCCESS at log level.
  2. `gh pr merge 167 --squash --delete-branch`.
  3. `git checkout main && git pull --ff-only origin main`.
  4. `git status --short` — resolve any verifier-created local
     changes before ending the verifier task.
  5. Update orchestration state files:
     - `ai/state.json`: `completed` += `"0022"`; advance
       `current_task` to `"0023"`; `task_agent` →
       `ai/reports/task-0022-verifier.md` (path, not role);
       `next_focus` → `"orun-component-catalog Milestone C0 code half — internal/catalogmodel + internal/sourcectx skeleton"`;
       `last_verified` → today's date.
     - `ai/context/current.md`: roll forward to Task 0023 frame
       (C0 code half), keep Phase 2 milestone table, move Task 0022
       into a "Just Completed" / "Last Verified" callout.
     - `ai/context/task-ledger.md`: append `## Task 0022 (Verifier
       pass)` below the existing Task 0022 entry — Status: PASS +
       merged, PR #167 squash commit SHA, CI evidence.
     - `ai/waiting_for_input.md`: keep as "no input currently
       requested" (Phase 2 has no human-decision blockers).
- If FAIL: leave PR #167 open. Do not merge. Document blockers in
  the verifier report and (only if every roadmap candidate is
  blocked, which is NOT the case here — Task 0023 is human-
  independent) `ai/waiting_for_input.md`. Otherwise let the
  orchestrator route a follow-up implementer fix-task on the same
  PR/branch.
- Never merge a PR with unresolved verification blockers or failing
  CI on final head SHA.

## When Done Report

Save to `ai/reports/task-0022-verifier.md` with sections (concise —
see `agents/orchestrator.md` budget):

- **Result:** PASS | FAIL
- **Checks** — bullets matching Verification Steps (1–12).
- **Issues** — blockers and non-blocking concerns; severity per item.
- **Risk Notes** — residual risks (e.g. monolith archived not
  deleted, deferred decision protocol newly introduced wording).
- **Spec Proposals** — `ai/proposals/task-0022-spec-update.md` link
  if any spec drift observed; otherwise "none".
- **CI Log Review** — `gh run view 26703506317 --log` evidence;
  matrix-leg skip-reason confirmation.
- **Local Resource Evidence** — `go build / vet / test`,
  `make test-state-redesign` outputs (coverage floors verbatim).
- **Recommended Next Move** — on PASS: emit Task 0023 (C0 code
  half). On FAIL: explicit blocker list and proposed implementer
  fix scope.
