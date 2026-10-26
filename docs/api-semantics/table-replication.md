<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table replication configuration (PutTableReplication, GetTableReplication, DeleteTableReplication)
_The configuration API of table-level replication: Put/Get/Delete triple, configuration versionToken vs table versionToken (exact 409/400 messages), role prerequisites and Put-time validation, re-Put semantics, 404 when absent; scope verdict for TableReplication._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-replication.md`; every claim cites the finding ids that carry the full evidence below.
The asynchronous side (status entries, replica tables, recovery, orphans) is the sibling `table-replication-status.md`.

### Orientation

This document covers the configuration API of table-level replication - PutTableReplication, GetTableReplication and
DeleteTableReplication keyed by a source table ARN, one rule with one role and 1-5 destination bucket ARNs
([S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), -011, -012, -037) - plus the tableArn identity and 404 shapes the two replication Gets share
with the other tableArn-keyed reads ([S3T-TABLE-173](#s3t-table-173), -174, -175, -176). Scope verdict: the controller implements Table,
Namespace and TableBucket only, `generator.yaml` carries no replication resource, operation or field, and the ground truth
records the family as not modeled (scope:defer, [GT-S3T-044](service.md#gt-s3t-044) (controller hooks catalog entry) via [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003)), so nothing here compares against
existing code - each finding is input for a would-be TableReplication CRD: a standalone resource keyed by the UUID-based
tableArn ([S3T-TABLE-173](#s3t-table-173)) whose create, update and delete are custom calls around the configuration versionToken
([S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), -010, -044), with MinItems=1 on rules and destinations (-011) and iam:PassRole on the caller
(-024). The write path is synchronous and token-locked: Put returns {versionToken (20 chars), status 'success'} - a literal
operation result, not one of pending/completed/failed; Get returns exactly [configuration, versionToken], echoes role and
destination byte-identically and is read-your-writes; every accepted Put rotates the token, a byte-identical re-Put
included, status transitions do not; the configuration token differs from GetTable.versionToken, which gets
ConflictException 409 on Put and Delete; the first Put needs no token, an update needs the current one
([S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), -004, -007, -010, -014, -015, -038, -039, -044). The 200 says nothing about the role, the
destination or the replica - the status is evaluated ~60 s later (-038) and lives in `table-replication-status.md`.
Delete(current token) -> 204 and both Gets -> 404 at once, but the service-managed replica stays listed in the destination
bucket and needs its own DeleteTable ([S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), -018, -046).

### Rules

1. Never forward GetTable.versionToken, a stale or a garbage token: Put and Delete answer ConflictException 409 on every
   retry (Put 5/5 at 2 s apart, Delete 5/5) with two messages - Put 'The specified version token does not match the
   configuration.', Delete 'Provided version token does not match the table version token.' (same text for the table token
   and a random 20-char token); Delete('') and a token-less Put on an existing configuration -> BadRequestException 400 'A
   version token is not specified.'; the configuration is kept with its token unchanged; read GetTableReplication before
   every write ([S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), -007, -015, -036, -044, -046).
2. Every accepted Put rotates the token, a byte-identical re-Put included: a re-Put 148 ms after the first -> 200 with a new
   token; 10 identical re-Puts chaining the returned token -> 10 x 200 in 1.02 s, token rotated on each, status 'completed'
   throughout, no duplicate replica; the token sampled after Put / at failed / at completed / after re-Put gave 3 distinct
   values over 4 samples - only Puts rotate it. A reconciler that re-sends its document each loop invalidates its own cached
   token every reconcile and must diff before it Puts ([S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), -039, -044).
3. The token is a real optimistic lock: two concurrent Puts ([D] and [E]) with the same current token -> exactly one 200 +
   one 409 in 5/5 rounds, the stored destinations are the winner's and Get's token matches the 200 response's; the loser
   must re-read and decide again, not retry ([S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), -016).
4. First Put and update differ: on a never-configured table token omitted or '' -> 200 while 'abc' -> NotFoundException 404
   'The specified replication configuration does not exist.' (existence check before token check); on an existing
   configuration token omitted or '' -> BadRequestException 400; Get, Status and Delete('abc') on a never-configured table
   -> 404 with the configuration message, to be mapped to 'no replication', not 'no table' ([S3T-TABLEREPLICATION-001](#s3t-tablereplication-001), -010,
   -015, -034, -044).
5. DeleteTableReplication(current token) -> 204, empty body; GetTableReplication -> 404 at +0.019 s and
   GetTableReplicationStatus at +0.044 s with 0 stale 200 bodies (without a bucket-level configuration; under one the status
   keeps a stale view ~60 s - [S3T-TABLEREPLICATION-022](table-replication-status.md#s3t-tablereplication-022) in `table-replication-status.md`); a repeat Delete -> 404;
   GetTable(source) -> 200 unchanged; the replica stays listed in the destination bucket (type 'customer', same
   namespace/name) and DeleteTable on it there -> 204 - the delete path must remove a table it did not create
   ([S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), -018, -046).
6. Put's 'success' is not a status: it is the literal 'success' on the first Put, on an identical re-Put and when adding a
   destination - none is in the GetTableReplicationStatus enum ['pending', 'completed', 'failed'] - so it cannot seed a
   condition; a Put 5.5 s after CreateRole + PutRolePolicy was accepted on attempt 1 and says nothing about the role working
   ([S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), -009, -014, -037, -038).
7. Request-time rejections are terminal (client validation off): rules=[], destinations=[], no rules, two rules, six
   destinations and an IAM user ARN as role -> ValidationException 400 with the constraint text; no role or a bare role name
   -> BadRequestException 400 'Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace,
   table name) together'; an other-account role ARN -> ForbiddenException 403 'Cross-account pass role is not allowed.';
   nothing is stored. Accepted with 200 and stored as sent: [D, D] (2 destinations), the source's own bucket, five
   destinations, an aws-cn role ARN; 'remove all destinations' cannot be expressed ([S3T-TABLEREPLICATION-011](#s3t-tablereplication-011), -012, -013).
8. Caller permissions: Put needs iam:PassRole on configuration.role (without it AccessDeniedException 403 naming iam:PassRole,
   Get -> 200) but no caller permission on the destination bucket; an explicit Deny on the three replication actions ->
   AccessDeniedException on each while GetTableReplicationStatus and GetTable -> 200; a tableArn carrying another account id
   -> ForbiddenException 403 'Unauthorized' on all four ops; a destinationTableBucketARN of another account -> 200
   ([S3T-TABLEREPLICATION-023](#s3t-tablereplication-023), -024).
9. tableArn identity: a bucket ARN, a name-based pseudo ARN or 'ns/table' -> BadRequestException 400 'The specified table
   ARN is not valid.' (3 forms x 5 ops, n=15) - the UUID ARN must be read from CreateTable/GetTable; a never-existing table
   UUID -> NotFoundException 404 'The specified table does not exist.', a never-existing bucket name -> 404 'The specified
   bucket does not exist.' on the two replication Gets, a foreign region segment -> BadRequestException 400 'The provided
   Region does not match the Region of the resource.' ([S3T-TABLE-173](#s3t-table-173), -174).
10. The replication Gets are 404 before and after DeleteTable and only the message changes ('...replication configuration
    does not exist.' -> '...table does not exist.') on the first call 0.04-0.05 s after the delete, so the code alone cannot
    tell 'config absent' from 'table gone'; after RenameTable the pre-rename ARN still resolves (GetTable(old ARN) -> 200
    with the new name, same tableARN and versionToken; GetTableReplication(old ARN) -> the same 404 as before) while the
    old triple -> 404 ([S3T-TABLE-175](#s3t-table-175), -176).
11. GetTable(source) is unchanged by PutTableReplication and only modifiedAt moved after DeleteTableReplication (~0.2 s
    before the Delete call; trigger ambiguous at n=1); managedByService and managedTableInformation stay absent on the
    source - a Table reconciler cannot detect replication from GetTable ([S3T-TABLE-177](#s3t-table-177)). Sequential sub-resource writes on a
    plain table (10 ordered pairs x 5, 0 s gap) gave 0 x 500 in 50 trials; the one 500 'An internal error occurred. Try
    again.' on a TagResource was on a table with an active table-level configuration and did not reproduce ([S3T-TABLE-304](#s3t-table-304)).

### Timing ranges

- PutTableReplication latency 92-109 ms per call; 10 identical re-Puts in 1.02 s wall ([S3T-TABLEREPLICATION-039](#s3t-tablereplication-039), n=10); an
  identical re-Put 148 ms after the first accepted ([S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), n=1); Put accepted 5.5 s after CreateRole +
  PutRolePolicy, 0 rejected attempts ([S3T-TABLEREPLICATION-009](#s3t-tablereplication-009), n=1 role).
- GetTableReplication read-your-writes: 10 reads within ~1 s of the Put, 1 distinct token ([S3T-TABLEREPLICATION-014](#s3t-tablereplication-014), n=10);
  stale-token retries: Put x5 (2 s apart) and Delete x5 -> 10/10 ConflictException ([S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), n=10);
  concurrent Put pairs: one winner in 5/5 rounds ([S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), n=5).
- Status seen from the configuration side: no-destinations 0.03 s, 'pending' 60.62 s, 'completed' 65.68 s after an identical
  re-Put on a fresh configuration ([S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), n=1); on a 'completed' configuration an identical re-Put left
  'completed' through 90 s at 1/s ([S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), n=1) and in 12 samples over 60 s after a 10x storm
  ([S3T-TABLEREPLICATION-039](#s3t-tablereplication-039), n=10); one configuration still 'pending' 300 s after its Put with the replica already listed
  ([S3T-TABLEREPLICATION-007](#s3t-tablereplication-007) re-verification, -046, n=1). Full ranges (62-72 s empty table, 436-521 s schema table,
  60.6-72.5 s to 'failed') are in `table-replication-status.md`.
- Get/Status -> 404 after DeleteTableReplication: 0.019 s / 0.044 s, 0 stale 200s at ~200 ms polling
  ([S3T-TABLEREPLICATION-018](#s3t-tablereplication-018), n=1); after DeleteTable the four tableArn-keyed Gets -> 404 on the first call at 0.01-0.05 s
  ([S3T-TABLE-176](#s3t-table-176), n=4); the source's modifiedAt moved to ~0.2 s before the Delete call ([S3T-TABLE-177](#s3t-table-177), n=1).
- Catalogues: 3 ARN forms x 5 ops, 15/15 BadRequestException ([S3T-TABLE-173](#s3t-table-173), n=15); ghost table / ghost bucket / foreign
  region x 5 ops ([S3T-TABLE-174](#s3t-table-174), n=15); 50 sequential write pairs, 0 x 500, second-write latencies 32-162 ms ([S3T-TABLE-304](#s3t-table-304),
  n=50).

### Handling gaps

The controller implements Table, Namespace and TableBucket only; `generator.yaml` carries no replication resource,
operation or field, so no finding here compares against existing controller code and none is marked `suspect-bug` -
nothing in this document renders as a defect. A future TableReplication implementation must handle at minimum, with the
evidence here: the two-token model - GetTableReplication before every write, never GetTable.versionToken, 409 means
're-read' not 'retry', the Delete message does not tell stale from garbage ([S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), -007, -015, -016,
-036, -046); first-Put vs update token rules and the 404 'configuration does not exist' as the unset read (-001, -010,
-034); token rotation on every accepted Put - diff before Put, no re-Put on a no-op, no token cached across reconciles
(-038, -039, -044); Put's literal 'success' is not a status (-003, -009, -014, -037); validation and role-form errors as
terminal codes at Put time, MinItems=1, duplicates and self-destination accepted as sent (-011, -012, -013); iam:PassRole
and AccessDeniedException vs ForbiddenException (-023, -024); delete = 204, both Gets 404 at once, repeat 404, plus the
leftover replica to DeleteTable (-007, -018, -046); tableArn as the key - the UUID ARN from CreateTable/GetTable, 400 on
derived forms, the message-only 404 split after DeleteTable, the pre-rename ARN still resolving ([S3T-TABLE-173](#s3t-table-173), -174, -175,
-176) - and GetTable(source) silent about the configuration ([S3T-TABLE-177](#s3t-table-177)).

### Where next

- `table-replication-status.md` - everything after the 200: destinations [] for ~60 s, per-destination 'pending' /
  'completed' / 'failed', the stale ~61 s view after every Put, 'failed' recovering only via a token-refreshing re-Put,
  replica tables and orphans, name collisions, dead destinations and the bucket-level precedence cells.
- `tablebucket-replication.md` - the bucket-level twin (PutTableBucketReplication token rules, role prerequisites); the join
  point is [S3T-TABLEREPLICATION-034](#s3t-tablereplication-034)'s caveat: under a bucket-level rule only, GetTableReplication(table) -> 404 while
  GetTableReplicationStatus -> 200.
- `table.md` - DeleteTable/RenameTable/ListTables behind [S3T-TABLE-175](#s3t-table-175), -176 ([S3T-TABLE-231](table.md#s3t-table-231) and -317 render there);
  `table-metadata-location.md` - the table's own versionToken and UpdateTableMetadataLocation CAS semantics the replication
  token must not be confused with ([S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), -007; [S3T-TABLE-031](table-metadata-location.md#s3t-table-031) and -315 render there); `table-policy.md`
  - AccessDeniedException vs ForbiddenException ([S3T-TABLEREPLICATION-023](#s3t-tablereplication-023), -024); `table-record-expiration.md` - the other
  tableArn-keyed family in the shared 400/404/405 catalogue ([S3T-TABLE-173](#s3t-table-173), -174, -176; [S3T-TABLE-221](table-record-expiration.md#s3t-table-221) renders there).
- Long catalogues worth their `details/<id>.md` page: [S3T-TABLEREPLICATION-011](#s3t-tablereplication-011)/-012/-013 (validation shapes and role-ARN
  forms with exact messages), -015 (token matrix), -023 (Deny / other-account / foreign-destination cells), [S3T-TABLE-174](#s3t-table-174)
  (ghost table, ghost bucket, foreign region x 5 ops), [S3T-TABLE-304](#s3t-table-304) (10 write pairs with latencies).
- Open experiments (STATUS Phase 8): the reverify set re-confirmed [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003) and -007 on 2026-10-15 and added
  -046 (same Delete 409 message for table and random token, replica left behind; still 'pending' at 300 s); F-followups-2
  settled nothing on the configuration API itself. Unwatched outcomes ([D, D], [S], aws-cn role: -012, -013), the n=1
  IAM-propagation sample (-009) and the one-off 500 ([S3T-TABLE-304](#s3t-table-304)) are in `table-replication.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 28 (high 10 / medium 14 / low 4); duplicates folded into the appendix: 0
- handling: handled 0 · partial 0 · tracked 0 · unhandled 28 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: delete-semantics 4, error-code 4, idempotency 4, request-validation 4, identity 3, other 3,
  prerequisite 2, quota-limit 1, read-gap 1, shape-mismatch 1, sub-resource-api 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableReplication | delete | tableArn, versionToken | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableRecordExpirationConfiguration | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableRecordExpirationJobStatus | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableReplication | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableReplicationStatus | read | tableArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| PutTableMaintenanceConfiguration | create | tableBucketARN, namespace, name, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTablePolicy | create | tableBucketARN, namespace, name, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableRecordExpirationConfiguration | create | tableArn, value | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| PutTableReplication | create | tableArn, configuration | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

Also referenced by findings but not in the s3tables model (other services or annotated variants): CreateRole

## State machine

- **JobStatus**: Not_Yet_Run, Successful, Failed, Disabled
- **ReplicationStatus**: pending, completed, failed (transitional: pending)
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)
- **TableRecordExpirationJobStatus**: NotYetRun, Successful, Failed, Disabled

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-table-173"></a>**S3T-TABLE-173** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **tableArn-keyed ops with a bucket ARN, name-based pseudo ARN or 'ns/table' -> BadRequestException 400 'The specified table ARN is not valid.'**
  All five tableArn-keyed operations (GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration(enabled, days=1),
  GetTableReplication, GetTableReplicationStatus) answered BadRequestException (HTTP 400) 'The specified table
  ARN is not valid.' for tableArn = (a) the table bucket ARN, (b) the name-based
  '<bucketARN>/table/<namespace>/<table>', (c) the bare '<namespace>/<table>'. botocore sent all three to the
  wire (it does not enforce the model's tableArn regex). For reference the real UUID-based table ARN gives:
  the three record-expiration ops -> MethodNotAllowedException (HTTP 405) 'Table record expiration
  configuration API calls to this table are not allowed'; GetTableReplication and GetTableReplicationStatus ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'
  - ACK: is_arn_primary_key, custom_find, terminal_codes · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration, GetTableReplication,
    GetTableReplicationStatus · fields: tableArn
  - repro: CreateTable -> call the five tableArn-keyed ops with the bucket ARN, a name-based pseudo ARN and a
    non-ARN (SDK client, then parameter_validation=False)
  - measurements: forms=3, ops=5, n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-030](table-encryption-storage.md#s3t-table-030), [S3T-TABLE-175](#s3t-table-175), [S3T-TABLE-174](#s3t-table-174), [S3T-TABLE-176](#s3t-table-176), [S3T-TABLE-146](table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-147](table-record-expiration.md#s3t-table-147) · hypotheses: H-U-103 · evidence: table/identity/expiration-arn-forms-rename-delete
  - notes: H-U-103 confirmed for the three forms tried (bucket ARN, name-based pseudo ARN, non-ARN):
    BadRequestException 400 about the ARN, never NotFoundException. The real ARN cannot be derived from
    bucket+namespace+name; a controller must read it from CreateTable/GetTable. 400 is terminal (retrying
    never...
  - full notes: [details/S3T-TABLE-173.md](details/S3T-TABLE-173.md)

- <a id="s3t-table-175"></a>**S3T-TABLE-175** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After RenameTable the pre-rename table ARN still resolves: GetTable(old ARN) name=new, tableARN unchanged; old triple -> NotFoundException**
  RenameTable(newName) -> HTTP 204. Then with the OLD (pre-rename) table ARN:
  GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405) (same as before the rename:
  the table still resolves); GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  GetTableReplication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.' (same as before); GetTable(tableArn=old ARN) -> HTTP 200 with the NEW name, tableARN identical to
  the pre-rename ARN, versionToken identical to the one returned by CreateTable.
  GetTable(bucketARN+namespace+OLD name) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(new triple) -> HTTP 200 with the same tableARN.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTable, GetTableRecordExpirationConfiguration,
    GetTableReplication · fields: tableArn, tableARN, name
  - repro: CreateTable -> RenameTable(newName) -> Get* with the old table ARN; GetTable with old and new
    triples
  - measurements: renames=1, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-001](table-metadata-location.md#s3t-table-001), [S3T-TABLE-030](table-encryption-storage.md#s3t-table-030), [S3T-TABLE-173](#s3t-table-173) · hypotheses: H-U-104 · evidence:
    table/identity/expiration-arn-forms-rename-delete
  - notes: H-U-104 qualified: the UUID-based ARN survives the rename (same ARN, GetTable by old ARN shows the
    new name, old triple 404); the 'unchanged expiration configuration' part is untestable on a customer table
    because the record-expiration ops answer 405 before and after the rename (the table still...
  - full notes: [details/S3T-TABLE-175.md](details/S3T-TABLE-175.md)

- <a id="s3t-tablereplication-004"></a>**S3T-TABLEREPLICATION-004** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replication versionToken != GetTable.versionToken; Put with the table token -> ConflictException...**
  After PutTableReplication the table's own GetTable.versionToken did not change (it differs from the
  replication token). PutTableReplication with versionToken=GetTable's token -> ConflictException (HTTP 409)
  'The specified version token does not match the configuration.'. PutTableReplication without a token while a
  configuration exists -> BadRequestException (HTTP 400) 'A version token is not specified.'. Re-Put of the
  identical configuration with the current replication token -> HTTP 200, token rotated.
  DeleteTableReplication with GetTable's token -> ConflictException (HTTP 409) 'Provided version token does
  not match the table version token.'.
  - ACK: custom_update, custom_delete · ops: PutTableReplication, GetTable, DeleteTableReplication · fields:
    versionToken
  - repro: PutTableReplication -> GetTable -> Put(versionToken=GetTable token) -> Put(no token) ->
    Put(current) -> DeleteTableReplication(GetTable token)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-006](tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018) ·
    hypotheses: H-U-137 · evidence: table/round-trip/repl-put-get
  - notes: H-U-137 confirmed (distinct tokens; the table token is rejected with ConflictException).
    Table-level token rules mirror the bucket level: token-less Put on existing -> BadRequestException,
    unchanged re-Put rotates.
  - full notes: [details/S3T-TABLEREPLICATION-004.md](details/S3T-TABLEREPLICATION-004.md)

## Idempotency

- <a id="s3t-tablereplication-015"></a>**S3T-TABLEREPLICATION-015** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Stale versionToken: Put x5, Delete x5 -> ConflictException/409 (5/5); Delete('x') -> 409; token-less Put on existing or Delete('') -> 400**
  With a configuration whose token rotated twice: PutTableReplication(token from 2 Puts ago) five times 2 s
  apart -> ['ConflictException/409', 'ConflictException/409', 'ConflictException/409',
  'ConflictException/409', 'ConflictException/409'] (message 'The specified version token does not match the
  configuration.'); DeleteTableReplication(same stale token) five times -> ['ConflictException/409',
  'ConflictException/409', 'ConflictException/409', 'ConflictException/409', 'ConflictException/409'] (message
  'Provided version token does not match the table version token.'). PutTableReplication WITHOUT versionToken
  while a configuration exists -> BadRequestException (HTTP 400) 'A version token is not specified.'.
  DeleteTableReplication(versionToken='x') -> ConflictException (HTTP 409) 'Provided version token does not
  match the table version token.'; DeleteTableReplication(versionToken='', client validation off) ->
  BadRequestException (HTTP 400) 'A version token is not specified.'. GetTableReplication afterwards -> 200
  (config intact) with the token unchanged.
  - ACK: custom_update, custom_delete, terminal_codes · ops: PutTableReplication, DeleteTableReplication,
    GetTableReplication · fields: versionToken
  - repro: Put (T1) -> Put (T2) -> Put/Delete with T1 x5 @2 s -> Put without token -> Delete('x') ->
    Delete('') -> Get
  - measurements: stale_retries=[5], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026),
    [S3T-TABLEBUCKETREPLICATION-006](tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008),
    [S3T-TABLEREPLICATION-036](#s3t-tablereplication-036), [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018),
    [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEREPLICATION-020](table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021) · hypotheses: H-U-140,
    H-U-141, H-U-142, H-U-174 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-142 confirmed: the stale-token ConflictException is deterministic (5/5 identical for Put, 5/5
    for Delete); 'You can retry the request' is misleading - re-read the token instead. H-U-140 (token-less
    Put = last-writer-wins) REFUTED and H-U-141 (token-less Put -> 409) REFUTED: it is...
  - full notes: [details/S3T-TABLEREPLICATION-015.md](details/S3T-TABLEREPLICATION-015.md)

- <a id="s3t-tablereplication-016"></a>**S3T-TABLEREPLICATION-016** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Two concurrent PutTableReplication with the same current token: 5 rounds of 5 gave exactly one 200 + one ConflictException**
  Five rounds: two threads Put [D] and Put [E] with the same current versionToken at the same instant. Per
  round codes [['200/200', 'ConflictException/409'], ['200/200', 'ConflictException/409'], ['200/200',
  'ConflictException/409'], ['200/200', 'ConflictException/409'], ['200/200', 'ConflictException/409']];
  stored destinations after each round [['D'], ['D'], ['D'], ['D'], ['D']]; the GetTableReplication token
  matched the 200 response's token in 5/5 rounds. Loser messages: ['The specified version token does not match
  the configuration.'].
  - ACK: custom_update, requeue · ops: PutTableReplication, GetTableReplication · fields: versionToken
  - repro: Get token T; fire Put [D] and Put [E] with T in parallel; Get; repeat 5x
  - measurements: rounds_one_winner=[5], rounds_two_ok=[0], n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018),
    [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010) · hypotheses: H-U-187 ·
    evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-187 confirmed: the token is a real optimistic lock.

- <a id="s3t-tablereplication-038"></a>**S3T-TABLEREPLICATION-038** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Identical PutTableReplication re-Put with the current token -> HTTP 200, versionToken ROTATES; status right after: no-destinations**
  PutTableReplication(first) -> HTTP 200 token1; identical re-Put with versionToken=token1 148 ms later ->
  HTTP 200, response token != token1; GetTableReplication then returns the re-Put's token: True.
  GetTableReplicationStatus immediately after: no-destinations. The initial watch then went [[0.03,
  "no-destinations"], [60.62, "pending"], [65.68, "completed"]].
  - ACK: custom_update, annotation-shadow-state · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken, status
  - repro: PutTableReplication -> PutTableReplication(same configuration, versionToken from the first
    response) -> GetTableReplication
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002), [S3T-TABLEREPLICATION-014](#s3t-tablereplication-014), [S3T-TABLEREPLICATION-039](#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEREPLICATION-019](table-replication-status.md#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-031](table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-025](table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-043](table-replication-status.md#s3t-tablereplication-043),
    [S3T-TABLEREPLICATION-005](table-replication-status.md#s3t-tablereplication-005), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013) · evidence:
    table/creative/replica-settings-fidelity
  - notes: Cross-service analogy (ElastiCache 'No modifications were requested' rejection vs DynamoDB no-op
    200): table-level counterpart of [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002) (bucket-level re-Put rotates the token). A
    controller that re-Puts on every reconcile must re-read the token each time if it rotates. n=1.

- <a id="s3t-tablereplication-044"></a>**S3T-TABLEREPLICATION-044** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Existing table-level config: Put token omitted/'' -> BadRequestException; table or stale token -> ConflictException; current -> 200, rotates**
  With a 'completed' table-level replication config: PutTableReplication(same document) with versionToken
  omitted -> BadRequestException/400 'A version token is not specified.'; '' -> BadRequestException/400 'A
  version token is not specified.'; GetTable.versionToken -> ConflictException/409 'The specified version
  token does not match the configuration.'; the token from the first Put (stale) -> ConflictException/409 'The
  specified version token does not match the configuration.'; the current GetTableReplication token -> HTTP
  200 (new token in the response: True). Status polled 1/s for 90 s after that identical re-Put:
  completed@0.0s. GetTableReplication versionToken sampled after Put/at failed/at completed/after re-Put: 3
  distinct values over 4 samples (['after-put-0', 'at-failed', 'at-completed', 'after-identical-reput']), i.e.
  status transitions do not rotate it; only Puts do.
  - ACK: custom_update, is_read_only, compare.is_ignored+delta_pre_compare · ops: PutTableReplication,
    GetTableReplication, GetTableReplicationStatus · fields: versionToken
  - repro: completed config -> Put(same doc) with each token source -> Put(current) -> poll status 90 s;
    GetTableReplication token sampled across phases
  - measurements: tokens_distinct=3, tokens_sampled=4, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018), [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007),
    [S3T-TABLEREPLICATION-020](table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-014](#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003) · evidence:
    tablereplication/creative/controller-loop
  - notes: extends [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010) (first-Put token rules) to the update path a controller runs on an
    existing config, and -004 with the token's stability across status transitions. Decision rule: on update
    always read GetTableReplication first and forward its token; never forward the table's token;...
  - full notes: [details/S3T-TABLEREPLICATION-044.md](details/S3T-TABLEREPLICATION-044.md)

## Errors

- <a id="s3t-table-174"></a>**S3T-TABLE-174** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Ghost table UUID in tableArn -> 404 'The specified table does not exist.' on all 5 ops; ghost bucket: replication Gets say 'bucket' instead**
  tableArn = <real bucket ARN>/table/<uuid4 that never existed>: GetTableRecordExpirationConfiguration ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableRecordExpirationJobStatus ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; PutTableRecordExpirationConfiguration ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'; GetTableReplication -> NotFoundException
  (HTTP 404) 'The specified table does not exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404)
  'The specified table does not exist.'. Same ARN shape but the table UUID of an existing table under a
  NON-existing bucket name: GetTableRecordExpirationConfiguration -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; GetTableRecordExpirationJobStatus -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; PutTableRecordExpirationConfiguration -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'; GetTableReplication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'. Existing table ARN with the region segment swapped to us-east-1, sent to the us-west-2 endpoint:
  GetTableRecordExpirationConfiguration -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; GetTableRecordExpirationJobStatus -> BadRequestException (HTTP 400) 'The
  provided Region does not match the Region of the resource.'; PutTableRecordExpirationConfiguration ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  GetTableReplication -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'; GetTableReplicationStatus -> BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'.
  - ACK: exceptions.404, terminal_codes · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration, GetTableReplication,
    GetTableReplicationStatus · fields: tableArn
  - repro: Call the five ops with <bucketARN>/table/<random uuid>, with a non-existing bucket name, and with a
    foreign region segment
  - measurements: n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-173](#s3t-table-173), [S3T-TABLE-176](#s3t-table-176), [S3T-TABLE-146](table-record-expiration.md#s3t-table-146), [S3T-TABLE-147](table-record-expiration.md#s3t-table-147) · hypotheses: H-U-109, H-U-102 ·
    evidence: table/identity/expiration-arn-forms-rename-delete
  - notes: H-U-109 confirmed: Put and GetJobStatus (and the other three) -> NotFoundException 404 with a
    message naming the table ('The specified table does not exist.'); with a non-existing BUCKET name in the
    ARN the record-expiration ops still say 'table does not exist' while the replication Gets say 'The...
  - full notes: [details/S3T-TABLE-174.md](details/S3T-TABLE-174.md)

- <a id="s3t-tablereplication-001"></a>**S3T-TABLEREPLICATION-001** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Existing table without replication config: GetTableReplication, GetTableReplicationStatus, DeleteTableReplication('abc') -> 404**
  On an existing table that never had a replication configuration: GetTableReplication -> NotFoundException
  (HTTP 404) 'The specified replication configuration does not exist.'. GetTableReplicationStatus ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'.
  DeleteTableReplication(versionToken='abc') -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'. (For a table ARN that does not exist, both Gets return NotFoundException
  'The specified table does not exist.')
  - ACK: exceptions.404, custom_find · ops: GetTableReplication, GetTableReplicationStatus,
    DeleteTableReplication · fields: destinations
  - repro: CreateTable -> GetTableReplication / GetTableReplicationStatus / DeleteTableReplication('abc')
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEREPLICATION-034](#s3t-tablereplication-034), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-037](#s3t-tablereplication-037),
    [S3T-TABLEBUCKETREPLICATION-018](tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003),
    [S3T-TABLEBUCKETREPLICATION-007](tablebucket-replication.md#s3t-tablebucketreplication-007), [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010),
    [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013) · hypotheses:
    H-U-138 · evidence: service/error-taxonomy/repl-permissions
  - notes: Complements the table-level round-trip probe: the 'unset' read shape a controller must map to 'no
    replication'. Full original title: Existing table without replication config: GetTableReplication,
    GetTableReplicationStatus and DeleteTableReplication all -> NotFoundException 404

- <a id="s3t-tablereplication-023"></a>**S3T-TABLEREPLICATION-023** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Explicit IAM Deny: Get/Put/Delete TableReplication -> AccessDeniedException; other-account tableArn -> ForbiddenException; foreign dest 200**
  (a) Session policy with an explicit Deny on the three replication actions: GetTableReplication ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableReplication on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a session '; GetTableReplicationStatus (not denied) -> HTTP 200;
  PutTableReplication -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableReplication
  on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a session '; DeleteTableReplication -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableReplication on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-e64755-fd-src/table/bb81936f-64b9-4a90-809d-4050f3e82620
  with an explicit deny in a sessi'; GetTable -> HTTP 200. (b) tableArn with the account segment replaced by
  <ACCOUNT>: GetTableReplication -> ForbiddenException (HTTP 403) 'Unauthorized'; GetTableReplicationStatus ->
  ForbiddenException (HTTP 403) 'Unauthorized'; PutTableReplication -> ForbiddenException (HTTP 403)
  'Unauthorized'; DeleteTableReplication -> ForbiddenException (HTTP 403) 'Unauthorized'. A
  destinationTableBucketARN of another account (same bucket name) -> HTTP 200.
  - ACK: terminal_codes, requeue · ops: GetTableReplication, GetTableReplicationStatus, PutTableReplication,
    DeleteTableReplication · fields: tableArn, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: AssumeRole(shard role, session policy Deny) -> each op; admin -> each op with the ARN account
    replaced by <ACCOUNT>
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLEREPLICATION-040](table-replication-status.md#s3t-tablereplication-040),
    [S3T-TABLEREPLICATION-042](table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](#s3t-tablereplication-012), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEREPLICATION-024](#s3t-tablereplication-024),
    [S3T-TABLEREPLICATION-013](#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028) · hypotheses: H-U-184 · evidence:
    table/error-taxonomy/repl-forbidden-vs-denied
  - notes: H-U-184 confirmed: IAM deny -> ['AccessDeniedException']; foreign account identifier ->
    ['ForbiddenException'].
  - full notes: [details/S3T-TABLEREPLICATION-023.md](details/S3T-TABLEREPLICATION-023.md)

## Request validation

- <a id="s3t-tablereplication-010"></a>**S3T-TABLEREPLICATION-010** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **First PutTableReplication with versionToken='abc' -> NotFoundException; with '' -> 200; omitted -> 200...**
  On a table that never had a replication configuration: PutTableReplication(versionToken='abc') ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  PutTableReplication(versionToken='' , client-side validation disabled) -> HTTP 200; GetTableReplication
  afterwards -> HTTP 200. Omitting versionToken on the first Put -> HTTP 200 (phase 1). Fresh table reads:
  GetTableReplication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; GetTableReplicationStatus -> NotFoundException (HTTP 404) 'The specified replication configuration
  does not exist.'; DeleteTableReplication(versionToken='abc') -> NotFoundException (HTTP 404) 'The specified
  replication configuration does not exist.'.
  - ACK: custom_create, exceptions.404 · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus, DeleteTableReplication · fields: versionToken
  - repro: CreateTable -> Get/Status/Delete('abc') -> Put(versionToken='abc') -> Put(versionToken='') -> Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013),
    [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEBUCKETREPLICATION-006](tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017),
    [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027), [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015),
    [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018) · hypotheses: H-U-191,
    H-U-157, H-U-175 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-191: token 'abc' on a first Put -> NotFoundException (hypothesis said ConflictException 409);
    empty token -> accepted (hypothesis said BadRequestException 400). H-U-157 confirmed (both Gets 404
    'replication configuration does not exist', not the table message). H-U-175 second half confirmed:...
  - full notes: [details/S3T-TABLEREPLICATION-010.md](details/S3T-TABLEREPLICATION-010.md)

- <a id="s3t-tablereplication-011"></a>**S3T-TABLEREPLICATION-011** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication (client validation off): rules=[], destinations=[], no rules, two rules -> ValidationException; no role -> BadRequest**
  With botocore parameter validation disabled (the SDK itself rejects these shapes: rules min/max 1,
  destinations min 1 max 5): configuration.rules=[] -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'configuration.rules' failed to satisfy constraint: Member must have length greater than
  or equal to 1'; rules=[{destinations: []}] -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'configuration.rules.1.member.destinations' failed to satisfy constraint: Member must have length
  greater than or equal to 1'; configuration without rules -> ValidationException (HTTP 400) '1 validation
  error detected: Value at 'configuration.rules' failed to satisfy constraint: Member must not be null';
  configuration without role -> BadRequestException (HTTP 400) 'Invalid input: Must provide either table arn
  alone OR all of (table bucket arn, namespace, table name) together'; two rules [{D}, {E}] ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'configuration.rules' failed to
  satisfy constraint: Member must have length less than or equal to 1'. GetTableReplication after the rejected
  calls -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.' (nothing
  stored).
  - ACK: terminal_codes, custom_delete · ops: PutTableReplication, GetTableReplication · fields:
    configuration.rules, configuration.rules[].destinations, configuration.role
  - repro: Fresh table; PutTableReplication with each malformed configuration (parameter_validation=False);
    GetTableReplication after
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-012](#s3t-tablereplication-012), [S3T-TABLEREPLICATION-013](#s3t-tablereplication-013), [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009),
    [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003) · hypotheses: H-U-143, H-U-144 · evidence:
    table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-143 confirmed: the empty-list shapes are rejected with ['ValidationException']. 'Remove all
    destinations' cannot be expressed through Put; the CRD needs MinItems=1. H-U-144 confirmed: two rules ->
    ValidationException (HTTP 400) '1 validation error detected: Value at 'configuration.rules' failed...
  - full notes: [details/S3T-TABLEREPLICATION-011.md](details/S3T-TABLEREPLICATION-011.md)

- <a id="s3t-tablereplication-012"></a>**S3T-TABLEREPLICATION-012** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication destinations: [D, D] -> 200 (2 stored); [source bucket] -> 200; 6 dests -> ValidationException; 5 dests -> 200** (hypothesis refuted; behavior confirmed)
  Fresh table, valid role: destinations [D, D] (same bucket twice) -> HTTP 200; GetTableReplication then -> 2
  destinations ['D', 'D']. destinations [S] (the source table's own bucket) -> HTTP 200; Get -> stored ['S'].
  Six well-formed ARNs of nonexistent same-region buckets -> ValidationException (HTTP 400) '1 validation
  error detected: Value at 'configuration.rules.1.member.destinations' failed to satisfy constraint: Member
  must have length less than or equal to 5'. Five of them -> HTTP 200; Get -> stored 5 destinations.
  - ACK: terminal_codes, compare.is_ignored+delta_pre_compare · ops: PutTableReplication, GetTableReplication
    · fields: configuration.rules[].destinations[].destinationTableBucketARN
  - repro: Fresh table: Put [D, D] -> Get -> Delete; Put [self] -> Get -> Delete; Put 6 nonexistent -> Put 5
    nonexistent -> Get -> Delete
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-011](#s3t-tablereplication-011), [S3T-TABLEREPLICATION-013](#s3t-tablereplication-013),
    [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEREPLICATION-040](table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](table-replication-status.md#s3t-tablereplication-042),
    [S3T-TABLEREPLICATION-023](#s3t-tablereplication-023), [S3T-TABLEBUCKET-118](tablebucket-replication.md#s3t-tablebucket-118) · hypotheses: H-U-145, H-U-146, H-U-147, H-U-148 · evidence:
    table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-146 REFUTED (duplicate accepted); H-U-147 refuted (stored twice, not de-duplicated); H-U-148
    REFUTED (self accepted); H-U-145 confirmed (6 -> ValidationException, 5 -> 200). Table level mirrors the
    bucket level ([S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009)).

- <a id="s3t-tablereplication-013"></a>**S3T-TABLEREPLICATION-013** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication role: IAM user ARN -> ValidationException; aws-cn role -> 200; bare name -> BadRequest; other-account role -> Forbidden**
  Fresh table: role = IAM user ARN -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'configuration.role' failed to satisfy constraint: Member must satisfy regular expression pattern:
  arn:.+:iam::[0-9]{12}:role/<PRINCIPAL>'. role = role ARN in the aws-cn partition -> HTTP 200 (echoed
  'arn:aws-cn:iam::<ACCOUNT>:role/<PRINCIPAL>'). role = bare role name -> BadRequestException (HTTP 400)
  'Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace, table name)
  together'. role = role ARN of another account -> ForbiddenException (HTTP 403) 'Cross-account pass role is
  not allowed.'. GetTableReplication afterwards -> NotFoundException (HTTP 404) 'The specified replication
  configuration does not exist.'. The shard role ARN carries a path (role/<PRINCIPAL>) and is echoed
  byte-identical by GetTableReplication (phase 1).
  - ACK: references, terminal_codes · ops: PutTableReplication, GetTableReplication · fields:
    configuration.role
  - repro: Fresh table: Put with each role form (parameter_validation=False); Get after each; Delete when
    accepted
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEREPLICATION-011](#s3t-tablereplication-011), [S3T-TABLEREPLICATION-012](#s3t-tablereplication-012),
    [S3T-TABLEBUCKETREPLICATION-009](tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEREPLICATION-023](#s3t-tablereplication-023), [S3T-TABLEREPLICATION-024](#s3t-tablereplication-024),
    [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-026](table-policy.md#s3t-table-026) · hypotheses: H-U-186 · evidence:
    table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-186 qualified: see behavior. The model pattern is arn:.+:iam::[0-9]{12}:role/<PRINCIPAL> (any
    partition).
  - full notes: [details/S3T-TABLEREPLICATION-013.md](details/S3T-TABLEREPLICATION-013.md)

## Update granularity and ordering

- <a id="s3t-tablereplication-009"></a>**S3T-TABLEREPLICATION-009** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication with a role created 6s earlier -> 200 on attempt 1 (0 rejected first); first-Put status='success'**
  iam:CreateRole (path /ackq/, trust replication.s3tables.amazonaws.com) + PutRolePolicy, then
  PutTableReplication(empty table -> destination bucket in another region) issued 5.5s after CreateRole:
  attempts [+5.5s 200]; first 200 at +5.5s with status='success'.
  - ACK: requeue, references · ops: PutTableReplication, CreateRole · fields: configuration.role
  - repro: CreateRole -> PutRolePolicy -> PutTableReplication every 2 s until 200
  - measurements: first_put_ok_since_role_create_s=[5.5], rejected_attempts_before_ok=[0], rejected_codes=[],
    n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEREPLICATION-002](table-replication-status.md#s3t-tablereplication-002), [S3T-TABLEREPLICATION-008](table-replication-status.md#s3t-tablereplication-008),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEBUCKETREPLICATION-028](tablebucket-replication.md#s3t-tablebucketreplication-028), [S3T-TABLEBUCKETREPLICATION-007](tablebucket-replication.md#s3t-tablebucketreplication-007),
    [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEBUCKETREPLICATION-010](tablebucket-replication.md#s3t-tablebucketreplication-010) · hypotheses: H-U-139 · evidence:
    table/dependencies/repl-prereqs-token-matrix
  - notes: IAM propagation sample for the table-level Put (n=1 role); the role is validated asynchronously
    (part a), so a rejected attempt here would be IAM eventual consistency of iam:PassRole, not role trust.
    H-U-139 first-Put status string recorded here; the rest in the token-matrix finding.

- <a id="s3t-tablereplication-024"></a>**S3T-TABLEREPLICATION-024** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Caller scoped to the source bucket only (+iam:PassRole): PutTableReplication 200 (no caller perms on dest); without PassRole -> AccessDenied** (hypothesis refuted; behavior confirmed)
  Assumed-role sessions on the shard role. Session policy allowing s3tables:* only on the source bucket and
  its tables plus iam:PassRole: GetTableReplication -> HTTP 200; PutTableReplication(t -> D, current token) ->
  HTTP 200. Session policy also allowing s3tables:* on D and its tables: Put -> HTTP 200. Session policy
  allowing s3tables:* everywhere but NO iam:PassRole: Put -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: iam:PassRole on resource:
  arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL> because no session policy allows the iam:PassRole action';
  GetTableReplication -> HTTP 200.
  - ACK: references, terminal_codes · ops: PutTableReplication, GetTableReplication · fields:
    configuration.role, configuration.rules[].destinations[].destinationTableBucketARN
  - repro: AssumeRole with a session policy scoped to S (+PassRole) -> Put t->D; widen to D -> Put; drop
    PassRole -> Put
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLEREPLICATION-023](#s3t-tablereplication-023), [S3T-TABLEREPLICATION-013](#s3t-tablereplication-013),
    [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028), [S3T-TABLE-026](table-policy.md#s3t-table-026) · hypotheses: H-U-185 · evidence:
    table/error-taxonomy/repl-forbidden-vs-denied
  - notes: H-U-185 REFUTED: the Put succeeded without any caller permission on the destination bucket (only
    the role needs it).
  - full notes: [details/S3T-TABLEREPLICATION-024.md](details/S3T-TABLEREPLICATION-024.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-177"></a>**S3T-TABLE-177** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTable(source) is unchanged by PutTableReplication (no field changed) and DeleteTableReplication (changed: ['modifiedAt'])...**
  GetTable(source) before the first Put, after the Put (+status populated) and after DeleteTableReplication:
  fields differing before->after Put: none; after Put->after Delete: ['modifiedAt']. managedByService=None and
  managedTableInformation=None on the source throughout.
  - ACK: custom_find, synced.when · ops: GetTable, PutTableReplication, DeleteTableReplication · fields:
    versionToken, modifiedAt, managedByService, managedTableInformation
  - repro: GetTable -> PutTableReplication -> GetTable -> ... -> DeleteTableReplication -> GetTable; diff
    bodies
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-031](table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-136](table-replication-status.md#s3t-table-136), [S3T-TABLE-223](table-replication-status.md#s3t-table-223), [S3T-TABLE-221](table-record-expiration.md#s3t-table-221), [S3T-TABLE-299](table-replication-status.md#s3t-table-299) ·
    hypotheses: H-U-188 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-188 qualified: versionToken, managedByService and managedTableInformation never changed;
    modifiedAt was identical before and after the Put but had moved (to a timestamp ~0.2 s before the
    DeleteTableReplication call, i.e. while the same-region destination E was completing its first
    replication)...
  - full notes: [details/S3T-TABLE-177.md](details/S3T-TABLE-177.md)

- <a id="s3t-tablereplication-014"></a>**S3T-TABLEREPLICATION-014** `shape-mismatch` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableReplication response status is the literal 'success' (first Put, re-Put, dest add), not a ReplicationStatus; Get read-your-writes** (hypothesis refuted; behavior confirmed)
  PutTableReplication response status: first Put -> 'success'; identical re-Put with the current token ->
  'success'; Put adding a second destination -> 'success'. None is in the GetTableReplicationStatus enum
  ['pending', 'completed', 'failed']. GetTableReplication x10 within ~1 s after the Put returned 1 distinct
  token(s), all equal to the Put's token.
  - ACK: synced.when, custom_update · ops: PutTableReplication, GetTableReplication · fields: status,
    versionToken
  - repro: Put -> Get x10 @100 ms -> re-Put identical (token) -> Put [D, E] (token); compare status strings
  - measurements: get_x10_distinct_tokens=[1], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002) · hypotheses: H-U-139,
    H-U-160 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-139 REFUTED: the Put status is a literal operation result ('success'), not the per-destination
    enum, so it cannot seed status conditions. H-U-160 first half (read-your-writes on GetTableReplication)
    confirmed (10/10).
  - full notes: [details/S3T-TABLEREPLICATION-014.md](details/S3T-TABLEREPLICATION-014.md)

## Sub-resources

- <a id="s3t-tablereplication-003"></a>**S3T-TABLEREPLICATION-003** `sub-resource-api` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **PutTableReplication -> 200 {versionToken (20 chars), status 'success'}; GetTableReplication echoes role (with path) and destination**
  PutTableReplication(tableArn, role ARN with a path segment, one destination bucket in another region) on a
  table with an Iceberg schema -> HTTP 200 with versionToken (20 chars) and status='success' (lowercase).
  GetTableReplication immediately after -> members [configuration, versionToken]; configuration.role and
  destinationTableBucketARN byte-identical to the input; versionToken equals the Put's. GetTableReplication /
  GetTableReplicationStatus on a table ARN whose table id does not exist -> NotFoundException 404 'The
  specified table does not exist.' (an existing table WITHOUT a configuration gives 404 'The specified
  replication configuration does not exist.', see [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001)).
  - ACK: custom_update, exceptions.404 · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken, status, configuration.role,
    configuration.rules[].destinations[].destinationTableBucketARN
  - repro: CreateTable(schema) -> PutTableReplication -> GetTableReplication; Get*/Status on a random table id
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-018](tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003),
    [S3T-TABLEBUCKETREPLICATION-007](tablebucket-replication.md#s3t-tablebucketreplication-007), [S3T-TABLEREPLICATION-037](#s3t-tablereplication-037), [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001),
    [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013),
    [S3T-TABLEBUCKETREPLICATION-010](tablebucket-replication.md#s3t-tablebucketreplication-010), [S3T-TABLEREPLICATION-009](#s3t-tablereplication-009), [S3T-TABLEREPLICATION-038](#s3t-tablereplication-038),
    [S3T-TABLEREPLICATION-039](#s3t-tablereplication-039), [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-014](#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029),
    [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002) · hypotheses: H-U-138 · evidence: table/round-trip/repl-put-get
  - notes: H-U-138 confirmed. Full original title: PutTableReplication -> 200 {versionToken (20 chars), status
    'success'}; GetTableReplication echoes role ARN (with path) and destination byte-identical

## Delete semantics

- <a id="s3t-table-176"></a>**S3T-TABLE-176** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable all four tableArn-keyed Gets return NotFoundException 404 'The specified table does not exist.' on the first call (<0.1 s)**
  Customer table, no replication. Before DeleteTable: GetTableRecordExpirationConfiguration and
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405; GetTableReplication and
  GetTableReplicationStatus -> NotFoundException 404 'The specified replication configuration does not
  exist.'. DeleteTable -> HTTP 204. First call per op right after the delete returned (0.01 / 0.03 / 0.04 /
  0.05 s): all four -> NotFoundException (HTTP 404) 'The specified table does not exist.'; no 200 or 405 was
  seen after the delete (one call each sufficed, so no 1/s polling followed).
  - ACK: exceptions.404, pre-delete-cleanup · ops: DeleteTable, GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, GetTableReplication, GetTableReplicationStatus · fields: tableArn
  - repro: CreateTable -> DeleteTable -> the four Gets immediately and every 1 s for up to 30 s
  - measurements: first_404_s.GetTableRecordExpirationConfiguration=0.01,
    first_404_s.GetTableRecordExpirationJobStatus=0.03, first_404_s.GetTableReplication=0.04,
    first_404_s.GetTableReplicationStatus=0.05, n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-173](#s3t-table-173), [S3T-TABLE-174](#s3t-table-174), [S3T-TABLE-146](table-record-expiration.md#s3t-table-146), [S3T-TABLE-147](table-record-expiration.md#s3t-table-147) · hypotheses: H-U-105 · evidence:
    table/identity/expiration-arn-forms-rename-delete
  - notes: H-U-105 confirmed for the four Gets (no stale window; first call < 0.1 s). Note the replication
    Gets are 404 both before and after the delete and only the message changes ('replication configuration
    does not exist' -> 'table does not exist'), so for those two ops a controller cannot use the code...
  - full notes: [details/S3T-TABLE-176.md](details/S3T-TABLE-176.md)

- <a id="s3t-tablereplication-007"></a>**S3T-TABLEREPLICATION-007** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **DeleteTableReplication: table-token -> 409 'does not match the table version token', current -> 204, repeat -> 404; Get/Status -> 404...**
  DeleteTableReplication(versionToken = the table's own CreateTable/GetTable token) -> ConflictException 409
  'Provided version token does not match the table version token.' (a different message than Put's 'The
  specified version token does not match the configuration.'). DeleteTableReplication(current replication
  token) -> 204; GetTableReplication -> NotFoundException 404 'The specified replication configuration does
  not exist.'; GetTableReplicationStatus -> the same 404; repeating the Delete -> the same 404. The source
  table itself: GetTable -> 200, unchanged. (In this run the replica had already been deleted in the previous
  step, so what remains in the destination after a table-level delete is recorded in the bucket-level probe.)
  DeleteTable(source) afterwards -> 204.
  - ACK: custom_delete, pre-delete-cleanup, exceptions.404 · ops: DeleteTableReplication, GetTableReplication,
    GetTableReplicationStatus, UpdateTableMetadataLocation · fields: versionToken
  - repro: completed replication -> DeleteTableReplication(stale) -> DeleteTableReplication(current) ->
    Get/Status -> ListTables(destination) -> commit on source -> GetTable(replica) after 120 s
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](tablebucket-replication.md#s3t-tablebucketreplication-004),
    [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-036](#s3t-tablereplication-036), [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015),
    [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005),
    [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003),
    [S3T-TABLEBUCKETREPLICATION-017](tablebucket-replication.md#s3t-tablebucketreplication-017), [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-027](tablebucket-replication.md#s3t-tablebucketreplication-027),
    [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEBUCKETREPLICATION-032](tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLE-138](table-replication-status.md#s3t-table-138), [S3T-TABLE-196](table-replication-status.md#s3t-table-196),
    [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018) · hypotheses: H-S-061,
    H-U-137 · evidence: table/round-trip/repl-put-get
  - notes: Token kinds are enforced on Delete too (H-U-137). H-S-061/H-S-087 at table level could not be
    judged here because the replica was gone; see tablebucket/sub-resources/repl-backfill-delete for the
    bucket-level answer (replicas remain, read-only).
  - full notes: [details/S3T-TABLEREPLICATION-007.md](details/S3T-TABLEREPLICATION-007.md)

- <a id="s3t-tablereplication-018"></a>**S3T-TABLEREPLICATION-018** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableReplication(current token), no bucket-level config -> 204; Get/Status 404 at +0.019s/+0.044s (0 stale 200s); Delete again -> 404**
  DeleteTableReplication with the current token -> HTTP 204. Polling GetTableReplication and
  GetTableReplicationStatus alternately every ~200 ms: first NotFoundException at +0.019s (Get) and +0.044s
  (Status); stale HTTP 200 bodies seen: {'get': 0, 'status': 0}. A second Delete with the same (consumed)
  token -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'.
  - ACK: custom_delete, exceptions.404 · ops: DeleteTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken
  - repro: Delete(current token) -> poll both Gets @200 ms -> Delete(same token)
  - measurements: get_404_after_delete_s=[0.019], status_404_after_delete_s=[0.044], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEBUCKETREPLICATION-024](tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](tablebucket-replication.md#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-015](tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-016](tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLE-225](table-replication-status.md#s3t-table-225),
    [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044),
    [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEREPLICATION-004](#s3t-tablereplication-004), [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021),
    [S3T-TABLEREPLICATION-022](table-replication-status.md#s3t-tablereplication-022), [S3T-TABLEBUCKETREPLICATION-013](tablebucket-replication.md#s3t-tablebucketreplication-013), [S3T-TABLEBUCKETREPLICATION-014](tablebucket-replication.md#s3t-tablebucketreplication-014) · hypotheses:
    H-U-173, H-U-175 · evidence: table/dependencies/repl-prereqs-token-matrix
  - notes: H-U-173 confirmed (no stale 200). H-U-175 first half confirmed (second Delete ->
    NotFoundException).
  - full notes: [details/S3T-TABLEREPLICATION-018.md](details/S3T-TABLEREPLICATION-018.md)

- <a id="s3t-tablereplication-046"></a>**S3T-TABLEREPLICATION-046** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **DeleteTableReplication: table token and garbage token -> the same 409 message...**
  With the configuration still 'pending' (300 s after Put): DeleteTableReplication(versionToken=<GetTable
  token of the source table>) and DeleteTableReplication(versionToken='0123456789abcdefghij') both ->
  ConflictException 409 'Provided version token does not match the table version token.';
  PutTableReplication(same configuration, versionToken=<table token>) -> ConflictException 409 'The specified
  version token does not match the configuration.'. DeleteTableReplication(current GetTableReplication token)
  -> HTTP 204 with an empty body; GetTableReplication, GetTableReplicationStatus and a repeated Delete ->
  NotFoundException 404 'The specified replication configuration does not exist.'; GetTable(source) -> 200
  with no changed member. Immediately afterwards ListTables(destination bucket) still listed the replica table
  (same namespace and name as the source, type 'customer'); DeleteTable on it in the destination region ->
  204, then DeleteNamespace and DeleteTableBucket -> 204.
  - ACK: custom_delete, pre-delete-cleanup · ops: DeleteTableReplication, PutTableReplication,
    GetTableReplication, GetTableReplicationStatus, ListTables, DeleteTable · fields: versionToken
  - repro: PutTableReplication -> wait 300 s (still pending) -> Delete(table token) -> Delete(garbage token)
    -> Put(table token) -> Delete(current) -> Get/Status/Delete again -> GetTable -> ListTables(destination)
    -> DeleteTable(replica)
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007) · evidence: tablereplication/creative/reverify-set-e-repl
  - notes: Precision gain while re-verifying [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007) (all its codes and messages reproduced,
    here with the replication still pending rather than completed): the 409 message does not distinguish a
    stale token from a random one, and deleting the configuration does not remove the replica.
  - full notes: [details/S3T-TABLEREPLICATION-046.md](details/S3T-TABLEREPLICATION-046.md)

## Quotas and rate limits

- <a id="s3t-tablereplication-039"></a>**S3T-TABLEREPLICATION-039** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **10 identical PutTableReplication re-Puts back-to-back (fresh token each): 10 x 200 in 1.0 s, token rotates each time, status stays completed**
  On a 'completed' table-level configuration: 10 sequential identical re-Puts, each sending the token returned
  by the previous response: codes {"HTTP 200": 10}, latencies [102, 102, 106, 107, 95, 109, 99, 101, 97, 92]
  ms, wall 1.02s; token rotated on 10 of the successful calls; GetTableReplication afterwards returns the last
  token: True. GetTableReplicationStatus right after the burst: 'completed'; 12 samples over the next 60 s:
  ['completed']. Destination bucket still lists ['ackq_5c8005_t'] (no duplicate replica).
  - ACK: custom_update, annotation-shadow-state, requeue · ops: PutTableReplication, GetTableReplication,
    GetTableReplicationStatus · fields: versionToken, destinations[].replicationStatus
  - repro: completed table-level replication -> 10 x PutTableReplication(same configuration, token from the
    previous response) with no pause -> status samples
  - measurements: storm_wall_s=1.02, latency_ms=[102, 102, 106, 107, 95, 109, 99, 101, 97, 92], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-038](#s3t-tablereplication-038), [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044), [S3T-TABLEREPLICATION-014](#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003) ·
    evidence: tablereplication/creative/reput-storm-and-lost-destination
  - notes: Cross-service analogy (DynamoDB SSE/TTL change cooldowns, DeletionProtection 15 s cooldown;
    ElastiCache no-op rejection): is a controller that re-Puts the replication configuration every reconcile
    (because the token rotates, [S3T-TABLEREPLICATION-038](#s3t-tablereplication-038)) throttled, rejected or does it reset the...
  - full notes: [details/S3T-TABLEREPLICATION-039.md](details/S3T-TABLEREPLICATION-039.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-173](#s3t-table-173) | tableArn-keyed ops with a bucket ARN, name-based pseudo ARN or 'ns/table' -> BadRequestException 400 'The specified table ARN is not valid.' | forms=3, ops=5 | 15 |
| [S3T-TABLE-175](#s3t-table-175) | After RenameTable the pre-rename table ARN still resolves: GetTable(old ARN) name=new, tableARN unchanged; old triple -> NotFoundException | renames=1 | 1 |
| [S3T-TABLE-176](#s3t-table-176) | After DeleteTable all four tableArn-keyed Gets return NotFoundException 404 'The specified table does not exist.' on the first call (<0.1 s) | first_404_s.GetTableRecordExpirationConfiguration=0.01, first_404_s.GetTableRecordExpirationJobStatus=0.03, first_404_s.GetTableReplication=0.04, first_404_s.GetTableReplicationStatus=0.05 | 4 |
| [S3T-TABLE-304](#s3t-table-304) | Single-writer back-to-back writes to one table (10 ordered pairs x 5, 0 s gap): 0 x 500; the 500 seen on a replication source not reproduced | pairs=10, second_write_500_total=0, latency_ms_second_maint->tag=[105, 106, 111, 107, 99], latency_ms_second_tag->maint=[36, 34, 33, 35, 32], latency_ms_second_policy->tag=[142, 131, 128, 129, 134], latency_ms_second_tag->tag=[94, 98, 105, 99, 162] | 50 |
| [S3T-TABLEREPLICATION-009](#s3t-tablereplication-009) | PutTableReplication with a role created 6s earlier -> 200 on attempt 1 (0 rejected first); first-Put status='success' | first_put_ok_since_role_create_s=[5.5], rejected_attempts_before_ok=[0], rejected_codes=[] | 1 |
| [S3T-TABLEREPLICATION-014](#s3t-tablereplication-014) | PutTableReplication response status is the literal 'success' (first Put, re-Put, dest add), not a ReplicationStatus; Get read-your-writes | get_x10_distinct_tokens=[1] | 10 |
| [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015) | Stale versionToken: Put x5, Delete x5 -> ConflictException/409 (5/5); Delete('x') -> 409; token-less Put on existing or Delete('') -> 400 | stale_retries=[5] | 10 |
| [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016) | Two concurrent PutTableReplication with the same current token: 5 rounds of 5 gave exactly one 200 + one ConflictException | rounds_one_winner=[5], rounds_two_ok=[0] | 5 |
| [S3T-TABLEREPLICATION-018](#s3t-tablereplication-018) | DeleteTableReplication(current token), no bucket-level config -> 204; Get/Status 404 at +0.019s/+0.044s (0 stale 200s); Delete again -> 404 | get_404_after_delete_s=[0.019], status_404_after_delete_s=[0.044] | 1 |
| [S3T-TABLEREPLICATION-039](#s3t-tablereplication-039) | 10 identical PutTableReplication re-Puts back-to-back (fresh token each): 10 x 200 in 1.0 s, token rotates each time, status stays completed | storm_wall_s=1.02, latency_ms=[102, 102, 106, 107, 95, 109, 99, 101, 97, 92] | 10 |
| [S3T-TABLEREPLICATION-044](#s3t-tablereplication-044) | Existing table-level config: Put token omitted/'' -> BadRequestException; table or stale token -> ConflictException; current -> 200, rotates | tokens_distinct=3, tokens_sampled=4 | 1 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-replication.md` (configuration API) - unverified, stale or single-run items worth a future probe
(one line each). Status-side items (recovery, dead destinations, orphans, rename propagation) moved to
`table-replication-status.open-questions.md`.

- [S3T-TABLEREPLICATION-012](#s3t-tablereplication-012): destinations [D, D] stored twice and [S] (the source's own bucket) accepted with 200 - the
  resulting status entries and whether a duplicate or self-destination ever completes were not followed.
- [S3T-TABLEREPLICATION-013](#s3t-tablereplication-013): an aws-cn role ARN is accepted and echoed byte-identically; its asynchronous outcome ('failed'
  message) was not watched.
- [S3T-TABLEREPLICATION-009](#s3t-tablereplication-009) (n=1 role, confidence medium): one IAM-propagation sample (Put accepted 5.5 s after CreateRole,
  0 rejections) - re-run with Puts at +0/+1/+2 s to find whether iam:PassRole propagation ever rejects a first attempt.
- [S3T-TABLEREPLICATION-024](#s3t-tablereplication-024) (status refuted = hypothesis refuted): the no-caller-permission-on-destination cell is n=1 per
  session policy; a cross-account destination bucket (-023 accepted the ARN with 200) was never watched to its status.
- [S3T-TABLEREPLICATION-016](#s3t-tablereplication-016) (n=5): concurrent Puts always produced exactly one winner; a Put racing a Delete with the same
  token was not tried.
- [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015) / -036: the 'You can retry the request' hint in the 409 response was never true in 10/10
  retries - no probe timed how long a stale token stays rejected (whether the lock ever expires).
- [S3T-TABLEREPLICATION-010](#s3t-tablereplication-010): 'abc' on a first Put -> 404 and '' -> 200 are n=1 each; a 20-char well-formed random token on a
  never-configured table was not tried (-046 tried one only on an existing configuration: 409 on Delete).
- [S3T-TABLE-177](#s3t-table-177) (n=1): the source's modifiedAt moved ~0.2 s before DeleteTableReplication - whether the trigger is the
  Delete or the same-region destination completing its first replication is ambiguous.
- [S3T-TABLE-175](#s3t-table-175) (n=1): RenameTable to a new namespace was not tried for the tableArn-keyed Gets; only a name change was.
- [S3T-TABLE-304](#s3t-table-304) (low): the HTTP 500 once seen on a TagResource ~100 ms after two PutTableMaintenanceConfiguration calls on
  a table with an active table-level replication configuration did not reproduce on a plain table (0/50) - the
  replication-active context was never re-tested.
- [S3T-TABLEREPLICATION-011](#s3t-tablereplication-011) covers rules/destinations/role absent; the 'configuration' member absent or null is a 500 with
  no config created, recorded service-wide ([S3T-SERVICE-007](service.md#s3t-service-007), -020 in `service.md`), not in this document.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-304"></a>**S3T-TABLE-304** | error-code | low | confirmed | Single-writer back-to-back writes to one table (10 ordered pairs x 5, 0 s gap): 0 x 500; the 500 seen on a replication source not reproduced | [S3T-TABLE-186](table.md#s3t-table-186), [S3T-TABLE-191](table-maintenance.md#s3t-table-191), [S3T-TABLE-124](table-metadata-location.md#s3t-table-124), [S3T-TABLE-206](table-maintenance.md#s3t-table-206), [S3T-SERVICE-007](service.md#s3t-service-007), [S3T-SERVICE-020](service.md#s3t-service-020), [S3T-TABLE-300](table-replication-status.md#s3t-table-300) | - |
| <a id="s3t-tablereplication-034"></a>**S3T-TABLEREPLICATION-034** | other | low | confirmed | Doc claim C073 TRUE: GetTableReplicationStatus with neither table- nor bucket-level replication config -> NotFoundException 404, not 200 | [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-037](#s3t-tablereplication-037), [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026) | - |
| <a id="s3t-tablereplication-036"></a>**S3T-TABLEREPLICATION-036** | other | low | confirmed | Doc claim C075 TRUE: DeleteTableReplication with a stale/garbage/table versionToken -> ConflictException 409, config kept; empty -> 400 | [S3T-TABLEREPLICATION-007](#s3t-tablereplication-007), [S3T-TABLEREPLICATION-015](#s3t-tablereplication-015), [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEBUCKETREPLICATION-006](tablebucket-replication.md#s3t-tablebucketreplication-006), [S3T-TABLEBUCKETREPLICATION-004](tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008) | - |
| <a id="s3t-tablereplication-037"></a>**S3T-TABLEREPLICATION-037** | other | low | confirmed | Doc claim C070 TRUE: PutTableReplication in a bucket without bucket-level replication -> 200 {versionToken, status 'success'}; Get echoes it | [S3T-TABLEREPLICATION-003](#s3t-tablereplication-003), [S3T-TABLEREPLICATION-001](#s3t-tablereplication-001), [S3T-TABLEREPLICATION-021](table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-008](tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEBUCKETREPLICATION-005](tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLEREPLICATION-034](#s3t-tablereplication-034), [S3T-TABLEBUCKETREPLICATION-018](tablebucket-replication.md#s3t-tablebucketreplication-018), [S3T-TABLEBUCKETREPLICATION-001](tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-TABLEBUCKETREPLICATION-007](tablebucket-replication.md#s3t-tablebucketreplication-007) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
