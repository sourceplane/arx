# Task 0038 — Verifier Report: C5 PR-2 Catalog Read Surface

**Spec:** specs/orun-component-catalog · **Milestone:** C5 · **PR:** #177 (squash `3811d18`)
**CI hotfix:** PR #178 (squash `53ec66a`) · **Verdict:** ✅ **PASS** (single-pass closure)
**Date:** 2026-06-01 · **Verifier:** inline (same session as implementer)

---

## 1. Scope verified

Task 0038 delivered the read side of the `orun catalog` CLI per cli-surface.md:

- `orun catalog list` (§3) — enumerate components from a resolved catalog
- `orun catalog describe <component>` (§4) — single-component detail, exit 4 on ambiguous
- `orun catalog tree` (§5) — dependency-graph walk with cycle protection
- `orun catalog history <component>` (§7) — read-only event listing (append is C7)
- `orun catalog validate` (§6/§8) — structural validation; `--rebuild-indexes` documented no-op (C8 reserved)
- `orun catalog diff` (§6) — registered stub, exit 5 (not-implemented; full impl is C8)

All gated through the §11 JSON envelope (`{apiVersion, kind, data, warnings}`) and the shared
`parseCatalogSelector → catalogstore.RefSelector` bridge introduced in PR-1.

## 2. Inspection findings (Phase 1–3)

Read-path review of all six command files + `catalog.go` root/envelope/selector bridge,
`reads.go` seam, and `main.go` exit-code plumbing:

- Exit-code contract correct: 0 success/reused · 1 validation · 2 resolver internal · 3 statestore ·
  4 ambiguous (describe) · 5 diff stub · 6 missing component. `main.go` unwraps via
  `errors.As` on `interface{ ExitCode() int }`.
- Determinism: list/tree/history sort outputs; tree walk has cycle protection (visited set).
- Read-only guarantee: `history` lists events, never appends. `validate --rebuild-indexes` is a
  documented no-op (no rebuild behavior wired) — complies with "do not wire it here" intent.
- Architecture rule held: `internal/catalogstore/reads.go` avoids `os`/`io`/`filepath` (no-raw-FS);
  CLI layer may use `filepath`. `ReadCatalogGraph`/`CatalogGraphPath` take BARE graph kind.
- Immutability: catalog objects byte-identical on idempotent re-read.

Coverage (pre-merge): catalogstore 90.7%, catalogresolve 90.9%, sourcectx 91.1% — all floors held.
cmd/orun package coverage at long-standing baseline (no enforced floor); new seams + refresh→refs
and read-surface E2E covered.

## 3. Gates (Phase 2/3)

Local: `go build ./...`, `go vet`, `go test -race -count=1` (catalogstore + cmd/orun) all green.
PR #177 CI green on HEAD `ce8e3d3` (mergeStateStatus CLEAN, all required checks pass).

## 4. Merge (Phase 5–6)

Merged #177 via `--squash --delete-branch` at **`3811d18`**; source branch
`task-0038-catalog-cli-c5-pr2-read-surface` deleted. Local main ff-only synced.

## 5. Post-merge CI triage + R-008 resolution (Finish-CI)

Post-merge main CI on `3811d18` showed **`state-redesign-tests` FAILED**:
`internal/executionstate` measured **89.6%** vs its zero-margin **90.0%** floor — a *coverage* gate
flap, not a logic failure (`CI` and `orun remote-state conformance` both passed). This is the
documented **R-008** carry-forward (the same floor also flapped once on PR #175 CI).

Diagnosis: locally `executionstate` measured a deterministic 90.0% (3× with `-race`), confirming an
environmental delta on identical source against a zero-margin floor — not a regression introduced by
task-0038 (which touches no Phase-1 package).

Resolution (per R-008 plan; floors must not be lowered): added **PR #178** with four deterministic
buffer tests over genuinely-reachable branches —
`scanForNextRunSeq` non-run-key skip + max tracking, `listExecutionKeys` multi-key dedup,
`resolveExactByIndex` + `resolvePrefixScan` stale-index error wraps. Lifted coverage to **90.6%**
(+0.6pp headroom, stable across 3 `-race` runs). Test-only; no production code touched.

Merged #178 via `--squash --delete-branch` at **`53ec66a`**. Post-merge main CI on `53ec66a` fully
green: `state-redesign-tests` ✅ · `CI` ✅ · `orun remote-state conformance` ✅. **R-008 CLOSED.**

## 6. Verdict

✅ **PASS.** Read surface complete and correct; all coverage floors held; main CI green at `53ec66a`.

## 7. Carry-forward

- **R-007 (still open):** `CatalogLocalIndexes` owner/system/domain/type axes remain empty
  (data-model §9 under-specifies). Do not block future PRs on them; resolve when the spec firms up.
- **R-008 (CLOSED):** executionstate zero-margin floor given +0.6pp headroom via buffer tests.
