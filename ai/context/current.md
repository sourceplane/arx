# Current Roadmap Position

## Active Spec
`specs/orun-state-redesign/` (Phase 1, local-only) — trigger-first revision-first
local state model. See `specs/orun-state-redesign/README.md` for the index and
read order.

## Active Milestone
**M3 — `internal/revision`**. M2 closed at PR #156 (`cd8b3e8`, verified PASS).
M3 PR-A (Task 0007 model + keys + writer skeleton) verified PASS and
squash-merged via PR **#157** at main commit `7f1e53d` on 2026-05-30T03:33:16Z.
**Next: M3 PR-B (Task 0010).**

## Last Completed Task (0009 — M3 PR-A Verifier)
- Result: **PASS** — PR #157 squash-merged → `main` `7f1e53d` at 2026-05-30T03:33:16Z; branch `impl/task-0007-m3-revision-pra` deleted.
- Verifier report: `ai/reports/task-0009-verifier.md`.
- Coverage on `internal/revision` 93.3 % (gate ≥ 90 %); `internal/statestore` 96.1 % (no regression).
- Both required CI checks SUCCESS at log level on the verifier-side commit (`Orun Plan` 59 s, `Harness dry-run guard` 19 s).
- Claim-first ordering deviation from `cli-surface.md` §1.2 ACCEPTED in-place; no spec proposal filed (rationale in verifier report).
- Risk Notes carried forward: `stateStoreVersionPath()` lives in `internal/revision` (deferred relocation); `writeCompatibilityMirror` is a no-op stub gated by default-true `Config.CompatibilityWrites` (M5 fills the body); `RevSummary.JobCount` = 0 in PR-A (PR-B threads it via `WriteManifest`).

## Last Implementer Task (0008 — delivery chore)
- Implementer prompts: `ai/tasks/task-0007.md`, `ai/tasks/task-0008.md`
- Implementer reports: `ai/reports/task-0007-implementer.md` (PR #157, coverage 93.3 % on `internal/revision`),
  `ai/reports/task-0008-implementer.md` (chore-only delivery; no production-code changes beyond Task 0007's tree)
- Branch `impl/task-0007-m3-revision-pra` pushed; PR **#157** opened.
  Implementer-side commits: `96621ed` (Task 0007 tree), `500218c` (PR-number
  backfill into reports).
- Outstanding flag for verifier: claim-first ordering deviation from
  `cli-surface.md` §1.2 step-7 (index slot reserved via `CreateIfAbsent`
  BEFORE body writes; rationale in
  `ai/reports/task-0007-implementer.md` § "Step-Order Deviation From
  cli-surface.md §1.2"). Implementer also flagged `version.json` helper
  location as a defer-or-relocate decision.

## Repo Checkpoint

| Attribute | Value |
|---|---|
| Branch (local checkout) | `main` (clean post-merge) |
| `main` tip | `7f1e53d` — Task 0007: M3 PR-A — internal/revision model + keys + writer skeleton (#157) |
| Open PRs (state-redesign lineage) | None |
| Repo health | 🟢 Green — M3 PR-A on main; ready for Task 0010 |
| Last verified | 2026-05-30 (Task 0009, PR #157) |
| Active milestone | M3 (`internal/revision`) — PR-A merged; PR-B next |
| Tasks completed | 0001, 0002, 0003, 0004, 0005, 0007, 0008, 0009 (8 total) |
| Current task | None (awaiting Task 0010 emission) |

## Roadmap (M0 → M6)
1. ✅ **M0 Foundation** — landed on main at `4ea1980` (PR #152).
2. ✅ **M1 `internal/triggerctx`** — landed on main at `db342dd` (PR #153).
3. ✅ **M2 `internal/statestore`** — closed at PR #156 (`cd8b3e8`, 2026-05-30).
   - PR A — frozen interface + local driver non-CAS ops (PR #154 → `9b0a39c`)
   - PR B — `CompareAndSwap`, `List`, atomicity/exclusivity/CAS/`rapid` property suite (PR #155 → `0fa2111`)
   - PR C — typed refs/indexes marshallers + `RebuildIndexes()` stub (PR #156 → `cd8b3e8`)
4. **M3 `internal/revision`** ← current
   - ✅ PR-A — model + keys + writer skeleton (PR #157 → `7f1e53d`, verified PASS via Task 0009 on 2026-05-30)
   - PR-B — `ResolveRevision` seven-branch resolver + `WriteManifest` + legacy `.orun/plans/**` mirror body (Task 0010, next)
5. M4 `internal/executionstate` + runner bridge
6. M5 CLI rewire (`orun plan/run/status/logs/describe/get plans` + hidden `state migrate`)
7. M6 End-to-end + property gates

## Next Task After 0009 (proposed)
**Task 0010 — M3 PR-B Implementer.** Adds `manifest.go`
(`WriteManifest`, `UpdateLatestExecutionSummary`), `resolver.go`
(`ResolveRevision` seven-branch resolver per
`compatibility-and-migration.md` §3), and promotes the `// TODO(m5)`
compatibility-mirror stub in `writer.go` to a real conditional write of
the legacy `.orun/plans/<checksum>.json` body gated by
`Config.CompatibilityWrites`. Coverage gate stays ≥ 90 %; resolver test
matrix MUST cover all 7 resolution branches; compat-writes flag MUST
exercise both true and false paths. Branch suggestion:
`impl/task-0010-m3-revision-prb`.

If Task 0009 verifier rejects the claim-first ordering and files
`ai/proposals/task-0007-spec-update.md`, the orchestrator will accept,
revise, defer, or query the user about it before generating Task 0010 —
and may interpose a small spec-update task between 0009 and 0010.

## Known Spec Drift / Open Questions
- **Claim-first vs `cli-surface.md` §1.2 step-7 ordering.** Task 0009 verifier
  ACCEPTED the deviation in-place (no proposal filed). Rationale:
  `CreateIfAbsent` is the only exclusive primitive in `state-store.md` §3,
  claim-first is the only ordering producing distinct revision keys before any
  body write occurs under concurrent `(TriggerKey, planHash)` duplicates, refs
  still land after bodies preserving `state-store.md` §6 crash-recovery, and
  `cli-surface.md` §1.2 is a high-level descriptive flow not a normative
  atomicity proof. RESOLVED.
- **`version.json` helper location.** Task 0009 verifier deferred relocation —
  `stateStoreVersionPath()` stays in `internal/revision`. If M5 migration
  tooling needs a statestore-side helper, that PR can lift the constant up.
  RESOLVED.
- **Half-shipped delivery anti-pattern.** Task 0007 was the first observed
  case of "implemented locally" with no PR (see `open-risks.md` R-005).
  Task 0008 corrective chore landed PR #157; future implementer prompts
  should keep the `PR Creation Requirement` section explicit and acceptance
  criteria should include a `gh pr list --head <branch>` check returning a
  non-empty array.
