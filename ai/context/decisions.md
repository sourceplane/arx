# Decisions

| Date | Decision | Rationale | Reference |
|---|---|---|---|
| 2026-05-30 | Pivot active spec from `.kiro/specs/orun-tui-cockpit/` to `specs/orun-state-redesign/` (Phase 1, local-only). | New authoritative engineering design pack landed; trigger-first revision-first local state model unblocks Phase 2 (R2/S3 driver) and Phase 3 (DO + Supabase coordination). TUI cockpit work paused until M5 lands. | `agents/orchestrator.md` (local edit), `specs/orun-state-redesign/README.md` |
| 2026-05-30 | Task numbering restarts at 0001 under the new spec lineage. | Clean ledger for the new spec; prior TUI cockpit lineage (0139–0147 + sub-tasks) preserved in git history. | `ai/context/task-ledger.md` |
| 2026-05-30 | Phase 1 is strictly local-only. R2/S3/Supabase/DO are deferred to Phase 2/3. | Matches `specs/orun-state-redesign/README.md` Phase boundaries table. | `specs/orun-state-redesign/README.md` |
| 2026-05-30 | Implementer agents have latitude to split or merge PRs within a milestone, as long as each PR stays reviewable and dependencies are respected. The Orchestrator names the milestone, not sub-task numbers. | Spec policy in `implementation-plan.md` and orun-saas-orchestration skill (flexible-spec-layout-and-milestones reference). | `specs/orun-state-redesign/implementation-plan.md` |
