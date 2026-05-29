Below is an implementation-ready **Phase 1 design** you can give to an agent. It keeps the first phase **local-only**, but makes the storage contract directly portable to R2/S3/remote backends later.

The design builds on Orun’s current model: trigger bindings already map provider events to environment activation and plan scope, the plan is already the immutable artifact consumed by runtime, and current local execution state already supports resume, retry, and logs under `.orun/executions/{exec-id}`.   

# Phase 1: Trigger-first local revision state model

## 1. Goal

Implement a new local state model where **every plan has a trigger context**, every trigger context produces a **PlanRevision**, and executions are stored under the revision.

Target lineage:

```text
TriggerOccurrence
  → PlanRevision
      → ExecutionRun
          → JobRun
              → JobAttempt
                  → StepAttempt
```

This replaces the current mental split:

```text
.orun/plans/{checksum}.json
.orun/executions/{exec-id}/...
```

with a revision-first layout:

```text
.orun/revisions/{revisionKey}/plan.json
.orun/revisions/{revisionKey}/executions/{executionKey}/...
```

The current docs say `orun plan` stores generated plans under `.orun/plans/{checksum}.json` and `.orun/plans/latest.json`, while `orun run` creates execution records under `.orun/executions/{exec-id}`.   Phase 1 should keep compatibility with those paths where needed, but introduce the new canonical layout.

## 2. Phase 1 scope

### In scope

Implement locally:

```text
1. TriggerOccurrence model
2. Built-in system triggers for non-trigger runs
3. PlanRevision model
4. Revision-first `.orun/revisions/` layout
5. ExecutionRun under PlanRevision
6. Short readable revision keys
7. Local StateStore abstraction
8. Backward-compatible plan/run/status/log resolution
9. Basic refs and indexes
10. Migration-compatible read path from old `.orun/plans` and `.orun/executions`
```

### Out of scope for Phase 1

Do **not** implement yet:

```text
R2/S3 object store driver
SaaS auth
Supabase metadata sync
DO coordination
remote state migration
cross-plan reuse/evidence
distributed object-store locking
```

But design the interfaces and layout so those can be added without changing the file structure.

## 3. Core concepts

### 3.1 TriggerBinding

Static rule declared in `intent.yaml` under `automation.triggerBindings`.

Existing behavior remains:

```text
Provider event → TriggerBinding match → Environment activation → Plan
```

When `--trigger` or `--from-ci` is used, Orun already normalizes the event, matches trigger bindings, activates referenced environments, and plans only subscribed components. 

### 3.2 TriggerOccurrence

A runtime trigger context. Every `orun plan` must produce exactly one `TriggerOccurrence`.

TriggerOccurrence can come from:

```text
declared trigger       --trigger github-pull-request
provider event         --from-ci github --event-file ...
system manual          orun plan
system manual changed  orun plan --changed
system replay          orun run revision/<id>
system api             future SaaS/API-created plan
```

### 3.3 PlanRevision

Immutable compiled execution contract produced from:

```text
intent + components + compositions + trigger context + CLI scope
```

The plan remains the compiled artifact. The docs already define the plan as the artifact produced by `orun plan` and consumed by `orun run`, with concrete jobs and execution DAG fields. 

### 3.4 ExecutionRun

One execution of a PlanRevision. Multiple executions can exist under the same revision.

Examples:

```text
run-001      first execution
run-002      rerun same revision
run-003      retry execution
```

### 3.5 JobRun, JobAttempt, StepAttempt

Phase 1 can keep the existing `state.json` behavior internally, but the new layout should leave room for:

```text
jobs/{jobKey}/job-run.json
jobs/{jobKey}/attempts/{attempt}/attempt.json
jobs/{jobKey}/attempts/{attempt}/steps/{stepKey}.json
```

The current state model already supports resumable execution, job-level retry, immutable logs, and parallel-safe exec IDs.  Phase 1 should preserve those behaviors.

## 4. ID and key model

Use short, readable storage keys. Do **not** put the full traceback into folder names.

### 4.1 Trigger key

Format:

```text
trg-<scope>-<shortSha>
```

Examples:

```text
trg-pr139-def456a
trg-main-a91ff00
trg-tag-v1-2-0-a91ff00
trg-manual-def456a
trg-local-dirty
trg-changed-def456a
```

The full trace goes in `trigger.json`.

### 4.2 Revision key

Format:

```text
rev-<scope>-<shortSha>-p<planHash8>
```

Examples:

```text
rev-pr139-def456a-p8f31c09
rev-main-a91ff00-pb91e72f
rev-v1-2-0-a91ff00-p31df90
rev-manual-def456a-p9aa021d
rev-local-dirty-p7c81a21
```

Recommended fields:

```text
scope       pr139 | main | v1-2-0 | manual | local-dirty
shortSha    7 chars from source revision, or dirty/worktree marker
planHash    8 chars from compiled plan checksum
```

If collision occurs:

```text
rev-pr139-def456a-p8f31c09
rev-pr139-def456a-p8f31c09d4e2
rev-pr139-def456a-p8f31c09d4e2-x2
```

### 4.3 Execution key

Inside a revision, execution keys can be short:

```text
run-001
run-002
run-003
```

For CI-pinned executions, allow user-provided key:

```text
gha-26562411779-1
```

But normalize it to a filesystem-safe key.

### 4.4 Machine IDs

Each object also gets a stable machine ID in JSON:

```text
triggerId:   trg_01J...
revisionId:  rev_01J...
executionId: exec_01J...
```

The folder key is for humans and storage. The JSON ID is for APIs and future SaaS.

## 5. Canonical local file layout

Implement this as the new canonical layout:

```text
.orun/
  version.json

  refs/
    latest-revision.json
    latest-execution.json

    triggers/
      system.manual/
        latest.json
      system.manual-changed/
        latest.json
      github-pull-request/
        latest.json
        pr-139.json
      github-push-main/
        latest.json
        branch-main.json
      github-tag-release/
        latest.json
        tag-v1-2-0.json

    named/
      release-candidate.json

  indexes/
    revisions/
      rev-pr139-def456a-p8f31c09.json
    executions/
      run-001.json
      gha-26562411779-1.json

  revisions/
    rev-pr139-def456a-p8f31c09/
      manifest.json
      trigger.json
      revision.json
      plan.json

      executions/
        run-001/
          execution.json
          snapshot.latest.json
          state.json              # compatibility summary
          metadata.json           # compatibility summary

          jobs/
            j-a8f31c09/
              job-run.json
              attempts/
                1/
                  attempt.json
                  steps/
                    s-setup.json
                    s-validate.json
                  logs/
                    setup.log
                    validate.log

          logs/
            j-a8f31c09/
              setup.log
              validate.log

          events/
            000000001-execution-created.json
            000000002-job-started.json
            000000003-step-started.json
            000000004-step-completed.json
            000000005-job-completed.json

          artifacts/
            j-a8f31c09/
              outputs.json
              summary.json
```

### Why job keys should be hashed

Plan job IDs can contain `@`, `.`, slashes, or future delimiters. For local FS and R2/S3 compatibility, create safe keys:

```text
jobKey = "j-" + shortHash(jobID)
stepKey = "s-" + slug(stepID)
```

`job-run.json` stores the original job ID.

Example:

```json
{
  "jobKey": "j-a8f31c09",
  "jobId": "api@dev.validate",
  "component": "api",
  "environment": "dev",
  "status": "completed"
}
```

## 6. JSON contracts

### 6.1 `version.json`

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "StateStoreVersion",
  "layout": "revision-first",
  "version": 1,
  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 6.2 `trigger.json`

For declared CI trigger:

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "TriggerOccurrence",

  "triggerId": "trg_01JABC",
  "triggerKey": "trg-pr139-def456a",
  "triggerType": "declared",
  "triggerName": "github-pull-request",

  "mode": "event-file",
  "provider": "github",
  "event": "pull_request",
  "action": "synchronize",

  "matchedBindings": ["github-pull-request"],

  "source": {
    "repo": "sourceplane/orun",
    "ref": "refs/pull/139/head",
    "sourceScope": "pr-139",
    "headRevision": "def456a1b2c3...",
    "baseRevision": "abc1239f8e7d...",
    "workingTree": "clean"
  },

  "planScope": {
    "mode": "changed",
    "base": "abc1239f8e7d...",
    "head": "def456a1b2c3...",
    "activeEnvironments": ["development"],
    "changedComponents": ["api-edge-worker"]
  },

  "createdAt": "2026-05-29T00:00:00Z"
}
```

For default non-trigger local run:

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "TriggerOccurrence",

  "triggerId": "trg_01JLOCAL",
  "triggerKey": "trg-manual-def456a",
  "triggerType": "system",
  "triggerName": "system.manual",

  "mode": "manual",
  "provider": "orun",
  "event": "manual",
  "action": "plan",

  "matchedBindings": ["system.manual"],

  "source": {
    "repo": "sourceplane/orun",
    "ref": "refs/heads/main",
    "sourceScope": "branch-main",
    "headRevision": "def456a1b2c3...",
    "workingTree": "clean"
  },

  "planScope": {
    "mode": "full",
    "activationMode": "all-environments"
  },

  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 6.3 `revision.json`

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "PlanRevision",

  "revisionId": "rev_01JABC",
  "revisionKey": "rev-pr139-def456a-p8f31c09",

  "triggerId": "trg_01JABC",
  "triggerKey": "trg-pr139-def456a",

  "planHash": "sha256:8f31c09d4e2...",
  "planShortHash": "8f31c09",

  "source": {
    "repo": "sourceplane/orun",
    "headRevision": "def456a1b2c3...",
    "baseRevision": "abc1239f8e7d..."
  },

  "summary": {
    "jobCount": 12,
    "scope": "changed",
    "activeEnvironments": ["development"],
    "changedComponents": ["api-edge-worker"]
  },

  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 6.4 `manifest.json`

This is the human and tool entrypoint.

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "RevisionManifest",

  "revision": {
    "id": "rev_01JABC",
    "key": "rev-pr139-def456a-p8f31c09",
    "planHash": "sha256:8f31c09d4e2...",
    "createdAt": "2026-05-29T00:00:00Z"
  },

  "trigger": {
    "id": "trg_01JABC",
    "key": "trg-pr139-def456a",
    "type": "declared",
    "name": "github-pull-request",
    "provider": "github",
    "event": "pull_request",
    "action": "synchronize",
    "scope": "changed"
  },

  "source": {
    "repo": "sourceplane/orun",
    "sourceScope": "pr-139",
    "headRevision": "def456a1b2c3...",
    "baseRevision": "abc1239f8e7d..."
  },

  "summary": {
    "jobCount": 12,
    "activeEnvironments": ["development"],
    "latestExecutionKey": "run-001",
    "latestExecutionStatus": "completed"
  },

  "objects": {
    "plan": "plan.json",
    "trigger": "trigger.json",
    "revision": "revision.json"
  }
}
```

### 6.5 `execution.json`

```json
{
  "apiVersion": "orun.io/v1alpha1",
  "kind": "ExecutionRun",

  "executionId": "exec_01JXYZ",
  "executionKey": "run-001",

  "revisionId": "rev_01JABC",
  "revisionKey": "rev-pr139-def456a-p8f31c09",

  "triggerId": "trg_01JABC",
  "triggerKey": "trg-pr139-def456a",

  "reason": "direct-run",
  "status": "running",
  "attempt": 1,

  "runner": {
    "mode": "local",
    "backend": "local",
    "platform": "darwin"
  },

  "summary": {
    "total": 12,
    "completed": 0,
    "failed": 0,
    "running": 0,
    "pending": 12
  },

  "createdAt": "2026-05-29T00:00:00Z",
  "startedAt": "2026-05-29T00:00:00Z",
  "finishedAt": null
}
```

## 7. System triggers

Add built-in triggers so every plan has a trigger.

Required built-ins:

```text
system.manual
system.manual-changed
system.ci-unmatched
system.replay
system.api
```

### Resolution rules

Before planning:

```text
if --from-ci:
  match declared trigger from provider event
  if no match:
    fail by default
    optionally allow system.ci-unmatched in future
else if --trigger:
  use named declared trigger
else if --changed:
  synthesize system.manual-changed
else:
  synthesize system.manual
```

Existing docs say trigger system is currently opt-in and non-trigger plans include all environments.  Preserve that user-facing behavior, but internally represent it as `system.manual`.

## 8. Planning flow changes

### Current simplified flow

```text
load intent
resolve trigger if flags present
compile plan
write .orun/plans/{checksum}.json
```

### New Phase 1 flow

```text
load intent
resolve TriggerOccurrence always
compile Plan with trigger metadata always
compute plan hash
derive revisionKey
write .orun/revisions/{revisionKey}/
  trigger.json
  revision.json
  manifest.json
  plan.json
update refs/latest-revision.json
update refs/triggers/{triggerName}/latest.json
optionally write compatibility .orun/plans/{checksum}.json
```

### Plan metadata

Every generated plan should include trigger metadata, not only triggered plans. The docs already show trigger metadata when a trigger is used, including mode, provider, event, action, matched bindings, active environments, scope, base, and head.  Extend that to system triggers.

Example:

```json
{
  "metadata": {
    "trigger": {
      "type": "system",
      "name": "system.manual",
      "mode": "manual",
      "provider": "orun",
      "event": "manual",
      "scope": "full"
    },
    "revision": {
      "key": "rev-manual-def456a-p8f31c09",
      "planHash": "sha256:8f31c09..."
    }
  }
}
```

## 9. Run flow changes

### Current simplified flow

```text
resolve plan from latest/hash/file/name
create .orun/executions/{execID}
execute jobs
write state/logs
```

### New Phase 1 flow

```text
resolve PlanRevision
if no explicit plan:
  generate PlanRevision first using system/declarative trigger
create executionKey under revision
write execution.json
write snapshot.latest.json
execute jobs
write job/step/log state
update revision manifest latest execution summary
update refs/latest-execution.json
optionally write compatibility .orun/executions/{execID}
```

For Phase 1, if changing the runner state writer is too risky, implement a compatibility bridge:

```text
1. Existing runner writes old `.orun/executions/{execID}`.
2. New state layer mirrors or moves the output into:
   `.orun/revisions/{revisionKey}/executions/{executionKey}`.
3. Later phases make the runner write directly through StateStore.
```

## 10. Local StateStore interface

Introduce this now, even if only local filesystem is implemented.

```go
type StateStore interface {
    Root() string

    Read(ctx context.Context, path string) ([]byte, ObjectMeta, error)
    Write(ctx context.Context, path string, data []byte, opts WriteOptions) (ObjectMeta, error)
    CreateIfAbsent(ctx context.Context, path string, data []byte) (ObjectMeta, error)
    CompareAndSwap(ctx context.Context, path string, oldRev string, data []byte) (ObjectMeta, error)
    List(ctx context.Context, prefix string) ([]ObjectInfo, error)
    Delete(ctx context.Context, path string) error
}
```

### Local behavior

```text
Read             os.ReadFile
Write            atomic temp file + rename
CreateIfAbsent   O_EXCL create
CompareAndSwap   local revision file or mtime/hash guarded write
List             filepath walk
Delete           os.Remove / RemoveAll
```

### Future remote compatibility

The same logical paths should map to:

```text
local: .orun/revisions/...
r2:    orgs/{org}/projects/{project}/.orun/revisions/...
s3:    orgs/{org}/projects/{project}/.orun/revisions/...
```

Do not introduce a path shape that only works locally.

## 11. Refs and indexes

### 11.1 `refs/latest-revision.json`

```json
{
  "revisionKey": "rev-pr139-def456a-p8f31c09",
  "revisionId": "rev_01JABC",
  "planHash": "sha256:8f31c09...",
  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 11.2 `refs/latest-execution.json`

```json
{
  "revisionKey": "rev-pr139-def456a-p8f31c09",
  "executionKey": "run-001",
  "executionId": "exec_01JXYZ",
  "status": "completed",
  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 11.3 Trigger refs

```text
.orun/refs/triggers/{triggerName}/latest.json
.orun/refs/triggers/{triggerName}/{sourceScope}.json
```

Example:

```text
.orun/refs/triggers/github-pull-request/pr-139.json
```

```json
{
  "triggerName": "github-pull-request",
  "triggerKey": "trg-pr139-def456a",
  "revisionKey": "rev-pr139-def456a-p8f31c09",
  "latestExecutionKey": "run-001",
  "headRevision": "def456a1b2c3...",
  "createdAt": "2026-05-29T00:00:00Z"
}
```

### 11.4 Execution index

Use this for `orun status --all` without scanning every revision.

```text
.orun/indexes/executions/{executionKey}.json
```

```json
{
  "executionKey": "run-001",
  "executionId": "exec_01JXYZ",
  "revisionKey": "rev-pr139-def456a-p8f31c09",
  "status": "completed",
  "createdAt": "2026-05-29T00:00:00Z",
  "path": "revisions/rev-pr139-def456a-p8f31c09/executions/run-001"
}
```

## 12. CLI behavior changes

### `orun plan`

Should now:

```text
1. Always resolve TriggerOccurrence.
2. Always write revision-first layout.
3. Print revision key.
4. Preserve `--output` behavior.
5. Preserve compatibility `.orun/plans`.
```

Example output:

```text
✓ Plan revision created

Revision: rev-pr139-def456a-p8f31c09
Trigger:  github-pull-request / pr-139 / def456a
Jobs:     12
Path:     .orun/revisions/rev-pr139-def456a-p8f31c09/plan.json
```

### `orun run`

Should resolve in this order:

```text
1. explicit file path
2. revision key
3. named ref
4. old plan hash compatibility
5. component name
6. no arg: generate fresh PlanRevision and run it
```

### `orun status`

Should default to `refs/latest-execution.json`.

Should support:

```bash
orun status
orun status --all
orun status --revision rev-pr139-def456a-p8f31c09
orun status --exec-id run-001
```

Current status already supports latest, all, detailed, JSON, and watch mode.  Rewire the source, do not redesign the display yet.

### `orun logs`

Should resolve:

```text
latest execution
specific revision + execution
specific job
specific step
```

Current logs are stored per step and are read through `orun logs`.  Preserve that behavior.

### `orun describe`

Add aliases:

```bash
orun describe revision latest
orun describe revision rev-pr139-def456a-p8f31c09
orun describe trigger latest
orun describe execution run-001
```

Existing describe already shows run metadata, plan reference, trigger, and per-job breakdown.  This change makes those fields explicit.

## 13. Backward compatibility

Phase 1 should not break existing workflows.

### Must preserve

```bash
orun plan -o /tmp/plan.json
orun run --plan /tmp/plan.json
orun run a1b2c3
orun run --exec-id my-id
orun status
orun logs
```

### Compatibility writes

For the first implementation, also write:

```text
.orun/plans/{checksum}.json
.orun/plans/latest.json
```

as compatibility aliases to the new revision plan.

### Compatibility reads

If no revision layout exists:

```text
1. Read old `.orun/plans`.
2. Read old `.orun/executions`.
3. Optionally synthesize system.manual trigger/revision metadata in memory.
```

Do not force migration during read-only commands.

## 14. Migration strategy

Add a new command later, but Phase 1 can include a hidden/internal migration function:

```bash
orun state migrate --dry-run
orun state migrate
```

Migration behavior:

```text
old .orun/plans/{checksum}.json
  → .orun/revisions/rev-migrated-<hash>/plan.json
  → synthesize trigger.json as system.migrated

old .orun/executions/{execID}
  → attach to best matching revision if plan hash known
  → otherwise attach to rev-migrated-unknown-p<hash>
```

Do not delete old state in Phase 1.

## 15. Implementation package plan

Recommended Go package structure:

```text
internal/triggerctx/
  context.go
  system.go
  github.go
  resolve.go
  ids.go

internal/revision/
  model.go
  keys.go
  manifest.go
  writer.go
  resolver.go

internal/statestore/
  store.go
  local.go
  paths.go
  refs.go
  indexes.go

internal/executionstate/
  model.go
  writer.go
  resolver.go
  compat.go
```

### `internal/triggerctx`

Responsibilities:

```text
- Resolve trigger context for every plan.
- Keep existing declared trigger behavior.
- Synthesize system triggers.
- Produce TriggerOccurrence.
- Derive triggerKey.
```

### `internal/revision`

Responsibilities:

```text
- Compute plan hash.
- Derive revisionKey.
- Write revision files.
- Update manifest.
- Resolve revision by key/hash/latest/name.
```

### `internal/statestore`

Responsibilities:

```text
- Abstract local file operations.
- Keep path layout remote-compatible.
- Atomic local writes.
- Refs and indexes.
```

### `internal/executionstate`

Responsibilities:

```text
- Create ExecutionRun under revision.
- Maintain snapshot/latest state.
- Bridge old runner state if needed.
- Update status/log lookup.
```

## 16. Key algorithms

### 16.1 Trigger resolution

```go
func ResolveTriggerContext(opts PlanOptions, intent Intent) (TriggerOccurrence, error) {
    switch {
    case opts.FromCI != "":
        return ResolveProviderEvent(opts.FromCI, opts.EventFile, intent)
    case opts.Trigger != "":
        return ResolveNamedTrigger(opts.Trigger, opts, intent)
    case opts.Changed:
        return NewSystemManualChanged(opts, intent)
    default:
        return NewSystemManual(opts, intent)
    }
}
```

### 16.2 Revision key generation

```go
func RevisionKey(trigger TriggerOccurrence, planHash string) string {
    scope := NormalizeScope(trigger.Source.SourceScope)
    sha := ShortSHA(trigger.Source.HeadRevision)
    if sha == "" {
        sha = WorktreeMarker(trigger.Source)
    }
    return fmt.Sprintf("rev-%s-%s-p%s", scope, sha, ShortHash(planHash, 8))
}
```

### 16.3 Execution key generation

```go
func NextExecutionKey(store StateStore, revisionKey string) string {
    existing := List("revisions/{revisionKey}/executions/")
    return fmt.Sprintf("run-%03d", len(existing)+1)
}
```

If `--exec-id` is provided, derive:

```text
executionKey = SanitizeExecID(value)
```

and store original value in `execution.json`.

## 17. Acceptance criteria

Give the agent these acceptance checks.

### Planning

```bash
orun plan --intent examples/intent.yaml
```

Must create:

```text
.orun/revisions/<rev-key>/plan.json
.orun/revisions/<rev-key>/trigger.json
.orun/revisions/<rev-key>/revision.json
.orun/revisions/<rev-key>/manifest.json
.orun/refs/latest-revision.json
```

For non-trigger plan, `trigger.json` must use:

```text
triggerName = system.manual
triggerType = system
```

### Triggered planning

```bash
orun plan --intent examples/intent.yaml --trigger github-pull-request --base main --head HEAD
```

Must create a declared trigger occurrence:

```text
triggerName = github-pull-request
triggerType = declared
```

and plan metadata must still include trigger metadata.

### Changed planning

```bash
orun plan --intent examples/intent.yaml --changed --base main
```

Must use:

```text
triggerName = system.manual-changed
planScope.mode = changed
```

### Running

```bash
orun run --intent examples/intent.yaml --dry-run
```

Must create an execution under the latest revision:

```text
.orun/revisions/<rev-key>/executions/run-001/execution.json
.orun/refs/latest-execution.json
.orun/indexes/executions/run-001.json
```

### Existing commands still work

```bash
orun status
orun logs
orun describe run latest
orun get plans
```

Must not fail because of the new layout.

### Compatibility

```bash
orun plan -o /tmp/orun-plan.json
orun run --plan /tmp/orun-plan.json --dry-run
```

Must continue to work.

## 18. Agent implementation prompt

Use this prompt directly:

```text
Implement Phase 1 of the Orun trigger-first local revision state model.

Goal:
Make every plan produce a TriggerOccurrence, make every compiled plan persist as a PlanRevision under `.orun/revisions/{revisionKey}`, and make every execution live under that revision. Preserve all existing CLI behavior and compatibility paths.

Current behavior to preserve:
- Trigger bindings under `automation.triggerBindings` still work with `--trigger` and `--from-ci`.
- `orun plan` still emits an immutable plan JSON.
- `orun run`, `orun status`, `orun logs`, `orun describe`, and existing `--plan` behavior must continue to work.
- Existing `.orun/plans` and `.orun/executions` should remain readable; do not delete old state.

New model:
- Add TriggerOccurrence model.
- Add built-in system triggers:
  - `system.manual`
  - `system.manual-changed`
  - `system.ci-unmatched` only as a model value; do not enable fallback unless explicit.
  - `system.replay`
  - `system.api`
- Resolve a TriggerOccurrence for every plan.
- Add PlanRevision model.
- Generate short revision keys:
  `rev-<sourceScope>-<shortSha>-p<planHash8>`
- Generate trigger keys:
  `trg-<sourceScope>-<shortSha>`
- Store full traceback in JSON, not in folder names.
- Create local canonical layout:

.orun/
  version.json
  refs/
    latest-revision.json
    latest-execution.json
    triggers/{triggerName}/latest.json
    named/{name}.json
  indexes/
    revisions/{revisionKey}.json
    executions/{executionKey}.json
  revisions/{revisionKey}/
    manifest.json
    trigger.json
    revision.json
    plan.json
    executions/{executionKey}/
      execution.json
      snapshot.latest.json
      state.json
      metadata.json
      jobs/
      logs/
      events/
      artifacts/

Implementation requirements:
1. Add `internal/triggerctx` package:
   - TriggerOccurrence model
   - system trigger constructors
   - declared trigger resolver wrapper around current trigger logic
   - trigger key generation
2. Add `internal/revision` package:
   - PlanRevision model
   - revision key generation
   - manifest writer
   - revision resolver by latest, key, plan hash, and named ref
3. Add `internal/statestore` package:
   - local filesystem StateStore
   - atomic writes
   - refs and indexes helpers
   - path helpers that are remote-compatible
4. Update `orun plan`:
   - always resolve trigger context
   - always write revision layout
   - preserve `--output`
   - preserve `.orun/plans/{checksum}.json` and latest compatibility writes
5. Update `orun run`:
   - resolve or create a PlanRevision
   - create ExecutionRun under revision
   - preserve old execution behavior, using a bridge if needed
6. Update `status`, `logs`, `describe`, and `get plans`:
   - prefer revision-first layout
   - fall back to old paths
7. Add tests:
   - non-trigger plan creates system.manual trigger
   - changed plan creates system.manual-changed trigger
   - named trigger creates declared TriggerOccurrence
   - execution lives under revision
   - latest refs are updated
   - old `--plan /tmp/plan.json` still works
   - old `.orun/plans` read path still works

Do not implement R2, S3, Supabase, Durable Objects, or remote state in this phase. Keep the layout and StateStore interface compatible with future object-store backends.
```

## Final Phase 1 shape

The final local-only model should be:

```text
Every plan has trigger.json.
Every plan is a PlanRevision.
Every execution lives under a PlanRevision.
Every folder key is short and readable.
Every full traceback is stored in JSON.
Refs/indexes make lookup fast.
StateStore hides local filesystem details.
Old plan/execution paths remain compatible.
```

That gives you a clean local foundation that can later be mapped directly to:

```text
R2/S3 object keys
DO-backed execution coordination
Supabase indexes
Orun Cloud project routing
```

