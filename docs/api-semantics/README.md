<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3Tables API semantics (s3tables-controller)

These documents describe observed AWS API behaviors that matter when writing or reviewing this controller:
state machines, error codes, defaults, normalization, read gaps, tag semantics, quotas and the like. Each
finding carries its ACK implication and the controller's current handling status (handled / unhandled /
suspected bug).

Provenance: generated from the aws-api-semantics lab
(https://github.com/aws-controllers-k8s/aws-api-semantics, path `services/s3tables/`), where every finding's
probe evidence lives. Re-render with `semantics render s3tables --controller-dir <this repo>`; the lab is the
source of truth, so edit findings there, not here.

Reading budget: the Documents table gives each file's line count. Long notes and compacted entries live in
`details/<finding id>.md` (409 files, one finding each; full entry plus full notes), linked from the entries
as `full notes:` or `see:`; read them on demand.

Escape hatch: text inside preserved blocks (`preserved:start` ... `preserved:end` comments in the Overview,
Open questions and Supplementary notes sections) survives re-renders, and files without the generated marker
(for example under `supplementary/`) are never touched. Controller references in `handling` point at files and
symbols; where a line range is given it refers to controller commit 1eda357 (see each document's marker), not
to the current tree.

## Load guide

| task | read |
| --- | --- |
| add a resource | service.md (inventory, scope verdicts) + every doc listing that resource in the table below (start with its catch-all: tablebucket.md, namespace.md, table.md) |
| fix a reconcile bug | service.md `Handling gaps summary` (every suspected bug and partial handling, with its doc) first, then the entry and its doc's `Handling gaps (bugs to file)` |
| add a field / review a PR | the catch-all's `Field matrix` (C/U/R per leaf) + the doc(s) whose title or description covers the field or operation (see Documents) |
| e2e test work | each relevant doc's Overview (`Timing ranges`, where present) then its `E2E timing` table (observed durations, n) |
| adoption / import of existing resources | the `Adoption and first-sync hazards` and `Delete semantics` sections of the resource docs |
| codegen / generator.yaml changes | service.md `Codegen notes` + the `Scope` and `Field behavior` sections |
| what the controller already handles | [service.md 'Ground truth catalog'](service.md#ground-truth) (controller hooks catalog, `GT-*` ids) |
| find the doc of a finding id | [index.md](index.md) (id -> document map); ids in the text are links already |
| full notes or full entry of one finding | `details/<finding id>.md` (linked from the entry) |

## How to read an entry

- Header line: `- <a id="<id lowercased>"></a>**<id>** <category> · impact <high|medium|low> · <handling
  label> · verified <date>[, re-verified][ · status: unverified|stale]` (the category is a code span). The
  anchor makes the entry addressable as `<doc>.md#<id lowercased>`; every finding id mentioned anywhere in
  these documents (overviews, `related:` lists, notes, tables) is rendered as a link to that anchor, in its
  own document or across documents. `index.md` is the plain id -> document map.
- Title line: the bold title; `(hypothesis refuted; behavior confirmed)` after it means the lab hypothesis
  behind the probe was refuted while the behavior described is confirmed (in this dataset `status: refuted` is
  about the hypothesis, never about the observation). `· status: unverified` in the header marks doc claims
  and untested items; they are also listed under `Open questions`.
- Handling labels (header) and the `- handling:` bullet: `handled` = `handled via <controller refs>`;
  `partially handled` = `partially handled via <refs>` (also listed under `Handling gaps`); `unhandled (not
  handled in controller)` = `not handled in the controller (as of commit <sha>)`; `SUSPECTED CONTROLLER BUG` =
  `suspected controller bug - see Handling gaps`; `n/a` = no handling line; `tracked in GitHub issue (not
  handled)` = the stored handling reference is an open issue, rendered `tracked in <url> (not handled)` with
  any code refs after it. Line ranges in refs refer to controller commit 1eda357.
- Body bullets: `ACK:` implications · `ops:` · `fields:`; `repro:`; `measurements:` (k=v, nested dicts as
  k.sub=v); `handling:`; `related:` (links) · `hypotheses:` · `evidence:` (probe ids under
  `services/s3tables/probes/`); `notes:` inline or an excerpt plus `full notes: details/<id>.md`. A probe's
  own truncation inside a stored behavior is shown as `[truncated in evidence]`; the renderer never cuts a
  high-impact behavior. Compacted medium entries (oversized docs only) are header + title + first sentence +
  `see: details/<id>.md`.
- A labeled behavior summary describes the observation without incidental sample values. The full recorded
  observation remains in the lab finding and its probe evidence.
- `GT-*` ids are controller hooks catalog entries (behaviors the controller already handles, with the
  mechanism and file:line): [service.md 'Ground truth catalog'](service.md#ground-truth); they render as links
  followed by `(controller hooks catalog entry)`.
- Overview imperatives (must/never/always) are evidence-derived recommendations from the lab, not project
  decisions.

## Documents

| file | title | covers | resources | canonical findings | high | unhandled | suspect-bug | last verified | lines |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| [service.md](service.md) | S3Tables: service facts and cross-cutting behaviors | Model facts, inventory, service-wide findings, handling gaps summary, ground truth catalog | all | 26 | 8 | 10 | 4 | 2026-10-15 | 969 |
| [tablebucket-replication.md](tablebucket-replication.md) | TableBucket replication (PutTableBucketReplication, GetTableBucketReplication, GetTableReplicationStatus at bucket scope) | Bucket-level replication rules: versionToken locking, role prerequisites, destination lifecycle, ~60 s status evaluation, replica creation and re-creation, scope verdict for a TableBucketReplication resource. | TableBucketReplication, TableBucket | 44 | 18 | 44 | 0 | 2026-10-15 | 1193 |
| [tablebucket-encryption-storage.md](tablebucket-encryption-storage.md) | TableBucket encryption (SSE-S3 / SSE-KMS) and storage class | Put/Get/DeleteTableBucketEncryption and Put/GetTableBucketStorageClass: defaults, KMS key policy and key-state prerequisites, caller permissions, inheritance by tables, the ValidationException-vs-BadRequestException split. | TableBucket | 61 | 18 | 46 | 1 | 2026-10-15 | 1345 |
| [tablebucket-policy.md](tablebucket-policy.md) | TableBucket resource policy and authorization errors | Put/Get/DeleteTableBucketPolicy semantics (whole-document replace, 404 when absent, idempotent delete, read-back normalization, principal propagation windows), the two 403 codes, and the first-sync deletion hazard; scope verdict for TableBucketPolicy. | TableBucketPolicy, TableBucket | 33 | 9 | 22 | 3 | 2026-10-14 | 988 |
| [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) | TableBucket maintenance configuration and metrics configuration | PutTableBucketMaintenanceConfiguration (merge-patch semantics, int32 clamping, ValidationException on days < 1, server defaults enabled/3/10) and the metrics configuration pair; scope verdict for TableBucketMaintenanceConfiguration. | TableBucketMaintenanceConfiguration, TableBucket | 28 | 2 | 16 | 1 | 2026-10-15 | 715 |
| [tablebucket.md](tablebucket.md) | TableBucket (lifecycle, identity, errors, tags, limits) | CreateTableBucket/GetTableBucket/DeleteTableBucket/ListTableBuckets: identity and ARN construction, the ~10 min post-delete name reservation, not-empty delete as 400, tags, throttling, account limits; catch-all for TableBucket findings not covered above. | TableBucket | 63 | 18 | 38 | 3 | 2026-10-15 | 1162 |
| [namespace.md](namespace.md) | Namespace | CreateNamespace/GetNamespace/DeleteNamespace/ListNamespaces: minimal Create echo, Get-only identifiers, naming rules, strong consistency, not-empty delete as 400, the ~16 req/s per account-region limiter, no tags or ARN. | Namespace | 40 | 14 | 19 | 1 | 2026-10-15 | 1007 |
| [table-replication-status.md](table-replication-status.md) | Table replication status and replica tables (GetTableReplicationStatus, service-managed replicas) | The asynchronous side of table replication: GetTableReplicationStatus entries per destination state (pending/completed/failed), the ~60 s evaluation after every Put, failed destinations recovering only via a token-refreshing re-Put, replica tables as service-managed resources (type, managedByService, orphans after DeleteTableReplication), same-name collisions, dead destinations. | TableReplication, Table | 39 | 23 | 39 | 0 | 2026-10-15 | 1355 |
| [table-replication.md](table-replication.md) | Table replication configuration (PutTableReplication, GetTableReplication, DeleteTableReplication) | The configuration API of table-level replication: Put/Get/Delete triple, configuration versionToken vs table versionToken (exact 409/400 messages), role prerequisites and Put-time validation, re-Put semantics, 404 when absent; scope verdict for TableReplication. | TableReplication, Table | 28 | 10 | 28 | 0 | 2026-10-15 | 841 |
| [table-record-expiration.md](table-record-expiration.md) | Table record expiration configuration (scope investigation) | Put/GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus: MethodNotAllowedException 405 on every creatable table, the check order before the 405 gate, ValidationException on days < 1; scope verdict (not implementable today). | TableRecordExpirationConfiguration, Table | 24 | 10 | 24 | 0 | 2026-10-14 | 713 |
| [table-maintenance.md](table-maintenance.md) | Table maintenance configuration and maintenance job status | PutTableMaintenanceConfiguration (merge-patch, defaults auto/512 + 1/120, strategy prerequisites, every Put bumps modifiedAt), GetTableMaintenanceJobStatus, KMS maintenance-principal prerequisite, authorization codes; scope verdict for TableMaintenanceConfiguration. | TableMaintenanceConfiguration, Table | 72 | 28 | 64 | 0 | 2026-10-15 | 1381 |
| [table-policy.md](table-policy.md) | Table resource policy and authorization errors | Put/Get/DeleteTablePolicy semantics and Resource-form binding, IAM action requirements per table operation (PutTableData for schemas, GetTableData for sort compaction, kms:DescribeKey for aws:kms), AccessDeniedException vs ForbiddenException; scope verdict for TablePolicy. | TablePolicy, Table | 43 | 21 | 25 | 0 | 2026-10-15 | 1256 |
| [table-encryption-storage.md](table-encryption-storage.md) | Table encryption and storage class | Create-only encryptionConfiguration/storageClassConfiguration, creation-time snapshot of bucket defaults, GetTableEncryption/GetTableStorageClass read paths, KMS key-state and permission effects on CreateTable. | Table | 31 | 10 | 21 | 0 | 2026-10-15 | 802 |
| [table-schema.md](table-schema.md) | Table metadata and schema (CreateTable metadata.iceberg) | CreateTable metadata.iceberg: write-only through the S3 Tables API (no read path), schema={} -> 500, field ids renumbered and types normalized, server-added properties, required round-trip, format rules, what metadata.json carries. | Table | 37 | 10 | 26 | 1 | 2026-10-15 | 811 |
| [table-metadata-location.md](table-metadata-location.md) | Table metadata location and versionToken (UpdateTableMetadataLocation, GetTableMetadataLocation) | metadataLocation as engine-managed state (set synchronously by CreateTable-with-schema, moved by every commit), GetTableMetadataLocation, UpdateTableMetadataLocation validation, versionToken as a function of metadataLocation, ConflictException CAS semantics on Delete/Rename/Update. | Table | 57 | 29 | 25 | 3 | 2026-10-15 | 1347 |
| [table.md](table.md) | Table (lifecycle, identity, rename, errors, tags) | CreateTable/GetTable/DeleteTable/RenameTable/ListTables: minimal Create echo, Get-only members, lookup by ARN or triple, rename across namespaces, 404 message shapes, tags (ValidationException on length, UntagResource 500 on vanished tables), concurrency; catch-all for Table findings not covered above. | Table | 69 | 20 | 33 | 2 | 2026-10-15 | 1432 |

`unhandled` counts handling unhandled plus handled/partial entries whose only reference is an open GitHub
issue (rendered `tracked in <url> (not handled)`). [index.md](index.md) maps every finding id to its document.

Medium-impact entries are compacted (header, title, first sentence, link to `details/`) in:
table-maintenance.md. High-impact entries are always rendered in full.

## Resources

| resource | scope verdict(s) | docs |
| --- | --- | --- |
| Namespace | implemented | [namespace.md](namespace.md) |
| Table | defer, field-on-parent, implemented | [table-replication-status.md](table-replication-status.md), [table-replication.md](table-replication.md), [table-record-expiration.md](table-record-expiration.md), [table-maintenance.md](table-maintenance.md), [table-policy.md](table-policy.md), [table-encryption-storage.md](table-encryption-storage.md), [table-schema.md](table-schema.md), [table-metadata-location.md](table-metadata-location.md), [table.md](table.md) |
| TableBucket | implemented | [tablebucket-replication.md](tablebucket-replication.md), [tablebucket-encryption-storage.md](tablebucket-encryption-storage.md), [tablebucket-policy.md](tablebucket-policy.md), [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md), [tablebucket.md](tablebucket.md) |
| TableBucketMaintenanceConfiguration | - | [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) |
| TableBucketPolicy | - | [tablebucket-policy.md](tablebucket-policy.md) |
| TableBucketReplication | - | [tablebucket-replication.md](tablebucket-replication.md) |
| TableMaintenanceConfiguration | - | [table-maintenance.md](table-maintenance.md) |
| TablePolicy | - | [table-policy.md](table-policy.md) |
| TableRecordExpirationConfiguration | defer | [table-record-expiration.md](table-record-expiration.md) |
| TableReplication | - | [table-replication-status.md](table-replication-status.md), [table-replication.md](table-replication.md) |
| service | - | - |

## Totals

- findings: 716 (695 canonical, 21 duplicates) · suspect-bug: 19 · documents: 15 + service.md · details pages: 409
- model: 2018-05-10 (service/s3tables v1.16.4) · controller commit: 1eda357 · render date: see the marker on line 1 of each document

## Supplementary documents

None (files in this directory without the generated marker are listed here and never touched).

## AGENTS.md

The repository root `AGENTS.md` carries a short pointer block to this directory (between `aws-api-semantics:start` and
`aws-api-semantics:end` comments, replaced on every render); `CLAUDE.md` includes it via `@AGENTS.md`.
