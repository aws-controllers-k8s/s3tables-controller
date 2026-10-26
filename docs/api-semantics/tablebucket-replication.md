<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# TableBucket replication (PutTableBucketReplication, GetTableBucketReplication, GetTableReplicationStatus at bucket scope)
_Bucket-level replication rules: versionToken locking, role prerequisites, destination lifecycle, ~60 s status evaluation, replica creation and re-creation, scope verdict for a TableBucketReplication resource._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `tablebucket-replication.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the bucket-level replication configuration - PutTableBucketReplication, GetTableBucketReplication
and DeleteTableBucketReplication on a table bucket ARN (one rule, one role, 1-5 destination bucket ARNs) - the per-table
GetTableReplicationStatus read that is the only progress signal under a bucket-level rule, the role and destination
prerequisites, and what happens to replicas when sources are renamed, deleted or re-created and when destinations or the
configuration itself disappear. The write path is synchronous and versionToken-locked: Put returns {versionToken, status:
'success'} where 'success' is a literal operation result, Get returns exactly ['configuration', 'versionToken'] with no
status, health or progress member, every accepted Put rotates the token (a byte-identical re-send included), and an
update or delete of an existing configuration needs the current token (token-less -> BadRequestException 400 'A version
token is not specified.', stale or garbage -> ConflictException 409; the model marks the token optional on both)
([S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007), -018, -001, -002, -004, -006, -019, -026). Replication itself is asynchronous and visible
only per table: GetTableReplicationStatus(table) works under the bucket-level rule while GetTableReplication(table) is
404, a fresh Put carries a ~60 s warm-up after which new tables complete in seconds, existing tables are backfilled, and
after a Delete the status is 404 while the replicas stay behind as read-only service-managed tables
([S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), -031, -014, -015). Put validates only ARN syntax and list lengths - a nonexistent role and
a nonexistent, duplicate, self or foreign-account destination are all 200 and echoed back, a missing or under-permissioned
role surfaces only as a per-table 'failed', and the one synchronous role check is the caller's own iam:PassRole
([S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), -021, -012, [S3T-TABLEBUCKET-021](#s3t-tablebucket-021)). Scope verdict: the findings show
the API shape of a standalone resource - a Put/Get/Delete triple keyed by the bucket ARN, NotFoundException 404 'The
specified replication configuration does not exist.' when absent (never-set and after-delete share the message;
Delete-when-absent is 404, not 204), optimistic locking on a server-held versionToken that must live in status and be
refreshed on every write, and no bucket-level status to reconcile - so a TableBucketReplication CRD is implementable on
this API, with the role as a prerequisite the API will not validate and health observable only by fanning out
GetTableReplicationStatus over the bucket's tables ([S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), -005, -023, -007, -029). The
controller implements no replication resource or field today (render.yaml partition note; generator.yaml declares Table,
Namespace and TableBucket and wires none of the *TableBucketReplication operations), so "Handling gaps" lists what a
future implementation must handle rather than defects.

### Rules

1. Never re-Put on every reconcile: a byte-identical Put with the current token is 200 and ROTATES the token, the previous
   token then fails with ConflictException 409 - a perpetual re-Put churns the token held by any other writer and makes
   the token in status stale each reconcile, while it does not re-replicate 'completed' tables (unchanged 150 s, replica
   GetTable byte-identical), restart an in-flight backfill or add a new ~60 s warm-up ([S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002),
   -027, -019).
2. Carry the token read from Get on every update: a token-less Put while a configuration exists -> BadRequestException
   400 'A version token is not specified.' (old destinations and token kept); a stale or garbage token ('abc') ->
   ConflictException 409 'The specified version token does not match the configuration.' with the token unchanged; 20
   Gets at 1/s returned one identical token ([S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), -002, -019, -020, -017).
3. Delete only with the current token and treat the model's "optional" versionToken as mandatory: token-less Delete on an
   existing configuration -> 400 'A version token is not specified.' twice in a row, configuration still readable for 5 s
   of 200 ms polls; a token from two Puts ago or 'abc' -> 409, configuration still present; the current token -> 204 and
   Get 404 on the first 200 ms poll ([S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006), -004, -026).
4. On a 404 from Put, clear the stored token and re-Put without one: after a Delete, Get -> 404 with no versionToken in
   the error body, Put with the pre-delete token -> NotFoundException 404 (not 409), token-less Put -> 200 with a token
   unlike every earlier one; on a never-configured bucket Get -> 404 and Delete -> 404 with or without a token (never 400;
   the existence check precedes the token check) - Delete-when-absent is NOT the idempotent 204 of the botocore
   'idempotent' trait, unlike policy and metrics ([S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), -008, -023, [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), -096,
   -098).
5. Deleting the configuration orphans the replicas, it does not remove them: after Delete(current token) both
   destinations still list all 6 replicas, each keeps managedByService='replication.s3tables.amazonaws.com' and
   UpdateTableMetadataLocation on it -> 400 'Updates to service managed tables are not allowed.'; the sources' status ->
   404; a new commit and a new source table were not replicated in 120 s; a Delete 2.1 s after a Put on a bucket with 10
   tables left 0 of 10 replicas after 300 s; the destination stays undeletable (400 'not empty') until its replicas and
   namespace are deleted ([S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), -024, -025, -016, [S3T-TABLEBUCKET-118](#s3t-tablebucket-118)).
6. A source Table deleted and re-created under the same namespace/name is permanently 'failed': DeleteTable(source) -> 204
   leaves the replica in D and the status by the old ARN -> NotFoundException; CreateTable(same name) -> 200 with a NEW
   ARN, status 'no-destinations' then 'failed' 'A table with an identical name already exists at the destination.' at
   +0.2/5.4 s (n=2) and +5.2 s (n=1), the replica's ARN unchanged and its replicationInformation.sourceTableARN still the
   OLD source; a replica deleted in the destination IS re-created: at +31.1/36.2 s with no source activity (-034, n=2) and at +46.7 s in a
   run with two source deletes in the same second (-033) (completed -> pending -> completed), correcting -032's 'not re-created in 120 s' clause ([S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), -033, -034).
7. RenameTable on a source is not followed by the replica: after RenameTable(t -> t2) -> 204 the destination listed the old
   name for 153 s and after RenameTable(t2 -> ns2/t2) -> 204 the old namespace for 122 s, GetTable(replica) kept the old
   name/namespace, the source status stayed 'completed', and a control CreateTable was listed in 5.2 s (empty tables,
   same-region destination, n=1) ([S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031)).
8. Do not expect Put to reject a bad destination list - pre-validate in the CRD: [D, D] -> 200 and stored twice, [V] (self)
   -> 200, [V, D] -> 200, a well-formed ARN of a nonexistent same-region or other-region bucket -> 200 and echoed, a
   foreign-account bucket ARN -> 200, three sources into one destination and a back-edge cycle -> 200; only a malformed
   ARN, >5 destinations, 2 rules, 0 rules or 0 destinations are rejected, with the wire code ValidationException 400 that
   the model does not list ([S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), -017, -011, -022, [S3T-TABLEBUCKET-021](#s3t-tablebucket-021)).
9. The role is a prerequisite the API never checks at Put time: a role ARN that does not exist -> 200 'success' and Get
   echoes it (5 destinations + nonexistent role -> 200 too); a role created 4.9 s earlier was accepted on the first try, 0
   rejections (n=1); insufficient role permissions appear only as a per-table 'failed' 'Insufficient permissions given to
   successfully complete replication.' (~66 s, table-level evidence cited by -021); a bare role name instead of an ARN ->
   BadRequestException 400 with the misleading 'Must provide either table arn alone OR all of (table bucket arn,
   namespace, table name) together' ([S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-021](#s3t-tablebucketreplication-021), -017, -003).
10. The caller needs iam:PassRole on the configured role: with s3tables:* allowed and iam:PassRole explicitly denied,
    PutTableBucketReplication and PutTableReplication -> AccessDeniedException 403 with the IAM-style 'User: <caller> is
    not authorized to perform: iam:PassRole on resource: <role> with an explicit deny in an identity-based policy'; the
    same caller's GetTableBucket -> 200 and GetTableBucketReplication -> 404 ([S3T-TABLEBUCKETREPLICATION-012](#s3t-tablebucketreplication-012)).
11. There is no bucket-level health and the read-back never reflects a broken state: Put's status='success' is a literal
    result, Get has only configuration+versionToken; with the destination bucket emptied and deleted, Get still echoed it
    2 s later, the 'completed' table stayed 'completed', a new table went 'none' -> 'failed' 'The specified destination
    bucket does not exist.' at 23.9 s and a commit on the old table 'pending' -> 'failed' at 2.1 s; a role deleted
    out-of-band was still echoed and tables created 5-15 s later still completed in 2.1-3.2 s (~20 s window, n=1)
    ([S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007), -018, -029, -028).
12. Recover a dead destination by re-pointing, not re-creating: CreateTableBucket(same destination name) 28.6 s after its
    deletion -> ConflictException 409 'The bucket is in a transitional state because of a previous deletion attempt.';
    Put(new destination, current token) -> 200, the failed table 'completed' at 63.3 s and the old 'completed' table
    'pending' again ([S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030)).
13. Rules cannot be disabled, only deleted: a raw PUT whose rule carries 'status':'Disabled', an unknown rule member or a
    top-level 'status' -> 200 'success' and the GET body has no status member; the SDK model has no rule status
    ([S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010)).
14. Classify 404s by message, 403s by code: a foreign-account bucket ARN as the target -> ForbiddenException 403
    'Unauthorized' on Get/Put/Delete replication (not 404); a never-existing or just-deleted bucket -> NotFoundException 404
    'The specified bucket does not exist.' on every sub-resource Get and Delete, the same code as the config-absent
    message; DeleteTableBucket -> 204 on an empty bucket still carrying a configuration (n=1) and a same-name re-created
    bucket reads replication 404 with a new tableBucketId ([S3T-TABLEBUCKET-021](#s3t-tablebucket-021), -097, -098, -101, -012, -116).
15. The ~60 s warm-up belongs to the configuration, not to each table: tables created 1-2 s after a fresh Put reported
    'none'/'no-destinations' until ~61 s and completed at 61.0-66.0 s (empty) or ~336 s (schema), existing tables were
    backfilled at 70.7 s (empty) and 443.2-484.5 s (schema), a Put during the backfill was accepted (200, new token), and
    tables created after the warm-up completed in 2.1-10.6 s ([S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), -031, -013, -014, -033,
    -034, -028).

### Timing ranges

- First replicas after a fresh Put (empty tables created 1-2 s after the Put or already present): listed/'completed' at
  10.7-70.7 s, n=12 tables across 7 probes - 10.7 s x2 ([S3T-TABLEBUCKETREPLICATION-034](#s3t-tablebucketreplication-034)), 16.0 s x3 (-033), 33.3 s (-029),
  61.0 s x2 (-032), 62.4 s with 'none' until 61.4 s (-027), 66.0 s with 'pending' at 60.9 s (-031), 70.7 s x2 existing
  tables (-013); -031/-032/-033/-034 polled every 5 s, so their values are upper bounds within 5 s.
- Backfill of tables with an Iceberg schema: 443.2-484.5 s to 'completed', n=3 (-013), and ~336 s for the schema table
  created 2 s after the Put, n=1 (-027) - together 336-485 s, n=4; replicas were listable in the same-region destination
  at 70.7 s, long before 'completed', n=1 listing (-013).
- Tables created after the warm-up: 'completed' at 2.1-10.6 s, n=8 - 4.13 s at +485 s after the Put (-014), 3.2 s at +6 s
  after a re-Put (-027), 3.2 and 2.1 s around a role delete/re-create (-028), 5.2 s (-031), 5.2 s (-032), 5.4 and 10.6 s
  (-033).
- Replica deleted in the destination re-created by the service: 31.1-36.2 s with no source activity (n=2, 180 s quiet window,
  -034); 46.7 s in a run with concurrent source deletes (n=1, -033); 5 s polls.
- Same-name source re-create -> 'failed' 'identical name': 0.2-5.4 s after 'no-destinations' at +0.1 s, n=3 (0.2/5.4 s,
  -033; 5.2 s, -032).
- Dead destination, single runs (no range - do not derive waits): new table 'none' -> 'failed' at 23.9 s and old-table
  commit 'pending' 0.0 s -> 'failed' 2.1 s, n=1 each (-029); re-pointed configuration: failed table 'completed' at 63.3 s,
  n=1, and same-name destination re-create -> 409 at 28.6 s, n=1 (-030); table-level re-Put toward a deleted destination
  'completed' 0.05 s -> 'pending' 86.81 s -> 'failed' 87.84 s, n=1, with DeleteTableBucket 400 'not empty' x2 and 204 once
  emptied x2 ([S3T-TABLEBUCKET-118](#s3t-tablebucket-118)).
- Delete -> read: Get 404 on the first 200 ms poll at 0.01 s, n=1 (-004); after a token-less Delete the configuration kept
  returning 200 for 5 s of 200 ms polls, n=1 (-006).
- Role prerequisite: first Put accepted 4.9 s after CreateRole, 0 rejected attempts, n=1 role ([S3T-TABLEBUCKET-011](#s3t-tablebucket-011)); role
  absent ~20 s with tables still 'completed' in 2.1-3.2 s, n=2 tables in one run (-028).
- Non-replication windows (nothing appeared): 0 of 10 replicas over 300 s at 5 s polls after a Delete 2.1 s past the Put,
  n=1 (-016); new table and commit absent 120 s after a Delete, n=1 (-015); renamed replica unchanged over 153.0 s and
  122.3 s, n=1 each (-031); 'completed' tables unchanged over 150 s after an identical re-Put, n=2 (-027).
- Token stability: 20 Gets at 1/s returned 1 distinct 20-char token, n=20 (-001).

### Handling gaps

No finding in this document is `handling: suspect-bug`: the controller implements no replication resource or field today
(generator.yaml declares Table, Namespace and TableBucket and wires none of PutTableBucketReplication,
GetTableBucketReplication, DeleteTableBucketReplication or GetTableReplicationStatus), so all 44 findings are `unhandled`
and there is no controller behavior to compare against. A future TableBucketReplication implementation must handle:

- Token lifecycle: keep the versionToken from every Put/Get in status, send it on every update and delete, treat a
  token-less 400 'A version token is not specified.' and a 409 'does not match' as "re-read the token" and a 404 on Put as
  "configuration gone - clear the token and Put without one" ([S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), -002, -004, -005, -006,
  -019, -020, -026).
- No-op re-Puts: compare desired vs Get.configuration before writing - an identical Put rotates the token even though it
  is a no-op for replication (-002, -027).
- Delete semantics: Delete-when-absent is 404 (not 204) and must count as already removed; the config-absent 404 must be
  told apart from 'The specified bucket does not exist.' (-008, -023, [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), -096, -097, -098, -101).
- Validation the API does not do: duplicate, self, nonexistent and foreign-account destinations and nonexistent roles are
  accepted; the only rejections are ValidationException 400 (unmodeled) for ARN syntax and list bounds and
  BadRequestException 400 with a table-identifier message for a bare role name (-009, -017, -003, -011, -022,
  [S3T-TABLEBUCKET-011](#s3t-tablebucket-011), -021).
- Role prerequisite: never validated by the Put; a caller lacking iam:PassRole gets AccessDeniedException 403, a code to
  map rather than a terminal spec error (-012, -021, -028).
- Status: none at bucket level; readiness has to be derived per table from GetTableReplicationStatus, interpreting the
  ~60 s warm-up, the minutes-long schema backfill, 'failed' on a deleted destination or identical-name collision, and the
  404 after a configuration Delete (-007, -013, -014, -015, -027, -029, -031, -032, -033, -034).
- Destination and replica lifecycle: replicas outlive the configuration read-only and block DeleteTableBucket on the
  destination; a deleted destination name is 409-blocked and recovery is a re-point; rules cannot be disabled; renames are
  not propagated (-015, -016, -024, -025, -030, -010, -031, [S3T-TABLEBUCKET-118](#s3t-tablebucket-118), -012, -116).

### Where next

- `table-replication.md` is the table-level twin and holds the evidence cited here by reference: configuration versionToken
  vs table versionToken, per-destination status entries and the ~60 s evaluation after every Put, 'failed' destinations
  recovering only via a token-refreshing re-Put ([S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008) per -027's contradiction note), 'Insufficient
  permissions given to successfully complete replication.' ([S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), -002 per -021), orphan replicas and
  same-name collisions ([S3T-TABLE-138](table-replication-status.md#s3t-table-138), -196, [S3T-TABLEREPLICATION-020](table-replication-status.md#s3t-tablereplication-020), -025 per -032's notes), the table-level half of the
  destination-deletion cell ([S3T-TABLEBUCKET-118](#s3t-tablebucket-118)) and the same-region [D]->[E] switch ([S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017) per -022).
- `tablebucket.md` owns DeleteTableBucket's not-empty 400, the ~10 min post-delete name reservation behind -030's 409
  ([S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115) per -030's note), same-name re-create identity ([S3T-TABLEBUCKET-116](#s3t-tablebucket-116)), the deleted- and
  never-existing-bucket 404 message ([S3T-TABLEBUCKET-097](#s3t-tablebucket-097), -098, -101) and the foreign-account ForbiddenException
  ([S3T-TABLEBUCKET-021](#s3t-tablebucket-021)).
- `tablebucket-policy.md`, `tablebucket-encryption-storage.md` and `tablebucket-maintenance-metrics.md` cover the sibling
  sub-resource pairs sharing the fresh-bucket matrices ([S3T-TABLEBUCKET-092](#s3t-tablebucket-092), -096, -097, -098): Delete-when-absent
  (policy/metrics 204, replication/encryption 404) and Get-when-absent (encryption, maintenance, storage class 200 with
  defaults; policy, metrics, replication 404).
- `table.md` covers the source-side operations whose replication consequences appear here - RenameTable (-031), DeleteTable
  and same-name CreateTable (-032, -033), the managedByService read-only replica ([S3T-TABLE-225](table-replication-status.md#s3t-table-225) per -024's note) - and
  `table-metadata-schema.md` the UpdateTableMetadataLocation commit that -015 and -029 use as a replication trigger.
- Long catalogues belong in `details/<id>.md`: the destination-validation cells with the full ARN regex message
  ([S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009)), the five list-bounds cells with their ValidationException texts (-017), the per-table
  backfill timeline for five tables and two destinations (-013), the C1-C4 delete/re-create matrix (-032, -033, -034), the
  post-delete read matrix (-015), the destination lifecycle (-029, -030), the six-operation 404/403 matrices
  ([S3T-TABLEBUCKET-097](#s3t-tablebucket-097), -098, -101, -021) and the IAM-style PassRole message (-012).
- Open experiments (STATUS.md): Phase 8 reverify corrected [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) - the deleted replica IS
  re-created in 31-47 s, recorded as -033/-034 on n=3 with 5 s polls; the F-followups-2 settlements (replacement
  destination at +61-64 s after EVERY Put, dead destination 'failed' only after a re-Put storm, replicationInformation
  nested under managedTableInformation) are table-level and live in `table-replication.md`; the account holds 0 ackq IAM
  roles and 0 other table buckets in us-west-2/us-east-1, so no replication fixture survives, and the undeletable leftover
  bucket is a policy matter (`tablebucket-policy.md`). From the findings: -028 tested a ~20 s role absence only (STS
  session expiry untested); -013 is confidence medium (the cross-region listing also held another probe's replicas); -029,
  -030, -031, -016, -012 and [S3T-TABLEBUCKET-011](#s3t-tablebucket-011)/-012 stand on single runs; the stored duplicate [D, D] (-009) and the
  accepted cycle (-011, removed right after) were never watched for their replication effect; -010's dropped
  'status':'Disabled' was probed on the raw wire only.
<!-- preserved:end -->

## At a glance

- canonical findings: 44 (high 18 / medium 17 / low 9); duplicates folded into the appendix: 0
- handling: handled 0 · partial 0 · tracked 0 · unhandled 44 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 1 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: delete-semantics 11, error-code 7, other 6, request-validation 4, async-state-machine 3,
  idempotency 3, prerequisite 3, codegen-artifact 1, identity 1, normalization 1, quota-limit 1, read-gap 1,
  requested-vs-effective 1, shape-mismatch 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketEncryption | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketMetricsConfiguration | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketPolicy | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketReplication | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableReplication | delete | tableArn, versionToken | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableBucketMaintenanceConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketMetricsConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketPolicy | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketReplication | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketStorageClass | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableReplication | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableReplicationStatus | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| PutTableBucketMetricsConfiguration | create | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketPolicy | create | tableBucketARN, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketReplication | create | tableBucketARN, configuration | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableReplication | create | tableArn, configuration | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

Also referenced by findings but not in the s3tables model (other services or annotated variants): CreateRole

## State machine

- **ReplicationStatus**: pending, completed, failed (transitional: pending)
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

- <a id="s3t-tablebucketreplication-013"></a>**S3T-TABLEBUCKETREPLICATION-013** `async-state-machine` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-level backfill: 5 existing tables terminal in 71 s (empty) and 443-485 s (schema); replicas listable at ~+71 s...**
  PutTableBucketReplication(B -> D in another region) on a bucket with 5 existing tables (b1-b3 created with
  an Iceberg schema, b4-b5 without metadata) -> 200. Under the bucket-level config GetTableReplication(b1) ->
  NotFoundException 404 'The specified replication configuration does not exist.' (no table-level config
  exists) while GetTableReplicationStatus(b1) works (destinations: [] at +0, populated at ~+61 s). A second
  Put at +5.2 s adding a same-region destination E with the current token -> 200 and Get showed both
  destinations (no 'backfill in progress' conflict). Per-table status polled every 2 s: b4/b5 (no metadata)
  completed at +70.7 s for both destinations; b1/b2 completed for D at +443.2 s, b3 at +484.5 s; E's entries
  completed between +436 s and +484 s. The same-region destination E listed all five replica tables at +70.7 s
  (created with the namespace at the start of the backfill, long before 'completed').
  - ACK: e2e-timing, synced.when, custom_update · ops: PutTableBucketReplication, GetTableReplicationStatus,
    GetTableReplication, ListTables · fields: destinations[].replicationStatus,
    destinations[].destinationTableArn
  - repro: 5 tables -> PutTableBucketReplication -> +5 s Put with an added destination -> poll per-table
    status 2 s for <= 10 min; ListTables on both destinations
  - measurements: existing_table_terminal_s_schema=[443.2, 443.2, 484.5],
    existing_table_terminal_s_empty=[70.7, 70.7], replicas_listable_in_dest_s=[70.7], n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015),
    [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-006](table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030),
    [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-019](table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-031](table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-025](table-replication-status.md#s3t-tablereplication-025),
    [S3T-TABLEREPLICATION-030](table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](table-replication-status.md#s3t-tablereplication-043), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](table-replication-status.md#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014) · hypotheses: H-S-058, H-S-063 · evidence:
    tablebucket/sub-resources/repl-backfill-delete
  - notes: H-S-058 confirmed for both same-region (E) and cross-region (D) destinations: existing tables are
    backfilled; tables with an Iceberg metadata file take ~7-8 min, tables without any metadata ~70 s. H-S-063
    confirmed: a Put during the initial backfill is accepted (200, new token). Note the D...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-013.md](details/S3T-TABLEBUCKETREPLICATION-013.md)

- <a id="s3t-tablebucketreplication-014"></a>**S3T-TABLEBUCKETREPLICATION-014** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Table created 485 s after the bucket-level Put (warm-up over): status 'completed' in 4.13 s (n=1); replica present in both destinations**
  CreateTable (no metadata) in the source bucket 485.2s after PutTableBucketReplication:
  GetTableReplicationStatus transitions 0.05s -> no-destinations; 2.09s -> pending|pending; 4.13s ->
  completed|completed; terminal after 4.13s. Destination D then lists ['ackq_19dd08_b1', 'ackq_19dd08_b2',
  'ackq_19dd08_b3', 'ackq_19dd08_b4', 'ackq_19dd08_b5', 'ackq_19dd08_b6', 'ackq_a5e7ff_t3']; same-region
  destination E lists ['ackq_19dd08_b1', 'ackq_19dd08_b2', 'ackq_19dd08_b3', 'ackq_19dd08_b4',
  'ackq_19dd08_b5', 'ackq_19dd08_b6'].
  - ACK: e2e-timing, post-create-nudge · ops: CreateTable, GetTableReplicationStatus, ListTables · fields:
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable -> poll GetTableReplicationStatus(new table)
  - measurements: new_table_to_terminal_s=[4.13], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-022](table-replication-status.md#s3t-tablereplication-022),
    [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015) · hypotheses:
    H-S-062, H-S-058 · evidence: tablebucket/sub-resources/repl-backfill-delete
  - notes: H-S-062 second half confirmed: tables created after the Put are replicated.
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-014.md](details/S3T-TABLEBUCKETREPLICATION-014.md)

- <a id="s3t-tablebucketreplication-029"></a>**S3T-TABLEBUCKETREPLICATION-029** `async-state-machine` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Destination bucket deleted under an active config: Get still echoes it, old table stays 'completed'; new table and source commit -> 'failed'**
  Bucket-level config to one other-region destination; table t1 'completed' at 33.3 s. Destination emptied
  (replicas ['ackq_704fdb_t1'] deleted 204, namespace 204) and DeleteTableBucket -> HTTP 204. 2 s later
  GetTableBucketReplication -> HTTP 200 and still names the deleted destination: True; t1 status completed.
  New table t2 created: status transitions [(0.1, 'none'), (23.9, 'failed')] (final entry
  {"replicationStatus": "failed", "failureMessage": "The specified destination bucket does not exist."}). t1
  status afterwards 'completed'. Metadata-only UpdateTableMetadataLocation on t1 -> HTTP 200; its status over
  the next 90 s: [(0.0, 'pending'), (2.1, 'failed')].
  - ACK: references, requeue · ops: DeleteTableBucket, GetTableBucketReplication, GetTableReplicationStatus,
    CreateTable, UpdateTableMetadataLocation · fields:
    configuration.rules[].destinations[].destinationTableBucketARN, destinations[].replicationStatus,
    destinations[].failureMessage
  - repro: Bucket-level replication -> table completed -> empty+delete the destination bucket ->
    GetTableBucketReplication, GetTableReplicationStatus, CreateTable, commit on the old table
  - measurements: t1_completed_s=33.3, t2_terminal_s=23.9, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEREPLICATION-006](table-replication-status.md#s3t-tablereplication-006),
    [S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030), [S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115) · evidence:
    tablebucketreplication/creative/destination-deleted
  - notes: Extends [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009) (Put validates nothing) with the post-hoc destruction of the
    destination: the configuration read-back never reflects a dead destination and already-replicated tables
    keep reporting 'completed', so a TableBucketReplication controller cannot detect the broken...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-029.md](details/S3T-TABLEBUCKETREPLICATION-029.md)

## Identity and lookup

- <a id="s3t-tablebucket-116"></a>**S3T-TABLEBUCKET-116** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucket with 5 sub-resources set -> 204; same-name re-create gets a new tableBucketId, every sub-resource reads fresh (+0/+30 s)**
  With encryption AES256, policy, maintenance enabled 7/3, INTELLIGENT_TIERING and metrics set (pre-delete
  reads {'encryption': "{'sseAlgorithm': 'AES256'}", 'policy': 'policy(235 chars)', 'metrics':
  'id=EntireBucket', 'replication': 'NotFoundException/404', 'maintenance': 'enabled/7/3', 'storage_class':
  'INTELLIGENT_TIERING'}), DeleteTableBucket -> HTTP 204; GetTableBucket -> NotFoundException after 0.0 s.
  CreateTableBucket with the same name (after the 409 window): the new ARN equals the old one, tableBucketId
  changed: True. Sub-resource Gets right after re-create: {'encryption': "{'sseAlgorithm': 'AES256'}",
  'policy': 'NotFoundException/404', 'metrics': 'NotFoundException/404', 'replication':
  'NotFoundException/404', 'maintenance': 'enabled/3/10', 'storage_class': 'STANDARD'}; at +30 s:
  {'encryption': "{'sseAlgorithm': 'AES256'}", 'policy': 'NotFoundException/404', 'metrics':
  'NotFoundException/404', 'replication': 'NotFoundException/404', 'maintenance': 'enabled/3/10',
  'storage_class': 'STANDARD'} (identical to +0).
  - ACK: is_arn_primary_key, custom_find · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket,
    GetTableBucketEncryption, GetTableBucketPolicy, GetTableBucketMetricsConfiguration,
    GetTableBucketReplication, GetTableBucketMaintenanceConfiguration, GetTableBucketStorageClass · fields:
    tableBucketId, arn
  - repro: set 5 sub-resources -> DeleteTableBucket -> wait 404 -> CreateTableBucket(same name) until 201 ->
    GetTableBucket -> six Gets at +0 and +30 s
  - measurements: get_bucket_404_after_delete_s=0.0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-097](#s3t-tablebucket-097), [S3T-TABLEBUCKET-101](#s3t-tablebucket-101), [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), [S3T-TABLEBUCKET-060](tablebucket.md#s3t-tablebucket-060),
    [S3T-TABLEBUCKET-117](tablebucket-encryption-storage.md#s3t-tablebucket-117) · hypotheses: H-S-022, H-S-099 · evidence:
    tablebucket/identity/subresources-across-recreate
  - notes: H-S-022 confirmed (no sub-resource blocked deletion; replication not set; same-name re-create
    needed 45 attempts because of the 409 window, see the delete-semantics finding). H-S-099 confirmed: no old
    policy/metrics/maintenance/storage-class/encryption value survived on the reused ARN.
  - full notes: [details/S3T-TABLEBUCKET-116.md](details/S3T-TABLEBUCKET-116.md)

## Idempotency

- <a id="s3t-tablebucketreplication-001"></a>**S3T-TABLEBUCKETREPLICATION-001** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket replication token: 20 Gets identical; garbage-token Put -> ConflictException (token unchanged)...**
  After PutTableBucketReplication (status='success') GetTableBucketReplication x20 at 1/s returned one
  identical versionToken (len 20, equal to the Put's token). Put with versionToken='abc' -> ConflictException
  (HTTP 409) 'The specified version token does not match the configuration.'; the Get token afterwards was
  unchanged. Put of a DIFFERENT configuration WITHOUT versionToken while a configuration exists ->
  BadRequestException (HTTP 400) 'A version token is not specified.'; Get then showed destinations [D] (old
  config kept) and the token unchanged.
  - ACK: custom_update, requeue · ops: PutTableBucketReplication, GetTableBucketReplication · fields:
    versionToken
  - repro: Put -> Get x20 -> Put(versionToken='abc') -> Get -> Put(other config, no token) -> Get
  - measurements: get_x20_distinct_tokens=[1], token_len=[20], n=20
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKET-012](#s3t-tablebucket-012), [S3T-SERVICE-016](service.md#s3t-service-016) · hypotheses: H-S-119, H-S-050 · evidence:
    tablebucket/dependencies/repl-setup
  - notes: H-S-119 confirmed (20 Gets identical, failed Put does not move the token). H-S-050 confirmed in its
    400 variant: a token-less Put on an existing configuration is rejected with BadRequestException 400 'A
    version token is not specified.' (not ConflictException 409); the stored configuration and token...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-001.md](details/S3T-TABLEBUCKETREPLICATION-001.md)

- <a id="s3t-tablebucketreplication-002"></a>**S3T-TABLEBUCKETREPLICATION-002** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Re-sending an unchanged bucket replication config with the current token -> 200, token ROTATES...**
  PutTableBucketReplication with byte-identical configuration and the current versionToken -> HTTP 200; the
  returned token DIFFERS from the one sent; GetTableBucketReplication afterwards returned the new Put token.
  Sending the same configuration again with the PREVIOUS token -> ConflictException (HTTP 409) 'The specified
  version token does not match the configuration.'.
  - ACK: custom_update, compare.is_ignored+delta_pre_compare · ops: PutTableBucketReplication,
    GetTableBucketReplication · fields: versionToken, configuration
  - repro: Put C (T1) -> Get -> Put C again with T1 -> Get -> Put C again with T1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEBUCKET-012](#s3t-tablebucket-012), [S3T-SERVICE-016](service.md#s3t-service-016), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008),
    [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEREPLICATION-038](table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](table-replication.md#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-044](table-replication.md#s3t-tablereplication-044), [S3T-TABLEREPLICATION-014](table-replication.md#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003) ·
    hypotheses: H-S-122 · evidence: tablebucket/dependencies/repl-setup
  - notes: H-S-122 confirmed: every no-change Put rotates the token and the previous token becomes stale
    (ConflictException). Full original title: Re-sending an unchanged bucket replication config with the
    current token -> 200, token ROTATES; resend with the previous token -> ConflictException 409

- <a id="s3t-tablebucketreplication-027"></a>**S3T-TABLEBUCKETREPLICATION-027** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Identical PutTableBucketReplication re-send (token rotates) does not re-replicate 'completed' tables (150 s) or restart a backfill**
  Bucket-level config (one other-region destination) put at T0; table t1 (no metadata) created at T0+1 s:
  status 'none' until +61.4 s, 'pending' +61.4 s, 'completed' +62.4 s; table t2 (schema) created at T0+2 s:
  'pending' from the first read, 'completed' only at ~+336 s (lastSuccessfulReplicatedUpdate.timestamp).
  Identical re-Put (GetTableBucketReplication.configuration == sent document) with the current token -> HTTP
  200, token rotated, response status 'success'. GetTableReplicationStatus polled at 1/s per table for 150 s
  after the re-Put: t1 stayed 'completed' (single value for 150 s) and its replica's GetTable (modifiedAt,
  versionToken) was byte-identical before/after; t2's in-flight backfill continued unchanged ('pending' ->
  'completed' at +95 s of the watch, i.e. not restarted). A table created 6 s after the re-Put reached
  'completed' in 3.2 s (no new ~60 s warm-up).
  - ACK: compare.is_ignored+delta_pre_compare, one-per-reconcile · ops: PutTableBucketReplication,
    GetTableBucketReplication, GetTableReplicationStatus · fields: versionToken, configuration,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable x2 -> wait 'completed' -> GetTableBucketReplication token
    -> identical Put -> poll GetTableReplicationStatus 150 s + GetTable(replica)
  - measurements: t1_completed_s=62.4, t2_completed_s=336, post_reput_watch_s=150,
    new_table_after_reput_completed_s=3.2, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-031](table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-004](table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-006](table-replication-status.md#s3t-tablereplication-006), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029),
    [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008) · evidence:
    tablebucketreplication/creative/identical-reput-retrigger
  - notes: Extends [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002) (identical re-Put rotates the token) with its effect on
    replication: none - a controller that re-Puts on every reconcile only churns the versionToken
    (invalidating any token a user or a second controller holds) and does not cause backfills, so a perpetual
    spec...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-027.md](details/S3T-TABLEBUCKETREPLICATION-027.md)

## Errors

- <a id="s3t-tablebucket-021"></a>**S3T-TABLEBUCKET-021** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Six bucket ops on a well-formed ARN of another account -> ForbiddenException/403 (not 404); foreign destination in Put -> 200 accepted**
  With tableBucketARN=arn:aws:s3tables:<region>:<ACCOUNT>:bucket/ackq-nonexistent: GetTableBucket ->
  ForbiddenException (HTTP 403) 'Unauthorized'; GetTableBucketEncryption -> ForbiddenException (HTTP 403)
  'Unauthorized'; GetTableBucketStorageClass -> ForbiddenException (HTTP 403) 'Unauthorized';
  GetTableBucketReplication -> ForbiddenException (HTTP 403) 'Unauthorized'; DeleteTableBucketReplication ->
  ForbiddenException (HTTP 403) 'Unauthorized'; PutTableBucketReplication -> ForbiddenException (HTTP 403)
  'Unauthorized'. Using the same foreign ARN as a replication DESTINATION on our own bucket -> HTTP 200.
  - ACK: exceptions.404, terminal_codes · ops: GetTableBucket, GetTableBucketEncryption,
    GetTableBucketStorageClass, GetTableBucketReplication, PutTableBucketReplication,
    DeleteTableBucketReplication
  - repro: Call each op with a syntactically valid ARN whose account id is another account
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-022](service.md#s3t-service-022), [S3T-SERVICE-026](service.md#s3t-service-026), [S3T-TABLEBUCKET-098](#s3t-tablebucket-098), [S3T-TABLEBUCKET-011](#s3t-tablebucket-011),
    [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-SERVICE-021](service.md#s3t-service-021) · hypotheses: H-S-064 · evidence:
    tablebucket/weird-inputs/repl-validation
  - notes: H-S-064 confirmed: all six ops return ForbiddenException 403 'Unauthorized' for a foreign-account
    ARN (no 404 leak, no AccessDeniedException). Extra: a foreign-account bucket ARN is accepted as a
    replication DESTINATION on our own bucket (200), consistent with lazy destination validation.

- <a id="s3t-tablebucket-092"></a>**S3T-TABLEBUCKET-092** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket: Get metrics 404 'No such metric configuration exists'; Delete metrics/policy -> 204; Delete replication/encryption -> 404** (hypothesis refuted; behavior confirmed)
  On a table bucket with no sub-resource ever set: GetTableBucketMetricsConfiguration -> NotFoundException
  (HTTP 404) 'No such metric configuration exists'. DeleteTableBucketMetricsConfiguration -> HTTP 204;
  DeleteTableBucketPolicy -> HTTP 204; DeleteTableBucketReplication -> NotFoundException (HTTP 404) 'The
  specified replication configuration does not exist.'; DeleteTableBucketEncryption -> NotFoundException (HTTP 404)
  'The specified encryption configuration does not exist.'. Get metrics again -> NotFoundException (HTTP 404)
  'No such metric configuration exists'.
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy, DeleteTableBucketReplication,
    DeleteTableBucketEncryption · fields: tableBucketARN
  - repro: CreateTableBucket -> Get metrics -> Delete metrics / policy / replication / encryption -> Get
    metrics
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-174](tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-097](#s3t-tablebucket-097), [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](#s3t-tablebucket-096), [S3T-TABLEBUCKET-095](tablebucket-maintenance-metrics.md#s3t-tablebucket-095),
    [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKETREPLICATION-023](#s3t-tablebucketreplication-023), [S3T-TABLEBUCKET-098](#s3t-tablebucket-098), [S3T-SERVICE-014](service.md#s3t-service-014),
    [S3T-TABLEBUCKET-101](#s3t-tablebucket-101), [S3T-TABLEBUCKET-116](#s3t-tablebucket-116) · hypotheses: H-S-043, H-S-088, H-S-098 · evidence:
    tablebucket/sub-resources/metrics-lifecycle
  - notes: H-S-043 qualified: the Get returns NotFoundException 404 with a metrics-specific message
    (confirmed) but DeleteTableBucketMetricsConfiguration on an absent configuration returns HTTP 204, not 404
    (refuted). H-S-088 qualified: DeleteTableBucketPolicy and DeleteTableBucketMetricsConfiguration are...
  - full notes: [details/S3T-TABLEBUCKET-092.md](details/S3T-TABLEBUCKET-092.md)

- <a id="s3t-tablebucket-097"></a>**S3T-TABLEBUCKET-097** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket Gets: encryption/maintenance/storage-class 200 with defaults, policy/metrics/replication 404; missing bucket -> 404 bucket msg** (hypothesis refuted; behavior confirmed)
  Fresh bucket (nothing configured): encryption -> HTTP 200; policy -> NotFoundException (HTTP 404) 'The
  specified bucket policy does not exist.'; metrics -> NotFoundException (HTTP 404) 'No such metric
  configuration exists'; replication -> NotFoundException (HTTP 404) 'The specified replication configuration
  does not exist.'; maintenance -> HTTP 200; storage_class -> HTTP 200. Same six Gets on an ARN whose bucket
  never existed: encryption -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; policy ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; metrics -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'; replication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; maintenance -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  storage_class -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'. The config-absent 404
  messages ({'policy': 'The specified bucket policy does not exist.', 'metrics': 'No such metric configuration
  exists', 'replication': 'The specified replication configuration does not exist.'}) differ from the
  bucket-missing message(s) ['The specified bucket does not exist.'], so the message text is the in-band
  discriminator.
  - ACK: exceptions.404, custom_find · ops: GetTableBucketEncryption, GetTableBucketPolicy,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketMaintenanceConfiguration,
    GetTableBucketStorageClass · fields: tableBucketARN
  - repro: CreateTableBucket -> six Get* sub-resource reads -> the same six on
    arn:...:bucket/<never-existed-name>
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-174](tablebucket-maintenance-metrics.md#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](tablebucket-maintenance-metrics.md#s3t-tablebucket-176), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094), [S3T-TABLEBUCKET-173](tablebucket-maintenance-metrics.md#s3t-tablebucket-173), [S3T-TABLEBUCKET-053](tablebucket-maintenance-metrics.md#s3t-tablebucket-053),
    [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKETREPLICATION-023](#s3t-tablebucketreplication-023), [S3T-TABLEBUCKET-101](#s3t-tablebucket-101), [S3T-TABLEBUCKET-116](#s3t-tablebucket-116) ·
    hypotheses: H-S-002, H-S-098 · evidence: tablebucket/error-taxonomy/subresource-arn-variants
  - notes: H-S-002 refuted: GetTableBucketEncryption on a never-configured bucket returns HTTP 200
    {sseAlgorithm: AES256}, not 404, so there is no message to match for that noun. H-S-098 confirmed:
    GetTableBucketMetricsConfiguration is 404 for both cases and the messages differ ('No such metric
    configuration...
  - full notes: [details/S3T-TABLEBUCKET-097.md](details/S3T-TABLEBUCKET-097.md)

- <a id="s3t-tablebucket-098"></a>**S3T-TABLEBUCKET-098** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Sub-resource Deletes on a never-existing bucket -> ['NotFoundException/404']...**
  Never-existing bucket ARN: policy -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  metrics -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; replication ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; encryption -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'. Fresh bucket, nothing configured: policy -> HTTP 204; metrics ->
  HTTP 204; replication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; encryption -> NotFoundException (HTTP 404) 'The specified encryption configuration does not
  exist.'.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketPolicy, DeleteTableBucketMetricsConfiguration,
    DeleteTableBucketReplication, DeleteTableBucketEncryption · fields: tableBucketARN
  - repro: four Delete* sub-resource calls on arn:...:bucket/<never-existed-name>, then on a fresh bucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-021](#s3t-tablebucket-021), [S3T-SERVICE-022](service.md#s3t-service-022), [S3T-SERVICE-026](service.md#s3t-service-026), [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), [S3T-TABLEBUCKET-096](#s3t-tablebucket-096),
    [S3T-SERVICE-014](service.md#s3t-service-014) · hypotheses: H-S-088 · evidence: tablebucket/error-taxonomy/subresource-arn-variants
  - notes: H-S-088 qualified: policy and metrics Deletes are idempotent 204 on an absent configuration,
    replication and encryption Deletes are 404 with a noun-specific message; on a missing bucket all four are
    ['NotFoundException/404'] with ['The specified bucket does not exist.']. So 404 on Delete...
  - full notes: [details/S3T-TABLEBUCKET-098.md](details/S3T-TABLEBUCKET-098.md)

- <a id="s3t-tablebucketreplication-005"></a>**S3T-TABLEBUCKETREPLICATION-005** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketReplication: Get -> 404; Put with the pre-delete token -> 404 (not 409); token-less Put -> 200 with a new token**
  After deleting an existing configuration with its current token: GetTableBucketReplication ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (error body carries
  no versionToken). PutTableBucketReplication WITH the token that was current before the Delete ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (a 404, not a 409).
  PutTableBucketReplication WITHOUT a token -> HTTP 200 with a versionToken different from every earlier
  token.
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketReplication, DeleteTableBucketReplication,
    PutTableBucketReplication · fields: versionToken
  - repro: Put -> Delete(current token) -> Get -> Put(old token) -> Put(no token)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-023](#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034),
    [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037),
    [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-024](#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016),
    [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006),
    [S3T-TABLEREPLICATION-004](table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLE-138](table-replication-status.md#s3t-table-138),
    [S3T-TABLE-196](table-replication-status.md#s3t-table-196) · hypotheses: H-S-120 · evidence: tablebucket/dependencies/repl-setup
  - notes: H-S-120 confirmed, with one precision: sending the old token after a delete yields
    NotFoundException 404 (not 409/400), so a controller that sees 404 on Put must clear its token and re-Put
    without one. The never-configured half of H-S-047 is probed in tablebucket/weird-inputs/repl-validation
    (in...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-005.md](details/S3T-TABLEBUCKETREPLICATION-005.md)

- <a id="s3t-tablebucketreplication-008"></a>**S3T-TABLEBUCKETREPLICATION-008** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Never-configured bucket: GetTableBucketReplication -> 404; DeleteTableBucketReplication -> 404 with or without a token (never 400)**
  On a fresh table bucket that never had a replication configuration: GetTableBucketReplication ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  DeleteTableBucketReplication without versionToken -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'; DeleteTableBucketReplication with versionToken='abc' -> NotFoundException
  (HTTP 404) 'The specified replication configuration does not exist.' (repeated: NotFoundException (HTTP 404)
  'The specified replication configuration does not exist.').
  - ACK: exceptions.404, custom_delete · ops: GetTableBucketReplication, DeleteTableBucketReplication ·
    fields: versionToken
  - repro: CreateTableBucket -> GetTableBucketReplication; DeleteTableBucketReplication (no token / token
    'abc') x2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-023](#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034),
    [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037),
    [S3T-TABLEBUCKETREPLICATION-026](#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004),
    [S3T-TABLEREPLICATION-036](table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003) · hypotheses:
    H-S-047 · evidence: tablebucket/weird-inputs/repl-validation
  - notes: H-S-047 confirmed: Get -> NotFoundException 404 and Delete-when-none -> NotFoundException 404 (not
    an idempotent 204), both with and without a token. Note the asymmetry with a CONFIGURED bucket, where a
    token-less Delete is rejected with BadRequestException 400 'A version token is not specified.'...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-008.md](details/S3T-TABLEBUCKETREPLICATION-008.md)

- <a id="s3t-tablebucketreplication-012"></a>**S3T-TABLEBUCKETREPLICATION-012** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Caller with s3tables:* but Deny iam:PassRole: PutTableBucketReplication and PutTableReplication -> AccessDeniedException 403 (iam:PassRole)**
  Assuming a role whose identity policy allows s3tables:* and explicitly denies iam:PassRole:
  PutTableBucketReplication(role=that same role) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> with an explicit deny in an identity-based policy'.
  PutTableReplication(same configuration) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> with an explicit deny in an identity-based policy'. Control reads by
  the same caller: GetTableBucket -> HTTP 200, GetTableBucketReplication -> NotFoundException (HTTP 404) 'The
  specified replication configuration does not exist.'.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketReplication, PutTableReplication · fields:
    configuration.role
  - repro: AssumeRole(role with Deny iam:PassRole) -> PutTableBucketReplication / PutTableReplication
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009), [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](table-replication.md#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLE-026](table-policy.md#s3t-table-026) · hypotheses:
    H-S-052, H-U-133 · evidence: service/error-taxonomy/repl-permissions
  - notes: H-S-052 confirmed; H-U-133 PassRole half confirmed: both Put operations return
    AccessDeniedException 403 whose message is the IAM-style 'User: <caller> is not authorized to perform:
    iam:PassRole on resource: <role> with an explicit deny in an identity-based policy'.
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-012.md](details/S3T-TABLEBUCKETREPLICATION-012.md)

## Request validation

- <a id="s3t-tablebucketreplication-009"></a>**S3T-TABLEBUCKETREPLICATION-009** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication: duplicate/self/nonexistent destinations -> 200; malformed ARN or >5 destinations -> ValidationException 400** (hypothesis refuted; behavior confirmed)
  On an empty source bucket V (role valid, no prior config): destinations [D, D] -> HTTP 200 (Get shows 2
  destination(s)). [V] (self) -> HTTP 200. [V, D] -> HTTP 200 (Get shows 2 destination(s)). A well-formed ARN
  of a bucket that does not exist (same account/region) -> HTTP 200 and Get echoes it. A different-region
  nonexistent bucket -> HTTP 200. Truncated 'arn:aws:s3tables:us-east-1:123:bucket' -> ValidationException
  (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations.1.member.destinationTableBucketARN' failed to satisfy constraint:
  Member must satisfy regular expression pattern:
  (arn:aws[-a-z0-9]*:[a-z0-9]+:[-a-z0-9]*:[0-9]{12}:bucket/[a-z0-9_-]{3,63})'. Six destinations ->
  ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length less than
  or equal to 5'.
  - ACK: terminal_codes, references · ops: PutTableBucketReplication, GetTableBucketReplication · fields:
    configuration.rules[].destinations[].destinationTableBucketARN
  - repro: Fresh bucket; one Put per case (clearing the config in between); Get after each
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-032](table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKETREPLICATION-022](#s3t-tablebucketreplication-022), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005),
    [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029), [S3T-TABLEREPLICATION-011](table-replication.md#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-013](table-replication.md#s3t-tablereplication-013) · hypotheses: H-S-121, H-S-123, H-S-055 · evidence:
    tablebucket/weird-inputs/repl-validation
  - notes: H-S-121 REFUTED: [D, D] accepted and stored twice (no 400, no dedup). H-S-123 REFUTED: the source
    bucket itself as the only destination is accepted (200) and stored; [self, D] stored with 2 entries.
    H-S-055 qualified: nonexistent destination accepted (as hypothesised) but self/duplicate are NOT...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-009.md](details/S3T-TABLEBUCKETREPLICATION-009.md)

- <a id="s3t-tablebucketreplication-017"></a>**S3T-TABLEBUCKETREPLICATION-017** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication rules/destinations bounds -> ValidationException 400 (code not in the model); 5 dests + non-existent role -> 200**
  PutTableBucketReplication on a fresh table bucket, role = a syntactically valid but NON-EXISTENT IAM role
  ARN, client validation off: 6_destinations -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length
  less than or equal to 5'; 2_rules -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules' failed to satisfy constraint: Member must have length less than or equal to 1';
  0_rules -> ValidationException (HTTP 400) '1 validation error detected: Value at 'configuration.rules'
  failed to satisfy constraint: Member must have length greater than or equal to 1'; rule_with_0_destinations
  -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length greater
  than or equal to 1'; 5_destinations -> HTTP 200; 1_destination -> BadRequestException (HTTP 400) 'A version
  token is not specified.'. The list-size checks fire first and use the wire code 'ValidationException' (HTTP
  400), which is NOT among the modeled errors (BadRequestException is). The 5-destination configuration with a
  role ARN that does not exist was ACCEPTED (HTTP 200). A subsequent DeleteTableBucketReplication and
  PutTableBucketReplication without versionToken were rejected with BadRequestException 400 'A version token
  is not specified.'; the bucket with the lingering replication configuration was still deletable with
  DeleteTableBucket.
  - ACK: terminal_codes, none · ops: PutTableBucketReplication · fields: configuration.rules,
    configuration.rules[].destinations, configuration.role
  - repro: PutTableBucketReplication(configuration={role: <nonexistent>, rules: <6 dest | 2 rules | [] |
    [{destinations: []}] | 5 dest>})
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEBUCKETREPLICATION-021](#s3t-tablebucketreplication-021), [S3T-TABLEBUCKET-011](#s3t-tablebucket-011),
    [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-032](table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002),
    [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006),
    [S3T-TABLEREPLICATION-004](table-replication.md#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011), [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010),
    [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-012](#s3t-tablebucketreplication-012),
    [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009) · hypotheses: H-S-056 · evidence:
    tablebucket/limits/bucket-quota-create-burst
  - notes: H-S-056 qualified: the limits exist (>5 destinations, >1 rule, [] rules, [] destinations all
    rejected) but as ValidationException, not BadRequestException; 5 destinations accepted. The IAM role was
    not validated at Put time (non-existent role ARN accepted). No IAM role was created (shard:...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-017.md](details/S3T-TABLEBUCKETREPLICATION-017.md)

- <a id="s3t-tablebucketreplication-019"></a>**S3T-TABLEBUCKETREPLICATION-019** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C062 PARTLY: Second Put (other destinations) replaces config + rotates token ONLY with the current versionToken; token-less -> 400**
  With the current token: Put [S2, D] then Put [D, S2] -> HTTP 200 each, Get shows exactly the new destination
  list ([S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003)) and every Put - even an unchanged re-send - returns a token different
  from the one sent ([S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002)). WITHOUT versionToken a Put of a different configuration
  while one exists -> BadRequestException (HTTP 400) 'A version token is not specified.'; the old destinations
  and token stay ([S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001); same 400 in -017). The model marks versionToken optional.
  - ACK: custom_update, requeue · ops: PutTableBucketReplication, GetTableBucketReplication · fields:
    versionToken, configuration.rules[].destinations
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017); no new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005) · evidence: service/static/doc-claims-2
  - notes: VERDICT: PARTLY - 'creates or updates' holds, but an update requires the current versionToken
    although the model marks it optional; a plain second Put is rejected.

- <a id="s3t-tablebucketreplication-026"></a>**S3T-TABLEBUCKETREPLICATION-026** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C069 PARTLY: DeleteTableBucketReplication on an existing config: token-less -> 400, stale -> 409; never-configured bucket -> 404**
  DeleteTableBucketReplication with no versionToken on an existing configuration -> BadRequestException (HTTP 400)
  'A version token is not specified.'; the configuration stays (Get kept returning 200 for 5 s); a second
  token-less Delete -> the same 400 ([S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006)). Delete(stale token from 2 Puts ago) ->
  ConflictException (HTTP 409) 'The specified version token does not match the configuration.'; Delete('abc')
  -> the same 409; Delete(current token) -> HTTP 204 ([S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004)). On a never-configured
  bucket every Delete -> 404 ([S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008)). The API model marks versionToken optional.
  - ACK: custom_delete, requeue · ops: DeleteTableBucketReplication · fields: versionToken
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008); no new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008),
    [S3T-TABLEREPLICATION-036](table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015), [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177),
    [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKETREPLICATION-023](#s3t-tablebucketreplication-023), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034) · evidence:
    service/static/doc-claims-2
  - notes: VERDICT: PARTLY - the stale-token half is true (409); the token-less half is false - the token is
    mandatory on an existing configuration despite being optional in the model.

## Update granularity and ordering

- <a id="s3t-tablebucket-011"></a>**S3T-TABLEBUCKET-011** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication accepts a role ARN that does not exist (200, status 'success'); a 5 s-old role was accepted on the first try** (hypothesis refuted; behavior confirmed)
  PutTableBucketReplication(configuration.role = ARN of a role name that does not exist in the account, one
  existing destination bucket in another region) -> HTTP 200 {versionToken, status: 'success'};
  GetTableBucketReplication then returned that configuration (role is not validated at Put time). With the
  shard role created by iam:CreateRole (trust replication.s3tables.amazonaws.com) the first Put issued 4.9 s
  after CreateRole returned 200 on attempt 1 (no IAM-propagation rejection observed, n=1). The source bucket
  was EMPTY (no namespace, no table) and the Put still returned 200.
  - ACK: requeue, references · ops: PutTableBucketReplication, CreateRole · fields: configuration.role
  - repro: iam:CreateRole (trust replication.s3tables.amazonaws.com) -> immediately PutTableBucketReplication
    on an empty bucket, retry every 2 s on 4xx
  - measurements: first_put_ok_since_role_create_s=[4.9], rejected_attempts_before_ok=[0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-021](#s3t-tablebucketreplication-021), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-032](table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKET-021](#s3t-tablebucket-021), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-SERVICE-021](service.md#s3t-service-021) · hypotheses: H-S-054, H-S-062 ·
    evidence: tablebucket/dependencies/repl-setup
  - notes: H-S-054 REFUTED: a nonexistent role is NOT rejected with 400 (accepted with 200); consequently no
    IAM-propagation 4xx can be seen on this path (n=1 role, first Put at +4.9 s). H-S-062 confirmed: Put on an
    empty source bucket succeeds. The role-existence check, if any, is asynchronous (see the...
  - full notes: [details/S3T-TABLEBUCKET-011.md](details/S3T-TABLEBUCKET-011.md)

- <a id="s3t-tablebucketreplication-021"></a>**S3T-TABLEBUCKETREPLICATION-021** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C064 FALSE: PutTableBucketReplication does not check the role (nonexistent role -> 200 'success'); failures surface per table**
  PutTableBucketReplication(configuration.role = ARN of a role that does not exist, one existing destination
  in another region) -> HTTP 200 {versionToken, status: 'success'} and Get returns that configuration
  ([S3T-TABLEBUCKET-011](#s3t-tablebucket-011); 5 destinations + nonexistent role -> 200, [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017)). Only ARN
  syntax and list lengths are validated at Put time (ValidationException 400)
  ([S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009)). A role allowed on destination B1 only: per-table GetTableReplicationStatus
  shows B2 'failed' with failureMessage 'Insufficient permissions given to successfully complete replication.'
  after ~66 s ([S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028); same text for a nonexistent role, [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002)).
  - ACK: requeue, references · ops: PutTableBucketReplication, GetTableReplicationStatus · fields:
    configuration.role
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002); no new API call for
    this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-032](table-replication-status.md#s3t-tablereplication-032), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008) ·
    evidence: service/static/doc-claims-2
  - notes: VERDICT: FALSE - no synchronous permission check: the Put is accepted and insufficient role
    permissions appear only as an asynchronous per-destination 'failed' status.

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-tablebucketreplication-003"></a>**S3T-TABLEBUCKETREPLICATION-003** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketReplication echoes destinations in the order sent and the role ARN with path byte-identical...**
  Put with destinations [S2 (us-west-2 bucket), D (us-east-1 bucket)] then Get -> order preserved; Put [D, S2]
  then Get -> order preserved (no sorting). configuration.role, a role ARN with a path segment
  (role/ackq/<name>), came back byte-identical. Put with configuration.role='S3TablesReplicationRole' (bare
  role name, 23 chars) -> BadRequestException (HTTP 400) with the MISLEADING message 'Invalid input: Must
  provide either table arn alone OR all of (table bucket arn, namespace, table name) together'.
  - ACK: compare.is_ignored+delta_pre_compare, references · ops: PutTableBucketReplication,
    GetTableBucketReplication · fields: configuration.rules[].destinations, configuration.role
  - repro: Put [B, A] -> Get; Put [A, B] -> Get; Put role='S3TablesReplicationRole'
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-022](#s3t-tablebucketreplication-022), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017), [S3T-SERVICE-026](service.md#s3t-service-026), [S3T-SERVICE-007](service.md#s3t-service-007), [S3T-SERVICE-015](service.md#s3t-service-015), [S3T-SERVICE-021](service.md#s3t-service-021),
    [S3T-TABLEBUCKET-011](#s3t-tablebucket-011), [S3T-TABLEBUCKET-021](#s3t-tablebucket-021), [S3T-TABLEREPLICATION-011](table-replication.md#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-013](table-replication.md#s3t-tablereplication-013) · hypotheses: H-S-116, H-S-066 · evidence: tablebucket/dependencies/repl-setup
  - notes: H-S-116 confirmed (order preserved, path kept, no rewriting). H-S-066 confirmed (bare name -> 400),
    with the surprise that the 400 message talks about table identifiers, not about the role.
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-003.md](details/S3T-TABLEBUCKETREPLICATION-003.md)

- <a id="s3t-tablebucketreplication-007"></a>**S3T-TABLEBUCKETREPLICATION-007** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketReplication returns status='success'; GetTableBucketReplication has only configuration+versionToken...**
  PutTableBucketReplication response members: versionToken and status='success' (a literal operation result).
  GetTableBucketReplication at +0 s and at the end of the probe (several minutes later) returned exactly the
  members ['configuration', 'versionToken'] - no status, health or progress member exists in the bucket-level
  read path.
  - ACK: synced.when, docs-only · ops: PutTableBucketReplication, GetTableBucketReplication · fields: status,
    versionToken, configuration
  - repro: Put -> inspect response; Get at +0 and +5 min -> list top-level members
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003),
    [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009) · hypotheses: H-S-048 · evidence:
    tablebucket/dependencies/repl-setup
  - notes: H-S-048 qualified: the Put response status is the literal lowercase 'success' (the user guide shows
    'Success'); it is an operation result, not a replication state, and the Get has no status member. Progress
    is only visible per table via GetTableReplicationStatus (probed separately).
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-007.md](details/S3T-TABLEBUCKETREPLICATION-007.md)

- <a id="s3t-tablebucketreplication-010"></a>**S3T-TABLEBUCKETREPLICATION-010** `shape-mismatch` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Raw PUT with rule 'status':'Disabled' -> HTTP 200 (status dropped by Get); unknown rule member -> HTTP 200...**
  A SigV4-signed raw PUT /table-bucket-replication whose single rule carries the user-guide member
  "status":"Disabled" -> HTTP 200 {"status":"success","versionToken":"2f12bf303ac159ca7044"}. GET afterwards
  -> HTTP 200, body has no status member. A rule with an unknown member fooUnknown -> HTTP 200
  {"status":"success","versionToken":"486a6438dc13486bff6d"}. A top-level "status":"Disabled" next to
  configuration -> HTTP 200 {"status":"success","versionToken":"073df15cc1fc7bc65435"}. The SDK model has no
  rule status member at all.
  - ACK: docs-only, scope:field-on-parent · ops: PutTableBucketReplication, GetTableBucketReplication ·
    fields: configuration.rules[].status
  - repro: Signed raw HTTP PUT with an extra rule member, then GET the raw body
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011),
    [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029), [S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009) · hypotheses: H-S-057 · evidence:
    tablebucket/weird-inputs/repl-validation
  - notes: H-S-057 confirmed (silently dropped): rules cannot be disabled through the API; only deleted. Full
    original title: Raw PUT with rule 'status':'Disabled' -> HTTP 200 (status dropped by Get); unknown rule
    member -> HTTP 200; the read shape has no rule status

- <a id="s3t-tablebucketreplication-031"></a>**S3T-TABLEBUCKETREPLICATION-031** `requested-vs-effective` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Bucket-level replication ignores RenameTable (name, namespace) on a source for 150/120 s: replica keeps the old name; new table in 5 s**
  PutTableBucketReplication(S -> D) -> HTTP 200; then CreateNamespace ns + CreateTable t in S: replica listed
  in D after 66.0 (status completed). RenameTable(t -> t2) -> HTTP 204; D polled every 5 s for 153s:
  transitions [[0.1, "completed", {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t"]}]]; final D {"ackq_6c3e6e_ns":
  ["ackq_6c3e6e_t"]}; GetTable(replica ARN) {"name": "ackq_6c3e6e_t", "namespace": ["ackq_6c3e6e_ns"]};
  GetTableReplicationStatus(source ARN) -> 'completed'. Control CreateTable t3 -> listed in D after 5.2s
  (status transitions [[0.1, "no-destinations"], [5.2, "completed"]]). RenameTable(t2 -> ns2/t2) -> HTTP 204;
  after 122s D = {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t", "ackq_6c3e6e_t3"]} (transitions [[0.1, "completed",
  {"ackq_6c3e6e_ns": ["ackq_6c3e6e_t", "ackq_6c3e6e_t3"]}]]).
  - ACK: custom_update, pre-delete-cleanup, requeue · ops: PutTableBucketReplication, RenameTable,
    CreateTable, ListTables, GetTableReplicationStatus · fields: newName, newNamespaceName,
    destinations[].replicationStatus
  - repro: PutTableBucketReplication -> CreateTable -> replica appears -> RenameTable(newName) -> poll D 150 s
    -> CreateTable (control) -> RenameTable(newNamespaceName) -> poll D 120 s
  - measurements: initial_replica_listed_s=66.0, b1_window_s=153.0, control_new_table_listed_s=5.2,
    b3_window_s=122.3, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLEREPLICATION-020](table-replication-status.md#s3t-tablereplication-020),
    [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) · evidence:
    tablebucketreplication/creative/rename-under-bucket-replication
  - notes: Cross-service analogy (DynamoDB replicas are addressed by the shared table name; bucket-level
    replication replicates new tables within seconds per [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014)): the controller's
    only Update (RenameTable) against a bucket-wide replication rule. Compare the table-level case in...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-031.md](details/S3T-TABLEBUCKETREPLICATION-031.md)

## Delete semantics

- <a id="s3t-tablebucket-101"></a>**S3T-TABLEBUCKET-101** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucket -> HTTP 204; GetTableBucket polled at 200 ms: NotFoundException/404 from the first read (0.009 s)...**
  After deleting the table and namespace, DeleteTableBucket -> HTTP 204. GetTableBucket polled every 200 ms:
  codes in order ['NotFoundException/404', 'NotFoundException/404', 'NotFoundException/404',
  'NotFoundException/404', 'NotFoundException/404'] (first poll at +0.009 s, message 'The specified bucket
  does not exist.'); no ConflictException and no DELETING-style intermediate. Then on the deleted ARN:
  encryption -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; policy ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; metrics -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'; replication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; maintenance -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  storage_class -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; PutTableBucketPolicy
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; PutTableBucketMetricsConfiguration
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: exceptions.404, deletable.when · ops: DeleteTableBucket, GetTableBucket, PutTableBucketPolicy,
    PutTableBucketMetricsConfiguration, GetTableBucketEncryption, GetTableBucketPolicy,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketMaintenanceConfiguration,
    GetTableBucketStorageClass · fields: tableBucketARN
  - repro: DeleteTable -> DeleteNamespace -> DeleteTableBucket -> GetTableBucket every 200 ms -> six
    sub-resource Gets + Put policy + Put metrics on the ARN
  - measurements: first_404_after_delete_s=0.009, polls=5, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-161](tablebucket.md#s3t-tablebucket-161), [S3T-TABLEBUCKET-162](tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-014](tablebucket.md#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](tablebucket.md#s3t-tablebucket-015),
    [S3T-TABLEBUCKET-097](#s3t-tablebucket-097), [S3T-TABLEBUCKET-116](#s3t-tablebucket-116), [S3T-TABLEBUCKET-092](#s3t-tablebucket-092) · hypotheses: H-S-071, H-S-098, H-S-002 ·
    evidence: tablebucket/error-taxonomy/subresource-arn-variants
  - notes: H-S-071 (no DELETING state, 404 on first read, sub-resource Put on the deleted ARN -> 404 not 409)
    confirmed. H-S-098 cell (b) (deleted bucket) and H-S-002: the deleted-bucket messages are ['The specified
    bucket does not exist.'] - identical to the never-existing-bucket message(s).
  - full notes: [details/S3T-TABLEBUCKET-101.md](details/S3T-TABLEBUCKET-101.md)

- <a id="s3t-tablebucket-118"></a>**S3T-TABLEBUCKET-118** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **DeleteTableBucket(destination) with a replica inside -> BadRequestException; once emptied -> 204...**
  With t replicated ('completed' at +62.22s) into the destination bucket D2: DeleteTableBucket(D2) ->
  BadRequestException (HTTP 400) 'The bucket that you tried to delete is not empty.'. Emptying D2 (the
  service-created replica and its namespace): 2 deletes, codes ['204']. DeleteTableBucket(D2) -> HTTP 204;
  GetTableBucket(D2) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'. Source side over
  the next 60 s (no new Put): GetTableReplicationStatus [(0.0, 'completed')]; GetTableReplication -> HTTP 200
  (configuration still stored). PutTableReplication(identical, current token) with the destination bucket gone
  -> HTTP 200; status transitions [(0.05, 'completed'), (86.81, 'pending'), (87.84, 'failed:The specified
  destination bucket does not exist.')]. DeleteTableReplication -> HTTP 204.
  - ACK: pre-delete-cleanup, synced.when, requeue · ops: DeleteTableBucket, DeleteTable, DeleteNamespace,
    GetTableReplicationStatus, GetTableReplication, PutTableReplication, DeleteTableReplication · fields:
    destinations[].replicationStatus, destinations[].failureMessage
  - repro: replica completed -> DeleteTableBucket(D) -> empty D -> DeleteTableBucket(D) -> poll source status
    60 s -> re-Put (token) -> poll -> DeleteTableReplication
  - measurements: put_to_terminal_s=[62.22], re_put_to_failed_s=[87.84], delete_nonempty_400_n=[2],
    delete_empty_204_n=[2], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](table-replication-status.md#s3t-table-139), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-196](table-replication-status.md#s3t-table-196), [S3T-TABLEBUCKET-012](#s3t-tablebucket-012),
    [S3T-TABLE-306](table-replication-status.md#s3t-table-306), [S3T-TABLEREPLICATION-040](table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023) · hypotheses: H-U-169, H-U-170 · evidence:
    tablebucket/dependencies/repl-dest-teardown
  - notes: Destination-bucket deletion cell (n=2 for the two DeleteTableBucket codes: run 1 on the shard's
    destination bucket gave the same 400 'not empty' then 204): the destination owner can delete replicas
    (['204']) and then the bucket; the source keeps its configuration and its last status until a new Put....
  - full notes: [details/S3T-TABLEBUCKET-118.md](details/S3T-TABLEBUCKET-118.md)

- <a id="s3t-tablebucketreplication-004"></a>**S3T-TABLEBUCKETREPLICATION-004** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketReplication: stale token -> ConflictException, garbage token -> ConflictException, current token -> HTTP 204...**
  With a configuration whose token had rotated twice: Delete(versionToken=<token from 2 Puts ago>) ->
  ConflictException (HTTP 409) 'The specified version token does not match the configuration.' (config still
  present); Delete(versionToken='abc') -> ConflictException (HTTP 409) 'The specified version token does not
  match the configuration.' (config still present); Delete(current token) -> HTTP 204;
  GetTableBucketReplication polled at 200 ms returned NotFoundException on the first poll (0.01s).
  - ACK: custom_delete, exceptions.404 · ops: DeleteTableBucketReplication, GetTableBucketReplication ·
    fields: versionToken
  - repro: Put (T1) -> Put (T2) -> Delete(T1) -> Get -> Delete('abc') -> Get -> Delete(T2) -> Get
  - measurements: get_404_after_delete_s=[0.01], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-020](#s3t-tablebucketreplication-020), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001),
    [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEBUCKETREPLICATION-026](#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](#s3t-tablebucketreplication-006),
    [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-036](table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015), [S3T-TABLEBUCKET-012](#s3t-tablebucket-012), [S3T-SERVICE-016](service.md#s3t-service-016) · hypotheses: H-S-129 · evidence:
    tablebucket/dependencies/repl-setup
  - notes: H-S-129 qualified: a stale token -> ConflictException 409 and the configuration remains
    (confirmed); a syntactically garbage token 'abc' -> ConflictException 409 as well, NOT BadRequestException
    400 (refuted for that half). Only the current token deletes (a token-less Delete is rejected with 400,...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-004.md](details/S3T-TABLEBUCKETREPLICATION-004.md)

- <a id="s3t-tablebucketreplication-006"></a>**S3T-TABLEBUCKETREPLICATION-006** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Configured bucket: DeleteTableBucketReplication WITHOUT versionToken -> BadRequestException 400 'A version token is not specified.'** (hypothesis refuted; behavior confirmed)
  DeleteTableBucketReplication with no versionToken on an existing configuration -> BadRequestException (HTTP 400)
  'A version token is not specified.'; the configuration stays (Get kept returning 200 for 5 s of 200 ms
  polls). A second token-less Delete -> the same 400. The same 400 is returned for a token-less Delete on a
  bucket whose configuration references a nonexistent role. The API model marks versionToken optional on this
  operation.
  - ACK: custom_delete, exceptions.404 · ops: DeleteTableBucketReplication, GetTableBucketReplication ·
    fields: versionToken
  - repro: Put -> Delete (no versionToken) -> Get @200 ms -> Delete again
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-026](#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008),
    [S3T-TABLEREPLICATION-036](table-replication.md#s3t-tablereplication-036), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEREPLICATION-015](table-replication.md#s3t-tablereplication-015),
    [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-004](table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005),
    [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027) · hypotheses: H-S-051 · evidence:
    tablebucket/dependencies/repl-setup
  - notes: H-S-051 REFUTED: the token is NOT optional on Delete; the removal path must Get the token first
    (and treat 404 on that Get as already removed). A fresh bucket that never had a configuration is probed in
    tablebucket/weird-inputs/repl-validation.

- <a id="s3t-tablebucketreplication-015"></a>**S3T-TABLEBUCKETREPLICATION-015** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketReplication: the 6 replicas remain (read-only, 400 on update), per-table status -> 404...**
  DeleteTableBucketReplication(current token) -> 204; GetTableBucketReplication -> NotFoundException 404;
  GetTableReplicationStatus(b1) -> NotFoundException 404 'The specified replication configuration does not
  exist.' (status is gone with the configuration); the source bucket GetTableBucket -> 200 unchanged. Both
  destinations still list all 6 replicas (b1-b6). The replica of b1 in D keeps
  managedByService='replication.s3tables.amazonaws.com' and managedTableInformation.replicationInformation;
  UpdateTableMetadataLocation on it -> BadRequestException 400 'Updates to service managed tables are not
  allowed.'. A new commit on source b1 and a new table b7 created after the Delete: after 120 s the replica's
  metadataLocation was unchanged and b7 did not appear in either destination; GetTableReplicationStatus(b7)
  and (b1) -> 404.
  - ACK: pre-delete-cleanup, custom_delete, docs-only · ops: DeleteTableBucketReplication,
    GetTableReplicationStatus, ListTables, UpdateTableMetadataLocation, GetTable · fields: managedByService,
    managedTableInformation.replicationInformation
  - repro: backfilled bucket config -> DeleteTableBucketReplication -> reads on source/destination -> commit +
    CreateTable on source -> wait 120 s -> reads
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-024](#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](#s3t-tablebucketreplication-025), [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035),
    [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](table-replication.md#s3t-tablereplication-010),
    [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003),
    [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-196](table-replication-status.md#s3t-table-196), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014) · hypotheses: H-S-061, H-S-087 · evidence:
    tablebucket/sub-resources/repl-backfill-delete
  - notes: H-S-061 confirmed: replicas remain, stay read-only and stop following the source; the source bucket
    is unaffected. H-S-087 REFUTED: the orphaned replicas are NOT released to writable tables
    (UpdateTableMetadataLocation -> 400); they can only be read or deleted.
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-015.md](details/S3T-TABLEBUCKETREPLICATION-015.md)

- <a id="s3t-tablebucketreplication-016"></a>**S3T-TABLEBUCKETREPLICATION-016** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketReplication 2.1s after a Put on a bucket with 10 tables -> HTTP 204; 0 replicas appeared in the destination within 5 min**
  PutTableBucketReplication(B2 with 10 empty tables -> same-region E) then
  DeleteTableBucketReplication(current token) at +2.1s -> HTTP 204. ListTables(E) polled every 5 s for 300 s:
  new tables over time [[2.1, 0], [32.3, 0], [62.6, 0], [92.8, 0], [123.0, 0], [153.3, 0], [183.5, 0], [213.7,
  0], [244.0, 0], [274.3, 0], [304.5, 0]]; final count 0 of 10. GetTableReplicationStatus on c1 / c10
  afterwards: ERR:NotFoundException / ERR:NotFoundException.
  - ACK: pre-delete-cleanup, e2e-timing · ops: PutTableBucketReplication, DeleteTableBucketReplication,
    ListTables, GetTableReplicationStatus
  - repro: 10 tables -> Put -> 2 s -> Delete -> ListTables(destination) every 5 s for 5 min
  - measurements: replicas_after_early_delete=[0], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-024](#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](#s3t-tablebucketreplication-025), [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035),
    [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007), [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-196](table-replication-status.md#s3t-table-196) · hypotheses:
    H-S-127 · evidence: tablebucket/sub-resources/repl-backfill-delete
  - notes: H-S-127 qualified: nothing was replicated after the early Delete.

- <a id="s3t-tablebucketreplication-030"></a>**S3T-TABLEBUCKETREPLICATION-030** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Deleted destination re-created under the same name 28.6 s later -> ConflictException 409; config re-pointed to a new bucket -> 200**
  CreateTableBucket(same destination name, other region) 28.6 s after its deletion -> ConflictException (HTTP 409)
  'The bucket is in a transitional state because of a previous deletion attempt. Try again later.'. Instead:
  new destination bucket + role policy, PutTableBucketReplication(new destination, current token) -> HTTP 200;
  the table that had failed/pended against the dead destination: [(0.0, 'failed'), (63.3, 'completed')]; the
  old 'completed' table afterwards 'pending'.
  - ACK: requeue, references · ops: CreateTableBucket, PutTableBucketReplication, GetTableReplicationStatus ·
    fields: configuration.rules[].destinations[].destinationTableBucketARN
  - repro: DeleteTableBucket(destination) -> CreateTableBucket(same name) -> PutTableBucketReplication(new
    destination) -> poll statuses
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115), [S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028),
    [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005) · evidence:
    tablebucketreplication/creative/destination-deleted
  - notes: Cross-links the 10-min 'transitional state' name block ([S3T-TABLEBUCKET-115](tablebucket.md#s3t-tablebucket-115)) with replication
    recovery: a controller that re-creates a deleted destination TableBucket under the same name is blocked by
    409 for ~10 min, while re-pointing the config to a new destination recovers at once. Single run.
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-030.md](details/S3T-TABLEBUCKETREPLICATION-030.md)

- <a id="s3t-tablebucketreplication-032"></a>**S3T-TABLEBUCKETREPLICATION-032** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Bucket rule: source delete leaves the replica; same-name re-create -> failed 'identical name' +5 s (replica re-creation: -033/-034)**
  Two empty source tables ta, tb replicated to D under PutTableBucketReplication (D listed both after 61.0s).
  C1 DeleteTable(replica of ta in D) -> HTTP 204; GetTableReplicationStatus(ta) right after 'completed'; 120 s
  poll (D tables, status): [[0.0, "completed", ["ackq_97f768_tb@ff92f100"]]]. C2 DeleteTable(source tb) ->
  HTTP 204; status by tb's ARN right after: 'ERR:NotFoundException'; 120 s poll: [[0.0,
  "ERR:NotFoundException", ["ackq_97f768_tb@ff92f100"]]]. C3 CreateTable(same namespace/name as tb) -> HTTP
  200 (new ARN: True); 150 s poll of the new table's status + D: [[0.1, "no-destinations",
  ["ackq_97f768_tb@ff92f100"]], [5.2, "failed:A table with an identical name already exists at the
  destination.", ["ackq_97f768_tb@ff92f100"]]]; replica entry before/after {"before":
  ["ackq_97f768_tb@ff92f100"], "after": ["ackq_97f768_tb@ff92f100"]}; the replica's
  replicationInformation.sourceTableARN now points to the old source. C4 control CreateTable tc -> listed in D
  after 5.2s, status 'completed'.
  - ACK: pre-delete-cleanup, requeue, custom_delete · ops: PutTableBucketReplication, DeleteTable,
    CreateTable, GetTableReplicationStatus, ListTables, GetTable · fields: destinations[].replicationStatus,
    destinations[].failureMessage, managedTableInformation.replicationInformation.sourceTableARN
  - repro: PutTableBucketReplication -> 2 tables replicated -> DeleteTable(replica) / DeleteTable(source) /
    CreateTable(same name) / CreateTable(new) with 5 s polling of D and status
  - measurements: c1_window_s=121.7, c2_window_s=121.4, c3_window_s=152.1, c4_listed_s=5.2, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-196](table-replication-status.md#s3t-table-196), [S3T-TABLEREPLICATION-020](table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014),
    [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031),
    [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-007](table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLE-224](table-replication-status.md#s3t-table-224) · evidence:
    tablebucketreplication/creative/replica-delete-and-source-recreate
  - notes: Cross-service analogy (DynamoDB: deleting a replica region is explicit and the table name is
    shared; a delete+recreate of a CR with the same name is the classic ACK adoption loop). Extends
    [S3T-TABLE-138](table-replication-status.md#s3t-table-138) (table-level: deleted replica not re-created; later commit -> failed 'destination table has
    been...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-032.md](details/S3T-TABLEBUCKETREPLICATION-032.md)

- <a id="s3t-tablebucketreplication-033"></a>**S3T-TABLEBUCKETREPLICATION-033** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Reverify [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032): C2-C4 reproduce (identical-name 'failed' at +0.2/5.4 s) but the deleted replica IS re-created...**
  Three empty source tables under a bucket-level rule were listed in D after 16.0 s (all 'completed' after
  16.0 s). C1 DeleteTable(replica of ta in D) -> HTTP 204; GetTableReplicationStatus(ta) right after ->
  'completed'; after a 123.7 s poll the replica was re-created: True. C2 DeleteTable(source tb, tb2) -> HTTP
  204/204; status by the old ARNs -> ERR:NotFoundException/ERR:NotFoundException; replicas still in D at the
  end of the window: True/True. C3 CreateTable(same namespace/name as tb, tb2) -> HTTP 200/200 (new ARN
  differs: True/True); status at +0 s 'no-destinations'; first 'failed' at +0.2/5.4 s with message ['A table
  with an identical name already exists at the destination.']; the replica's ARN unchanged: True, its
  replicationInformation.sourceTableARN still the OLD source: True (points at the new source: False). C4
  control tables tc, td -> listed in D at +5.4/10.6 s, 'completed' at +5.4/10.6 s.
  - ACK: pre-delete-cleanup, requeue, e2e-timing · ops: PutTableBucketReplication, DeleteTable, CreateTable,
    GetTableReplicationStatus, ListTables, GetTable · fields: destinations.replicationStatus,
    destinations.failureMessage, managedTableInformation.replicationInformation.sourceTableARN
  - repro: PutTableBucketReplication(S->D) -> CreateTable ta,tb,tb2 (empty) -> wait listed in D ->
    DeleteTable(replica ta in D) + DeleteTable(source tb,tb2) -> poll D + statuses 120 s -> CreateTable(tb,tb2
    same names) + CreateTable(tc,td) -> poll 150 s -> GetTable(replica tb)
  - measurements: initial_listed_all_s=16.0, initial_completed_all_s=16.0, c1_c2_window_s=123.7,
    c3_c4_window_s=151.5, same_name_failed_at_s=[0.2, 5.4], control_listed_in_d_s=[5.4, 10.6],
    control_completed_s=[5.4, 10.6], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) · evidence: tablebucketreplication/creative/reverify-set-d
  - notes: Reverify set D for [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) with n=2 for the same-name failure and the
    control timing (original n=1); poll interval 5 s so +X s values are upper bounds within 5 s. CORRECTS
    clause C1 ('deleted replica not re-created within 120 s'): the replica of ta reappeared in
    ListTables(D)...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-033.md](details/S3T-TABLEBUCKETREPLICATION-033.md)

- <a id="s3t-tablebucketreplication-034"></a>**S3T-TABLEBUCKETREPLICATION-034** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Bucket rule: a replica deleted in D IS re-created with no source activity, at +31.1/36.2 s (n=2, 5 s polls)**
  Two empty sources ta1, ta2 replicated to D (listed+completed after 10.7 s), then 20 s quiet.
  DeleteTable(replica ta1/ta2 in D) -> HTTP 204/204; GetTableReplicationStatus(source) right after ->
  'completed'/'completed'. Replica reappearance in ListTables(D): quiet 180 s: reappeared {'ta1': 31.1, 'ta2':
  36.2}. Status strings seen during the quiet window: ['completed', 'pending'].
  - ACK: pre-delete-cleanup, requeue · ops: DeleteTable, ListTables, GetTableReplicationStatus,
    PutTableBucketReplication, CreateTable · fields: destinations.replicationStatus
  - repro: PutTableBucketReplication(S->D) -> CreateTable ta1,ta2 -> wait completed -> DeleteTable(replicas in
    D) -> poll D 180 s quiet -> (if absent) CreateTable tx -> poll 120 s -> DeleteTable tx -> poll 120 s ->
    re-Put config -> poll 120 s
  - measurements: quiet_reappeared_at_s.ta1=31.1, quiet_reappeared_at_s.ta2=36.2,
    reappeared_by_phase.ta1.phase=p1, reappeared_by_phase.ta1.at_s=31.1, reappeared_by_phase.ta2.phase=p1,
    reappeared_by_phase.ta2.at_s=36.2, n=2, poll_interval_s=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) · evidence: tablebucketreplication/creative/reverify-set-d-c1
  - notes: Reverify set D follow-up isolating clause C1 of [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) (in the main
    reverify probe the replica reappeared at +46.7 s while two other sources were deleted in the same second).
    REFUTES clause C1 of [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032): with n=3 across the two probes the deleted...
  - full notes: [details/S3T-TABLEBUCKETREPLICATION-034.md](details/S3T-TABLEBUCKETREPLICATION-034.md)

## Codegen notes

- <a id="s3t-tablebucket-096"></a>**S3T-TABLEBUCKET-096** `codegen-artifact` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Model marks the 4 bucket sub-resource Deletes idempotent; on an absent config metrics/policy -> 204 but replication/encryption -> 404**
  botocore model 'idempotent': true on ['DeleteTableBucketMetricsConfiguration', 'DeleteTableBucketPolicy',
  'DeleteTableBucketEncryption', 'DeleteTableBucketReplication']. Live on a fresh bucket with nothing
  configured: DeleteTableBucketMetricsConfiguration -> HTTP 204, DeleteTableBucketPolicy -> HTTP 204,
  DeleteTableBucketReplication -> NotFoundException/404, DeleteTableBucketEncryption -> NotFoundException/404;
  a second DeleteTableBucketMetricsConfiguration right after a successful one -> HTTP 204.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy,
    DeleteTableBucketEncryption, DeleteTableBucketReplication
  - repro: inspect the service model; Delete each sub-resource on a fresh bucket; Delete metrics twice
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKET-093](tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-095](tablebucket-maintenance-metrics.md#s3t-tablebucket-095), [S3T-TABLEBUCKET-092](#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-098](#s3t-tablebucket-098), [S3T-SERVICE-014](service.md#s3t-service-014) · hypotheses: H-S-102 · evidence:
    tablebucket/sub-resources/metrics-lifecycle
  - notes: H-S-102 confirmed (trait misleading): the trait does not predict the per-noun outcome; read the
    Delete-when-absent code from this probe, not from the model.
  - full notes: [details/S3T-TABLEBUCKET-096.md](details/S3T-TABLEBUCKET-096.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLEBUCKET-011](#s3t-tablebucket-011) | PutTableBucketReplication accepts a role ARN that does not exist (200, status 'success'); a 5 s-old role was accepted on the first try | first_put_ok_since_role_create_s=[4.9], rejected_attempts_before_ok=[0] | 1 |
| [S3T-TABLEBUCKET-101](#s3t-tablebucket-101) | DeleteTableBucket -> HTTP 204; GetTableBucket polled at 200 ms: NotFoundException/404 from the first read (0.009 s)... | first_404_after_delete_s=0.009, polls=5 | 1 |
| [S3T-TABLEBUCKET-118](#s3t-tablebucket-118) | DeleteTableBucket(destination) with a replica inside -> BadRequestException; once emptied -> 204... | put_to_terminal_s=[62.22], re_put_to_failed_s=[87.84], delete_nonempty_400_n=[2], delete_empty_204_n=[2] | 2 |
| [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001) | Bucket replication token: 20 Gets identical; garbage-token Put -> ConflictException (token unchanged)... | get_x20_distinct_tokens=[1], token_len=[20] | 20 |
| [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004) | DeleteTableBucketReplication: stale token -> ConflictException, garbage token -> ConflictException, current token -> HTTP 204... | get_404_after_delete_s=[0.01] | 1 |
| [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013) | Bucket-level backfill: 5 existing tables terminal in 71 s (empty) and 443-485 s (schema); replicas listable at ~+71 s... | existing_table_terminal_s_schema=[443.2, 443.2, 484.5], existing_table_terminal_s_empty=[70.7, 70.7], replicas_listable_in_dest_s=[70.7] | 5 |
| [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014) | Table created 485 s after the bucket-level Put (warm-up over): status 'completed' in 4.13 s (n=1); replica present in both destinations | new_table_to_terminal_s=[4.13] | 1 |
| [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027) | Identical PutTableBucketReplication re-send (token rotates) does not re-replicate 'completed' tables (150 s) or restart a backfill | t1_completed_s=62.4, t2_completed_s=336, post_reput_watch_s=150, new_table_after_reput_completed_s=3.2 | 2 |
| [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028) | Replication role deleted out-of-band: GetTableBucketReplication still echoes it; tables created 5-15 s later still replicate (n=1) | role_absent_window_s=20, t3_completed_s=3.2, t4_completed_s=2.1 | 1 |
| [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029) | Destination bucket deleted under an active config: Get still echoes it, old table stays 'completed'; new table and source commit -> 'failed' | t1_completed_s=33.3, t2_terminal_s=23.9 | 1 |
| [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031) | Bucket-level replication ignores RenameTable (name, namespace) on a source for 150/120 s: replica keeps the old name; new table in 5 s | initial_replica_listed_s=66.0, b1_window_s=153.0, control_new_table_listed_s=5.2, b3_window_s=122.3 | 1 |
| [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032) | Bucket rule: source delete leaves the replica; same-name re-create -> failed 'identical name' +5 s (replica re-creation: -033/-034) | c1_window_s=121.7, c2_window_s=121.4, c3_window_s=152.1, c4_listed_s=5.2 | 1 |
| [S3T-TABLEBUCKETREPLICATION-033](#s3t-tablebucketreplication-033) | Reverify [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032): C2-C4 reproduce (identical-name 'failed' at +0.2/5.4 s) but the deleted replica IS re-created... | initial_listed_all_s=16.0, initial_completed_all_s=16.0, c1_c2_window_s=123.7, c3_c4_window_s=151.5, same_name_failed_at_s=[0.2, 5.4], control_listed_in_d_s=[5.4, 10.6], control_completed_s=[5.4, 10.6] | 2 |
| [S3T-TABLEBUCKETREPLICATION-034](#s3t-tablebucketreplication-034) | Bucket rule: a replica deleted in D IS re-created with no source activity, at +31.1/36.2 s (n=2, 5 s polls) | quiet_reappeared_at_s.ta1=31.1, quiet_reappeared_at_s.ta2=36.2, reappeared_by_phase.ta1.phase=p1, reappeared_by_phase.ta1.at_s=31.1, reappeared_by_phase.ta2.phase=p1, reappeared_by_phase.ta2.at_s=36.2, poll_interval_s=5 | 2 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `tablebucket-replication.md` - unverified or stale items worth a future probe, one line each.

- [S3T-TABLEBUCKETREPLICATION-028](#s3t-tablebucketreplication-028): the role was absent ~20 s only; whether replication fails after STS session expiry (~1 h) or IAM propagation, and what the per-table status then says, is untested (n=1).
- [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013): confidence medium - the cross-region destination listing also contained another probe's replicas, so only the same-region 'first listed' time (70.7 s) is clean; re-run with an isolated destination.
- [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031): RenameTable not followed by the replica is n=1 on empty tables with a same-region destination; untested with a schema table, a commit after the rename, or a cross-region destination.
- [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029) / -030: destination deleted under an active configuration and the re-point recovery are single runs; the old 'completed' table going 'pending' after the re-point was not watched to a terminal state.
- [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009) / -011: a stored duplicate destination [D, D], a self-destination [V] and the back-edge cycle D -> V were accepted (200) but never watched for their replication effect (the cycle configuration was removed right after).
- [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016): Delete 2.1 s after a Put left 0 of 10 replicas in 300 s (n=1); the window between Put and Delete in which replicas do appear is unknown.
- [S3T-TABLEBUCKETREPLICATION-034](#s3t-tablebucketreplication-034): replica re-creation at +31.1/36.2 s was observed with 180 s quiet on empty tables (n=2, plus 46.7 s in -033); untested for a schema replica or when the source has pending commits.
- [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010): 'status':'Disabled' and unknown rule members were dropped on the raw wire (n=1 each); whether the SDK's typed rule shape ever gains a status member is a model-bump watch item.
- [S3T-TABLEBUCKETREPLICATION-012](#s3t-tablebucketreplication-012): the iam:PassRole 403 was probed with an explicit Deny only; an implicit (missing) PassRole permission was not probed.
- [S3T-TABLEBUCKET-011](#s3t-tablebucket-011): the 5 s-old role accepted on the first try is n=1 and the Put never validates the role, so no IAM-propagation window exists on this path; the per-table 'failed' timing for a role that becomes valid later is table-level evidence only ([S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002)/-008 per -021).
- [S3T-TABLEBUCKET-012](#s3t-tablebucket-012): DeleteTableBucket succeeded (204) on an empty bucket carrying a replication configuration once (n=1, configuration referencing a nonexistent role); not re-tried with a valid role or with tables ever replicated.
- [S3T-TABLEBUCKET-021](#s3t-tablebucket-021): a foreign-account destination ARN was accepted (200) but no cross-account replication was attempted, so the per-table status for that destination is unknown.
- [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027) vs [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008): an identical bucket-level re-Put was shown not to touch 'completed' or in-flight tables; whether it re-triggers a 'failed' table (as the table-level re-Put does) was not probed at bucket level.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-tablebucket-012"></a>**S3T-TABLEBUCKET-012** | delete-semantics | low | confirmed | DeleteTableBucket succeeds (204) on an empty bucket that still carries a replication configuration | [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-139](table-replication-status.md#s3t-table-139), [S3T-TABLE-196](table-replication-status.md#s3t-table-196), [S3T-TABLEBUCKET-118](#s3t-tablebucket-118), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-SERVICE-016](service.md#s3t-service-016) | - |
| <a id="s3t-tablebucketreplication-011"></a>**S3T-TABLEBUCKETREPLICATION-011** | quota-limit | low | confirmed | Fan-in: three source buckets -> one destination all 200; destination bucket as source pointing back (cycle) -> 200 | [S3T-TABLEBUCKETREPLICATION-022](#s3t-tablebucketreplication-022), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-010](#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-029](#s3t-tablebucketreplication-029) | - |
| <a id="s3t-tablebucketreplication-018"></a>**S3T-TABLEBUCKETREPLICATION-018** | other | low | confirmed | Doc claim C061 TRUE: PutTableBucketReplication -> 200 {versionToken, status 'success'}; Get echoes role, destinations and the same token | [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-007](#s3t-tablebucketreplication-007), [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEREPLICATION-003](table-replication.md#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001) | - |
| <a id="s3t-tablebucketreplication-020"></a>**S3T-TABLEBUCKETREPLICATION-020** | other | low | confirmed | Doc claim C063 TRUE: PutTableBucketReplication with a stale versionToken -> ConflictException 409 'token does not match'; config unchanged | [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002), [S3T-TABLEBUCKETREPLICATION-001](#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-019](#s3t-tablebucketreplication-019), [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005) | - |
| <a id="s3t-tablebucketreplication-022"></a>**S3T-TABLEBUCKETREPLICATION-022** | other | low | confirmed | Doc claim C065 TRUE: Same-region, same-account destination bucket is accepted by PutTableBucketReplication (200), incl. self and cycles | [S3T-TABLEBUCKETREPLICATION-003](#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-009](#s3t-tablebucketreplication-009), [S3T-TABLEBUCKETREPLICATION-011](#s3t-tablebucketreplication-011), [S3T-TABLEREPLICATION-017](table-replication-status.md#s3t-tablereplication-017) | - |
| <a id="s3t-tablebucketreplication-023"></a>**S3T-TABLEBUCKETREPLICATION-023** | other | low | confirmed | Doc claim C066 TRUE: GetTableBucketReplication on a bucket without a replication configuration -> NotFoundException 404 'does not exist' | [S3T-TABLEBUCKETREPLICATION-008](#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEBUCKET-092](#s3t-tablebucket-092), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKET-097](#s3t-tablebucket-097), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034), [S3T-TABLEREPLICATION-001](table-replication.md#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](table-replication.md#s3t-tablereplication-037), [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKETREPLICATION-026](#s3t-tablebucketreplication-026) | - |
| <a id="s3t-tablebucketreplication-024"></a>**S3T-TABLEBUCKETREPLICATION-024** | other | low | confirmed | Doc claim C067 TRUE: After DeleteTableBucketReplication Get -> 404 while the replicas stay in the destination (listable, read-only) | [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005), [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-004](#s3t-tablebucketreplication-004), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-025](#s3t-tablebucketreplication-025), [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018) | - |
| <a id="s3t-tablebucketreplication-025"></a>**S3T-TABLEBUCKETREPLICATION-025** | other | low | confirmed | Doc claim C068 TRUE: After DeleteTableBucketReplication new/in-flight source tables are not replicated (0 replicas in 120 s/300 s seen) | [S3T-TABLEBUCKETREPLICATION-015](#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-024](#s3t-tablebucketreplication-024), [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEREPLICATION-018](table-replication.md#s3t-tablereplication-018), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-005](#s3t-tablebucketreplication-005) | - |
| <a id="s3t-tablebucketreplication-028"></a>**S3T-TABLEBUCKETREPLICATION-028** | prerequisite | low | confirmed | Replication role deleted out-of-band: GetTableBucketReplication still echoes it; tables created 5-15 s later still replicate (n=1) | [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEBUCKETREPLICATION-012](#s3t-tablebucketreplication-012), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008), [S3T-TABLEBUCKETREPLICATION-017](#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009), [S3T-TABLEBUCKETREPLICATION-014](#s3t-tablebucketreplication-014), [S3T-TABLEBUCKETREPLICATION-027](#s3t-tablebucketreplication-027), [S3T-TABLEBUCKETREPLICATION-031](#s3t-tablebucketreplication-031), [S3T-TABLEBUCKETREPLICATION-032](#s3t-tablebucketreplication-032), [S3T-TABLEBUCKETREPLICATION-013](#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-030](#s3t-tablebucketreplication-030), [S3T-TABLEBUCKETREPLICATION-002](#s3t-tablebucketreplication-002) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
