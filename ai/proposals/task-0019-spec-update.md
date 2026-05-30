# Spec Proposal — Task 0019 follow-up

**Filed:** 2026-05-30
**Source:** Verifier walk against PR #163 (Task 0019, M5.c)
**Severity:** Low (documentation / resolver-arg semantics; non-blocking)
**Decision:** PASS-with-followup — surface for M5.d

## Drift summary

`specs/orun-state-redesign/cli-surface.md` §5.1 documents the canonical
describe aliases as:

```
orun describe revision latest
orun describe revision <revKey>
orun describe trigger latest
orun describe trigger <triggerName>
orun describe execution <execKey>
```

Behavior on PR #163 head (`fb364f1`) in a real workspace:

| Command | Outcome |
|---|---|
| `describe revision` (no arg) | ✓ resolves latest revision |
| `describe revision <revKey>` | ✓ resolves by exact revision key |
| `describe revision latest` | ✕ `revision: ambiguous or unknown run target: "latest"` |
| `describe trigger` (no arg) | ✓ renders trigger of latest revision |
| `describe trigger <revKey>` | ✓ renders trigger by revision-key arg |
| `describe trigger latest` | ✕ `revision: ambiguous or unknown run target: "latest"` |
| `describe trigger <triggerName>` | ✕ `revision: ambiguous or unknown run target: "system.manual"` |
| `describe execution latest` | ✓ resolves latest execution |
| `describe execution <execKey>` | ✓ resolves by exact execution key |

Two distinct gaps:

1. **`latest` literal** — `revision.ResolveRevision` does not translate the
   sentinel string `"latest"` to "use `refs/latest-revision.json`". Empty-arg
   works (branch 1 of the seven-branch resolver), but the spec uses the
   literal `latest`. Same gap on the `describe trigger latest` path.
2. **`<triggerName>` lookup on `describe trigger`** — the CLI wires
   `describe trigger <arg>` through `revision.ResolveRevision(arg, …)`, which
   does not understand a trigger name (e.g. `system.manual`,
   `github-pull-request`) as a lookup key. Spec §5.1 documents this as
   canonical syntax.

## Root cause

CLI is faithful — it routes `describe trigger <ref>` and `describe revision <ref>`
straight through `revision.ResolveRevision(ref, …)`. The resolver's
seven-branch ladder (compat §3) is keyed on revision identifiers — empty,
revision key, named-ref, plan-path, plan-checksum, component-path — with no
`"latest"` alias and no trigger-name branch.

The implementer report does not flag this; the new tests use empty-arg
(`describeRevision("")`) which short-circuits to branch 1 and bypasses the
gap entirely.

## Proposed remediation (out of scope for M5.c)

Two options for M5.d (or a small hotfix task on the same PR slot):

**Option A — CLI-side normalization** (smaller blast radius)
- In `cmd/orun/command_describe.go::describeRevision` and `describeTrigger`,
  normalize `ref == "latest"` → `ref = ""` before calling
  `revision.ResolveRevision`.
- For `describeTrigger`, add a trigger-name lookup pass before falling
  through to the resolver: read `refs/triggers/<name>/latest.json` (the
  writer side already lands these per cli-surface.md §1 line 57).

**Option B — Resolver-side branch** (cleaner; matches the resolver pattern)
- Add a branch to `revision.ResolveRevision` that recognizes `arg == "latest"`
  and routes to branch 1 (latest-revision ref).
- Add a trigger-name resolution branch that reads
  `refs/triggers/<name>/latest.json` and follows the embedded revision key.
- Document in `compatibility-and-migration.md` §3 the expanded resolution
  chain.

Recommendation: **Option A** for the `latest` literal (trivial, CLI-local),
**Option B** for the trigger-name lookup (touches resolver semantics that
will also affect future cockpit/API consumers).

## Why not a verifier-side fix

The verifier prompt allows a tiny verification-only fix when essential.
This drift is non-blocking: the four read commands all work via the documented
empty-arg / revision-key / execution-key syntaxes that the new tests cover.
The `latest` literal and `<triggerName>` lookup are refinements rather than
breakage, and they cross a resolver boundary that should be reviewed by an
implementer with full M3/M4 context. M5.d (`orun state migrate`) is the
natural follow-up slot.

## Suggested next move

Track as a small refinement in M5.d's task scope (or a discrete Task 0021
preamble), citing this proposal. No change to `cli-surface.md` is required —
the spec is the authority and the implementation should catch up.
