No human input currently requested. **M5.a is closed**:

- PR #161 verified PASS by Task 0016 verifier (single-pass closure) and
  squash-merged into `main` as `7a9c494` on 2026-05-30T12:31:56Z.
- Required CI both PASS at log level on final head SHA after verifier-side
  commit `01e75bd`: `CI / Orun Plan` run `26683860043`; `Harness dry-run
  guard` run `26683860052`.
- Coverage: `internal/statestore` 95.7 %, `internal/revision` 90.4 %,
  `internal/executionstate` 90.0 % (exact floor held — package not touched
  in M5.a).
- Verifier report: `ai/reports/task-0016-verifier.md`.

`orun plan` now writes the canonical revision-first layout end-to-end via
`internal/triggerctx` + `internal/revision.WriteRevision`, embeds
`metadata.trigger` + `metadata.revision`, retains byte-identical compat
aliases, preserves `-o`, and emits the new §1.1 summary block. Repo health
🟢 green; `main` clean at `7a9c494`; no open state-redesign PRs.

Active milestone: **M5** (CLI rewire), 1/4 slices closed. Next emission per
`specs/orun-state-redesign/implementation-plan.md` §M5.b will be **Task 0018
= M5.b (`orun run` rewire + bridge wiring + `--revision` flag) implementer**.
Suggested PR scope: 1 PR covering `orun run` only — keep `orun status` /
`logs` / `describe` / `get plans` (M5.c) and hidden `orun state migrate`
(M5.d) out of scope. Branch base: `main` @ `7a9c494`. M5.b must also pin the
`bridge-mirror-failed` payload schema in `data-model.md` §9 before any
second consumer.

No outstanding orchestrator questions; no blocked verifier or implementer
work. `ai/deferred.md` is empty.
