No human input currently requested. Milestone M3 (`internal/revision`) is the
active milestone. M3 PR-A (Task 0007 — model + keys + writer skeleton) was
delivered into PR **#157** by the corrective Task 0008 chore. PR #157 is
OPEN, MERGEABLE, mergeStateStatus CLEAN, head SHA `500218c`, with both
required CI checks SUCCESS (`CI / Orun Plan` run `26672937657`,
`Harness dry-run guard` run `26672937641`).

Current orchestrator decision (2026-05-30): emit **Task 0009 = M3 PR-A
verifier** at `ai/tasks/task-0009-verifier.md`. Verifier validates Task 0007
against `specs/orun-state-redesign/implementation-plan.md` Milestone M3
"Done when" criteria (≥90 % coverage, revision-key uniqueness +
collision-suffix property test, leaf-clean imports, all exported symbols
documented), adjudicates the claim-first ordering deviation from
`cli-surface.md` §1.2 step-7 (accept-and-document inline OR file
`ai/proposals/task-0007-spec-update.md`), inspects PR CI logs at log level,
and merges per the Verifier Merge Protocol on PASS.

Next implementer after Task 0009 = **Task 0010 = M3 PR-B implementer**
(`manifest.go` + `resolver.go` seven-branch resolver + legacy
`.orun/plans/<checksum>.json` mirror body promoting the `// TODO(m5)` stub
to a real conditional write gated by `Config.CompatibilityWrites`).
