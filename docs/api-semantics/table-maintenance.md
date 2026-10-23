<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table maintenance configuration and maintenance job status
_PutTableMaintenanceConfiguration (merge-patch, defaults auto/512 + 1/120, strategy prerequisites, every Put bumps modifiedAt), GetTableMaintenanceJobStatus, KMS maintenance-principal prerequisite, authorization codes; scope verdict for TableMaintenanceConfiguration._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-maintenance.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the table-scoped maintenance surface: PutTableMaintenanceConfiguration / GetTableMaintenanceConfiguration
on (tableBucketARN, namespace, name) with one `type` path member per call - `icebergCompaction` or `icebergSnapshotManagement`;
the bucket-only `icebergUnreferencedFileRemoval` is BadRequestException 400 'Invalid table maintenance type' here ([S3T-TABLE-160](#s3t-table-160),
-229) - and a `value` {status, settings.<type>} tagged union ([S3T-TABLE-211](#s3t-table-211)); the read-only GetTableMaintenanceJobStatus, which
lists all three job types per table including the bucket-level one ([S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-198](#s3t-table-198)); the
KMS maintenance-principal prerequisite that CreateTable enforces ([S3T-TABLE-054](#s3t-table-054), -055, -056); and the identity/authorization
codes of these calls ([S3T-TABLE-178](#s3t-table-178), -180, -179, -233, -327). Mental model for the configuration: the write is synchronous and
merge-patch - one type per Put, status and settings merged field by field into a value that already exists with server defaults
on every fresh table (compaction enabled/512 MB/strategy auto, snapshots enabled/1 kept/120 h), there is no Delete and no reset
input, the Get right after the Put was never stale, Put returns 204 with an empty body and both Gets echo the tableARN
([S3T-TABLE-197](#s3t-table-197), -200, -202, -203, -205, -208, -210, -212, -298). Every Put - including a status-only re-Put that changes nothing -
advances GetTable.modifiedAt while versionToken stays the same ([S3T-TABLE-110](#s3t-table-110), -206, [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)). The
job status is the asynchronous half: a status flip shows as Disabled / Not_Yet_Run on the first read, but runs appeared only
minutes after a metadata commit and never on untouched tables, and a revoked key grant surfaces there as Failed, not on any Put
([S3T-TABLEMAINTENANCECONFIGURATION-002](#s3t-tablemaintenanceconfiguration-002), -006, -007, [S3T-TABLE-236](#s3t-table-236), -237, -238). Scope verdict for a would-be
TableMaintenanceConfiguration CRD: the findings show a Put/Get pair with no Delete and no absent state to create or delete into
(a never-configured table already returns the full map; a recreated same-name table starts at the defaults again: [S3T-TABLE-212](#s3t-table-212),
-205, -183), a merge-patch Put whose omitted sub-fields keep their stored values so only declared sub-fields can be compared
([S3T-TABLE-200](#s3t-table-200), -202, -203, -296), strategy `sort`/`z-order` prerequisites rejected at Put time as 400 (table metadata) or 403
(caller lacks s3tables:GetTableData) ([S3T-TABLE-204](#s3t-table-204), -232, -234, -297, [S3T-TABLEMAINTENANCECONFIGURATION-013](#s3t-tablemaintenanceconfiguration-013)), modifiedAt churn
on every Put that must not be read as drift ([S3T-TABLE-206](#s3t-table-206), [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)), and model hazards - `type`
is a required path member, the enum literal `z-order` carries a hyphen ([S3T-TABLE-211](#s3t-table-211)). The controller implements no such CRD and
no maintenance field on Table today, and GetTableMaintenanceJobStatus has no controller surface at all (see Handling gaps).

### Rules

1. A nil or empty spec must mean "unmanaged", never "reset" or "write defaults": a never-configured table already returns the
   full map (compaction enabled/512, snapshots enabled/1/120) and no API distinguishes 'never set' from 'user set the defaults';
   a same-name table recreated after DeleteTable has a new tableARN and starts at those defaults again with policy 404 and all
   jobs Not_Yet_Run ([S3T-TABLE-212](#s3t-table-212), -197, -298, -309, -183). Contradiction recorded on -212: its "strategy absent on a fresh
   table" is superseded by -197/-199/-205, which read `auto` on the fresh table; -212 stands for the design claim only.
2. PutTableMaintenanceConfiguration is a merge-patch per type, never a replace: {status} only keeps the stored settings,
   {settings} only keeps the status, {targetFileSizeMB 128} keeps strategy binpack and {strategy binpack} keeps 128 MB, a
   compaction Put leaves snapshots 5/48 intact, value {} and settings {} are 204 no-ops, {status: disabled} without settings
   keeps targetFileSizeMB/strategy (doc claim C054 FALSE), {disabled, {256}} stores both - compare only the sub-fields the spec
   declares and never expect reset-by-omission ([S3T-TABLE-200](#s3t-table-200), -202, -203, -296, -201, -295, -294).
3. The defaults are reachable only by sending them: Put {auto, 512} and {1, 120} read back fresh-identical, but Put {512} alone
   leaves a previously sent binpack in place, so the fresh state needs strategy `auto` sent explicitly; settings survive a
   disabled -> enabled flip ([S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012), [S3T-TABLE-205](#s3t-table-205)).
4. Every Put advances GetTable.modifiedAt and leaves versionToken unchanged - three status-only re-Puts left the configuration
   unchanged and modifiedAt changed, a CAS with the pre-Put token still succeeded, the same value re-sent or flipped 10 times
   back-to-back is 204 every time (final = last Put) - while Put/DeleteTablePolicy change neither; a reconciler that re-Puts
   unconditionally keeps moving modifiedAt and must not read it as drift ([S3T-TABLE-110](#s3t-table-110), -206, -155, -207,
   [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)).
5. A settings member that does not match `type` is silently dropped with 204: only the status is applied to the typed entry,
   the foreign member is ignored, the other type's entry is untouched, and with both members present only the typed one is
   stored - a mis-keyed union never errors and never converges ([S3T-TABLE-228](#s3t-table-228)).
6. Compaction and snapshot Puts on one table conflict when fired together: 10/10 rounds gave one 204 and one ConflictException
   409 'The resource was modified by a concurrent request. Please retry your request.'; 10 concurrent PutTablePolicy gave
   2 x 200, 7 x 409, 1 x InternalServerErrorException 500, and every rejected call succeeded on sequential retry - serialize the
   two types and treat 409/500 here as retryable ([S3T-TABLE-191](#s3t-table-191)).
7. Strategy `sort`/`z-order` has Put-time prerequisites: a table without metadata -> BadRequestException 400 'The table metadata
   location is not specified.' for any caller (3/3 runs); metadata without a sort order -> 400 'requires sort order in the table
   metadata'; a role lacking s3tables:GetTableData -> AccessDeniedException 403 'Insufficient access...' while its binpack/size
   Puts are 204 - the 400s are a table-state fact, the 403 a caller permission ([S3T-TABLE-204](#s3t-table-204), -232, -234, -297, -233,
   [S3T-TABLEMAINTENANCECONFIGURATION-013](#s3t-tablemaintenanceconfiguration-013)).
8. Bounds are server-side BadRequestException 400 naming the bound: targetFileSizeMB 64..512 (-1, 0, 1, 32, 63, 513, 1024, 10240,
   2147483647 -> 'Minimum value: 64 MB, maximum value: 512 MB.'; 100.5 stored as 100; "256"/true -> 'Invalid input...');
   minSnapshotsToKeep/maxSnapshotAgeHours 0 and -1 -> 'Value must be a positive integer.' with no upper bound (2147483647 kept,
   2147483648 silently clamped to 2147483647) ([S3T-TABLE-226](#s3t-table-226), -227).
9. Literals are exact lowercase and `type` is validated as a path segment: status ENABLED/Enabled/on -> 400; strategy
   zorder/ZORDER/BINPACK/Auto -> 400 while binpack/z-order/sort pass the enum; type icebergUnreferencedFileRemoval /
   IcebergCompaction / foo -> 400 'Invalid table maintenance type: <value>'; an empty type -> 404 ([S3T-TABLE-230](#s3t-table-230), -229, -160, -211).
10. Identity errors split by shape, not by level: a missing bucket, namespace or table -> NotFoundException 404 with the identical
    'The specified table does not exist.' on all six policy/maintenance ops; a table ARN in the tableBucketARN slot -> 400 'The
    specified bucket name isn't valid...'; another account's ARN -> ForbiddenException 403 'Unauthorized'; another region's ARN
    -> 400 'The provided Region does not match the Region of the resource.' ([S3T-TABLE-178](#s3t-table-178), -180, -179).
11. IAM denials are AccessDeniedException 403 with the IAM message naming the action, and each sub-resource read needs its own
    action: a role allowed only s3tables:GetTable got GetTable 200 and 30/30 sub-resource Gets denied; a role allowed
    Get/PutTableMaintenanceConfiguration still got GetTableMaintenanceJobStatus and DeleteTable 403; a table-policy Deny on
    PutTableMaintenanceConfiguration denied the Put while the Get stayed 200 - 403 is neither "table gone" nor "config absent"
    ([S3T-TABLE-327](#s3t-table-327), -233, -215, -213).
12. Maintenance configuration and job status follow a RenameTable (new name: compaction 128 MB and status 200; old name: 404) and
    vanish with DeleteTable, which needs no prior cleanup (204 with policy and custom maintenance in place; Gets 404 from 129 ms)
    ([S3T-TABLE-182](#s3t-table-182), -183).
13. The KMS maintenance principal is checked by CreateTable and by the jobs, nowhere else: aws:kms with a key whose policy lacks
    the maintenance.s3tables.amazonaws.com Allow -> BadRequestException 400 'Insufficient access to perform table maintenance...'
    - also via a bucket default (bucket Put 200, CreateTable 400), with RSA_2048/HMAC_256 keys, and with a Condition scoped to
    <bucket>/table/<name>/* (<bucket>/table/*, <bucket>/*, no Condition -> 200); an explicit AES256 on CreateTable bypasses it;
    after revocation UpdateTableMetadataLocation and warehouse GetObject/PutObject stay 200 ([S3T-TABLE-054](#s3t-table-054), -256, -055, -246,
    -056, -301, -235).
14. Job status mirrors configuration writes at once but runs only after a metadata commit: Put(disabled) -> 'Disabled' on the
    first read, re-enable -> 'Not_Yet_Run' on the first read even after 60 min Disabled; untouched tables stayed Not_Yet_Run for
    31-60 min and a compaction settings Put did not trigger a run in 29 min; a metadata-only UpdateTableMetadataLocation drove
    all three jobs to Successful within 7 min; a bucket-level UFR Put(disabled) reads 'Disabled' for that job in every table's
    status, newborns included ([S3T-TABLEMAINTENANCECONFIGURATION-002](#s3t-tablemaintenanceconfiguration-002), -010, -006, -007, -008, [S3T-TABLE-209](#s3t-table-209), -265, -198).
15. Failed is a job-status fact, minutes late and erasable: after a key-grant revocation the three jobs went Not_Yet_Run -> Failed
    5.0/6.0/8.1 min after a metadata-only commit with the CreateTable maintenance text as failureMessage (n=1 each); entry
    members are Not_Yet_Run {status}, Disabled {status}, Successful {status, lastRunTimestamp}; a Put(disabled) on a Failed job
    collapses the entry to {status: Disabled}, dropping lastRunTimestamp and failureMessage ([S3T-TABLE-236](#s3t-table-236), -237, -238,
    [S3T-TABLEMAINTENANCECONFIGURATION-009](#s3t-tablemaintenanceconfiguration-009), -011).

### Timing ranges

- Configuration read-after-write: no wait exists - 0 stale reads in 20 Put -> Get trials, write-to-read gap 40-64 ms, n=20
  ([S3T-TABLE-208](#s3t-table-208)); the 12-step reset sequence read every Put back on the next Get, n=1 sequence
  ([S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)).
- Sub-resource writes after CreateTable: PutTableMaintenanceConfiguration 108-182 ms after CreateTable returned -> 204 in 20/20
  (PutTablePolicy at 0 ms -> 200 in 20/20), CreateTable latency 307-416 ms, n=20 ([S3T-TABLE-219](#s3t-table-219)).
- Sub-resource reads after CreateTable: GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable all 200,
  0 stale, n=20 ([S3T-TABLE-220](#s3t-table-220)); fresh job status sampled at table age 0.7-1.4 s, n=4 ([S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001));
  the 619 ms first Get in -197/-198 is a single value (n=1, includes the CreateTable latency); the job status stayed Not_Yet_Run
  at offsets 0/1/5/30/60 s, n=5 polls ([S3T-TABLE-198](#s3t-table-198)).
- Disabled visible in GetTableMaintenanceJobStatus: 0.0-0.2 s after the Put, first read every time - 0.0 s, n=2
  ([S3T-TABLEMAINTENANCECONFIGURATION-002](#s3t-tablemaintenanceconfiguration-002)), 0.2 s, n=1 ([S3T-TABLE-209](#s3t-table-209)); a bucket-level UFR Put(disabled) was reflected in one
  table's job status on its first read 0.545 s after the Put and the re-enable at 0.49 s, n=1 table ([S3T-TABLE-265](#s3t-table-265)).
- Post-delete 404: GetTablePolicy/MaintenanceConfiguration/MaintenanceJobStatus were NotFoundException at 129, 154 and 500 ms
  after DeleteTable's 204, n=3 ([S3T-TABLE-183](#s3t-table-183)).
- Job runs after a metadata-only commit: Successful (lastRunTimestamp set) first seen 2.8-6.8 min after the commit (compaction
  2.8, snapshot management 4.8, unreferenced-file removal 6.8; sampled every 60 s; one table, n=3 job types; table age
  168-411 s) ([S3T-TABLEMAINTENANCECONFIGURATION-007](#s3t-tablemaintenanceconfiguration-007), -003, -004, -005). Failed after a key-grant revocation: 5.0-8.1 min after
  the commit (5.0 UFR, 6.0 compaction, 8.1 snapshots), n=1 per job ([S3T-TABLE-236](#s3t-table-236), -237, -238).
- Quiet-table soaks: Not_Yet_Run held 30.8-60.0 min untouched, n=61 samples, and a compaction settings Put did not start a run
  within 29 min, n=1 ([S3T-TABLEMAINTENANCECONFIGURATION-006](#s3t-tablemaintenanceconfiguration-006)); Disabled held the full 60 min, n=61 (-008); re-enable after 60 min
  -> Not_Yet_Run on 1 poll, n=1 (-010).
- Concurrency on one table: compaction+snapshot Put pairs 1.04-1.08 s wall per round, 10 rounds (n=30 measurement set); the
  10-way PutTablePolicy race took 6.22 s wall with latencies 155-6210 ms, one burst ([S3T-TABLE-191](#s3t-table-191)). 30 concurrent
  GetTablePolicy+GetTableMaintenanceConfiguration x2: 0 TooManyRequests, burst wall 0.11-0.13 s ([S3T-TABLE-187](#s3t-table-187)). 10 concurrent
  CreateTable into a fresh namespace: one run 424-909 ms for 9 calls and 10453 ms for one (wall 10.46 s), the other run's wall
  1.63 s, n=2 runs ([S3T-TABLE-303](#s3t-table-303)).
- Resource-policy Deny windows (single runs, n=1 each): a Deny on GetTableMaintenanceConfiguration/DeleteTable enforced after
  0.03 s and lifted after 0.01 s ([S3T-TABLE-213](#s3t-table-213)); a Deny on PutTableMaintenanceConfiguration enforced after 0.03 s and lifted
  after 0.04 s ([S3T-TABLE-215](#s3t-table-215)).
- KMS key-policy flips seen by CreateTable: 400 -> 200 at 0.51-0.62 s after the grant and 200 -> 400 at 0.37-0.42 s after the
  revoke, first call each time, n=2 ([S3T-TABLE-253](#s3t-table-253)); the rejected CreateTable itself took 509 ms, n=1 ([S3T-TABLE-054](#s3t-table-054)), 411-414 ms
  for RSA/HMAC keys and 652 ms for the symmetric key, n=1 each ([S3T-TABLE-246](#s3t-table-246)).
- IAM propagation on this surface: an inline role policy was effective 0.0 s after put_role_policy with AssumeRole 9.1 s after
  CreateRole, n=1 ([S3T-TABLE-233](#s3t-table-233)); three roles denied 30/30 at 30.1 s and again at 238.2 s after PutRolePolicy ([S3T-TABLE-327](#s3t-table-327)) -
  no allow-side propagation window was observed, so none is quoted.

### Handling gaps

No finding in this document is marked HANDLING-GAP as `suspect-bug`; the two marked `partial` come first, then the state of the
rest against the controller as it is today.

- [S3T-TABLE-021](table-encryption-storage.md#s3t-table-021) and [S3T-TABLE-058](table-encryption-storage.md#s3t-table-058) (partial, [GT-S3T-034](service.md#gt-s3t-034) (controller hooks catalog entry)/-038): the creation-time snapshot of the bucket's storage class and
  encryption defaults (the controller marks both fields immutable but never reads the effective value back) is documented
  with its evidence in `table-encryption-storage.md`, where both findings now render.
- Maintenance surface ([S3T-TABLE-197](#s3t-table-197), -199, -200, -202, -203, -205, -212, -228, [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012) and the
  job-status findings, unhandled): generator.yaml and hooks.go carry no `MaintenanceConfiguration` field on Table and no
  Put/GetTableMaintenanceConfiguration or GetTableMaintenanceJobStatus call, so these are scope input for a future field or CRD,
  not defects; -212 lists `late_initialize`, `compare.is_ignored+delta_pre_compare` and `custom_delete` as the implications a
  TableBucket-style implementation would need.
- Scored `handled` only through `terminal_codes: [BadRequestException]` (generator.yaml:103-104, [GT-S3T-039](service.md#gt-s3t-039) (controller hooks catalog entry)): [S3T-TABLE-226](#s3t-table-226),
  -227, -160, -204, -229, -230, -232 would go Terminal if a future field sent those values; nothing sends them today.
  [S3T-TABLE-318](#s3t-table-318) (create-only inputs) is handled by the `is_immutable` markers at generator.yaml:41-50. AccessDeniedException /
  ForbiddenException ([S3T-TABLE-233](#s3t-table-233), -327, -179, -213, -215, unhandled) are not in `terminal_codes`, so a caller lacking a
  sub-resource action would be retried, not Terminal - a watch item for the future field, since no sub-resource call exists yet.

### Where next

- `tablebucket-maintenance-metrics.md` is the bucket-level twin: the one bucket type (`icebergUnreferencedFileRemoval`), its own
  merge-patch Put and enabled/3/10 defaults, int32 clamping and the ValidationException wire code the table-level Put never
  showed; [S3T-TABLE-265](#s3t-table-265) here is where that bucket-level status becomes visible per table.
- `table-policy.md` owns the authorization details behind rules 7, 10 and 11: Put/Get/DeleteTablePolicy semantics, the IAM action
  per table operation (s3tables:GetTableData for sort compaction, PutTableData for schemas), AccessDeniedException vs
  ForbiddenException, and the Deny probes whose canary was a maintenance call ([S3T-TABLE-213](#s3t-table-213), -215).
- `table-encryption-storage.md` and `tablebucket-encryption-storage.md` carry the encryption side of rule 13 and of the two
  partial findings: creation-time snapshot of bucket defaults ([S3T-TABLE-021](table-encryption-storage.md#s3t-table-021), -058, -318), KMS key-state and permission effects
  on CreateTable, the bucket key-policy prerequisites.
- `table-metadata-schema.md` covers UpdateTableMetadataLocation and versionToken, which rule 4 (token unchanged by maintenance
  Puts) and rule 14 (metadata-only commit as the job trigger) rest on, and the metadata location / sort order rule 7 requires.
- `table.md` carries the lifecycle behind rules 10 and 12 (GetTable/DeleteTable/RenameTable, 404 message shapes, tags, concurrency:
  [S3T-TABLE-303](#s3t-table-303), -187, recreated-table identity: -183); `table-replication.md` and `table-record-expiration.md` are the other
  table sub-resource families - no finding here shows them interacting with maintenance.
- Long catalogues belong in `details/<id>.md`: the targetFileSizeMB cells with the bound message ([S3T-TABLE-226](#s3t-table-226)) and the
  snapshot-count cells with the clamp (-227), the partial-value and union-mismatch matrices (-202, -228), the 12-step reset
  sequence ([S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)), the key-policy Condition matrix ([S3T-TABLE-056](#s3t-table-056)), the 30 denied sub-resource
  cells (-327) and the 11-operation role matrix (-233), the six-operation 404 table (-178), the post-delete read table (-183)
  and the two race outcomes (-191).
- Open experiments: STATUS.md "Account state" and "Phase 8 result" list none for this stem. Phase 8 re-verified
  [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012) (tag `reverified`) and CORRECTED [S3T-TABLE-233](#s3t-table-233) (sort/z-order on a table without metadata
  is 400 for any caller; the 403 needs a metadata location and names s3tables:GetTableData - now carried by
  [S3T-TABLEMAINTENANCECONFIGURATION-013](#s3t-tablemaintenanceconfiguration-013)). The undeletable leftover bucket in "Account state" belongs to the policy stem, not to
  maintenance. Single-run items worth a future probe are in `table-maintenance.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 72 (high 28 / medium 38 / low 6); duplicates folded into the appendix: 1
- handling: handled 8 · partial 0 · tracked 0 · unhandled 64 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: async-state-machine 12, error-code 8, prerequisite 8, request-validation 7, response-fidelity 7,
  server-default 6, other 4, eventual-consistency 3, delete-semantics 2, idempotency 2, identity 2,
  quota-limit 2, update-granularity 2, codegen-artifact 1, first-sync-destructive 1, immutable-field 1,
  requested-vs-effective 1, scope 1, sub-resource-api 1, unsettable-field 1
- medium-impact entries are compacted in this document (header, title, first sentence, link) because its full
  form exceeds 1500 lines (render.yaml `size_warn_lines`); each links to its full entry under `details/`.
  High-impact entries are never compacted.

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateNamespace | create | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTablePolicy | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableBucketMaintenanceConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketPolicy | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketStorageClass | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMaintenanceConfiguration | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMaintenanceJobStatus | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTablePolicy | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| PutTableBucketEncryption | create | tableBucketARN, encryptionConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketMaintenanceConfiguration | create | tableBucketARN, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketStorageClass | create | tableBucketARN, storageClassConfiguration | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableMaintenanceConfiguration | create | tableBucketARN, namespace, name, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTablePolicy | create | tableBucketARN, namespace, name, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

Also referenced by findings but not in the s3tables model (other services or annotated variants): PutKeyPolicy

## State machine

- **JobStatus**: Not_Yet_Run, Successful, Failed, Disabled
- **MaintenanceStatus**: enabled, disabled
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)
- **TableRecordExpirationJobStatus**: NotYetRun, Successful, Failed, Disabled

- <a id="s3t-table-209"></a>**S3T-TABLE-209** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceJobStatus.icebergCompaction shows Disabled 0.2 s after Put status=disabled and
  Not_Yet_Run 0.2 s after re-enable** - Put icebergCompaction {disabled, 512}; the first
  GetTableMaintenanceJobStatus poll 0.2 s later already reported icebergCompaction.status=Disabled. - see:
  [details/S3T-TABLE-209.md](details/S3T-TABLE-209.md)

- <a id="s3t-table-236"></a>**S3T-TABLE-236** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key grant revoked: UnreferencedFileRemoval job Not_Yet_Run -> Failed 5.0 min after a metadata-only commit; CreateTable maintenance text, n=1**
  Table in an SSE-KMS bucket; key state 'Enabled' (phase revoked). GetTableMaintenanceJobStatus every 60 s:
  icebergUnreferencedFileRemoval changed Not_Yet_Run -> Failed+ts+msg at +302 s into the phase (5.0 min after
  the metadata-only commit made under the revoked key). failureMessage: {"icebergUnreferencedFileRemoval":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'."}. Raw body: {"status": {"icebergCompaction": {"status":
  "Not_Yet_Run"}, "icebergUnreferencedFileRemoval": {"failureMessage": "Insufficient access to perform table
  maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.",
  "lastRunTimestamp": "2026-10-13T23:51:56.672Z", "status": "Failed"}, "icebergSnapshotManagement": {"status":
  "Not_Yet_Run"}}, "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-abb4b0-kmsjob/table/7752f10a-bdd2-4877-a789-a56ea35d13ce"}.
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus,
    UpdateTableMetadataLocation, PutKeyPolicy · fields: status.*.status, status.*.failureMessage,
    status.*.lastRunTimestamp, encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS table -> PutKeyPolicy(root only) -> UpdateTableMetadataLocation(copy of metadata.json) ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: since_commit_min=[5.0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-235](#s3t-table-235), [S3T-TABLE-237](#s3t-table-237), [S3T-TABLE-238](#s3t-table-238), [S3T-TABLEMAINTENANCECONFIGURATION-011](#s3t-tablemaintenanceconfiguration-011),
    [S3T-TABLE-242](table-encryption-storage.md#s3t-table-242), [S3T-TABLEMAINTENANCECONFIGURATION-003](#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-009](#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-126, H-U-062, H-U-060 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: Transition recorded immediately during the 'revoked' soak (key Enabled, policy = account root
    only). Failed entry members: status, lastRunTimestamp, failureMessage. The wire value is 'Failed' (model
    casing), not the user guide's 'FAILED'. The message is identical to CreateTable's synchronous...
  - full notes: [details/S3T-TABLE-236.md](details/S3T-TABLE-236.md)

- <a id="s3t-table-237"></a>**S3T-TABLE-237** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key grant revoked: Compaction job Not_Yet_Run -> Failed 6.0 min after a metadata-only commit; failureMessage = CreateTable maintenance text**
  Table in an SSE-KMS bucket; key state 'Enabled' (phase revoked). GetTableMaintenanceJobStatus every 60 s:
  icebergCompaction changed Not_Yet_Run -> Failed+ts+msg at +362 s into the phase (6.0 min after the
  metadata-only commit made under the revoked key). failureMessage: {"icebergCompaction": "Insufficient access
  to perform table maintenance. Ensure that the key used to encrypt the table still exists, is active, and has
  a resource policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'."}. Raw
  body: {"status": {"icebergCompaction": {"failureMessage": "Insufficient access to perform table maintenance.
  Ensure that the key used to encrypt the table still exists, is active, and has a resource policy granting
  access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.", "lastRunTimestamp":
  "2026-10-13T23:53:03.778Z", "status": "Failed"}, "icebergUnreferencedFileRemoval": {"failureMessage":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.", "lastRunTimestamp": "2026-10-13T23:51:56.672Z", "status": "Failed"},
  "icebergSnapshotManagement": {"status": "Not_Yet_Run"}}, "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-abb4b0-kmsjob/table/7752f10a-bdd2-4877-a789-a56ea35d13ce"}.
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus,
    UpdateTableMetadataLocation, PutKeyPolicy · fields: status.*.status, status.*.failureMessage,
    status.*.lastRunTimestamp, encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS table -> PutKeyPolicy(root only) -> UpdateTableMetadataLocation(copy of metadata.json) ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: since_commit_min=[6.0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-235](#s3t-table-235), [S3T-TABLE-236](#s3t-table-236), [S3T-TABLE-238](#s3t-table-238), [S3T-TABLEMAINTENANCECONFIGURATION-011](#s3t-tablemaintenanceconfiguration-011),
    [S3T-TABLE-242](table-encryption-storage.md#s3t-table-242), [S3T-TABLEMAINTENANCECONFIGURATION-003](#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-009](#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-126, H-U-062, H-U-060 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: Transition recorded immediately during the 'revoked' soak (key Enabled, policy = account root
    only). Failed entry members: status, lastRunTimestamp, failureMessage. The wire value is 'Failed' (model
    casing), not the user guide's 'FAILED'. The message is identical to CreateTable's synchronous...
  - full notes: [details/S3T-TABLE-237.md](details/S3T-TABLE-237.md)

- <a id="s3t-table-238"></a>**S3T-TABLE-238** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Key grant revoked: SnapshotManagement job Not_Yet_Run -> Failed 8.1 min after a metadata-only commit; CreateTable maintenance text (n=1)**
  Table in an SSE-KMS bucket; key state 'Enabled' (phase revoked). GetTableMaintenanceJobStatus every 60 s:
  icebergSnapshotManagement changed Not_Yet_Run -> Failed+ts+msg at +483 s into the phase (8.1 min after the
  metadata-only commit made under the revoked key). failureMessage: {"icebergSnapshotManagement":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'."}. Raw body: {"status": {"icebergCompaction": {"failureMessage":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.", "lastRunTimestamp": "2026-10-13T23:53:03.778Z", "status": "Failed"},
  "icebergUnreferencedFileRemoval": {"failureMessage": "Insufficient access to perform table maintenance.
  Ensure that the key used to encrypt the table still exists, is active, and has a resource policy granting
  access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.", "lastRunTimestamp":
  "2026-10-13T23:51:56.672Z", "status": "Failed"}, "icebergSnapshotManagement": {"failureMessage":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource poli [truncated raw body]
  - ACK: synced.when, requeue, hook:sdk_read_one_post_set_output · ops: GetTableMaintenanceJobStatus,
    UpdateTableMetadataLocation, PutKeyPolicy · fields: status.*.status, status.*.failureMessage,
    status.*.lastRunTimestamp, encryptionConfiguration.kmsKeyArn
  - repro: SSE-KMS table -> PutKeyPolicy(root only) -> UpdateTableMetadataLocation(copy of metadata.json) ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: since_commit_min=[8.1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-235](#s3t-table-235), [S3T-TABLE-236](#s3t-table-236), [S3T-TABLE-237](#s3t-table-237), [S3T-TABLEMAINTENANCECONFIGURATION-011](#s3t-tablemaintenanceconfiguration-011),
    [S3T-TABLE-242](table-encryption-storage.md#s3t-table-242), [S3T-TABLEMAINTENANCECONFIGURATION-003](#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-009](#s3t-tablemaintenanceconfiguration-009),
    [S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001) · hypotheses: H-U-126, H-U-062, H-U-060 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: Transition recorded immediately during the 'revoked' soak (key Enabled, policy = account root
    only). Failed entry members: status, lastRunTimestamp, failureMessage. The wire value is 'Failed' (model
    casing), not the user guide's 'FAILED'. The message is identical to CreateTable's synchronous...
  - full notes: [details/S3T-TABLE-238.md](details/S3T-TABLE-238.md)

- <a id="s3t-tablemaintenanceconfiguration-002"></a>**S3T-TABLEMAINTENANCECONFIGURATION-002** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration(disabled) -> job status 'Disabled' on the first read (<1 s) for snapshot mgmt and compaction...** (hypothesis refuted; behavior confirmed)
  t_dis (fresh table with schema, all jobs Not_Yet_Run).
  PutTableMaintenanceConfiguration(type=icebergSnapshotManagement, value={status: disabled, settings
  {minSnapshotsToKeep 1, maxSnapshotAgeHours 120}}) -> HTTP 204; GetTableMaintenanceJobStatus 95 ms later ->
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run;
  1/s sampling: [[0.15,
  "icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run"]].
  GetTableMaintenanceConfiguration afterwards -> HTTP 200 {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}},
  "icebergSnapshotManagement": {"status": "disabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}. Then
  PutTableMaintenanceConfiguration(type=icebergCompaction, {status: disabled, settings {512, auto}}) -> HTTP
  204; first read ->
  icebergCompaction=Disabled;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run;
  1/s sampling: [[0.15,
  "icebergCompaction=Disabled;icebergSnapshotManagement=Disabled;icebergUnreferencedFileRemoval=Not_Yet_Run"]].
  Raw member presence in the Disabled state: {"icebergCompaction": ["status"], "icebergSnapshotManagement":
  ["status"], "icebergUnreferencedFileRemoval": ["status"]} (lastRunTimestamp absent, failureMessage absent).
  - ACK: synced.when, hook:sdk_read_one_post_set_output · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus, GetTableMaintenanceConfiguration · fields: type, value.status,
    status.icebergSnapshotManagement.status, status.icebergCompaction.status
  - repro: CreateTable(schema) -> Put icebergSnapshotManagement disabled -> GetTableMaintenanceJobStatus 1/s x
    30 s -> Put icebergCompaction disabled -> same
  - measurements: disabled_visible_after_put_s=[0.0, 0.0], burst_polls=[1, 1], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-008](#s3t-tablemaintenanceconfiguration-008), [S3T-TABLEMAINTENANCECONFIGURATION-010](#s3t-tablemaintenanceconfiguration-010),
    [S3T-TABLEMAINTENANCECONFIGURATION-011](#s3t-tablemaintenanceconfiguration-011), [S3T-TABLE-265](#s3t-table-265) · hypotheses: H-U-057, H-U-062 · evidence:
    tablemaintenanceconfiguration/state-machine/job-status-soak
  - notes: H-U-057 (Disabled appears only after the next scheduled pass, hours later) REFUTED: the job status
    reflects the configuration write immediately. H-U-062 Disabled cell on a never-run job: no
    lastRunTimestamp (there was no previous run) and no failureMessage.
  - full notes: [details/S3T-TABLEMAINTENANCECONFIGURATION-002.md](details/S3T-TABLEMAINTENANCECONFIGURATION-002.md)

- <a id="s3t-tablemaintenanceconfiguration-003"></a>**S3T-TABLEMAINTENANCECONFIGURATION-003** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **t_prop: compaction Not_Yet_Run -> Successful (lastRunTimestamp set) seen 2.8 min after a metadata-only
  UpdateTableMetadataLocation** - Cell t_prop (CreateTable with schema, then ONE metadata-only commit: copy of
  metadata.json + properties history.expire.max-snapshot-age-ms=1, no snapshot, UpdateTableMetadataLocation ->
  200). - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-003.md](details/S3T-TABLEMAINTENANCECONFIGURATION-003.md)

- <a id="s3t-tablemaintenanceconfiguration-004"></a>**S3T-TABLEMAINTENANCECONFIGURATION-004** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **t_prop: snapshot management Not_Yet_Run -> Successful (lastRunTimestamp set) seen 4.8 min after a
  metadata-only UpdateTableMetadataLocation** - Cell t_prop (CreateTable with schema, then ONE metadata-only
  commit: copy of metadata.json + properties history.expire.max-snapshot-age-ms=1, no snapshot,
  UpdateTableMetadataLocation -> 200). - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-004.md](details/S3T-TABLEMAINTENANCECONFIGURATION-004.md)

- <a id="s3t-tablemaintenanceconfiguration-005"></a>**S3T-TABLEMAINTENANCECONFIGURATION-005** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **t_prop: unreferenced-file removal Not_Yet_Run -> Successful (lastRunTimestamp set) 6.8 min after a
  metadata-only UpdateTableMetadataLocation** - Cell t_prop (CreateTable with schema, then ONE metadata-only
  commit: copy of metadata.json + properties history.expire.max-snapshot-age-ms=1, no snapshot,
  UpdateTableMetadataLocation -> 200). - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-005.md](details/S3T-TABLEMAINTENANCECONFIGURATION-005.md)

- <a id="s3t-tablemaintenanceconfiguration-006"></a>**S3T-TABLEMAINTENANCECONFIGURATION-006** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Untouched tables stay Not_Yet_Run (no-metadata 60 min, schema 31 min); a compaction settings Put did not
  trigger a run in 29 min (n=1 each)** - GetTableMaintenanceJobStatus every 60 s for 60 min. - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-006.md](details/S3T-TABLEMAINTENANCECONFIGURATION-006.md)

- <a id="s3t-tablemaintenanceconfiguration-007"></a>**S3T-TABLEMAINTENANCECONFIGURATION-007** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Metadata-only UpdateTableMetadataLocation triggers all 3 maintenance jobs within 7 min (Successful +2.8/+4.8/+6.8 min)...** (hypothesis refuted; behavior confirmed)
  t_prop: CreateTable(schema) wrote metadata.json with keys ['current-schema-id', 'current-snapshot-id',
  'default-sort-order-id', 'default-spec-id', 'format-version', 'last-column-id', 'last-partition-id',
  'last-sequence-number', 'last-updated-ms', 'location', 'metadata-log', 'partition-specs',
  'partition-statistics', 'properties', 'refs', 'schemas', 'snapshot-log', 'snapshots', 'sort-orders',
  'statistics', 'table-uuid'] (snapshots: 0, properties: {"write.parquet.compression-codec": "zstd"}, refs:
  {}). A copy with properties['history.expire.max-snapshot-age-ms']='1' (+ metadata-log entry, still no
  snapshot) was written to <warehouse>/metadata/00001-<uuid>.metadata.json and
  UpdateTableMetadataLocation(versionToken) -> HTTP 200 (GetTable then shows the new metadataLocation: True,
  versionToken changed: True). Job status right after the commit:
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run'.
  Sampled every 60 s: icebergCompaction -> Successful+ts first seen 2.8 min after the commit (lastRunTimestamp
  2026-10-13T23:29:52.066Z); icebergSnapshotManagement -> Successful+ts first seen 4.8 min after the commit
  (lastRunTimestamp 2026-10-13T23:31:07.881Z); icebergUnreferencedFileRemoval -> Successful+ts first seen 6.8
  min after the commit (lastRunTimestamp 2026-10-13T23:33:02.148Z). Final raw body: {"status":
  {"icebergCompaction": {"lastRunTimestamp": "2026-10-13T23:29:52.066Z", "status": "Successful"},
  "icebergUnreferencedFileRemoval": {"lastRunTimestamp": "2026-10-13T23:33:02.148Z", "status": "Successful"},
  "icebergSnapshotManagement": {"lastRunTimestamp": "2026-10-13T23:31:07.881Z", "status": "Successful"}},
  "tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-cf508b-mjob/table/e707e087-85b8-4393-aa1e-fd9550ae04f2"}.
  Member presence per job type in the final state: {"icebergCompaction": ["lastRunTimestamp", "status"],
  "icebergUnreferencedFileRemoval": ["lastRunTimestamp", "status"], "icebergSnapshotManagement":
  ["lastRunTimestamp", "status"]}. Soak window 60 min, 61 samples.
  - ACK: synced.when, e2e-timing, requeue · ops: UpdateTableMetadataLocation, GetTableMaintenanceJobStatus,
    GetTable · fields: metadataLocation, versionToken, status.*.status, status.*.lastRunTimestamp,
    status.*.failureMessage
  - repro: CreateTable(schema) -> GetObject(metadata.json) -> PutObject(copy +
    history.expire.max-snapshot-age-ms=1) -> UpdateTableMetadataLocation -> GetTableMaintenanceJobStatus every
    60 s
  - measurements: commit_to_first_successful_min=[2.8, 4.8, 6.8], soak_window_min=[60.0], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEMAINTENANCECONFIGURATION-003](#s3t-tablemaintenanceconfiguration-003), [S3T-TABLEMAINTENANCECONFIGURATION-004](#s3t-tablemaintenanceconfiguration-004),
    [S3T-TABLEMAINTENANCECONFIGURATION-005](#s3t-tablemaintenanceconfiguration-005), [S3T-TABLEMAINTENANCECONFIGURATION-006](#s3t-tablemaintenanceconfiguration-006) · hypotheses: H-U-060,
    H-U-059, H-U-062 · evidence: tablemaintenanceconfiguration/state-machine/job-status-soak
  - notes: H-U-060 REFUTED on the property variant: snapshot management did not report Failed - it reported
    'Successful+ts' with a lastRunTimestamp 4.8 min after the commit (the hypothesis expected Failed within 48
    h); the branch/tag-ref variant was not attempted (a ref needs an existing snapshot). H-U-059...
  - full notes: [details/S3T-TABLEMAINTENANCECONFIGURATION-007.md](details/S3T-TABLEMAINTENANCECONFIGURATION-007.md)

- <a id="s3t-tablemaintenanceconfiguration-008"></a>**S3T-TABLEMAINTENANCECONFIGURATION-008** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Disabled job status is stable: compaction + snapshot mgmt stayed 'Disabled' 60 min after Put(disabled);
  unconfigured UFR stayed Not_Yet_Run** - After PutTableMaintenanceConfiguration disabled for
  icebergSnapshotManagement and icebergCompaction, GetTableMaintenanceJobStatus every 60 s for 60 min: final
  'icebergCompaction=Disabled;icebergSnapshotManagement=Disa... - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-008.md](details/S3T-TABLEMAINTENANCECONFIGURATION-008.md)

- <a id="s3t-tablemaintenanceconfiguration-010"></a>**S3T-TABLEMAINTENANCECONFIGURATION-010** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Re-enabling snapshot management after 60 min Disabled (never run) -> 'Not_Yet_Run' on the first read;
  compaction left Disabled** - t_dis had both types disabled for the whole soak (status Disabled, no run
  ever). - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-010.md](details/S3T-TABLEMAINTENANCECONFIGURATION-010.md)

## Identity and lookup

- <a id="s3t-table-180"></a>**S3T-TABLE-180** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **A table ARN in the tableBucketARN slot -> BadRequestException/400 for GetTablePolicy (all 6 ops rejected)...**
  tableBucketARN=<table ARN> with the correct namespace/name: GetTablePolicy -> BadRequestException/400 'The
  specified bucket name isn't valid. Specify a different bucket name, and then try again.'; DeleteTablePolicy
  -> BadRequestException/400 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; GetTableMaintenanceConfiguration -> BadRequestException/400 'The specified bucket name isn't
  valid. Specify a different bucket name, and then try again.'; GetTableMaintenanceJobStatus ->
  BadRequestException/400 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; PutTablePolicy -> BadRequestException/400 'The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.'; PutTableMaintenanceConfiguration -> BadRequestException/400
  'The specified bucket name isn't valid. Specify a different bucket name, and then try again.'. With a wrong
  namespace: GetTablePolicy -> BadRequestException/400 'The specified bucket name isn't valid. Specify a
  different bucket name, and then try again.'. With a wrong name: BadRequestException/400. Passing tableArn=
  to get_table_policy raises ParamValidationError client-side (the model has no such member);
  GetTable(tableArn) -> HTTP 200. Policy visible via the triple after the table-ARN-form Put: HTTP 200.
  - ACK: is_arn_primary_key, custom_find · ops: GetTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, PutTablePolicy,
    PutTableMaintenanceConfiguration, GetTable · fields: tableBucketARN, tableArn
  - repro: GetTablePolicy etc. with tableBucketARN=<tableARN>; get_table_policy(tableArn=...)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-178](#s3t-table-178), [S3T-TABLE-179](#s3t-table-179), [S3T-TABLE-182](#s3t-table-182), [S3T-TABLE-183](#s3t-table-183) · hypotheses: H-U-023, H-U-024,
    H-U-071 · evidence: table/identity/subresource-identifiers
  - notes: H-U-023 confirmed, H-U-024 refuted; H-U-071 confirmed at the model level (policy/maintenance ops
    have only tableBucketARN+namespace+name; GetTable also takes tableArn).
  - full notes: [details/S3T-TABLE-180.md](details/S3T-TABLE-180.md)

- <a id="s3t-table-182"></a>**S3T-TABLE-182** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After RenameTable the policy and maintenance config follow the table (new name returns them); old name -> NotFoundException/404...**
  With policy Sid NS1 and compaction targetFileSizeMB=128 stored: RenameTable(same ns, new name) -> HTTP 204.
  Under the OLD name: GetTablePolicy NotFoundException/404, GetTableMaintenanceConfiguration
  NotFoundException/404, GetTableMaintenanceJobStatus NotFoundException/404. Under the NEW name: policy Sid
  NS1, compaction 128 MB, job status HTTP 200. RenameTable into ns2 with a third name -> HTTP 204: policy Sid
  NS1, compaction 128. Rename back to the original ns1/name -> HTTP 204: policy Sid NS1, compaction 128.
  tableARN values across the 4 GetTable snapshots: 1 distinct;
  GetTable/GetTableMaintenanceConfiguration/GetTableMaintenanceJobStatus echo the same tableARN.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTablePolicy, GetTableMaintenanceConfiguration,
    GetTableMaintenanceJobStatus, GetTable · fields: namespace, name, tableARN
  - repro: PutTablePolicy + PutTableMaintenanceConfiguration -> RenameTable (same ns) -> Gets under old/new ->
    RenameTable (other ns) -> rename back
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-178](#s3t-table-178), [S3T-TABLE-179](#s3t-table-179), [S3T-TABLE-180](#s3t-table-180), [S3T-TABLE-183](#s3t-table-183) · hypotheses: H-U-021, H-U-022,
    H-U-082 · evidence: table/identity/subresource-identifiers
  - notes: H-U-021 confirmed, H-U-022 refuted; H-U-082 confirmed (tableARN identical across the three Gets and
    unchanged by renames and maintenance Put).
  - full notes: [details/S3T-TABLE-182.md](details/S3T-TABLE-182.md)

## Idempotency

- <a id="s3t-table-110"></a>**S3T-TABLE-110** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **versionToken is NOT changed by TagResource, PutTableMaintenanceConfiguration or RenameTable (newName / newNamespaceName)...** (hypothesis refuted; behavior confirmed)
  Table with metadata. TagResource -> HTTP 200: token changed False, modifiedAt changed True,
  UpdateTableMetadataLocation with the pre-tag token -> HTTP 200.
  PutTableMaintenanceConfiguration(icebergSnapshotManagement enabled) -> HTTP 204: token changed False,
  modifiedAt changed True, CAS with the pre-call token -> HTTP 200. RenameTable(newName) -> HTTP 204: token
  changed False, modifiedAt changed True, CAS with the pre-rename token -> HTTP 200.
  RenameTable(newNamespaceName) -> HTTP 204: token changed False, modifiedAt changed True, CAS with the
  pre-rename token -> HTTP 200.
  - ACK: hook:sdk_update_pre_build_request, requeue · ops: TagResource, PutTableMaintenanceConfiguration,
    RenameTable, UpdateTableMetadataLocation, GetTable · fields: versionToken, modifiedAt
  - repro: CreateTable with schema -> GetTable token -> TagResource / PutTableMaintenanceConfiguration /
    RenameTable, each followed by GetTable and UpdateTableMetadataLocation with the pre-mutation token
  - measurements: rename_latency_ms=[130, 141], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-155](#s3t-table-155), [S3T-TABLE-206](#s3t-table-206) · hypotheses: H-T-139 · evidence:
    table/identity/version-token-rename
  - notes: H-T-139 REFUTED: none of the four mutations changed GetTable.versionToken (each did advance
    modifiedAt), and UpdateTableMetadataLocation with the token read before each mutation -> HTTP 200 every
    time. Together with the metadata-location-update probe (same-location re-send echoes the same token;...
  - full notes: [details/S3T-TABLE-110.md](details/S3T-TABLE-110.md)

## Errors

- <a id="s3t-table-178"></a>**S3T-TABLE-178** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Policy/maintenance ops on a missing bucket, namespace or table: NotFoundException/404 sharing 'The specified table does not exist.'**
  Six operations (GetTablePolicy, DeleteTablePolicy, GetTableMaintenanceConfiguration,
  GetTableMaintenanceJobStatus, PutTablePolicy, PutTableMaintenanceConfiguration) against a nonexistent bucket
  (well-formed ARN, own account), a nonexistent namespace and a nonexistent table all return NotFoundException
  (HTTP 404) with the identical message 'The specified table does not exist.' - the missing level is not
  distinguishable. The only different 404 message is the policy-absent case on an existing table: 'The
  specified table policy does not exist.' (an upper-cased namespace additionally yields 'The specified
  namespace does not exist.').
  - ACK: exceptions.404, custom_find · ops: GetTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, PutTablePolicy,
    PutTableMaintenanceConfiguration
  - repro: each op with a bad bucket ARN / bad namespace / bad table name; GetTablePolicy on a real
    policy-less table
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-179](#s3t-table-179), [S3T-TABLE-180](#s3t-table-180), [S3T-TABLE-182](#s3t-table-182), [S3T-TABLE-183](#s3t-table-183) · hypotheses: H-U-019 · evidence:
    table/identity/subresource-identifiers
  - notes: H-U-019 qualified: code and HTTP status are uniform (404 NotFoundException) as hypothesized, but
    there are only two messages, not four: bucket/namespace/table missing share 'The specified table does not
    exist.'; only the policy-absent case differs.
  - full notes: [details/S3T-TABLE-178.md](details/S3T-TABLE-178.md)

- <a id="s3t-table-179"></a>**S3T-TABLE-179** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **tableBucketARN of another account -> ForbiddenException/403; other region in ARN ->
  BadRequestException/400...** - GetTablePolicy with a well-formed ARN whose account id is not the caller's ->
  ForbiddenException/403 'Unauthorized'; GetTableMaintenanceConfiguration same ARN -> ForbiddenException/403
  'Unauthorized'. - see: [details/S3T-TABLE-179.md](details/S3T-TABLE-179.md)

- <a id="s3t-table-191"></a>**S3T-TABLE-191** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Concurrent writes to one table's policy/maintenance -> ConflictException 409 (1 of 10 also 500); compaction and snapshot Puts conflict**
  10 PutTablePolicy with distinct Sids fired concurrently on one table (wall 6.22s, latencies [155, 6210] ms)
  -> {"200": 2, "ConflictException/409": 7, "InternalServerErrorException/500": 1} messages
  {"ConflictException": "The resource was modified by a concurrent request. Please retry your request.",
  "InternalServerErrorException": "An internal error occurred. Try again."}; GetTablePolicy then held Sid P04,
  which is one of the 2 successful puts; rejected puts retried sequentially -> [('P01', 'ConflictException',
  'HTTP 200'), ('P02', 'ConflictException', 'HTTP 200'), ('P03', 'ConflictException', 'HTTP 200'), ('P05',
  'InternalServerErrorException', 'HTTP 200'), ('P06', 'ConflictException', 'HTTP 200'), ('P07',
  'ConflictException', 'HTTP 200'), ('P08', 'ConflictException', 'HTTP 200'), ('P09', 'ConflictException',
  'HTTP 200')]. Ten rounds of PutTableMaintenanceConfiguration(icebergCompaction) and
  (icebergSnapshotManagement) fired together (distinct values per round) -> {"HTTP 204": 10,
  "ConflictException (HTTP 409)": 10}, per-round wall 1.04-1.08 s messages {"snapshot": "The resource was
  modified by a concurrent request. Please retry your request.", "compaction": "The resource was modified by a
  concurrent request. Please retry your request."}; GetTableMaintenanceConfiguration afterwards listed types
  ['icebergCompaction', 'icebergSnapshotManagement'] with compaction {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 73, "strategy": "binpack"}}} and snapshot {"status": "enabled",
  "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 8, "maxSnapshotAgeHours": 31}}} (the last
  round sent targetFileSizeMB=73, minSnapshotsToKeep=10, maxSnapshotAgeHours=33).
  - ACK: requeue, one-per-reconcile, custom_update · ops: PutTablePolicy, GetTablePolicy,
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration · fields: resourcePolicy,
    configuration.icebergCompaction, configuration.icebergSnapshotManagement
  - repro: 10 threads x PutTablePolicy(distinct Sid); 10 x
    [PutTableMaintenanceConfiguration(icebergCompaction) || (icebergSnapshotManagement)]; Get* after
  - measurements: policy_race_wall_s=6.22, policy_latency_ms_min_max=[155, 6210], pair_wall_s=[1.08, 1.04,
    1.05, 1.05, 1.04, 1.05, 1.04, 1.04, 1.04, 1.04], n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-207](#s3t-table-207), [S3T-TABLE-208](#s3t-table-208), [S3T-TABLE-209](#s3t-table-209) · hypotheses: H-U-064 · evidence:
    table/limits/policy-size-write-races
  - notes: H-U-064 refuted: (1) 10 concurrent PutTablePolicy -> 2 x 200, 7 x ConflictException 409 'The
    resource was modified by a concurrent request. Please retry your request.', 1 x
    InternalServerErrorException 500 'An internal error occurred. Try again.'; the stored policy was one of
    the two successes; all...
  - full notes: [details/S3T-TABLE-191.md](details/S3T-TABLE-191.md)

- <a id="s3t-table-215"></a>**S3T-TABLE-215** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Table policy Deny PutTableMaintenanceConfiguration: Put -> AccessDeniedException/403 after 0.03 s;
  GetTableMaintenanceConfiguration -> 200** - Deny s3tables:PutTableMaintenanceConfiguration for the caller's
  role via PutTablePolicy -> HTTP 200. - see: [details/S3T-TABLE-215.md](details/S3T-TABLE-215.md)

- <a id="s3t-table-233"></a>**S3T-TABLE-233** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Role lacking the IAM action: Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus, DeleteTable -> AccessDeniedException 403, IAM message**
  Assumed role with an inline Allow for only ['s3tables:GetTable',
  's3tables:GetTableMaintenanceConfiguration', 's3tables:PutTableMaintenanceConfiguration'] (effective 0.0 s
  after put_role_policy): GetTable -> HTTP 200; GetTablePolicy -> AccessDeniedException/403; PutTablePolicy ->
  AccessDeniedException/403; DeleteTablePolicy -> AccessDeniedException/403; GetTableMaintenanceConfiguration
  -> HTTP 200; PutTableMaintenanceConfiguration(binpack,256) -> HTTP 204;
  PutTableMaintenanceConfiguration(128) -> HTTP 204; PutTableMaintenanceConfiguration(sort) ->
  AccessDeniedException/403; PutTableMaintenanceConfiguration(z-order) -> AccessDeniedException/403;
  GetTableMaintenanceJobStatus -> AccessDeniedException/403; DeleteTable -> AccessDeniedException/403. Message
  form: 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTablePolicy on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-984'.
  - ACK: terminal_codes · ops: GetTablePolicy, PutTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceJobStatus, PutTableMaintenanceConfiguration, DeleteTable
  - repro: iam CreateRole + inline policy (GetTable, Get/PutTableMaintenanceConfiguration) -> sts AssumeRole
    -> each op
  - measurements: assume_role_after_s=9.1, policy_effective_after_s=0.0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-232](#s3t-table-232), [S3T-TABLE-234](#s3t-table-234), [S3T-TABLE-230](#s3t-table-230), [S3T-TABLE-251](table-policy.md#s3t-table-251) · hypotheses: H-U-017 · evidence:
    table/error-taxonomy/iam-deny-subresource-codes
  - notes: H-U-017 confirmed: identity-policy denials on GetTablePolicy, PutTablePolicy, DeleteTablePolicy,
    GetTableMaintenanceJobStatus and DeleteTable all surface as AccessDeniedException (HTTP 403) with the
    standard IAM message, although these operations declare only ForbiddenException;...
  - full notes: [details/S3T-TABLE-233.md](details/S3T-TABLE-233.md)

- <a id="s3t-table-253"></a>**S3T-TABLE-253** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms) 400 'Insufficient access to perform table maintenance' flips on the FIRST call after
  PutKeyPolicy (0.4-0.6 s, n=2)** - Key created with a root-only policy. - see:
  [details/S3T-TABLE-253.md](details/S3T-TABLE-253.md)

- <a id="s3t-table-327"></a>**S3T-TABLE-327** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Role allowed only s3tables:GetTable: every table sub-resource Get -> AccessDeniedException 403 naming its own action (30/30, 3 fresh roles)**
  Three fresh roles with one inline policy Allow s3tables:GetTable on * (trust: this account), first calls
  30.1 s after PutRolePolicy, second pass 238.2 s after; per operation {outcomes, action named in the denial,
  cells}: {"get_table": {"denied_actions": ["None"], "n": 6, "outcomes": ["HTTP 200"]}, "get_table_bucket":
  {"denied_actions": ["s3tables:GetTableBucket"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_encryption": {"denied_actions": ["s3tables:GetTableEncryption"], "n": 6, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_maintenance_configuration": {"denied_actions":
  ["s3tables:GetTableMaintenanceConfiguration"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_metadata_location": {"denied_actions": ["s3tables:GetTableMetadataLocation"], "n": 6, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_policy": {"denied_actions": ["s3tables:GetTablePolicy"],
  "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]}, "get_table_storage_class": {"denied_actions":
  ["s3tables:GetTableStorageClass"], "n": 6, "outcomes": ["AccessDeniedException (HTTP 403)"]}}. Missing table
  name: {"get_table": ["NotFoundException (HTTP 404) action=None"], "get_table_encryption":
  ["AccessDeniedException (HTTP 403) action=s3tables:GetTableEncryption"]}. One fresh role allowed only
  s3tables:GetTableBucket: {"get_table": {"denied_actions": ["s3tables:GetTable"], "n": 2, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket": {"denied_actions": ["None"], "n": 2, "outcomes":
  ["HTTP 200"]}, "get_table_bucket_encryption": {"denied_actions": ["s3tables:GetTableBucketEncryption"], "n":
  2, "outcomes": ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket_maintenance_configuration":
  {"denied_actions": ["s3tables:GetTableBucketMaintenanceConfiguration"], "n": 2, "outcomes":
  ["AccessDeniedException (HTTP 403)"]}, "get_table_bucket_policy": {"denied_actions":
  ["s3tables:GetTableBucketPolicy"], "n": 2, "outcomes": ["AccessDeniedException (HTTP 403)"]},
  "get_table_bucket_storage_class": {"denied_actions": ["s3tables:GetTableBucketStorageClass"], "n": 2,
  "outcomes": ["AccessDeniedException (HTTP 403)"]}}.
  - ACK: terminal_codes, hook:sdk_read_one_post_set_output · ops: GetTable, GetTableEncryption,
    GetTableStorageClass, GetTableMetadataLocation, GetTablePolicy, GetTableMaintenanceConfiguration,
    GetTableBucket, GetTableBucketEncryption, GetTableBucketStorageClass, GetTableBucketPolicy,
    GetTableBucketMaintenanceConfiguration
  - repro: CreateRole + PutRolePolicy(Allow s3tables:GetTable only) x3 -> wait >= 30 s -> AssumeRole -> each
    table Get on an existing table and on a missing name; repeat after 60 s; same with Allow
    s3tables:GetTableBucket only for the bucket Gets
  - measurements: roles=3, passes=2, sub_read_cells=30, first_pass_after_put_role_policy_s=30.1,
    second_pass_after_put_role_policy_s=238.2, n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-142](table-policy.md#s3t-table-142), [S3T-TABLE-132](table-policy.md#s3t-table-132) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa
  - notes: FU2-08: 132 holds - each table sub-resource Get (GetTableEncryption, GetTableStorageClass,
    GetTableMetadataLocation, GetTablePolicy, GetTableMaintenanceConfiguration) is authorized as its own
    s3tables action: AccessDeniedException 403 'not authorized to perform: s3tables:<Op> on resource:
    <table...
  - full notes: [details/S3T-TABLE-327.md](details/S3T-TABLE-327.md)

- <a id="s3t-tablemaintenanceconfiguration-013"></a>**S3T-TABLEMAINTENANCECONFIGURATION-013** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **PutTableMaintenanceConfiguration sort/z-order: no metadata location -> 400 for any caller...** - Assumed
  role allowed only s3tables:GetTable, GetTableMaintenanceConfiguration, PutTableMaintenanceConfiguration
  (Resource '*'), 3 runs. - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-013.md](details/S3T-TABLEMAINTENANCECONFIGURATION-013.md)

## Request validation

- <a id="s3t-table-202"></a>**S3T-TABLE-202** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration merges partial values: {status} only, {settings} only, {} and settings {} all HTTP 204**
  PutTableMaintenanceConfiguration type=icebergCompaction with value {"status":"disabled"} (no settings) ->
  HTTP 204, map after {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 64,
  "strategy": "auto"}}}. {"status":"enabled"} only -> HTTP 204, map after {"status": "enabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 64, "strategy": "auto"}}}.
  {"settings":{"icebergCompaction":{"targetFileSizeMB":256}}} (no status) -> HTTP 204, map after {"status":
  "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}. value {} ->
  HTTP 204. {"status":"enabled","settings":{}} -> HTTP 204.
  {"status":"enabled","settings":{"icebergCompaction":{}}} -> HTTP 204, map after {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}.
  - ACK: terminal_codes, custom_update · ops: PutTableMaintenanceConfiguration · fields: value.status,
    value.settings
  - repro: Put compaction with each partial value (boto3 parameter validation disabled)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](#s3t-table-294), [S3T-TABLE-295](#s3t-table-295), [S3T-TABLE-296](#s3t-table-296), [S3T-TABLE-298](#s3t-table-298), [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-201](#s3t-table-201),
    [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-200](#s3t-table-200), [S3T-TABLE-203](#s3t-table-203), [S3T-TABLE-199](#s3t-table-199) · hypotheses: H-U-037, H-U-043, H-U-049 ·
    evidence: table/mutation-matrix/maintenance-config
  - notes: H-U-037 REFUTED: settings is not conditionally required; {status:disabled} alone flips status and
    keeps the stored settings (64 MB stayed), {settings} alone keeps the stored status, value {} and settings
    {} are accepted as no-ops (nothing changed). H-U-043 empty-settings cell: accepted (204), not...
  - full notes: [details/S3T-TABLE-202.md](details/S3T-TABLE-202.md)

- <a id="s3t-table-226"></a>**S3T-TABLE-226** `request-validation` · impact high · handled · verified 2026-10-13
  **targetFileSizeMB: 64..512 enforced server-side (400 names the bounds); 100.5 accepted and stored as 100; "256"/true -> 400**
  PutTableMaintenanceConfiguration icebergCompaction with parameter validation disabled: accepted (HTTP 204,
  read back unchanged unless noted) for 64, 65, 100, 511, 512, float_100.5; rejected: {"-1":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "0": "BadRequestException/400 'The specified compaction file size is invalid. Minimum
  value: 64 MB, maximum value: 512 MB.'", "1": "BadRequestException/400 'The specified compaction file size is
  invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "32": "BadRequestException/400 'The specified
  compaction file size is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "63":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "513": "BadRequestException/400 'The specified compaction file size is invalid. Minimum
  value: 64 MB, maximum value: 512 MB.'", "1024": "BadRequestException/400 'The specified compaction file size
  is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "10240": "BadRequestException/400 'The specified
  compaction file size is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "2147483647":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "2147483648": "BadRequestException/400 'The specified compaction file size is invalid.
  Minimum value: 64 MB, maximum value: 512 MB.'", "string_256": "BadRequestException/400 'Invalid input.
  Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints on request parameters.'",
  "bool_true": "BadRequestException/400 'Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API
  reference for constraints on request parameters.'", "sdk_validation_0": "ParamValidationError/None
  'Parameter validation failed:\nInvalid value for parameter
  value.settings.icebergCompaction.targetFileSizeMB, value: 0, valid min value: 1'"}. Read-backs of accepted
  values: {"64": {"targetFileSizeMB": 64, "strategy": "auto"}, "65": {"targetFileSizeMB": 65, "strategy":
  "auto"}, "100": {"targetFileSizeMB": 100, "strategy": "auto"}, "511": {"targetFileSizeMB": 511, "strategy":
  "auto"}, "512": {"targetFileSizeMB": 512, "strategy": "auto"}, "float_100.5": {"targetFileSizeMB": 100,
  "strategy": "auto"}}. With boto3 validation enabled, 0 -> ParamValidationError/None before any request
  (model min=1).
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration · fields:
    value.settings.icebergCompaction.targetFileSizeMB
  - repro: Put compaction targetFileSizeMB in [-1,0,1,32,63,64,65,100,511,512,513,1024,10240,2^31-1,2^31] with
    parameter_validation=False; Get after each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-227](#s3t-table-227), [S3T-TABLE-228](#s3t-table-228), [S3T-TABLE-229](#s3t-table-229), [S3T-TABLE-230](#s3t-table-230) · hypotheses: H-U-040, H-U-041 ·
    evidence: table/weird-inputs/maintenance-bounds
  - notes: H-U-040 confirmed (0, 63, 513, 10240 and -1 -> BadRequestException 400 'The specified compaction
    file size is invalid. Minimum value: 64 MB, maximum value: 512 MB.'; 64 and 512 accepted and read back
    unchanged); H-U-041 refuted (1024/10240/32 rejected). Extra cells: a float 100.5 is accepted and...
  - full notes: [details/S3T-TABLE-226.md](details/S3T-TABLE-226.md)

- <a id="s3t-table-227"></a>**S3T-TABLE-227** `request-validation` · impact high · handled · verified 2026-10-13
  **Snapshot minSnapshotsToKeep/maxSnapshotAgeHours: 0 and -1 -> 400 'must be a positive integer'; no upper bound (2^31-1 ok, 2^31 clamped)**
  PutTableMaintenanceConfiguration icebergSnapshotManagement (minSnapshotsToKeep, maxSnapshotAgeHours) with
  SDK validation off: accepted ['1_1', '10_2500', '1000000_1000000', '2147483647_2147483647', '2147483648_1',
  '1_24', '100_1', 'min_only_3', 'max_only_48', 'empty_member']; rejected {"0_120": "BadRequestException/400
  'The specified minSnapshotsToKeep is invalid. Value must be a positive integer.'", "1_0":
  "BadRequestException/400 'The specified maxSnapshotAgeHours is invalid. Value must be a positive integer.'",
  "-1_120": "BadRequestException/400 'The specified minSnapshotsToKeep is invalid. Value must be a positive
  integer.'", "0_0": "BadRequestException/400 'The specified minSnapshotsToKeep is invalid. Value must be a
  positive integer.'"}. Read-backs: {"1_1": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 1}, "10_2500":
  {"minSnapshotsToKeep": 10, "maxSnapshotAgeHours": 2500}, "1000000_1000000": {"minSnapshotsToKeep": 1000000,
  "maxSnapshotAgeHours": 1000000}, "2147483647_2147483647": {"minSnapshotsToKeep": 2147483647,
  "maxSnapshotAgeHours": 2147483647}, "2147483648_1": {"minSnapshotsToKeep": 2147483647,
  "maxSnapshotAgeHours": 1}, "1_24": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 24}, "100_1":
  {"minSnapshotsToKeep": 100, "maxSnapshotAgeHours": 1}, "min_only_3": {"minSnapshotsToKeep": 3,
  "maxSnapshotAgeHours": 1}, "max_only_48": {"minSnapshotsToKeep": 3, "maxSnapshotAgeHours": 48},
  "empty_member": {"minSnapshotsToKeep": 3, "maxSnapshotAgeHours": 48}}.
  - ACK: terminal_codes, compare.nil_equals_zero_value · ops: PutTableMaintenanceConfiguration · fields:
    value.settings.icebergSnapshotManagement.minSnapshotsToKeep,
    value.settings.icebergSnapshotManagement.maxSnapshotAgeHours
  - repro: Put snapshot management with each (min,max) pair with parameter_validation=False; Get after each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-226](#s3t-table-226), [S3T-TABLE-228](#s3t-table-228), [S3T-TABLE-229](#s3t-table-229), [S3T-TABLE-230](#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)
    · hypotheses: H-U-042 · evidence: table/weird-inputs/maintenance-bounds
  - notes: H-U-042 confirmed: 0 or negative -> BadRequestException 400 'The specified minSnapshotsToKeep is
    invalid. Value must be a positive integer.' (resp. maxSnapshotAgeHours); 1/1, 10/2500, 1e6/1e6 and
    2147483647 accepted and read back unchanged. Surprises: 2147483648 is accepted and stored as 2147483647...
  - full notes: [details/S3T-TABLE-227.md](details/S3T-TABLE-227.md)

- <a id="s3t-table-228"></a>**S3T-TABLE-228** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Union member that does not match `type` is silently ignored (204): only status is applied; both members -> only the typed one stored**
  Baseline before each case: compaction {enabled,128,auto}, snapshots {enabled,2/24}.
  type=icebergSnapshotManagement with value {disabled, settings.icebergCompaction 256} -> HTTP 204; config
  after: {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB":
  128, "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "disabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}.
  type=icebergCompaction with {disabled, settings.icebergSnapshotManagement 7/48} -> HTTP 204; config after:
  {"icebergCompaction": {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}. Both members under
  type=icebergCompaction {enabled, compaction 256 + snapshot 7/48} -> HTTP 204; config after:
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2, "maxSnapshotAgeHours": 24}}}}. {disabled, settings
  {}} -> HTTP 204; config after: {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 2,
  "maxSnapshotAgeHours": 24}}}}. An unknown member (icebergFoo) or an extra key never leaves boto3
  (ClientSerializationError/None).
  - ACK: terminal_codes, custom_update · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: type, value.settings
  - repro: Put with each mismatched union payload (parameter_validation=False); Get after
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-226](#s3t-table-226), [S3T-TABLE-227](#s3t-table-227), [S3T-TABLE-229](#s3t-table-229), [S3T-TABLE-230](#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-012](#s3t-tablemaintenanceconfiguration-012)
    · hypotheses: H-U-043, H-U-044 · evidence: table/weird-inputs/maintenance-bounds
  - notes: H-U-044 confirmed / H-U-043 refuted: no 400 for a mismatched or empty union member; the request's
    status is applied to the typed entry and the foreign member is dropped (the other type's entry is
    untouched). With both members present only the member matching `type` is stored. A member unknown to
    the...
  - full notes: [details/S3T-TABLE-228.md](details/S3T-TABLE-228.md)

- <a id="s3t-table-229"></a>**S3T-TABLE-229** `request-validation` · impact medium · handled · verified 2026-10-13
  **type path segment: icebergUnreferencedFileRemoval / IcebergCompaction / foo -> 400 'Invalid table
  maintenance type: <value>'; empty -> 404** - PutTableMaintenanceConfiguration with an out-of-enum type:
  icebergUnreferencedFileRemoval -> BadRequestException/400 'Invalid table maintenance type:
  icebergUnreferencedFileRemoval'; IcebergCompaction -> BadRequestExc... - see:
  [details/S3T-TABLE-229.md](details/S3T-TABLE-229.md)

- <a id="s3t-table-230"></a>**S3T-TABLE-230** `request-validation` · impact medium · handled · verified 2026-10-13
  **Enums exact lowercase: status ENABLED/Enabled/on -> 400; strategy zorder/ZORDER/BINPACK/Auto -> 400;
  binpack ok; z-order/sort pass enum** - status values: ENABLED -> BadRequestException/400; Enabled ->
  BadRequestException/400; enabled -> HTTP 204; DISABLED -> BadRequestException/400; disabled -> HTTP 204; on
  -> BadRequestException/400; empty -> BadRequest... - see:
  [details/S3T-TABLE-230.md](details/S3T-TABLE-230.md)

- <a id="s3t-table-246"></a>**S3T-TABLE-246** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms) with an RSA_2048 or HMAC_256 key -> BadRequestException 400 maintenance-principal
  text; bucket Put accepts them** - Three Enabled keys with the same policy (account root +
  maintenance.s3tables.amazonaws.com kms:GenerateDataKey/kms:Decrypt). - see:
  [details/S3T-TABLE-246.md](details/S3T-TABLE-246.md)

## Update granularity and ordering

- <a id="s3t-table-054"></a>**S3T-TABLE-054** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable(aws:kms, key without maintenance.s3tables principal) -> BadRequestException HTTP 400**
  Key policy = KMS default only (account root kms:*; no maintenance.s3tables.amazonaws.com statement).
  CreateTable(ICEBERG, encryptionConfiguration={aws:kms, kmsKeyArn=<key ARN>}) -> BadRequestException (HTTP 400)
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.'; x-amzn-ErrorType header
  'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'; error body {"Message":
  "Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting access to the S3 service principal
  'maintenance.s3tables.amazonaws.com'.", "Code": "BadRequestException"}; latency 509 ms. Afterwards
  GetTable(triple) -> NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableEncryption
  -> NotFoundException (HTTP 404) 'The specified table does not exist.'; ListTables(namespace) lists [].
  - ACK: terminal_codes · ops: CreateTable, GetTable · fields: encryptionConfiguration.kmsKeyArn,
    encryptionConfiguration.sseAlgorithm
  - repro: CreateKey (no Policy) -> CreateTable(aws:kms, that key ARN) -> GetTable
  - measurements: create_table_rejected_latency_ms=[509], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-139](tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](tablebucket-encryption-storage.md#s3t-tablebucket-142), [S3T-TABLE-256](#s3t-table-256), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](tablebucket-encryption-storage.md#s3t-tablebucket-134),
    [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-055](#s3t-table-055), [S3T-TABLE-056](#s3t-table-056), [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057), [S3T-TABLE-141](table-policy.md#s3t-table-141), [S3T-TABLE-253](#s3t-table-253),
    [S3T-TABLE-301](#s3t-table-301), [S3T-TABLE-254](#s3t-table-254), [S3T-TABLE-246](#s3t-table-246), [S3T-TABLE-248](table-encryption-storage.md#s3t-table-248), [S3T-TABLE-239](table-record-expiration.md#s3t-table-239) · hypotheses: H-T-163, H-T-164,
    H-T-165 · evidence: table/dependencies/kms-maintenance-principal
  - notes: H-T-163 (BadRequestException 400, no half-created table) confirmed; H-T-164 (AccessDeniedException
    403, unmodeled) refuted; H-T-165 (create succeeds, failure only later in maintenance jobs) refuted (create
    rejected synchronously). The wire code IS in CreateTable's modeled error list.

- <a id="s3t-table-055"></a>**S3T-TABLE-055** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket default SSE-KMS with a key lacking the maintenance principal: PutTableBucketEncryption 200, CreateTable(no encryption config) 400**
  PutTableBucketEncryption(aws:kms, key with default policy only) -> HTTP 200; GetTableBucketEncryption
  afterwards -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}. CreateTable(ICEBERG, no
  encryptionConfiguration) under that bucket -> BadRequestException (HTTP 400) 'Insufficient access to perform
  table maintenance. Ensure that the key used to encrypt the table still exists, is active, and has a resource
  policy granting access to the S3 service principal 'maintenance.s3tables.amazonaws.com'.' (header
  'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'). GetTable(triple) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableEncryption -> NotFoundException
  (HTTP 404) 'The specified table does not exist.' .
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, GetTableBucketEncryption, CreateTable,
    GetTable · fields: encryptionConfiguration.kmsKeyArn
  - repro: PutTableBucketEncryption(aws:kms, key without maintenance statement) -> CreateTable without
    encryptionConfiguration -> GetTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-139](tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](tablebucket-encryption-storage.md#s3t-tablebucket-142), [S3T-TABLE-256](#s3t-table-256), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](tablebucket-encryption-storage.md#s3t-tablebucket-134),
    [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-054](#s3t-table-054), [S3T-TABLE-056](#s3t-table-056), [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057), [S3T-TABLE-141](table-policy.md#s3t-table-141), [S3T-TABLE-248](table-encryption-storage.md#s3t-table-248),
    [S3T-TABLE-246](#s3t-table-246), [S3T-TABLE-301](#s3t-table-301), [S3T-TABLE-254](#s3t-table-254) · hypotheses: H-T-155 · evidence:
    table/dependencies/kms-maintenance-principal
  - notes: H-T-155 confirmed (same failure as the explicit request, table absent). The bucket-level put does
    NOT run the maintenance check itself.
  - full notes: [details/S3T-TABLE-055.md](details/S3T-TABLE-055.md)

- <a id="s3t-table-056"></a>**S3T-TABLE-056** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance-principal key-policy Condition at CreateTable: <bucket>/table/<name>/* -> 400; <bucket>/table/*, <bucket>/*, no Condition -> 200**
  Key policy = root statement + Allow maintenance.s3tables.amazonaws.com kms:GenerateDataKey/kms:Decrypt with
  Condition StringLike kms:EncryptionContext:aws:s3:arn = X, 10 s after each PutKeyPolicy;
  CreateTable(aws:kms, key ARN) with a fresh name each time. X=<bucketARN>/table/<planned table name>/* ->
  BadRequestException (HTTP 400) 'Insufficient access to perform table maintenance. Ensure that the key used
  to encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.' after 3 attempts (41.4 s after PutKeyPolicy);
  X=<bucketARN>/table/* -> HTTP 200 (first attempt, 10.6 s after PutKeyPolicy); GetTableEncryption
  {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}; X=<bucketARN>/* -> HTTP 200
  (first attempt, 10.5 s after PutKeyPolicy); GetTableEncryption {"encryptionConfiguration": {"sseAlgorithm":
  "aws:kms", "kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}; no
  Condition -> HTTP 200 (first attempt, 10.6 s after PutKeyPolicy); GetTableEncryption
  {"encryptionConfiguration": {"sseAlgorithm": "aws:kms", "kmsKeyArn":
  "arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a"}}.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: PutKeyPolicy(maintenance statement with the given Condition) -> sleep 10 s -> CreateTable(aws:kms,
    key) (up to 3 attempts 15 s apart)
  - measurements: attempts_per_cell=[3, 1, 1, 1], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-054](#s3t-table-054), [S3T-TABLE-055](#s3t-table-055), [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057), [S3T-TABLE-141](table-policy.md#s3t-table-141) · hypotheses: H-T-166, H-T-163 ·
    evidence: table/dependencies/kms-maintenance-principal
  - notes: H-T-166 confirmed: the bucket-scoped condition passed and the name-scoped condition failed (note
    the real table ARN carries a UUID, not the name, so a name-scoped pattern cannot match any table ARN
    either); the <bucket>/table/* pattern passed, which contradicts the 'bucket-scoped zero-byte object'...
  - full notes: [details/S3T-TABLE-056.md](details/S3T-TABLE-056.md)

- <a id="s3t-table-200"></a>**S3T-TABLE-200** `update-granularity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration is a per-type upsert: Put compaction leaves snapshot 5/48 intact...**
  Put snapshot {enabled, 5/48} -> HTTP 204; then Put compaction {enabled, 256} -> HTTP 204 and the map
  afterwards is {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction":
  {"targetFileSizeMB": 256, "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled",
  "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5, "maxSnapshotAgeHours": 48}}}}. Then Put
  snapshot {disabled, 5/48} -> HTTP 204; map afterwards {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}},
  "icebergSnapshotManagement": {"status": "disabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 5, "maxSnapshotAgeHours": 48}}}}. Each Put touched only its own type.
  - ACK: custom_update, one-per-reconcile · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: type, value, configuration
  - repro: Put snapshot 5/48; Put compaction 256; Get; Put snapshot disabled; Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-203](#s3t-table-203), [S3T-TABLE-199](#s3t-table-199), [S3T-TABLE-205](#s3t-table-205) · hypotheses: H-U-034,
    H-U-035 · evidence: table/mutation-matrix/maintenance-config
  - notes: H-U-034 confirmed, H-U-035 refuted. Full original title: PutTableMaintenanceConfiguration is a
    per-type upsert: Put compaction leaves snapshot 5/48 intact; Put snapshot disabled leaves compaction 256
    intact

- <a id="s3t-table-204"></a>**S3T-TABLE-204** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Put compaction strategy sort/z-order on a table without metadata -> BadRequestException 400 'The table
  metadata location is not specified.'** - Table created with format=ICEBERG and no metadata (no schema, no
  sort order): Put {enabled, {strategy sort}} -> BadRequestException (HTTP 400) 'The table metadata location
  is not specified.', Get settings {"icebergCom... - see: [details/S3T-TABLE-204.md](details/S3T-TABLE-204.md)

- <a id="s3t-table-232"></a>**S3T-TABLE-232** `prerequisite` · impact medium · handled · verified 2026-10-13
  **Admin caller: Put strategy sort / z-order on a table WITH metadata but no sort order -> 400 'requires sort
  order in the table metadata'** - Table created with metadata.iceberg.schema (2 fields, no sort order;
  GetTable.metadataLocation present=True): Put icebergCompaction {enabled, strategy sort} ->
  BadRequestException (HTTP 400) 'The specified compaction... - see:
  [details/S3T-TABLE-232.md](details/S3T-TABLE-232.md)

- <a id="s3t-table-234"></a>**S3T-TABLE-234** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Put strategy sort/z-order by a role without s3tables:GetTableData -> AccessDeniedException 403
  'Insufficient access...'; binpack/size ok** - Role allowed Get/PutTableMaintenanceConfiguration + GetTable
  but NOT s3tables:GetTableData, on a table with Iceberg metadata (schema only): Put {binpack,256} -> HTTP
  204; Put {targetFileSizeMB 128} -> HTTP 204; Put {s... - see:
  [details/S3T-TABLE-234.md](details/S3T-TABLE-234.md)

- <a id="s3t-table-235"></a>**S3T-TABLE-235** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance grant revoked: UpdateTableMetadataLocation + warehouse GetObject/PutObject still 200; grant
  checked only by CreateTable and jobs** - After PutKeyPolicy(root statement only; the maintenance principal
  grant removed) the caller (under the root statement) read the current metadata.json: GetObject -> HTTP 200
  (ServerSideEncryption aws:kms, SSEKMSKeyId i... - see: [details/S3T-TABLE-235.md](details/S3T-TABLE-235.md)

- <a id="s3t-table-296"></a>**S3T-TABLE-296** `update-granularity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C054 FALSE: Put icebergCompaction {status: disabled} w/o settings keeps the stored
  targetFileSizeMB/strategy (merge), no reset** - With 64 MB stored: Put {"status":"disabled"} (no settings)
  -> HTTP 204, map after {"status": "disabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 64,
  "strategy": "auto"}}}; {"status":"enabled"} only -> e... - see:
  [details/S3T-TABLE-296.md](details/S3T-TABLE-296.md)

- <a id="s3t-table-297"></a>**S3T-TABLE-297** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C055 PARTLY: strategy=sort/z-order -> 204 ONLY on a table whose metadata has a sort order;
  otherwise 400 ([S3T-TABLE-204](#s3t-table-204)/232)** - Live: CreateTable with schema + writeOrder -> HTTP 200
  (metadataLocation present afterwards: True); fresh GetTableMaintenanceConfiguration compaction entry
  {"status": "enabled", "settings": {"icebergCompaction": {"tar... - see:
  [details/S3T-TABLE-297.md](details/S3T-TABLE-297.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-197"></a>**S3T-TABLE-197** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table GetTableMaintenanceConfiguration: compaction enabled/512 MB strategy auto; snapshots enabled 1/120h...**
  Table created in a bucket whose bucket-level icebergUnreferencedFileRemoval was set to status=disabled,
  unreferencedDays=7, nonCurrentDays=5 BEFORE the table existed. GetTableMaintenanceConfiguration 619 ms after
  the CreateTable request was issued (includes CreateTable latency) -> HTTP 200; the full configuration map is
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512,
  "strategy": "auto"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}. Nothing from the
  bucket-level config appears in the table-level map.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: GetTableMaintenanceConfiguration,
    CreateTable, PutTableBucketMaintenanceConfiguration · fields: configuration,
    configuration.icebergCompaction.settings.icebergCompaction.strategy, targetFileSizeMB, minSnapshotsToKeep,
    maxSnapshotAgeHours
  - repro: PutTableBucketMaintenanceConfiguration(disabled, 7/5) -> CreateTable ->
    GetTableMaintenanceConfiguration
  - measurements: first_get_after_create_ms=619, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-294](#s3t-table-294), [S3T-TABLE-295](#s3t-table-295), [S3T-TABLE-296](#s3t-table-296), [S3T-TABLE-298](#s3t-table-298), [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-202](#s3t-table-202),
    [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-160](#s3t-table-160), [S3T-TABLE-198](#s3t-table-198), [S3T-TABLE-204](#s3t-table-204), [S3T-TABLE-183](#s3t-table-183), [S3T-TABLE-211](#s3t-table-211), [S3T-TABLE-199](#s3t-table-199),
    [S3T-TABLE-212](#s3t-table-212) · hypotheses: H-U-031, H-U-032, H-U-033, H-U-067 · evidence:
    table/mutation-matrix/maintenance-config
  - notes: H-U-031 confirmed for the values (512; 1/120; both enabled; no bucket inheritance; no
    icebergUnreferencedFileRemoval key) but the compaction strategy IS present as 'auto' on a never-configured
    table. H-U-032 refuted (full map, not empty/404). H-U-033 fresh-table half REFUTED: strategy is reported
    as...
  - full notes: [details/S3T-TABLE-197.md](details/S3T-TABLE-197.md)

- <a id="s3t-table-198"></a>**S3T-TABLE-198** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table GetTableMaintenanceJobStatus: keys ['icebergCompaction', 'icebergSnapshotManagement',
  'icebergUnreferencedFileRemoval']...** - GetTableMaintenanceJobStatus 619 ms after the CreateTable request
  was issued -> HTTP 200 with status map {"tableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f1afd0-maint/table/8e81c95a-2317-46eb-8e1d-e6a7a1... -
  see: [details/S3T-TABLE-198.md](details/S3T-TABLE-198.md)

- <a id="s3t-table-203"></a>**S3T-TABLE-203** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Compaction settings MERGE field by field: Put {targetFileSizeMB 128} keeps strategy binpack; Put {strategy binpack} keeps 128 MB**
  Put {enabled, {strategy binpack, targetFileSizeMB 256}} -> settings {"targetFileSizeMB": 256, "strategy":
  "binpack"}. Put {enabled, {targetFileSizeMB 128}} -> settings {"targetFileSizeMB": 128, "strategy":
  "binpack"}. Put {enabled, {strategy binpack}} -> settings {"targetFileSizeMB": 128, "strategy": "binpack"}
  (Put results HTTP 204 / HTTP 204 / HTTP 204).
  - ACK: custom_update, compare.nil_equals_zero_value · ops: PutTableMaintenanceConfiguration,
    GetTableMaintenanceConfiguration · fields: value.settings.icebergCompaction.strategy,
    value.settings.icebergCompaction.targetFileSizeMB
  - repro: Put {binpack,256}; Get; Put {128}; Get; Put {binpack}; Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-200](#s3t-table-200), [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-199](#s3t-table-199), [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-204](#s3t-table-204),
    [S3T-TABLE-211](#s3t-table-211), [S3T-TABLE-016](table-schema.md#s3t-table-016) · hypotheses: H-U-038, H-U-039 · evidence:
    table/mutation-matrix/maintenance-config
  - notes: H-U-039 (merge) confirmed, H-U-038 (wholesale replace) refuted: an omitted field keeps its stored
    value; a controller must send strategy=auto explicitly to revert it.

- <a id="s3t-table-205"></a>**S3T-TABLE-205** `unsettable-field` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **No reset for maintenance config: Put of the default values leaves a previously sent strategy (binpack) in
  place; fresh table says auto** - There is no Delete or reset input for the maintenance configuration (value
  {} / settings {} outcomes in the request-validation finding). - see:
  [details/S3T-TABLE-205.md](details/S3T-TABLE-205.md)

- <a id="s3t-table-265"></a>**S3T-TABLE-265** `requested-vs-effective` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Bucket-level Put(disabled) on icebergUnreferencedFileRemoval -> only that job reads 'Disabled' in each
  table's job status; newborns too** - Two fresh tables read
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run /
  Compaction=Not_Yet_Run;SnapshotManagement=Not_Yet_Run;UnreferencedFileRemoval=Not_Yet_Run. - see:
  [details/S3T-TABLE-265.md](details/S3T-TABLE-265.md)

- <a id="s3t-table-301"></a>**S3T-TABLE-301** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Explicit CreateTable AES256 overrides a bucket SSE-KMS default (200, reads AES256) even when the bucket
  key lacks the maintenance principal** - Bucket default PutTableBucketEncryption(aws:kms K) -> HTTP 200 with
  K's policy lacking maintenance.s3tables.amazonaws.com. - see:
  [details/S3T-TABLE-301.md](details/S3T-TABLE-301.md)

- <a id="s3t-table-309"></a>**S3T-TABLE-309** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable with iceberg.schema: metadataLocation present at once; GetTableMaintenanceConfiguration 200
  with compaction + snapshot defaults** - Right after CreateTable with a schema (HTTP 200, versionToken
  present: True) GetTable showed metadataLocation=present, versionToken equal to the Create response: True,
  type=customer, format=ICEBERG, managedByService p... - see:
  [details/S3T-TABLE-309.md](details/S3T-TABLE-309.md)

- <a id="s3t-table-318"></a>**S3T-TABLE-318** `immutable-field` · impact high · handled · verified 2026-10-14
  **Table format, metadata, encryptionConfiguration and storageClassConfiguration are CreateTable-only inputs: no per-table Put/Update exists**
  botocore s3tables model (apiVersion 2018-05-10, 49 operations): the input members
  ['encryptionConfiguration', 'storageClassConfiguration', 'format', 'metadata'] appear only in
  {"CreateTable": ["format", "metadata", "encryptionConfiguration", "storageClassConfiguration"],
  "CreateTableBucket": ["encryptionConfiguration", "storageClassConfiguration"], "PutTableBucketEncryption":
  ["encryptionConfiguration"], "PutTableBucketStorageClass": ["storageClassConfiguration"]}. The table-scoped
  mutating operations are ['DeleteTable', 'DeleteTablePolicy', 'PutTableMaintenanceConfiguration',
  'PutTablePolicy', 'RenameTable', 'UpdateTableMetadataLocation']; none of them accepts encryption, storage
  class, format or metadata, so a table's encryption/storage class can be changed only by delete+recreate (the
  bucket-level Put*Encryption/Put*StorageClass operations apply to the bucket default, not to an existing
  table).
  - ACK: is_immutable, custom_update · ops: CreateTable, CreateTableBucket, PutTableBucketEncryption,
    PutTableBucketStorageClass, DeleteTable, DeleteTablePolicy, PutTableMaintenanceConfiguration,
    PutTablePolicy, RenameTable, UpdateTableMetadataLocation · fields: format, metadata,
    encryptionConfiguration, storageClassConfiguration
  - repro: walk every operation's input shape in the botocore s3tables model for the four member names
  - handling: handled via `generator.yaml:40-49; pkg/resource/table/sdk.go:104-181; generator.yaml:24-49; pkg/resource/table/hooks.go:69-75`
  - related: [S3T-TABLE-319](table-encryption-storage.md#s3t-table-319), [S3T-TABLE-312](table-schema.md#s3t-table-312) · evidence: table/sub-resources/fu-07-rename-kms-table
  - notes: FU-07 (static half). Confirms the sketch expectation: CreateTable is the only operation taking
    format/metadata, and the only table-level carrier of encryptionConfiguration/storageClassConfiguration
    (CreateTableBucket, PutTableBucketEncryption and PutTableBucketStorageClass carry them for the...
  - full notes: [details/S3T-TABLE-318.md](details/S3T-TABLE-318.md)

- <a id="s3t-tablemaintenanceconfiguration-012"></a>**S3T-TABLEMAINTENANCECONFIGURATION-012** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Maintenance defaults CAN be restored: Put auto/512 and 1/120 read back fresh-identical; status-only re-Put is a no-op but bumps modifiedAt**
  Fresh table: compaction {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}},
  "status": "enabled"}, snapshots {"settings": {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 120,
  "minSnapshotsToKeep": 1}}, "status": "enabled"}. Put {binpack,128} -> HTTP 204, reads {"settings":
  {"icebergCompaction": {"strategy": "binpack", "targetFileSizeMB": 128}}, "status": "enabled"}. Put {strategy
  auto} only -> HTTP 204, reads {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB":
  128}}, "status": "enabled"}. Put {binpack,128} again then Put {auto, 512} (the fresh defaults) -> HTTP 204,
  reads {"settings": {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status":
  "enabled"} (identical to fresh: True). Three Put {status enabled} only on that state -> HTTP 204; compaction
  changed: False, GetTable.modifiedAt changed: True. Put value {} -> HTTP 204, reads {"settings":
  {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status": "enabled"}. Snapshots Put
  {5,48} -> {"settings": {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 48, "minSnapshotsToKeep": 5}},
  "status": "enabled"} then Put {1,120} -> reads {"settings": {"icebergSnapshotManagement":
  {"maxSnapshotAgeHours": 120, "minSnapshotsToKeep": 1}}, "status": "enabled"} (identical to fresh: True). Put
  compaction {status disabled} -> reads {"settings": {"icebergCompaction": {"strategy": "auto",
  "targetFileSizeMB": 512}}, "status": "disabled"}; then {status enabled} -> reads {"settings":
  {"icebergCompaction": {"strategy": "auto", "targetFileSizeMB": 512}}, "status": "enabled"} (settings survive
  the flip: True).
  - ACK: late_initialize, compare.nil_equals_zero_value, custom_update · ops:
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration, GetTable · fields:
    icebergCompaction.strategy, icebergCompaction.targetFileSizeMB, icebergSnapshotManagement
  - repro: fresh table -> Put binpack/128 -> Put strategy auto -> Put binpack/128 -> Put auto/512 -> 3x Put
    status-only -> Put {} -> snapshots 5/48 -> 1/120 -> disabled -> enabled; Get after each
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-203](#s3t-table-203), [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-206](#s3t-table-206), [S3T-TABLE-227](#s3t-table-227), [S3T-TABLE-228](#s3t-table-228),
    [S3T-TABLE-230](#s3t-table-230), [S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001), [S3T-TABLE-232](#s3t-table-232), [S3T-TABLE-319](table-encryption-storage.md#s3t-table-319), [S3T-TABLE-265](#s3t-table-265) ·
    evidence: tablemaintenanceconfiguration/creative/reset-to-default
  - notes: qualifies [S3T-TABLE-205](#s3t-table-205) ('no reset'): the explicit enum value 'auto' is accepted and applied (alone
    it keeps the stored 128 MB; with 512 MB the read-back equals the fresh table byte-for-byte), so a
    controller CAN converge a removed spec field by sending the fresh-table defaults; there is still no...
  - full notes: [details/S3T-TABLEMAINTENANCECONFIGURATION-012.md](details/S3T-TABLEMAINTENANCECONFIGURATION-012.md)

## Response fidelity and consistency

- <a id="s3t-table-155"></a>**S3T-TABLE-155** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy/DeleteTablePolicy do NOT change GetTable versionToken (modifiedAt unchanged) across 4
  snapshots** - GetTable snapshots before Put, after 2 Puts, after 4 Puts, after Delete: versionToken distinct
  values=1, modifiedAt changed=False, modifiedBy changed=False. - see:
  [details/S3T-TABLE-155.md](details/S3T-TABLE-155.md)

- <a id="s3t-table-199"></a>**S3T-TABLE-199** `response-fidelity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Compaction strategy is always reported ('auto' on a fresh table); Put without strategy keeps the stored strategy (merge)**
  GetTableMaintenanceConfiguration strategy on a fresh table: auto. After Put icebergCompaction {enabled,
  {targetFileSizeMB 512, strategy auto}} -> Put HTTP 204, strategy read back auto. After Put {enabled,
  {targetFileSizeMB 512}} (no strategy) -> strategy read back auto. Full compaction entries: fresh {"status":
  "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 128, "strategy": "binpack"}}}; after auto
  {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}};
  after plain {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy":
  "auto"}}}.
  - ACK: compare.nil_equals_zero_value, compare.is_ignored+delta_pre_compare · ops:
    PutTableMaintenanceConfiguration, GetTableMaintenanceConfiguration · fields:
    configuration.icebergCompaction.settings.icebergCompaction.strategy
  - repro: Get fresh; Put compaction {512, auto}; Get; Put compaction {512}; Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-212](#s3t-table-212), [S3T-TABLE-183](#s3t-table-183), [S3T-TABLE-200](#s3t-table-200), [S3T-TABLE-201](#s3t-table-201),
    [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-203](#s3t-table-203), [S3T-TABLE-204](#s3t-table-204), [S3T-TABLE-211](#s3t-table-211), [S3T-TABLE-016](table-schema.md#s3t-table-016) · hypotheses: H-U-033 ·
    evidence: table/mutation-matrix/maintenance-config
  - notes: H-U-033 refuted: 'auto' is present on a never-configured table, so writing strategy=auto does not
    create a read-back difference; and omitting strategy in a later Put does not remove it (see the merge
    finding).
  - full notes: [details/S3T-TABLE-199.md](details/S3T-TABLE-199.md)

- <a id="s3t-table-206"></a>**S3T-TABLE-206** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableMaintenanceConfiguration does NOT change GetTable versionToken but DOES advance modifiedAt
  (modifiedBy stays absent)** - GetTable before the first maintenance Put and after 21 Puts (both types,
  enabled/disabled, partial values): versionToken identical (2c50...), modifiedAt advanced from the create
  time to the time of the last PutTableMa... - see: [details/S3T-TABLE-206.md](details/S3T-TABLE-206.md)

- <a id="s3t-table-208"></a>**S3T-TABLE-208** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceConfiguration immediately after Put: 0 stale reads in 20 trials (write->read gap 40-64
  ms)** - 20 alternating Puts of targetFileSizeMB 128/256 each followed immediately by a Get: 0 stale reads;
  the Get returned the new value 40-64 ms after the Put started. - see:
  [details/S3T-TABLE-208.md](details/S3T-TABLE-208.md)

- <a id="s3t-table-219"></a>**S3T-TABLE-219** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy / PutTableMaintenanceConfiguration 0-182 ms after CreateTable returned: HTTP 200 / HTTP 204
  in 20/20 trials** - 20 fresh tables: PutTablePolicy issued 0-0 ms after CreateTable returned -> ['HTTP
  200']; PutTableMaintenanceConfiguration issued 108-182 ms after -> ['HTTP 204']. - see:
  [details/S3T-TABLE-219.md](details/S3T-TABLE-219.md)

- <a id="s3t-table-220"></a>**S3T-TABLE-220** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable: all 200, 0 stale
  maintenance / 0 stale policy reads (n=20)** - First reads 165-244 ms after CreateTable returned (and right
  after the two Puts): GetTableMaintenanceConfiguration -> ['HTTP 200'] with (config keys, job keys) = [(2,
  3)]; GetTableMaintenanceJobStatus -> ['HTTP 200'];... - see:
  [details/S3T-TABLE-220.md](details/S3T-TABLE-220.md)

- <a id="s3t-tablemaintenanceconfiguration-001"></a>**S3T-TABLEMAINTENANCECONFIGURATION-001** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh table: GetTableMaintenanceJobStatus -> 200 listing all 3 job types (incl. bucket-level UFR)
  Not_Yet_Run, entry = {status} only (n=4)** - GetTableMaintenanceJobStatus 1-1 s after CreateTable on four
  fresh tables (with schema, without metadata, both kinds): t_schema -> HTTP 200
  icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUn... - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-001.md](details/S3T-TABLEMAINTENANCECONFIGURATION-001.md)

- <a id="s3t-tablemaintenanceconfiguration-009"></a>**S3T-TABLEMAINTENANCECONFIGURATION-009** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTableMaintenanceJobStatus entry members: Not_Yet_Run {status}, Disabled {status}, Successful {status,
  lastRunTimestamp}; absent = omitted** - Raw HTTP bodies captured before parsing. - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-009.md](details/S3T-TABLEMAINTENANCECONFIGURATION-009.md)

- <a id="s3t-tablemaintenanceconfiguration-011"></a>**S3T-TABLEMAINTENANCECONFIGURATION-011** `response-fidelity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Put(disabled) on a Failed job -> entry collapses to {status: Disabled} (lastRunTimestamp AND
  failureMessage dropped)...** (hypothesis refuted; behavior confirmed) - Compaction job had run and failed
  (KMS key disabled): raw entry {"failureMessage": "Insufficient access to perform table maintenance. - see:
  [details/S3T-TABLEMAINTENANCECONFIGURATION-011.md](details/S3T-TABLEMAINTENANCECONFIGURATION-011.md)

## Sub-resources

- <a id="s3t-table-201"></a>**S3T-TABLE-201** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Put compaction status=disabled with settings stores both: Get -> {"status": "disabled", "settings":
  {"icebergCompaction"...** - Put {disabled, {targetFileSizeMB 256}} -> HTTP 204, Get {"status": "disabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 256, "strategy": "auto"}}}. - see:
  [details/S3T-TABLE-201.md](details/S3T-TABLE-201.md)

## Delete semantics

- <a id="s3t-table-183"></a>**S3T-TABLE-183** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTable with a policy and custom maintenance -> HTTP 204; sub-resource Gets at 129 ms -> NotFoundException...**
  Table with policy Sid NS1, compaction 128 MB and snapshot management disabled (3/24):
  DeleteTable(versionToken) -> HTTP 204 with no prerequisite to remove the policy. Reads after delete:
  [{"at_ms": 129, "GetTablePolicy": {"code": "NotFoundException", "http": 404, "message": "The specified table
  does not exist."}, "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message":
  "The specified table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException",
  "http": 404, "message": "The specified table does not exist."}}, {"at_ms": 154, "GetTablePolicy": {"code":
  "NotFoundException", "http": 404, "message": "The specified table does not exist."},
  "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message": "The specified
  table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException", "http": 404,
  "message": "The specified table does not exist."}}, {"at_ms": 500, "GetTablePolicy": {"code":
  "NotFoundException", "http": 404, "message": "The specified table does not exist."},
  "GetTableMaintenanceConfiguration": {"code": "NotFoundException", "http": 404, "message": "The specified
  table does not exist."}, "GetTableMaintenanceJobStatus": {"code": "NotFoundException", "http": 404,
  "message": "The specified table does not exist."}}]. CreateTable with the same bucket/namespace/name -> new
  tableARN differs from the old: True. On the recreated table: GetTablePolicy -> NotFoundException/404 'The
  specified table policy does not exist.'; GetTableMaintenanceConfiguration compaction targetFileSizeMB=512,
  snapshot entry={"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 1,
  "maxSnapshotAgeHours": 120}}}; GetTableMaintenanceJobStatus statuses={"icebergCompaction": "Not_Yet_Run",
  "icebergUnreferencedFileRemoval": "Not_Yet_Run", "icebergSnapshotManagement": "Not_Yet_Run"}.
  - ACK: pre-delete-cleanup, exceptions.404, is_arn_primary_key · ops: DeleteTable, CreateTable,
    GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus
  - repro: PutTablePolicy + 2x PutTableMaintenanceConfiguration -> DeleteTable -> Gets at 0/100/500 ms ->
    CreateTable same triple -> Gets
  - measurements: post_delete_read_offsets_ms=[129, 154, 500], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-160](#s3t-table-160), [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-198](#s3t-table-198), [S3T-TABLE-204](#s3t-table-204), [S3T-TABLE-211](#s3t-table-211), [S3T-TABLE-199](#s3t-table-199),
    [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-212](#s3t-table-212), [S3T-TABLE-178](#s3t-table-178), [S3T-TABLE-179](#s3t-table-179), [S3T-TABLE-180](#s3t-table-180), [S3T-TABLE-182](#s3t-table-182), [S3T-TABLE-213](#s3t-table-213),
    [S3T-TABLE-214](table-policy.md#s3t-table-214), [S3T-TABLE-215](#s3t-table-215), [S3T-TABLE-216](table-policy.md#s3t-table-216) · hypotheses: H-U-020, H-U-030, H-U-083 · evidence:
    table/identity/subresource-identifiers
  - notes: H-U-030 confirmed (delete allowed; post-delete reads NotFoundException). H-U-020 refuted
    (sub-resources do not leak on the recreated triple). H-U-083 confirmed (job status after delete
    NotFoundException; recreated statuses {"icebergCompaction": "Not_Yet_Run",
    "icebergUnreferencedFileRemoval":...
  - full notes: [details/S3T-TABLE-183.md](details/S3T-TABLE-183.md)

- <a id="s3t-table-213"></a>**S3T-TABLE-213** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Table policy Deny s3tables:DeleteTable for the caller: own DeleteTable -> AccessDeniedException/403 after 0.03s...**
  PutTablePolicy (Deny DeleteTable + GetTableMaintenanceConfiguration, Principal = caller's role ARN,
  DateLessThan guard) -> HTTP 200. The read canary GetTableMaintenanceConfiguration polled at 1/s went
  ['AccessDeniedException'] and was first denied after 0.03 s (1 polls). DeleteTable(versionToken) while
  denied -> AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:DeleteTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3e8f-deny/tab'; GetTable afterwards -> HTTP 200.
  DeleteTablePolicy by the same (denied-for-delete) caller -> HTTP 204. Canary restored to 200 after 0.01 s
  (codes ['OK']). DeleteTable then -> HTTP 204 (after 1 attempt(s)).
  - ACK: terminal_codes, pre-delete-cleanup, requeue · ops: PutTablePolicy, DeleteTable, DeleteTablePolicy,
    GetTableMaintenanceConfiguration · fields: resourcePolicy
  - repro: PutTablePolicy Deny DeleteTable for own role -> poll a denied read at 1/s -> DeleteTable ->
    DeleteTablePolicy -> poll -> DeleteTable
  - measurements: deny_enforced_after_s=0.03, deny_lifted_after_s=0.01, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-214](table-policy.md#s3t-table-214), [S3T-TABLE-215](#s3t-table-215), [S3T-TABLE-216](table-policy.md#s3t-table-216), [S3T-TABLE-183](#s3t-table-183) · hypotheses: H-U-016 · evidence:
    table/dependencies/policy-deny-enforcement
  - notes: H-U-016 confirmed: denial code AccessDeniedException (HTTP 403), enforced within 0.03 s,
    DeleteTablePolicy still allowed=True, delete after removal=True. Exact message: 'User:
    arn:aws:sts::<account>:assumed-role/<role>/<session> is not authorized to perform: s3tables:DeleteTable on
    resource: <table...
  - full notes: [details/S3T-TABLE-213.md](details/S3T-TABLE-213.md)

## Quotas and rate limits

- <a id="s3t-table-187"></a>**S3T-TABLE-187** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **30 concurrent table reads x2 (GetTable, ListTables, GetTableEncryption+StorageClass,
  GetTablePolicy+Maintenance): 0 TooManyRequests** - Bursts of 30 concurrent reads (20 threads) against one
  table, two runs 20 s apart: GetTable -> {"200": 30} / {"200": 30} (wall 0.16 / 0.15 s); ListTables ->
  {"200": 30} / {"200": 30} (wall 0.10 / 0.08 s); GetTableEncr... - see:
  [details/S3T-TABLE-187.md](details/S3T-TABLE-187.md)

- <a id="s3t-table-303"></a>**S3T-TABLE-303** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **10 concurrent CreateTable into a FRESH namespace: wall 10.5 s (n=2; median 0.66 s, one call 10.45 s); warm
  namespace 0.7 s, 5 fresh <= 1.6 s** - Rounds of concurrent calls (ThreadPool, 6 s apart); contention scope:
  first-use only (second burst in the same namespace fast). - see:
  [details/S3T-TABLE-303.md](details/S3T-TABLE-303.md)

## Adoption and first-sync hazards

- <a id="s3t-table-212"></a>**S3T-TABLE-212** `first-sync-destructive` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance config always exists with server defaults and policy may exist out-of-band: a nil spec must mean 'unmanaged', not 'reset/delete'**
  A never-configured table already returns a full maintenance map (compaction enabled/512, snapshots
  enabled/1/120, strategy absent) and no API distinguishes 'never set' from 'user set the defaults'; an
  explicit Put of strategy makes 'auto' appear permanently (auto). A table policy is a separate sub-resource
  that returns 404 when absent. A reconcile that treats an empty spec as 'write defaults' or 'delete policy'
  alters state the user never declared.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare, custom_delete · ops:
    GetTableMaintenanceConfiguration, PutTableMaintenanceConfiguration, GetTablePolicy, DeleteTablePolicy
  - repro: derived from the fresh-table defaults and strategy findings in this probe
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-199](#s3t-table-199), [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-183](#s3t-table-183) · hypotheses: H-U-079 · evidence:
    table/mutation-matrix/maintenance-config
  - notes: H-U-079 confirmed as a design consequence of H-U-031/H-U-033 observations in this probe (not a
    separate experiment).
  - full notes: [details/S3T-TABLE-212.md](details/S3T-TABLE-212.md)

## Scope

- <a id="s3t-table-160"></a>**S3T-TABLE-160** `scope` · impact medium · handled · verified 2026-10-13
  **PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval) -> BadRequestException (HTTP
  400)...** **Scope verdict: field-on-parent** - On a fresh ICEBERG table:
  PutTableMaintenanceConfiguration(type=icebergUnreferencedFileRemoval, value={status: enabled}) ->
  BadRequestException (HTTP 400) 'Invalid table maintenance type: icebergUnreferencedFileRemova... - see:
  [details/S3T-TABLE-160.md](details/S3T-TABLE-160.md)

## Codegen notes

- <a id="s3t-table-211"></a>**S3T-TABLE-211** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Maintenance model hazards: enum value "z-order" (hyphen), required path member named "type", settings is a
  tagged union keyed by type** - botocore model: PutTableMaintenanceConfiguration input members
  tableBucketARN, namespace, name, type (enum TableMaintenanceType: icebergCompaction,
  icebergSnapshotManagement), value {status: enabled|disabled, settings... - see:
  [details/S3T-TABLE-211.md](details/S3T-TABLE-211.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-054](#s3t-table-054) | CreateTable(aws:kms, key without maintenance.s3tables principal) -> BadRequestException HTTP 400 | create_table_rejected_latency_ms=[509] | 1 |
| [S3T-TABLE-056](#s3t-table-056) | Maintenance-principal key-policy Condition at CreateTable: <bucket>/table/<name>/* -> 400; <bucket>/table/*, <bucket>/*, no Condition -> 200 | attempts_per_cell=[3, 1, 1, 1] | 4 |
| [S3T-TABLE-110](#s3t-table-110) | versionToken is NOT changed by TagResource, PutTableMaintenanceConfiguration or RenameTable (newName / newNamespaceName)... | rename_latency_ms=[130, 141] | 2 |
| [S3T-TABLE-183](#s3t-table-183) | DeleteTable with a policy and custom maintenance -> HTTP 204; sub-resource Gets at 129 ms -> NotFoundException... | post_delete_read_offsets_ms=[129, 154, 500] | 3 |
| [S3T-TABLE-187](#s3t-table-187) | 30 concurrent table reads x2 (GetTable, ListTables, GetTableEncryption+StorageClass, GetTablePolicy+Maintenance): 0 TooManyRequests | burst_wall_s.GetTable=[0.16, 0.15], burst_wall_s.ListTables=[0.1, 0.08], burst_wall_s.GetTableEncryption+GetTableStorageClass=[0.05, 0.05], burst_wall_s.GetTablePolicy+GetTableMaintenanceConfiguration=[0.13, 0.11], latency_ms_min_max.GetTable=[[9,149], [10,125]], latency_ms_min_max.ListTables=[[12,80], [11,59]], latency_ms_min_max.GetTableEncryption+GetTableStorageClass=[[6,40], [6,37]], latency_ms_min_max.GetTablePolicy+GetTableMaintenanceConfiguration=[[6,112], [6,90]] | 270 |
| [S3T-TABLE-191](#s3t-table-191) | Concurrent writes to one table's policy/maintenance -> ConflictException 409 (1 of 10 also 500); compaction and snapshot Puts conflict | policy_race_wall_s=6.22, policy_latency_ms_min_max=[155, 6210], pair_wall_s=[1.08, 1.04, 1.05, 1.05, 1.04, 1.05, 1.04, 1.04, 1.04, 1.04] | 30 |
| [S3T-TABLE-197](#s3t-table-197) | Fresh table GetTableMaintenanceConfiguration: compaction enabled/512 MB strategy auto; snapshots enabled 1/120h... | first_get_after_create_ms=619 | 1 |
| [S3T-TABLE-198](#s3t-table-198) | Fresh table GetTableMaintenanceJobStatus: keys ['icebergCompaction', 'icebergSnapshotManagement', 'icebergUnreferencedFileRemoval']... | poll_offsets_s=[0, 1, 5, 30, 60] | 5 |
| [S3T-TABLE-207](#s3t-table-207) | Same maintenance value re-sent and 10 enabled/disabled flips back-to-back: codes ['HTTP 204']; final status enabled (last Put enabled) | put_latency_ms_min=105, put_latency_ms_max=159 | 10 |
| [S3T-TABLE-208](#s3t-table-208) | GetTableMaintenanceConfiguration immediately after Put: 0 stale reads in 20 trials (write->read gap 40-64 ms) | trials=20, stale_reads=0, gap_ms_min=40, gap_ms_max=64 | 20 |
| [S3T-TABLE-209](#s3t-table-209) | GetTableMaintenanceJobStatus.icebergCompaction shows Disabled 0.2 s after Put status=disabled and Not_Yet_Run 0.2 s after re-enable | disabled_visible_after_s=0.2, back_after_s=null | 1 |
| [S3T-TABLE-213](#s3t-table-213) | Table policy Deny s3tables:DeleteTable for the caller: own DeleteTable -> AccessDeniedException/403 after 0.03s... | deny_enforced_after_s=0.03, deny_lifted_after_s=0.01 | 1 |
| [S3T-TABLE-215](#s3t-table-215) | Table policy Deny PutTableMaintenanceConfiguration: Put -> AccessDeniedException/403 after 0.03 s; GetTableMaintenanceConfiguration -> 200 | deny_enforced_after_s=0.03, deny_lifted_after_s=0.04 | 1 |
| [S3T-TABLE-219](#s3t-table-219) | PutTablePolicy / PutTableMaintenanceConfiguration 0-182 ms after CreateTable returned: HTTP 200 / HTTP 204 in 20/20 trials | put_policy_offset_ms_min=0, put_policy_offset_ms_max=0, put_maint_offset_ms_min=108, put_maint_offset_ms_max=182, create_latency_ms_min=307, create_latency_ms_max=416 | 20 |
| [S3T-TABLE-220](#s3t-table-220) | GetTableMaintenanceConfiguration/JobStatus/Policy 165-244 ms after CreateTable: all 200, 0 stale maintenance / 0 stale policy reads (n=20) | first_read_offset_ms_min=165, first_read_offset_ms_max=244, stale_maint=0, stale_policy=0 | 20 |
| [S3T-TABLE-233](#s3t-table-233) | Role lacking the IAM action: Get/Put/DeleteTablePolicy, GetTableMaintenanceJobStatus, DeleteTable -> AccessDeniedException 403, IAM message | assume_role_after_s=9.1, policy_effective_after_s=0.0 | 1 |
| [S3T-TABLE-236](#s3t-table-236) | Key grant revoked: UnreferencedFileRemoval job Not_Yet_Run -> Failed 5.0 min after a metadata-only commit; CreateTable maintenance text, n=1 | since_commit_min=[5.0] | 1 |
| [S3T-TABLE-237](#s3t-table-237) | Key grant revoked: Compaction job Not_Yet_Run -> Failed 6.0 min after a metadata-only commit; failureMessage = CreateTable maintenance text | since_commit_min=[6.0] | 1 |
| [S3T-TABLE-238](#s3t-table-238) | Key grant revoked: SnapshotManagement job Not_Yet_Run -> Failed 8.1 min after a metadata-only commit; CreateTable maintenance text (n=1) | since_commit_min=[8.1] | 1 |
| [S3T-TABLE-246](#s3t-table-246) | CreateTable(aws:kms) with an RSA_2048 or HMAC_256 key -> BadRequestException 400 maintenance-principal text; bucket Put accepts them | create_table_latency_ms.rsa=414, create_table_latency_ms.hmac=411, create_table_latency_ms.symmetric=652 | 1 |
| [S3T-TABLE-253](#s3t-table-253) | CreateTable(aws:kms) 400 'Insufficient access to perform table maintenance' flips on the FIRST call after PutKeyPolicy (0.4-0.6 s, n=2) | grant_to_200_s=[0.51, 0.62], revoke_to_400_s=[0.37, 0.42], attempts_grant=[1, 1], attempts_revoke=[1, 1], poll_interval_s=1.0 | 2 |
| [S3T-TABLE-265](#s3t-table-265) | Bucket-level Put(disabled) on icebergUnreferencedFileRemoval -> only that job reads 'Disabled' in each table's job status; newborns too | t1_first_read_after_put_s=0.545, t1_change_at_s=0.11, t1_reenable_at_s=0.49 | 1 |
| [S3T-TABLE-303](#s3t-table-303) | 10 concurrent CreateTable into a FRESH namespace: wall 10.5 s (n=2; median 0.66 s, one call 10.45 s); warm namespace 0.7 s, 5 fresh <= 1.6 s | R1_create_table_one_namespace.wall_s=10.46, R1_create_table_one_namespace.lat_ms_sorted=[424, 452, 476, 524, 524, 788, 816, 900, 909, 10453], R1_create_table_one_namespace.codes.200=10, R2_create_table_two_buckets.wall_s=1.63, R2_create_table_two_buckets.lat_ms_sorted=[328, 358, 377, 390, 401, 612, 663, 795, 856, 1616], R2_create_table_two_buckets.codes.200=10, R3_create_table_two_namespaces.wall_s=0.87, R3_create_table_two_namespaces.lat_ms_sorted=[325, 337, 352, 375, 459, 540, 549, 606, 655, 862], R3_create_table_two_namespaces.codes.200=10, R1b_create_table_same_namespace_again.wall_s=0.65, R1b_create_table_same_namespace_again.lat_ms_sorted=[346, 359, 409, 426, 426, 485, 516, 570, 631, 642], R1b_create_table_same_namespace_again.codes.200=10, R8_create_table_10_in_ns2.wall_s=0.74, R8_create_table_10_in_ns2.lat_ms_sorted=[319, 332, 341, 350, 379, 569, 588, 598, 627, 730], R8_create_table_10_in_ns2.codes.200=10, R9_create_table_7_in_b1.wall_s=0.67, R9_create_table_7_in_b1.lat_ms_sorted=[333, 343, 370, 408, 427, 496, 669], R9_create_table_7_in_b1.codes.200=7, R4_put_maintenance_10_tables.wall_s=0.09, R4_put_maintenance_10_tables.lat_ms_sorted=[52, 55, 59, 61, 65, 65, 69, 70, 71, 82], R4_put_maintenance_10_tables.codes.204=10, R5_tag_10_tables.wall_s=0.17, R5_tag_10_tables.lat_ms_sorted=[124, 127, 128, 131, 132, 145, 146, 153, 158, 165], R5_tag_10_tables.codes.200=10, R6_create_namespace_x10.wall_s=0.15, R6_create_namespace_x10.lat_ms_sorted=[33, 36, 38, 42, 67, 97, 107, 112, 116, 137], R6_create_namespace_x10.codes.200=10, R7_delete_table_x10.wall_s=0.44, R7_delete_table_x10.lat_ms_sorted=[134, 155, 162, 173, 195, 278, 307, 324, 366, 429], R7_delete_table_x10.codes.204=10, n_R1=2, R1_wall_s_runs=[10.49, 10.46] | 1 |
| [S3T-TABLE-327](#s3t-table-327) | Role allowed only s3tables:GetTable: every table sub-resource Get -> AccessDeniedException 403 naming its own action (30/30, 3 fresh roles) | roles=3, passes=2, sub_read_cells=30, first_pass_after_put_role_policy_s=30.1, second_pass_after_put_role_policy_s=238.2 | 30 |
| [S3T-TABLEMAINTENANCECONFIGURATION-001](#s3t-tablemaintenanceconfiguration-001) | Fresh table: GetTableMaintenanceJobStatus -> 200 listing all 3 job types (incl. bucket-level UFR) Not_Yet_Run, entry = {status} only (n=4) | age_at_sample_s=[1.4, 1.1, 0.9, 0.7] | 4 |
| [S3T-TABLEMAINTENANCECONFIGURATION-002](#s3t-tablemaintenanceconfiguration-002) | PutTableMaintenanceConfiguration(disabled) -> job status 'Disabled' on the first read (<1 s) for snapshot mgmt and compaction... | disabled_visible_after_put_s=[0.0, 0.0], burst_polls=[1, 1] | 2 |
| [S3T-TABLEMAINTENANCECONFIGURATION-003](#s3t-tablemaintenanceconfiguration-003) | t_prop: compaction Not_Yet_Run -> Successful (lastRunTimestamp set) seen 2.8 min after a metadata-only UpdateTableMetadataLocation | commit_to_successful_seen_min=[2.8], transition_at_table_age_s=[168] | 1 |
| [S3T-TABLEMAINTENANCECONFIGURATION-004](#s3t-tablemaintenanceconfiguration-004) | t_prop: snapshot management Not_Yet_Run -> Successful (lastRunTimestamp set) seen 4.8 min after a metadata-only UpdateTableMetadataLocation | commit_to_successful_seen_min=[4.8], transition_at_table_age_s=[290] | 1 |
| [S3T-TABLEMAINTENANCECONFIGURATION-005](#s3t-tablemaintenanceconfiguration-005) | t_prop: unreferenced-file removal Not_Yet_Run -> Successful (lastRunTimestamp set) 6.8 min after a metadata-only UpdateTableMetadataLocation | commit_to_successful_seen_min=[6.8], transition_at_table_age_s=[411] | 1 |
| [S3T-TABLEMAINTENANCECONFIGURATION-006](#s3t-tablemaintenanceconfiguration-006) | Untouched tables stay Not_Yet_Run (no-metadata 60 min, schema 31 min); a compaction settings Put did not trigger a run in 29 min (n=1 each) | soak_window_min=[60.0], untouched_min=[30.8, 60.0], nudge_to_first_run_min=[] | 61 |
| [S3T-TABLEMAINTENANCECONFIGURATION-007](#s3t-tablemaintenanceconfiguration-007) | Metadata-only UpdateTableMetadataLocation triggers all 3 maintenance jobs within 7 min (Successful +2.8/+4.8/+6.8 min)... | commit_to_first_successful_min=[2.8, 4.8, 6.8], soak_window_min=[60.0] | 3 |
| [S3T-TABLEMAINTENANCECONFIGURATION-008](#s3t-tablemaintenanceconfiguration-008) | Disabled job status is stable: compaction + snapshot mgmt stayed 'Disabled' 60 min after Put(disabled); unconfigured UFR stayed Not_Yet_Run | soak_window_min=[60.0] | 61 |
| [S3T-TABLEMAINTENANCECONFIGURATION-010](#s3t-tablemaintenanceconfiguration-010) | Re-enabling snapshot management after 60 min Disabled (never run) -> 'Not_Yet_Run' on the first read; compaction left Disabled | disabled_for_min=[60.0], burst_polls=[1] | 1 |
| [S3T-TABLEMAINTENANCECONFIGURATION-013](#s3t-tablemaintenanceconfiguration-013) | PutTableMaintenanceConfiguration sort/z-order: no metadata location -> 400 for any caller... | runs=3, schema_table_runs=2 | 3 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-maintenance.md` - unverified or single-run items worth a future probe, one line each.

- [S3T-TABLE-236](#s3t-table-236) / -237 / -238 (n=1 each): the Not_Yet_Run -> Failed transitions after a key-grant revocation were observed on one table in one phase; whether a re-granted key turns Failed back to Successful on the next run, and after how long, was not watched.
- [S3T-TABLEMAINTENANCECONFIGURATION-006](#s3t-tablemaintenanceconfiguration-006) (n=1): "a compaction settings Put did not trigger a run in 29 min" is one table; a status flip disabled -> enabled (rather than a settings change) as a trigger was not tried on an untouched table.
- [S3T-TABLEMAINTENANCECONFIGURATION-007](#s3t-tablemaintenanceconfiguration-007) / -003 / -004 / -005 (one table): the 2.8 / 4.8 / 6.8 min run offsets come from a single metadata-only commit sampled every 60 s; the per-job order and spacing may be scheduling noise, not a property.
- [S3T-TABLEMAINTENANCECONFIGURATION-010](#s3t-tablemaintenanceconfiguration-010) (n=1): only snapshot management was re-enabled after 60 min Disabled; compaction was left Disabled, so the compaction re-enable path after a long Disabled is untested.
- [S3T-TABLE-209](#s3t-table-209) (n=1): `back_after_s` is null - the Not_Yet_Run after re-enable was seen on the first read but the sampling could not bound it from below.
- [S3T-TABLE-212](#s3t-table-212) vs -197 / -199 / -205: the consolidation note keeps -212 for its design claim while its "strategy absent on a fresh table" text is superseded; a title fix on -212 would remove the contradiction from the rendered page.
- [S3T-TABLE-191](#s3t-table-191) (one burst each): whether the 409 loser between the compaction and snapshot Puts is deterministic (it alternated in 10 rounds) and whether the 1-in-10 InternalServerErrorException 500 on the policy race reproduces.
- [S3T-TABLE-228](#s3t-table-228): the unknown-union-member cell (icebergFoo) never left boto3 (ClientSerializationError), so the raw-wire response to a member the model does not know is untested.
- [S3T-TABLE-226](#s3t-table-226) / -227 (one cell each): the 100.5 -> 100 truncation and the 2147483648 -> 2147483647 clamp were seen once with SDK validation off; unreachable from an integer CRD field but unconfirmed as stable.
- [S3T-TABLE-213](#s3t-table-213) / -215 (n=1 each): the 0.03 s Deny enforcement and 0.01-0.04 s lift are single runs with a DateLessThan guard; no sweep at 1 s steps.
- [S3T-TABLE-253](#s3t-table-253) (n=2): the first-call flip of the CreateTable maintenance-principal check after PutKeyPolicy rests on two grant/revoke cycles at 1 s polling.
- [S3T-TABLE-265](#s3t-table-265) (n=1 table): the bucket-level UFR Put(disabled) -> per-table 'Disabled' propagation was watched on one existing table and one newborn; a bucket with many tables was not tried.
- [S3T-TABLE-303](#s3t-table-303) (n=2 runs): the 10.45 s outlier among 10 concurrent CreateTable calls into a fresh namespace appeared in one of two runs and was not isolated from the PutTableMaintenanceConfiguration calls issued in the same probe.
- [S3T-TABLE-197](#s3t-table-197) / -198 (n=1): the 619 ms first Get after the CreateTable request is a single sample that includes the CreateTable latency; -220 (165-244 ms, n=20) is the range to quote.
- [S3T-TABLE-021](table-encryption-storage.md#s3t-table-021) / -058: both landed in this stem through the `snapshot` include on "creation-time snapshot"; whether they should render with `table-encryption-storage.md` instead is a render.yaml question for the operator, not an API probe.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-207"></a>**S3T-TABLE-207** | idempotency | low | confirmed | Same maintenance value re-sent and 10 enabled/disabled flips back-to-back: codes ['HTTP 204']; final status enabled (last Put enabled) | [S3T-TABLE-208](#s3t-table-208), [S3T-TABLE-209](#s3t-table-209), [S3T-TABLE-191](#s3t-table-191) | - |
| <a id="s3t-table-210"></a>**S3T-TABLE-210** | response-fidelity | low | confirmed | Wire shapes: PutTableMaintenanceConfiguration HTTP 204 body ''; GetTableMaintenanceConfiguration/JobStatus 200 with tableARN echoed | [S3T-TABLE-159](table-policy.md#s3t-table-159) | - |
| <a id="s3t-table-254"></a>**S3T-TABLE-254** | prerequisite | medium | confirmed | Bucket default aws:kms on a grant-less key: inheriting CreateTable -> BadRequestException (HTTP 400)... | [S3T-TABLE-055](#s3t-table-055), [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-253](#s3t-table-253), [S3T-TABLE-301](#s3t-table-301), [S3T-TABLE-246](#s3t-table-246), [S3T-TABLE-248](table-encryption-storage.md#s3t-table-248), [S3T-TABLE-054](#s3t-table-054), [S3T-TABLE-239](table-record-expiration.md#s3t-table-239) | [S3T-TABLE-301](#s3t-table-301) |
| <a id="s3t-table-256"></a>**S3T-TABLE-256** | other | low | confirmed | Doc claim C014 TRUE: CreateTable{aws:kms, key WITHOUT maintenance grant} -> BadRequestException 400 synchronously; GetTable -> 404 | [S3T-TABLE-054](#s3t-table-054), [S3T-TABLEBUCKET-139](tablebucket-encryption-storage.md#s3t-tablebucket-139), [S3T-TABLEBUCKET-142](tablebucket-encryption-storage.md#s3t-tablebucket-142), [S3T-TABLE-257](table-encryption-storage.md#s3t-table-257), [S3T-TABLEBUCKET-134](tablebucket-encryption-storage.md#s3t-tablebucket-134), [S3T-TABLEBUCKET-119](tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-055](#s3t-table-055) | - |
| <a id="s3t-table-294"></a>**S3T-TABLE-294** | other | low | confirmed | Doc claim C052 TRUE: PutTableMaintenanceConfiguration icebergCompaction enabled/256/binpack -> 204; Get returns exactly those values | [S3T-TABLE-203](#s3t-table-203), [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-226](#s3t-table-226), [S3T-TABLE-230](#s3t-table-230), [S3T-TABLE-295](#s3t-table-295), [S3T-TABLE-296](#s3t-table-296), [S3T-TABLE-298](#s3t-table-298), [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-297](#s3t-table-297), [S3T-TABLE-204](#s3t-table-204), [S3T-TABLE-232](#s3t-table-232), [S3T-TABLE-234](#s3t-table-234) | - |
| <a id="s3t-table-295"></a>**S3T-TABLE-295** | other | low | confirmed | Doc claim C053 TRUE: Put icebergSnapshotManagement after Put icebergCompaction leaves BOTH entries in the config map (per-type upsert) | [S3T-TABLE-200](#s3t-table-200), [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-294](#s3t-table-294), [S3T-TABLE-296](#s3t-table-296), [S3T-TABLE-298](#s3t-table-298), [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-202](#s3t-table-202), [S3T-TABLE-205](#s3t-table-205) | - |
| <a id="s3t-table-298"></a>**S3T-TABLE-298** | other | low | confirmed | Doc claim C056 TRUE: Fresh table GetTableMaintenanceConfiguration -> 200: compaction enabled/512/auto, snapshots enabled/1/120h | [S3T-TABLE-197](#s3t-table-197), [S3T-TABLE-205](#s3t-table-205), [S3T-TABLE-220](#s3t-table-220), [S3T-TABLE-294](#s3t-table-294), [S3T-TABLE-295](#s3t-table-295), [S3T-TABLE-296](#s3t-table-296), [S3T-TABLE-201](#s3t-table-201), [S3T-TABLE-202](#s3t-table-202) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
