<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3Tables: service facts and cross-cutting behaviors
_Model facts, resource inventory and scope verdicts, findings tagged `service-wide`, tag API, codegen notes, out-of-scope resources and the handling-gaps summary across all documents._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `service.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document holds the s3tables findings that are the same for every resource: the per-error-code catalogues, the
credential, signature and clock-skew failures, the IAM propagation windows and the throttling shape. The model has 49
operations (each catalogue counts its code against 49/49, [S3T-SERVICE-014](#s3t-service-014), -015, -018, -020, -022); the controller
implements three nouns as CRDs - TableBucket, Namespace and Table, each with `terminal_codes: [BadRequestException]`
and nothing else (generator.yaml) - and the seven sub-resource nouns it does not implement (TableBucketPolicy,
TablePolicy, TableMaintenanceConfiguration, TableBucketMaintenanceConfiguration, TableReplication,
TableBucketReplication, TableRecordExpirationConfiguration) render inside their parent-family documents (render.yaml).
The modeled error set is small and generic - NotFoundException 404, BadRequestException 400, ConflictException 409,
ForbiddenException and AccessDeniedException 403, TooManyRequestsException 429, InternalServerErrorException 500, and
MethodNotAllowedException 405 on the three record-expiration operations only ([S3T-SERVICE-017](#s3t-service-017)) - but the wire differs
from it in four ways that decide terminal-vs-requeue: a ValidationException 400 with no SDK type comes from 10
operations ([S3T-SERVICE-021](#s3t-service-021)), AccessDeniedException is declared on 15/49 operations yet produced by 49/49 and so is
absent from any model-generated terminal list ([S3T-SERVICE-019](#s3t-service-019)), one InternalServerErrorException text covers a
transient collision and three deterministic request errors ([S3T-SERVICE-020](#s3t-service-020)), and transport codes without an s3tables
shape (UnknownOperationException '404', UnrecognizedClientException, InvalidSignatureException,
MissingAuthenticationTokenException) reach the same error switch ([S3T-SERVICE-001](#s3t-service-001), -003, -023). The catalogues label
every message class RETRY, WAIT-FOR-STATE, PERMANENT-SPEC, PERMANENT-QUOTA, AUTHZ or AMBIGUOUS (one code+message,
several causes with different remediations); the rules below use the labels as the catalogues do ([S3T-SERVICE-014](#s3t-service-014)..023).

### Rules

1. Branch on the wire code and HTTP status and treat the message as diagnostic text only - several texts echo the
   caller's input or differ by a curly apostrophe; for the AMBIGUOUS classes below the catalogue names the confirming
   read (GetTable(tableArn), GetTableBucket on the constructed ARN) or the exact message to key on
   ([S3T-SERVICE-014](#s3t-service-014), -015).
2. InternalServerErrorException 500 "An internal error occurred. Try again." is one text with four causes - the single
   most dangerous code for a retry loop: (1) RETRY, write collisions on one resource (10 concurrent TagResource -> 8-9 x
   500, 10 concurrent PutTableBucketPolicy -> 3 x 500, 20 concurrent maintenance Puts -> 1 x 500; losing writes are not
   applied, sequential retries succeed); (2) UntagResource on a deleted or never-existing table ARN -> 500 every time
   while TagResource/ListTagsForResource on the same ARN are 404; (3) CreateTable with metadata.iceberg.schema={} -> 500
   4/4, no table; (4) Put*Replication with `configuration` absent -> 500 6/6, no config. Retry with backoff a bounded
   number of times (3 is enough for cause 1); if it repeats in a stable state surface it; for a finalizer untagging a
   resource that reads 404 elsewhere treat the 500 as gone ([S3T-SERVICE-020](#s3t-service-020), -007).
3. ValidationException (HTTP 400) is NOT in the s3tables model - no SDK type - and is produced by 10 operations for
   required/enum/length/pattern violations enforced by the framework: PutTableBucketMaintenanceConfiguration (status
   enum, type path, days >= 1, null value), PutTableBucketStorageClass (enum, null), Put*Policy (resourcePolicy empty, or
   > 20480 bytes = PERMANENT-QUOTA), TagResource/UntagResource (key 1..128, value <= 256), UpdateTableMetadataLocation
   (versionToken ''), Put*Replication (exactly one rule, 1..5 destinations, ARN/role patterns),
   PutTableRecordExpirationConfiguration (status enum, days >= 1); all 17 classes are PERMANENT-SPEC or PERMANENT-QUOTA
   and the same violations on other operations are BadRequestException - two layers, one class ([S3T-SERVICE-021](#s3t-service-021), -002).
4. ConflictException 409 is never produced by a read and means four things: the ~10 min name cooldown after
   DeleteTableBucket ("transitional state", WAIT-FOR-STATE, nothing accelerates it); "already exists" duplicates on
   Create* (PERMANENT-SPEC: adopt via the Get); a stale or malformed versionToken ("Provided version token does not
   match ..." / "does not match the configuration", WAIT-FOR-STATE: re-read the token and retry in the same reconcile -
   any malformed token is 409, never 400, and DeleteTableReplication wants the REPLICATION token, not the table's); and
   concurrent-write collisions ("modified by a concurrent request" on PutTableMaintenanceConfiguration/PutTablePolicy,
   "modified concurrently" on RenameTable, RETRY as-is). RenameTable to the CURRENT name is not a no-op but the
   "identical name already exists" 409 - skip RenameTable when nothing changes ([S3T-SERVICE-016](#s3t-service-016)).
5. NotFoundException 404 reads "The specified X does not exist." on 49/49 operations and 7 of its 9 classes are
   AMBIGUOUS: the same 404 covers gone, renamed out-of-band (old triple 404 while GetTable(tableArn) is 200 with the new
   name - confirm by ARN before recreating), parent gone (bucket or namespace missing for a child operation; namespaces
   are not auto-created), sub-resource unset (replication config, policy, metrics config - the normal absent state) and
   a stale token with no config (Put*Replication with a versionToken when none exists -> 404, not 409 - drop the token
   and re-Put). The one PERMANENT-SPEC class is DeleteTableBucketEncryption's non-idempotent 404 (swallow as success);
   the one WAIT-FOR-STATE class is RenameTable's "destination namespace does not exist" ([S3T-SERVICE-014](#s3t-service-014)).
6. BadRequestException 400 is not uniformly terminal: of 57 classes 45 are PERMANENT-SPEC, but 5 are WAIT-FOR-STATE
   ("... that you tried to delete is not empty.", "The table metadata location is not specified.", the sort-order
   message, "The specified KMS key ARN is disabled." - same text for Disabled and PendingDeletion), 1 is RETRY ("The
   continuation token is no longer valid." - restart the listing), 1 is AUTHZ ("Insufficient access to perform table
   maintenance ..." = KMS key policy lacks the maintenance principal), 1 is PERMANENT-QUOTA (50 user tags, atomic) and
   4 are AMBIGUOUS ("Invalid principal in policy", the deserializer catch-all "Invalid input ...", the GetTable
   identifier text returned for an empty replication configuration, "The specified KMS key ARN does not exist." for a
   key the caller cannot see) ([S3T-SERVICE-015](#s3t-service-015), -026).
7. "Invalid principal in policy" is transient when the principal is seconds old: PutTableBucketPolicy/PutTablePolicy
   naming a role from CreateRole in the same pass returned 13-14 consecutive 400s at 0.5 s polling and the first 200 at
   7.22-7.93 s (n=3); sts:AssumeRole on the same role first succeeded at 8.1-8.75 s - requeue on this message for ~10 s
   before treating it as the permanent typo/deleted-principal case ([S3T-SERVICE-010](#s3t-service-010)).
8. Two 403 codes, opposite handling. AccessDeniedException is AUTHZ in all 9 classes (one template, five reason
   suffixes: identity explicit deny, no identity policy, resource-policy deny, session-policy deny, no permissions
   boundary); the denied ACTION may differ from the operation (Create* with encryption/storage class/tags is denied on
   the sub-resource Put action, Put*Replication on iam:PassRole); it is produced by 49/49 operations although declared
   on 15 and is what every call returns during the ~86-100 s IAM propagation window - it must stay retryable
   ([S3T-SERVICE-019](#s3t-service-019), -011, -012). ForbiddenException never heals by retrying unchanged: "Unauthorized" for any
   well-formed ARN whose account id is not the caller's (existence never disclosed - an account-id typo in a referenced
   ARN looks like a permission problem), two pass-through texts needing an external fix (KMS kms:DescribeKey key policy;
   S3-style "Access Denied" for CreateTable WITH metadata by a caller lacking s3tables:PutTableData), and "Cross-account
   pass role is not allowed." on PutTableReplication - a PERMANENT-SPEC error behind a 403, checked before the rules
   validation ([S3T-SERVICE-022](#s3t-service-022), -026).
9. Check order: ARN region/partition/name/syntax BadRequestException 400 precede authorization (a role denied by its
   permissions boundary still gets 400 for an other-region ARN, an invalid name inside the ARN or a non-ARN, 403
   otherwise); ValidationException is evaluated before the table lookup and the 405 gate; MethodNotAllowedException 405
   comes from exactly the 3 record-expiration operations as a scope signal (table type) ordered after shape validation,
   authz and lookup - so neither a 400 nor a 405 says anything about permissions or existence
   ([S3T-SERVICE-011](#s3t-service-011), -021, -017).
10. A raw HTTP 404 with body `<UnknownOperationException/>` (XML, no x-amzn-ErrorType) is a routing miss for an unknown
    path, wrong method or extra sub-resource segment (HEAD: 404 with an empty body); the SDK surfaces code '404', and an
    error-code switch must not read it as NotFoundException ([S3T-SERVICE-001](#s3t-service-001), -023).
11. Credential and signature failures are HTTP 403 with non-s3tables codes: UnrecognizedClientException for a bogus
    access key or a session token garbled in its head or middle, truncated or empty (tail edits - last char flipped,
    last 4 -> AAAA, one char appended - still 200), InvalidSignatureException for a wrong secret or a request signed for
    another region or for service 's3', MissingAuthenticationTokenException for an unsigned request; none is an s3tables
    permission problem, and only UnrecognizedClientException has a transient case - credentials issued ~2 s after a
    fresh role's first AssumeRole OK, see rule 15 ([S3T-SERVICE-003](#s3t-service-003), -006, -023, -025).
12. SigV4 clock-skew tolerance is +/-5 min, tighter than the 15 min most AWS services allow: X-Amz-Date shifted
    -4.5/+4.5 min is accepted, -5.5/-6/+5.5/+6 min is InvalidSignatureException 403 "Signature expired" / "Signature not
    yet current" with the window stated in the message; botocore does not auto-correct skew for this code
    ([S3T-SERVICE-013](#s3t-service-013), -023).
13. TooManyRequestsException 429 is RETRY with no server hint: no Retry-After or rate-limit header (only content-length,
    content-type, date, x-amzn-errortype, x-amzn-requestid), body 'Too Many Requests'; produced by 6 operations (all
    writes except one metrics Get) with no side effects; the CreateNamespace/DeleteNamespace budget is shared across
    buckets (~16 per burst, account/principal-wide: 12/12 OK into bucket A, then 429 from the 5th call into bucket B;
    no 429'd create or delete left partial state) ([S3T-SERVICE-018](#s3t-service-018), -004, -008).
14. An empty body `{}` on the 16 write operations hits two validation layers and three odd outcomes: 6 ValidationException
    (framework), 7 BadRequestException (handler), PutTableBucketMetricsConfiguration {} -> HTTP 204, Put*Replication {}
    -> 500; stock SDKs reject these client-side, so they are reachable only through raw requests or a nil pointer that
    serializes as absent ([S3T-SERVICE-002](#s3t-service-002), -007, -020). Likewise maxTables/maxNamespaces/maxBuckets=0 is stopped by the
    client before the server's BadRequestException "limit ... positive number", while values above the modeled maximum
    1001 are accepted with 200 ([S3T-SERVICE-009](#s3t-service-009)).
15. IAM changes propagate slowly and uniformly: an identity-policy Allow on an assumed role became visible to s3tables
    86.1-100.8 s after PutRolePolicy and a revoke 95.3-102.0 s after DeleteRolePolicy (n=5 over three runs, 1 s polls,
    both flips stable for the next 5 s), with s3 ListBuckets flipping in the same second as s3tables GetTableBucket - an
    IAM-wide delay, not an s3tables cache; a fresh role is assumable 8.1-9.08 s after CreateRole (n=4), a role re-created
    under a just-deleted name flapped OK -> AccessDenied for 2-3 calls -> OK, and credentials issued ~2 s after that
    first OK were rejected once by s3tables with UnrecognizedClientException and accepted 10 s later
    ([S3T-SERVICE-012](#s3t-service-012), -027, -028, -025).

### Timing ranges

- Put*Policy with a just-created role as Principal: "Invalid principal in policy" 13-14 times at 0.5 s polling, first
  HTTP 200 at 7.22-7.93 s after CreateRole, n=3 ([S3T-SERVICE-010](#s3t-service-010)).
- sts:AssumeRole on a fresh role: first success 8.1-8.75 s, n=3 ([S3T-SERVICE-010](#s3t-service-010)); 8.1-9.08 s, n=4 ([S3T-SERVICE-025](#s3t-service-025)).
  Flap after the first OK: 6 consecutive OK at 1 s polls, AccessDenied on the next 2-3 calls, then OK (one re-created
  role name); first STS credentials of that role rejected once by s3tables with UnrecognizedClientException ~2 s after
  the OK, accepted 10 s later with fresh credentials (n=1, [S3T-SERVICE-025](#s3t-service-025)).
- IAM identity-policy Allow visible to s3tables: 86.13 s, n=1 ([S3T-SERVICE-012](#s3t-service-012)); 86.1-95.2 s, n=2 ([S3T-SERVICE-028](#s3t-service-028));
  99.8-100.8 s, n=2 ([S3T-SERVICE-027](#s3t-service-027)) - 86.1-100.8 s over n=5 at 1 s polls. Revoke visible: 100.5 s, n=1 (-012);
  95.3-95.4 s, n=2 (-028); 101.8-102.0 s, n=2 (-027) - 95.3-102.0 s over n=5; both flips stable for the next 5 s in
  every cycle (-028); s3 ListBuckets in the same loop flipped at the same seconds, n=2 (-027).
- CreateNamespace burst into one bucket: 16 x 200 before the first 429 at 0.51 s, 8 x 429 in 24 calls (n=1,
  [S3T-SERVICE-004](#s3t-service-004)). Two buckets back-to-back, 0.6 s wall: A 12/12 x 200, B 4 x 200 then 429 from the 5th call; delete
  burst 0.4 s wall: A 12/12 x 204, B 4 x 204 then 429 (n=1, [S3T-SERVICE-008](#s3t-service-008)).
- SigV4 skew: accepted at -4.5 and +4.5 min, rejected at -5.5, -6, +5.5 and +6 min (6 shifted requests, one run,
  [S3T-SERVICE-013](#s3t-service-013)); the message quotes the window as (now - 5 min.) / (now + 5 min.).
- Deterministic 500s: Put*Replication with `configuration` absent 6/6 ([S3T-SERVICE-007](#s3t-service-007)); CreateTable schema={} 4/4
  ([S3T-SERVICE-020](#s3t-service-020), citing [S3T-TABLE-036](table-schema.md#s3t-table-036)/-133). Collision 500 rates: 8-9 of 10 concurrent TagResource, 3 of 10 concurrent
  PutTableBucketPolicy, 1 of 20 concurrent maintenance Puts ([S3T-SERVICE-020](#s3t-service-020), citing [S3T-TABLE-186](table.md#s3t-table-186),
  [S3T-TABLEBUCKET-070](tablebucket.md#s3t-tablebucket-070)/-082/-078).
- CreateTableBucket same-name cooldown after a delete: ~10 min per the ConflictException catalogue ([S3T-SERVICE-016](#s3t-service-016),
  citing [S3T-TABLEBUCKET-047](tablebucket.md#s3t-tablebucket-047)/-016/-115; the per-run range with n lives in tablebucket.md).
- Not measured here: the refill time of a 429 budget ([S3T-SERVICE-004](#s3t-service-004), -008) and the IAM delay of a resource policy as
  opposed to an identity policy ([S3T-SERVICE-019](#s3t-service-019) defers to table-policy.md / tablebucket-policy.md).

### Handling gaps

The controller's only error classification is `terminal_codes: [BadRequestException]` on each of TableBucket, Namespace
and Table (generator.yaml); every other code is Recoverable and requeued - which is also why AccessDeniedException,
produced by 49/49 operations but declared on 15, needs no entry ([S3T-SERVICE-019](#s3t-service-019), handled) while the propagation
findings [S3T-SERVICE-027](#s3t-service-027)/-028 are `unhandled` (nothing waits or surfaces a condition during the window). Four findings
here are confirmed controller bugs against that setting (rendered as known defects, not hidden):

- [S3T-SERVICE-010](#s3t-service-010) (suspect-bug, high) - today every Put*Policy BadRequestException is Terminal; the evidence shows
  "Invalid principal in policy" is transient for 7.22-7.93 s after CreateRole (n=3, 13-14 x 400 then 200) and only
  afterwards a permanent spec error, so a TableBucket or Table whose policy names a role created in the same pass is
  parked in ACK.Terminal and never retried until the spec changes. Not yet handled ([GT-S3T-010](#gt-s3t-010), [GT-S3T-039](#gt-s3t-039) (controller hooks catalog entries)).
- [S3T-SERVICE-015](#s3t-service-015) (suspect-bug, high) - today BadRequestException is terminal on every path; the catalogue shows 5
  WAIT-FOR-STATE classes (not-empty parent delete, missing metadata location, missing sort order, disabled KMS key), 1
  RETRY (continuation token), 1 AUTHZ (KMS key policy) and 4 AMBIGUOUS classes that heal without a spec change, so
  CreateTable, DeleteTableBucket/DeleteNamespace, PutTableBucketEncryption and maintenance updates are stranded in
  ACK.Terminal where a requeue would have succeeded. Not yet handled ([GT-S3T-022](#gt-s3t-022), [GT-S3T-008](#gt-s3t-008), [GT-S3T-039](#gt-s3t-039) (controller hooks catalog entries)).
- [S3T-SERVICE-021](#s3t-service-021) (suspect-bug, medium) - today ValidationException is not in terminal_codes (no SDK type) and is
  requeued as Recoverable; the evidence shows all 17 classes are permanent spec or quota errors on 10 operations (tag
  key/value lengths, resourcePolicy empty or > 20480 bytes, storage class enum, bucket maintenance enum/days, replication
  rule/destination bounds), so a permanent spec error loops forever instead of setting ACK.Terminal, while the same
  violation at Create (BadRequestException) is terminal. Not yet handled ([GT-S3T-005](#gt-s3t-005), [GT-S3T-019](#gt-s3t-019), [GT-S3T-039](#gt-s3t-039) (controller hooks catalog entries)).
- [S3T-SERVICE-022](#s3t-service-022) (suspect-bug, medium) - today ForbiddenException is requeued indefinitely with no condition naming the
  cause; the evidence shows it never heals by retrying unchanged - "Unauthorized" (account id in a referenced ARN is not
  the caller's) and "Cross-account pass role is not allowed." are spec errors, the KMS key-policy and
  s3tables:PutTableData texts need an external fix; the finding's rule: treat the first two as terminal, surface the
  others as a Recoverable condition, keep AccessDeniedException retryable (~86-100 s after IAM changes,
  [S3T-SERVICE-019](#s3t-service-019)). Not yet handled ([GT-S3T-039](#gt-s3t-039) (controller hooks catalog entry)).

### Where next

The 13 per-resource documents (render.yaml) carry the behavior this document only classifies:

- `tablebucket.md` - lifecycle, deterministic ARN, ~10 min post-delete name reservation (measured), not-empty delete 400.
- `tablebucket-policy.md` - Put/Get/DeleteTableBucketPolicy, principal propagation, first-sync deletion hazard; scope.
- `tablebucket-encryption-storage.md` - KMS key policy/key-state prerequisites, caller permissions, Validation-vs-400.
- `tablebucket-maintenance-metrics.md` - maintenance merge-patch, int32 clamping, ValidationException on days < 1.
- `tablebucket-replication.md` - versionToken locking, role prerequisites, ~60 s status evaluation, replica re-creation.
- `namespace.md` - minimal Create echo, naming rules, not-empty delete 400, ~16 req/s limiter, no tags or ARN.
- `table.md` - lookup by ARN or triple, rename, 404 message shapes, tag ValidationException, UntagResource 500.
- `table-metadata-schema.md` - metadata.iceberg (schema={} -> 500), UpdateTableMetadataLocation, versionToken, 409 CAS.
- `table-encryption-storage.md` - create-only encryption/storage class, snapshot of bucket defaults, KMS on CreateTable.
- `table-policy.md` - Put/Get/DeleteTablePolicy, IAM action per table op (PutTableData, GetTableData, kms:DescribeKey).
- `table-maintenance.md` - merge-patch, defaults, metadata-location/sort-order prerequisites, KMS maintenance principal.
- `table-replication.md` - config vs table versionToken, per-destination status, ~60 s evaluation, orphan replicas.
- `table-record-expiration.md` - 405 on every creatable table, check order before the 405 gate; not implementable today.

The catalogues' full class tables live on `details/<id>.md` pages (render.yaml `notes_policy`: 300-char inline excerpt,
full notes in `details/`): [S3T-SERVICE-014](#s3t-service-014) (404, 9 classes), -015 (400, 57), -016 (409, 8), -017 (405, 1), -018 (429,
1), -019 (AccessDeniedException, 9), -020 (500, 1 text / 4 causes), -021 (ValidationException, 17), -022
(ForbiddenException, 4), -023 (sibling-service and transport codes, 11); each row names the operations, HTTP status,
classification and controller rule and cites the per-resource finding that produced it.

Open items from STATUS.md relevant to this stem: the undeletable leftover table bucket `ackq-9ba36e-polnorm` (us-west-2,
namespace `ackq_9ba36e_ns`, table `ackq_9ba36e_t`) carries a Deny+NotAction bucket AND table policy for the account
root principal, so every IAM principal gets AccessDeniedException on all operations except GetTableBucket, including
Delete* and DeleteTableBucketPolicy - no owner override exists ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007); needs root / AWS Support);
8 KMS keys from this run are in PendingDeletion until 2026-10-20/21. Headline numbers (scorecard.md): 44 ground-truth
entries, 37 found + 7 partial, 0 missed; 12 suspected controller bugs confirmed, of which the four above are the
service-wide ones; Phase 8 re-verification confirmed 22 sampled findings and corrected 3 outside this stem. Every
finding in this document was observed in us-west-2 only; `service.open-questions.md` lists what a second region or a
repeat run should settle.
<!-- preserved:end -->

## Service facts

- API version: 2018-05-10 · SDK module: service/s3tables v1.16.4 · model fingerprint: 70ac978561f0cee0
- controller: aws-controllers-k8s/s3tables-controller @ 1eda357
- regions probed: us-west-2, us-east-1
- probes: 165 · findings: 716 (695 canonical, 21 duplicates) · service-wide findings: 26

## Resource inventory

| noun | ack_inferred | create | read | update | delete | list | sub_resource_ops | scope verdict |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Namespace | yes | CreateNamespace | GetNamespace | - | DeleteNamespace | ListNamespaces | - | implemented |
| Table | yes | CreateTable | GetTable | RenameTable | DeleteTable | ListTables | DeleteTableBucket, DeleteTableBucketEncryption, DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy, DeleteTableBucketReplication, DeleteTablePolicy, DeleteTableReplication, GetTableBucket, GetTableBucketEncryption, GetTableBucketMaintenanceConfiguration, GetTableBucketMetricsConfiguration, GetTableBucketPolicy, GetTableBucketReplication, GetTableBucketStorageClass, GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetTableMetadataLocation, GetTablePolicy, GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus, GetTableReplication, GetTableReplicationStatus, GetTableStorageClass, UpdateTableMetadataLocation | defer, field-on-parent, implemented |
| TableBucket | yes | CreateTableBucket | GetTableBucket | - | DeleteTableBucket | ListTableBuckets | DeleteTableBucketEncryption, DeleteTableBucketMetricsConfiguration, DeleteTableBucketPolicy, DeleteTableBucketReplication, GetTableBucketEncryption, GetTableBucketMaintenanceConfiguration, GetTableBucketMetricsConfiguration, GetTableBucketPolicy, GetTableBucketReplication, GetTableBucketStorageClass | implemented |
| TableBucketEncryption | no | - | GetTableBucketEncryption | - | DeleteTableBucketEncryption | - | - | - |
| TableBucketMaintenanceConfiguration | no | - | GetTableBucketMaintenanceConfiguration | - | - | - | - | - |
| TableBucketMetricsConfiguration | no | - | GetTableBucketMetricsConfiguration | - | DeleteTableBucketMetricsConfiguration | - | - | - |
| TableBucketPolicy | no | - | GetTableBucketPolicy | - | DeleteTableBucketPolicy | - | - | - |
| TableBucketReplication | no | - | GetTableBucketReplication | - | DeleteTableBucketReplication | - | - | - |
| TableBucketStorageClass | no | - | GetTableBucketStorageClass | - | - | - | - | - |
| TableMaintenanceConfiguration | no | - | GetTableMaintenanceConfiguration | - | - | - | - | - |
| TableMetadataLocation | no | - | GetTableMetadataLocation | UpdateTableMetadataLocation | - | - | - | - |
| TablePolicy | no | - | GetTablePolicy | - | DeleteTablePolicy | - | - | - |
| TableRecordExpirationConfiguration | no | - | GetTableRecordExpirationConfiguration | - | - | - | - | defer |
| TableReplication | no | - | GetTableReplication | - | DeleteTableReplication | - | GetTableReplicationStatus | - |

## Cross-cutting behaviors

Findings tagged `service-wide` (resource-specific findings live in the per-resource documents listed in
README.md).

### Errors

- <a id="s3t-service-002"></a>**S3T-SERVICE-002** `error-code` · impact medium · handled · verified 2026-10-14
  **Empty body {} on 16 write ops: 6 ValidationException (framework), 7 BadRequestException (handler), 3 other - two validation layers**
  Signed raw request with body '{}' and no other members: CreateTableBucket -> HTTP 400 BadRequestException
  "{\"message\":\"The specified bucket name isn't valid. Specify a different bucket name, and then try
  again.\"}"; CreateNamespace -> HTTP 400 BadRequestException "{\"message\":\"The specified namespace name
  isn\u2019t valid. Specify a different namespace name, and then try again.\"; CreateTable -> HTTP 400
  BadRequestException "{\"message\":\"The specified open table format is not valid\"}"; PutTableBucketPolicy
  -> HTTP 400 ValidationException "{\"message\":\"1 validation error detected: Value at 'resourcePolicy'
  failed to satisfy constraint: Member must not be ; PutTableBucketEncryption -> HTTP 400 BadRequestException
  "{\"message\":\"An encryption configuration is not specified\"}"; PutTableBucketStorageClass -> HTTP 400
  ValidationException "{\"message\":\"1 validation error detected: Value at 'storageClassConfiguration' failed
  to satisfy constraint: Member m; PutTableBucketMaintenanceConfiguration -> HTTP 400 ValidationException
  "{\"message\":\"1 validation error detected: Value at 'value' failed to satisfy constraint: Member must not
  be null\"}"; PutTableBucketMetricsConfiguration -> HTTP 204 - ""; PutTableBucketReplication -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableMaintenanceConfiguration -> HTTP 400 BadRequestException "{\"message\":\"The table maintenance
  configuration value is not specified.\"}"; UpdateTableMetadataLocation -> HTTP 400 ValidationException
  "{\"message\":\"2 validation errors detected: Value at 'versionToken' failed to satisfy constraint: Member
  must not be n; RenameTable -> HTTP 400 BadRequestException "{\"message\":\"Neither a new namespace name nor
  a new table name is specified.\"}"; PutTablePolicy -> HTTP 400 ValidationException "{\"message\":\"1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must not be ;
  PutTableReplication -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred. Try
  again.\"}"; PutTableRecordExpirationConfiguration -> HTTP 400 ValidationException "{\"message\":\"1
  validation error detected: Value at 'value' failed to satisfy constraint: Member must not be null\"}";
  TagResource -> HTTP 400 BadRequestException "{\"message\":\"The provided tags are invalid.\"}".
  - ACK: terminal_codes, docs-only · ops: CreateTableBucket, CreateNamespace, CreateTable,
    PutTableBucketPolicy, PutTableBucketEncryption, PutTableBucketStorageClass,
    PutTableBucketMaintenanceConfiguration, PutTableBucketMetricsConfiguration, PutTableBucketReplication,
    PutTableMaintenanceConfiguration, UpdateTableMetadataLocation, RenameTable, PutTablePolicy,
    PutTableReplication, PutTableRecordExpirationConfiguration, TagResource
  - repro: For each write op: SigV4 PUT/POST <modeled path> with body '{}' (content-type application/json)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-007](#s3t-service-007), [S3T-SERVICE-020](#s3t-service-020) · evidence: service/creative/raw-wire-codes
  - notes: ValidationException ops: ['PutTableBucketMaintenanceConfiguration', 'PutTableBucketPolicy',
    'PutTableBucketStorageClass', 'PutTablePolicy', 'PutTableRecordExpirationConfiguration',
    'UpdateTableMetadataLocation']. BadRequestException ops: ['CreateNamespace', 'CreateTable',
    'CreateTableBucket',...
  - full notes: [details/S3T-SERVICE-002.md](details/S3T-SERVICE-002.md)

- <a id="s3t-service-003"></a>**S3T-SERVICE-003** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GET /buckets credential failures (403): UnrecognizedClientException (bogus key), InvalidSignatureException (secret/region/service), unsigned**
  bogus access key -> HTTP 403 UnrecognizedClientException "{\"message\":\"The security token included in the
  request is invalid.\"}"; valid key, wrong secret -> HTTP 403 InvalidSignatureException "{\"message\":\"The
  request signature we calculated does not match the signature you provided. Check your AWS Secret Acc; valid
  key+secret, session token garbled -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  signed for us-east-1 -> HTTP 403 InvalidSignatureException "{\"message\":\"Credential should be scoped to a
  valid region. \"}"; signed for service s3 -> HTTP 403 InvalidSignatureException "{\"message\":\"Credential
  should be scoped to correct service: 's3tables'. \"}"; clock skew -20 min -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  clock skew -4 min (within tolerance?) -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  no Authorization header -> HTTP 403 MissingAuthenticationTokenException "{\"message\":\"Missing
  Authentication Token\"}".
  - ACK: terminal_codes, requeue · ops: ListTableBuckets
  - repro: SigV4-sign GET /buckets with a bogus key / wrong secret / wrong region / wrong service / skewed
    X-Amz-Date / no Authorization
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-005](#s3t-service-005), [S3T-SERVICE-013](#s3t-service-013), [S3T-SERVICE-006](#s3t-service-006), [S3T-SERVICE-025](#s3t-service-025), [S3T-SERVICE-023](#s3t-service-023) · evidence:
    service/creative/raw-wire-codes
  - notes: Creative round, seed error-messages, gap 'UnrecognizedClientException and sibling codes'. All are
    HTTP 403 (or 400) with no s3tables shape; none is a permission problem: they mean the controller's
    credentials/endpoint configuration is wrong (or the pod clock is off) and recur until that is fixed -...
  - full notes: [details/S3T-SERVICE-003.md](details/S3T-SERVICE-003.md)

- <a id="s3t-service-007"></a>**S3T-SERVICE-007** `error-code` · impact medium · handled · verified 2026-10-14
  **PutTableReplication/PutTableBucketReplication with body {} -> InternalServerErrorException 500 in 6/6 sends (deterministic, no side effect)**
  Raw signed PUT with body '{}' (no 'configuration' member): PutTableReplication {} #1 -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableBucketReplication {} #1 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error
  occurred. Try again.\"}"; PutTableReplication {} #2 -> HTTP 500 InternalServerErrorException
  "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableBucketReplication {} #2 -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableReplication
  {} #3 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableBucketReplication {} #3 -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error
  occurred. Try again.\"}"; PutTableReplication {configuration:{}} -> HTTP 400 BadRequestException
  "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace,
  tab; PutTableReplication {configuration:null} -> HTTP 500 InternalServerErrorException "{\"message\":\"An
  internal error occurred. Try again.\"}"; PutTableReplication {configuration:{rules:[]}} -> HTTP 400
  BadRequestException "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table
  bucket arn, namespace, tab; PutTableReplication {configuration:{role:x}} (rules omitted) -> HTTP 403
  ForbiddenException "{\"message\":\"Cross-account pass role is not allowed.\"}"; PutTableReplication {} no
  tableArn query -> HTTP 400 BadRequestException "{\"message\":\"A table ARN is not specified\"}";
  PutTableReplication {} ghost table uuid -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal
  error occurred. Try again.\"}"; PutTableReplication {} bucket ARN as tableArn -> HTTP 400
  BadRequestException "{\"message\":\"The specified table ARN is not valid.\"}"; PutTableReplication
  {versionToken:abc} only -> HTTP 500 InternalServerErrorException "{\"message\":\"An internal error occurred.
  Try again.\"}"; PutTableBucketReplication {configuration:{}} -> HTTP 400 BadRequestException
  "{\"message\":\"Invalid input: Must provide either table arn alone OR all of (table bucket arn, namespace,
  tab; PutTableBucketReplication {configuration:null} -> HTTP 500 InternalServerErrorException
  "{\"message\":\"An internal error occurred. Try again.\"}"; PutTableBucketReplication {} no query -> HTTP
  400 BadRequestException "{\"message\":\"The specified table bucket ARN is not valid.\"}";
  PutTableBucketReplication {} ghost bucket -> HTTP 500 InternalServerErrorException "{\"message\":\"An
  internal error occurred. Try again.\"}"; PutTableBucketReplication {} table ARN as bucket -> HTTP 400
  BadRequestException "{\"message\":\"The specified bucket name isn't valid. Specify a different bucket name,
  and then try again.\"}; PutTableBucketReplication {versionToken:abc} only -> HTTP 500
  InternalServerErrorException "{\"message\":\"An internal error occurred. Try again.\"}";
  PutTableRecordExpirationConfiguration {} (control, ValidationException expected) -> HTTP 400
  ValidationException "{\"message\":\"1 validation error detected: Value at 'value' failed to satisfy
  constraint: Member must not be. Side effects: GetTableReplication -> NotFoundException,
  GetTableBucketReplication -> NotFoundException, GetTable -> 200.
  - ACK: terminal_codes, requeue · ops: PutTableReplication, PutTableBucketReplication · fields: configuration
  - repro: SigV4 PUT /table-replication?tableArn=<arn> body '{}' x3; PUT
    /table-bucket-replication?tableBucketARN=<arn> body '{}' x3; variants with configuration {} / null
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-002](#s3t-service-002), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-026](#s3t-service-026), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003),
    [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-021](#s3t-service-021), [S3T-TABLE-300](table-replication-status.md#s3t-table-300), [S3T-TABLE-304](table-replication.md#s3t-table-304) · evidence:
    service/creative/sig-skew-and-empty-500
  - notes: Creative round, seed error-messages: a request-shape error (missing required 'configuration', which
    the SDK would reject client-side) surfaces as a 500 'An internal error occurred. Try again.' -
    PERMANENT-SPEC behind a RETRY code. Variant bodies in behavior show which member triggers it...
  - full notes: [details/S3T-SERVICE-007.md](details/S3T-SERVICE-007.md)

- <a id="s3t-service-011"></a>**S3T-SERVICE-011** `error-code` · impact medium · handled · verified 2026-10-14
  **Permissions-boundary deny -> AccessDeniedException 403 on Get/List/Create; ARN region/name/syntax BadRequestException 400 precede authz**
  Role with inline Allow s3tables:* but PermissionsBoundary=AmazonS3ReadOnlyAccess (no s3tables):
  GetTableBucket(existing) -> AccessDeniedException/403 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL>
  is not authorized to perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:'; GetTableBucket(ghost)
  -> AccessDeniedException/403 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:'; GetTableBucket(other-region ARN) ->
  BadRequestException/400 'The provided Region does not match the Region of the resource.';
  GetTableBucket(invalid name in ARN) -> BadRequestException/400 "The specified bucket name isn't valid.
  Specify a different bucket name, and then try again."; GetTableBucket(not an ARN) -> BadRequestException/400
  'The specified table bucket ARN is not valid.'; ListTableBuckets -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:ListTableBuckets on
  resource: arn:aws:s3table'; GetTable(tableArn) -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-wes'; CreateNamespace -> AccessDeniedException/403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:CreateNamespace on
  resource: arn:aws:s3tables'.
  - ACK: terminal_codes, requeue · ops: GetTableBucket, ListTableBuckets, GetTable, CreateNamespace · fields:
    tableBucketARN
  - repro: CreateRole(PermissionsBoundary=AmazonS3ReadOnlyAccess) + inline Allow s3tables:*; AssumeRole;
    GetTableBucket with existing/ghost/other-region/invalid/non-ARN values
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-024](#s3t-service-024), [S3T-SERVICE-027](#s3t-service-027), [S3T-SERVICE-012](#s3t-service-012), [S3T-TABLE-234](table-maintenance.md#s3t-table-234), [S3T-TABLEBUCKET-127](tablebucket.md#s3t-tablebucket-127),
    [S3T-TABLEBUCKET-022](tablebucket.md#s3t-tablebucket-022) · evidence: service/creative/iam-authz-variants
  - notes: Creative round, seed error-messages: fills the AccessDeniedException message-variant gap (boundary)
    and fixes the check order for the catalogue: which of ARN-syntax 400 / region 400 / authz 403 / lookup 404
    wins is in behavior. AUTHZ class regardless of the reason suffix.
  - full notes: [details/S3T-SERVICE-011.md](details/S3T-SERVICE-011.md)

- <a id="s3t-service-012"></a>**S3T-SERVICE-012** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **Role without s3tables permissions: 15/15 never-denied ops -> AccessDeniedException 403; IAM allow visible after 86.1 s, revoke after 100.5 s**
  Assumed role with no identity policy: AccessDeniedException/403 'User: <ARN> is not authorized to perform:
  s3tables:<Op> on resource: <ARN> because no identity-based policy allows the s3tables:<Op> action' on
  ['DeleteTableBucketEncryption', 'DeleteTableBucketMetricsConfiguration', 'DeleteTableBucketReplication',
  'DeleteTableReplication', 'GetTableBucket (control)', 'GetTableBucketMetricsConfiguration',
  'GetTableBucketReplication', 'GetTableBucketStorageClass', 'GetTableMaintenanceJobStatus',
  'GetTableReplication', 'GetTableReplicationStatus', 'PutTableBucketEncryption',
  'PutTableBucketMaintenanceConfiguration', 'PutTableBucketMetricsConfiguration',
  'PutTableBucketStorageClass']. Exceptions: none. Then PutRolePolicy Allow s3tables:GetTableBucket:
  GetTableBucket codes by second ['AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException', 'AccessDeniedException',
  'AccessDeniedException', 'AccessDeniedException', 'OK'] (first 200 at 86.13 s); DeleteRolePolicy: codes
  ['OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK',
  'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'OK', 'AccessDeniedException'] (first AccessDeniedException
  at 100.5 s).
  - ACK: terminal_codes, requeue · ops: GetTableBucketStorageClass, GetTableBucketReplication,
    GetTableBucketMetricsConfiguration, PutTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration, PutTableBucketEncryption, DeleteTableBucketEncryption,
    PutTableBucketStorageClass, PutTableBucketMaintenanceConfiguration, DeleteTableBucketReplication,
    GetTableReplication, GetTableReplicationStatus, DeleteTableReplication, GetTableMaintenanceJobStatus,
    GetTableBucket (control)
  - repro: CreateRole (no policy) -> AssumeRole -> call each op; PutRolePolicy Allow GetTableBucket and poll
    at 1 s; DeleteRolePolicy and poll at 1 s
  - measurements: allow_visible_s=86.13, revoke_visible_s=100.5, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-010](#s3t-service-010), [S3T-SERVICE-025](#s3t-service-025), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](table-policy.md#s3t-table-164),
    [S3T-SERVICE-024](#s3t-service-024), [S3T-SERVICE-027](#s3t-service-027), [S3T-SERVICE-011](#s3t-service-011), [S3T-TABLE-234](table-maintenance.md#s3t-table-234) · evidence:
    service/creative/iam-authz-variants
  - notes: Creative round, seed error-messages: AccessDeniedException is declared on only 15 of 49 operations
    in the model yet is reachable on every one (the structured probes saw ~35; this adds the remaining ones
    listed in operations). It is AUTHZ but can be transient for seconds after an IAM change...
  - full notes: [details/S3T-SERVICE-012.md](details/S3T-SERVICE-012.md)

- <a id="s3t-service-014"></a>**S3T-SERVICE-014** `error-code` · impact high · handled · verified 2026-10-14
  **Error catalogue NotFoundException (HTTP 404): 9 message classes - 7 AMBIGUOUS, 1 PERMANENT-SPEC, 1 WAIT-FOR-STATE**
  NotFoundException: declared on 49/49 operations; 1407 occurrences. 9 message classes; classification counts
  {'AMBIGUOUS': 7, 'PERMANENT-SPEC': 1, 'WAIT-FOR-STATE': 1}. Every message is 'The specified X does not
  exist.' except the metrics one; 7 of 9 are AMBIGUOUS because the same 404 covers 'gone', 'renamed', 'parent
  gone', 'sub-resource unset' and 'stale token with no config'.
  - ACK: terminal_codes, exceptions.404, requeue · ops: CreateNamespace, CreateTable, DeleteNamespace,
    DeleteTable, DeleteTableBucket, DeleteTableBucketEncryption, DeleteTableBucketReplication,
    DeleteTablePolicy, DeleteTableReplication, GetNamespace, GetTable, GetTableBucket,
    GetTableBucketMetricsConfiguration, GetTableBucketPolicy, GetTableBucketReplication, GetTableEncryption,
    GetTableMaintenanceJobStatus, GetTableMetadataLocation, GetTablePolicy, GetTableReplication,
    GetTableReplicationStatus, GetTableStorageClass, ListNamespaces, ListTables, ListTagsForResource,
    PutTableBucketReplication, PutTableMaintenanceConfiguration, PutTableReplication, RenameTable,
    TagResource, UntagResource, UpdateTableMetadataLocation
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-016](#s3t-service-016), [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-021](#s3t-service-021),
    [S3T-SERVICE-022](#s3t-service-022), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-098](tablebucket-replication.md#s3t-tablebucket-098) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-014.md](details/S3T-SERVICE-014.md)

- <a id="s3t-service-015"></a>**S3T-SERVICE-015** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Error catalogue BadRequestException (400): 57 message classes - 45 PERMANENT-SPEC, 5 WAIT-FOR-STATE, 4 AMBIGUOUS, 1 AUTHZ, 1 QUOTA, 1 RETRY**
  BadRequestException: declared on 49/49 operations; 988 occurrences, 60 message classes in the seed + 3 new.
  57 message classes; classification counts {'AMBIGUOUS': 4, 'AUTHZ': 1, 'PERMANENT-QUOTA': 1,
  'PERMANENT-SPEC': 45, 'RETRY': 1, 'WAIT-FOR-STATE': 5}. Mostly PERMANENT-SPEC, but 5 classes are
  state-dependent (not-empty deletes, missing metadata/sort order, disabled KMS key), 1 is RETRY (continuation
  token), 1 is a KMS key-policy AUTHZ problem and 4 are AMBIGUOUS (fresh-principal propagation, catch-all
  'Invalid input', identifier message for an empty configuration, invisible foreign key).
  - ACK: terminal_codes, docs-only · ops: CreateNamespace, CreateTable, CreateTableBucket, DeleteNamespace,
    DeleteTableBucket, DeleteTableBucketReplication, DeleteTableReplication, GetNamespace, GetTable,
    GetTableBucket, GetTableReplication, ListNamespaces, ListTableBuckets, ListTables, ListTagsForResource,
    PutTableBucketEncryption, PutTableBucketMetricsConfiguration, PutTableBucketPolicy,
    PutTableBucketReplication, PutTableMaintenanceConfiguration, PutTablePolicy,
    PutTableRecordExpirationConfiguration, PutTableReplication, RenameTable, TagResource, UntagResource,
    UpdateTableMetadataLocation
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-SERVICE-026](#s3t-service-026), [S3T-SERVICE-007](#s3t-service-007), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-021](#s3t-service-021),
    [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-016](#s3t-service-016), [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-022](#s3t-service-022) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-015.md](details/S3T-SERVICE-015.md)

- <a id="s3t-service-016"></a>**S3T-SERVICE-016** `error-code` · impact high · handled · verified 2026-10-14
  **Error catalogue ConflictException (HTTP 409): 8 message classes - 1 AMBIGUOUS, 2 PERMANENT-SPEC, 2 RETRY, 3 WAIT-FOR-STATE**
  ConflictException: declared on 42/49 operations, produced by 8 (CreateTableBucket, CreateNamespace,
  CreateTable, RenameTable, DeleteTable, UpdateTableMetadataLocation, Put/DeleteTable(Bucket)Replication,
  PutTableMaintenanceConfiguration, PutTablePolicy); never by a read. 8 message classes; classification counts
  {'AMBIGUOUS': 1, 'PERMANENT-SPEC': 2, 'RETRY': 2, 'WAIT-FOR-STATE': 3}. Three different things share the
  code: a ~10 min name cooldown (WAIT), 'already exists' duplicates (adopt), stale versionTokens (re-read and
  retry) and true concurrent-write collisions (RETRY as-is).
  - ACK: terminal_codes, docs-only · ops: CreateNamespace, CreateTable, CreateTableBucket, DeleteTable,
    DeleteTableBucketReplication, DeleteTableReplication, PutTableBucketReplication,
    PutTableMaintenanceConfiguration, PutTablePolicy, PutTableReplication, RenameTable,
    UpdateTableMetadataLocation
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-021](#s3t-service-021),
    [S3T-SERVICE-022](#s3t-service-022), [S3T-TABLEBUCKETREPLICATION-001](tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-002](tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-004](tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKET-012](tablebucket-replication.md#s3t-tablebucket-012) · evidence: service/creative/raw-wire-codes,
    service/creative/sig-skew-and-empty-500, service/creative/iam-authz-variants,
    service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-016.md](details/S3T-SERVICE-016.md)

- <a id="s3t-service-017"></a>**S3T-SERVICE-017** `error-code` · impact medium · handled · verified 2026-10-14
  **Error catalogue MethodNotAllowedException (HTTP 405): 1 message classes - 1 PERMANENT-SPEC**
  MethodNotAllowedException: declared on 3/49 operations (record expiration), produced by exactly those 3. 1
  message classes; classification counts {'PERMANENT-SPEC': 1}. A scope signal (table type), terminal; ordered
  after shape validation, authz and lookup.
  - ACK: terminal_codes, docs-only · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-018](#s3t-service-018), [S3T-SERVICE-023](#s3t-service-023) · evidence: service/creative/raw-wire-codes,
    service/creative/sig-skew-and-empty-500, service/creative/iam-authz-variants,
    service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-017.md](details/S3T-SERVICE-017.md)

- <a id="s3t-service-018"></a>**S3T-SERVICE-018** `error-code` · impact medium · handled · verified 2026-10-14
  **Error catalogue TooManyRequestsException (HTTP 429): 1 message classes - 1 RETRY**
  TooManyRequestsException: declared on 49/49; produced by 6 (all writes except one metrics Get). 1 message
  classes; classification counts {'RETRY': 1}. RETRY with no server hint; budget shared across buckets; no
  side effects.
  - ACK: terminal_codes, requeue · ops: CreateNamespace, CreateTableBucket, DeleteNamespace, DeleteTable,
    GetTableBucketMetricsConfiguration, PutTableMaintenanceConfiguration
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-004](#s3t-service-004), [S3T-SERVICE-008](#s3t-service-008), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018), [S3T-SERVICE-017](#s3t-service-017), [S3T-SERVICE-023](#s3t-service-023) ·
    evidence: service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-018.md](details/S3T-SERVICE-018.md)

- <a id="s3t-service-019"></a>**S3T-SERVICE-019** `error-code` · impact high · handled · verified 2026-10-14
  **Error catalogue AccessDeniedException (HTTP 403): 9 message classes - 9 AUTHZ**
  AccessDeniedException: declared on only 15/49 operations but produced by 49/49 (seed ~35 + 15 live); NOT in
  terminal lists generated from the model. 9 message classes; classification counts {'AUTHZ': 9}. AUTHZ in
  every case, one message template with 5 reason suffixes; the denied action may differ from the operation;
  visible ~86-100 s after an IAM allow/revoke ([S3T-SERVICE-012](#s3t-service-012)) and from a fresh role's first call.
  - ACK: terminal_codes, requeue · ops: CreateNamespace, CreateTable, CreateTableBucket, DeleteNamespace,
    GetNamespace, GetTable, GetTableBucket, GetTableMaintenanceJobStatus, ListNamespaces, ListTableBuckets,
    ListTables, PutTableBucketReplication, PutTableMaintenanceConfiguration, PutTableReplication
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-016](#s3t-service-016), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-021](#s3t-service-021),
    [S3T-SERVICE-022](#s3t-service-022), [S3T-SERVICE-012](#s3t-service-012) · evidence: service/creative/raw-wire-codes,
    service/creative/sig-skew-and-empty-500, service/creative/iam-authz-variants,
    service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-019.md](details/S3T-SERVICE-019.md)

- <a id="s3t-service-020"></a>**S3T-SERVICE-020** `error-code` · impact high · handled · verified 2026-10-14
  **Error catalogue InternalServerErrorException (HTTP 500): 1 message classes - 1 AMBIGUOUS**
  InternalServerErrorException: declared on 49/49; produced by 9 ops, one message text. 1 message classes;
  classification counts {'AMBIGUOUS': 1}. One text for a transient collision and for three deterministic
  request/state errors - the single most dangerous code for a retry loop.
  - ACK: terminal_codes, requeue · ops: CreateTable, PutTableBucketMaintenanceConfiguration,
    PutTableBucketPolicy, PutTableBucketReplication, PutTablePolicy, PutTableReplication, RenameTable,
    TagResource, UntagResource
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-002](#s3t-service-002), [S3T-SERVICE-007](#s3t-service-007), [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-016](#s3t-service-016),
    [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-021](#s3t-service-021), [S3T-SERVICE-022](#s3t-service-022), [S3T-TABLE-300](table-replication-status.md#s3t-table-300), [S3T-TABLE-304](table-replication.md#s3t-table-304) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-020.md](details/S3T-SERVICE-020.md)

- <a id="s3t-service-021"></a>**S3T-SERVICE-021** `error-code` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Error catalogue ValidationException (HTTP 400): 17 message classes - 1 PERMANENT-QUOTA, 16 PERMANENT-SPEC**
  ValidationException: NOT in the s3tables model (no SDK type); produced by 10 operations whose
  required/enum/length/pattern constraints are enforced by the framework. 17 message classes; classification
  counts {'PERMANENT-QUOTA': 1, 'PERMANENT-SPEC': 16}. Framework validation (unmodeled code) for 10 ops; the
  same violations on the other ops are BadRequestException - two layers, one class (PERMANENT-SPEC, plus the
  20 KB policy limit).
  - ACK: terminal_codes, docs-only · ops: PutTableBucketMaintenanceConfiguration, PutTableBucketPolicy,
    PutTableBucketReplication, PutTableBucketStorageClass, PutTablePolicy,
    PutTableRecordExpirationConfiguration, PutTableReplication, TagResource, UntagResource,
    UpdateTableMetadataLocation
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-SERVICE-026](#s3t-service-026), [S3T-SERVICE-007](#s3t-service-007), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-015](#s3t-service-015),
    [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-016](#s3t-service-016), [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-020](#s3t-service-020), [S3T-SERVICE-022](#s3t-service-022), [S3T-TABLEBUCKET-011](tablebucket-replication.md#s3t-tablebucket-011),
    [S3T-TABLEBUCKET-021](tablebucket-replication.md#s3t-tablebucket-021) · evidence: service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-021.md](details/S3T-SERVICE-021.md)

- <a id="s3t-service-022"></a>**S3T-SERVICE-022** `error-code` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Error catalogue ForbiddenException (HTTP 403): 4 message classes - 3 AUTHZ, 1 PERMANENT-SPEC**
  ForbiddenException: declared on 49/49; produced by 26+ ops (cross-account ARNs) + 3 special texts. 4 message
  classes; classification counts {'AUTHZ': 3, 'PERMANENT-SPEC': 1}. Cross-account 'Unauthorized' plus three
  pass-through texts (KMS key policy, S3 'Access Denied', cross-account PassRole); one is really a spec error.
  - ACK: terminal_codes, docs-only · ops: CreateTable, CreateTableBucket, PutTableReplication
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-SERVICE-014](#s3t-service-014), [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-016](#s3t-service-016), [S3T-SERVICE-019](#s3t-service-019), [S3T-SERVICE-020](#s3t-service-020),
    [S3T-SERVICE-021](#s3t-service-021), [S3T-TABLEBUCKET-021](tablebucket-replication.md#s3t-tablebucket-021), [S3T-SERVICE-026](#s3t-service-026), [S3T-TABLEBUCKET-098](tablebucket-replication.md#s3t-tablebucket-098) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-022.md](details/S3T-SERVICE-022.md)

- <a id="s3t-service-023"></a>**S3T-SERVICE-023** `error-code` · impact medium · handled · verified 2026-10-14
  **Error catalogue sibling-service and transport codes: 11 message classes - 2 AMBIGUOUS, 8 PERMANENT-SPEC, 1 WAIT-FOR-STATE**
  sibling-service and transport codes: codes without an s3tables shape: STS/KMS/IAM/S3 siblings, SigV4 and
  credential failures, routing 404s, client-side validation. 11 message classes; classification counts
  {'AMBIGUOUS': 2, 'PERMANENT-SPEC': 8, 'WAIT-FOR-STATE': 1}. Codes without an s3tables shape; mostly client
  configuration, two transient windows after IAM changes.
  - ACK: terminal_codes, docs-only
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-003](#s3t-service-003), [S3T-SERVICE-005](#s3t-service-005), [S3T-SERVICE-013](#s3t-service-013), [S3T-SERVICE-006](#s3t-service-006), [S3T-SERVICE-025](#s3t-service-025),
    [S3T-SERVICE-001](#s3t-service-001), [S3T-TABLEBUCKET-039](tablebucket.md#s3t-tablebucket-039), [S3T-SERVICE-017](#s3t-service-017), [S3T-SERVICE-018](#s3t-service-018) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue
  - notes: Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the
    request is right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion
    cooldown, or the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC =...
  - full notes: [details/S3T-SERVICE-023.md](details/S3T-SERVICE-023.md)

- <a id="s3t-service-025"></a>**S3T-SERVICE-025** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Fresh IAM role: assumable 8.1-9.1 s after CreateRole, AssumeRole flaps OK->AccessDenied->OK ~3 s; first STS creds -> UnrecognizedClient 403**
  Roles created by the probe became assumable 8.1 / 8.75 / 8.6 / 9.1 s after CreateRole (AccessDenied 'not
  authorized to perform: sts:AssumeRole' before). For a role re-created under a name deleted seconds earlier
  AssumeRole answered OK on 6 consecutive 1-s polls and then AccessDenied on the next 2-3 calls before OK
  again; a GetTableBucket with credentials obtained ~2 s after that OK was rejected by s3tables with
  UnrecognizedClientException 403 'The security token included in the request is invalid.' (the probe paused;
  the identical call 10 s later with fresh credentials -> the expected AccessDeniedException).
  - ACK: requeue · ops: GetTableBucket
  - repro: CreateRole; poll AssumeRole at 1 s; immediately use the credentials for GetTableBucket
  - measurements: assume_role_first_ok_s=[8.1, 8.75, 8.6, 9.08], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-010](#s3t-service-010), [S3T-SERVICE-012](#s3t-service-012), [S3T-SERVICE-003](#s3t-service-003), [S3T-SERVICE-005](#s3t-service-005), [S3T-SERVICE-013](#s3t-service-013),
    [S3T-SERVICE-006](#s3t-service-006), [S3T-SERVICE-023](#s3t-service-023), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](table-policy.md#s3t-table-164) · evidence:
    service/creative/iam-authz-variants, service/creative/error-catalogue
  - notes: Seed rows 'AccessDenied sts:AssumeRole (n=29)' and 'UnrecognizedClientException (n=2)' are
    explained: both are IAM propagation of a role created seconds earlier, i.e. RETRY, not configuration
    errors. A controller that creates an IAM Role and immediately assumes it (or hands it to another
    controller)...
  - full notes: [details/S3T-SERVICE-025.md](details/S3T-SERVICE-025.md)

- <a id="s3t-service-027"></a>**S3T-SERVICE-027** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **IAM Allow seen by s3tables after 100.8/99.8 s, revoke after 101.8/102.0 s, same as s3 (n=2): IAM-wide propagation, not an s3tables cache**
  One assumed role, one inline policy granting s3tables:GetTableBucket and s3:ListAllMyBuckets, polled at 1 s.
  After PutRolePolicy the first 200: s3tables GetTableBucket at [100.8, 99.8] s, s3 ListBuckets at [100.8,
  99.8] s. After DeleteRolePolicy the first 403: s3tables at [101.8, 102.0] s, s3 at [101.8, 102.0] s. Codes
  while waiting: s3tables ['AccessDeniedException', 'OK'] / s3 ['AccessDenied', 'OK'] (allow),
  ['AccessDeniedException', 'OK'] / ['AccessDenied', 'OK'] (revoke).
  - ACK: requeue · ops: GetTableBucket
  - repro: CreateRole; AssumeRole; PutRolePolicy(Allow s3tables:GetTableBucket + s3:ListAllMyBuckets); poll
    both at 1 s; DeleteRolePolicy; poll both; x2
  - measurements: allow_s3tables_s=[100.8, 99.8], allow_s3_s=[100.8, 99.8], revoke_s3tables_s=[101.8, 102.0],
    revoke_s3_s=[101.8, 102.0], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-012](#s3t-service-012), [S3T-SERVICE-024](#s3t-service-024), [S3T-TABLE-234](table-maintenance.md#s3t-table-234), [S3T-SERVICE-011](#s3t-service-011) · evidence:
    service/creative/iam-propagation-compare
  - notes: Creative round, seed error-messages: decides whether the AccessDeniedException row's '~90 s after
    an IAM change' is a property of s3tables (authorization decision cache) or of IAM. Verdict: IAM-wide
    propagation (s3 lags as much). Controller rule stays: requeue AccessDeniedException with backoff for...
  - full notes: [details/S3T-SERVICE-027.md](details/S3T-SERVICE-027.md)

### Response fidelity and consistency

- <a id="s3t-service-010"></a>**S3T-SERVICE-010** `eventual-consistency` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Fresh IAM role as policy Principal: Put*Policy -> BadRequestException 'Invalid principal in policy' 7.2/7.9/7.8 s after CreateRole, then OK**
  CreateRole, then PutTableBucketPolicy (roles 0,2) / PutTablePolicy (role 1) with Principal.AWS=<new role
  ARN> every 0.5 s: first HTTP 200 at [7.22, 7.93, 7.79] s ([13, 14, 14] x BadRequestException before it,
  codes seen ['BadRequestException'], messages ['Invalid principal in policy']). sts:AssumeRole on the same
  new role polled in the same loop: first success at [8.1, 8.75, 8.6] s; before that ['User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: sts:AssumeRole'].
  - ACK: requeue, references · ops: PutTableBucketPolicy, PutTablePolicy · fields: resourcePolicy
  - repro: iam CreateRole(trust=account root) -> immediately loop PutTableBucketPolicy/PutTablePolicy with the
    role as Principal at 0.5 s until 200
  - measurements: put_policy_first_ok_s=[7.22, 7.93, 7.79], assume_role_first_ok_s=[8.1, 8.75, 8.6], n=3
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](table-policy.md#s3t-table-164), [S3T-SERVICE-025](#s3t-service-025), [S3T-SERVICE-012](#s3t-service-012) · evidence:
    service/creative/iam-authz-variants
  - notes: Creative round, seed error-messages: the BadRequestException 'Invalid principal in policy' class is
    AMBIGUOUS - PERMANENT-SPEC for a typo/deleted principal but transient (WAIT-FOR-STATE on IAM propagation)
    for a principal created seconds earlier, e.g. an ACK Role and a TableBucketPolicy applied in...
  - full notes: [details/S3T-SERVICE-010.md](details/S3T-SERVICE-010.md)

- <a id="s3t-service-028"></a>**S3T-SERVICE-028** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **Re-verified: identity-policy Allow on an assumed role becomes visible to s3tables after 86.1/95.2 s, revoke after 95.4/95.3 s (n=2...**
  Role with no identity policy, one STS session: 15/15 ops -> AccessDeniedException/403 'User: <ARN> is not
  authorized to perform: s3tables:<Op> on resource: <ARN> because no identity-based policy allows the
  s3tables:<Op> action' (exceptions: none). PutRolePolicy(Allow s3tables:GetTableBucket) -> GetTableBucket
  polled at 1/s flipped from AccessDeniedException to 200 after [86.1, 95.2] s; DeleteRolePolicy -> flipped
  back to AccessDeniedException after [95.4, 95.3] s; both flips were stable for the next 5 s in every cycle
  (True / True).
  - ACK: requeue, e2e-timing · ops: GetTableBucket, DeleteTableBucketEncryption,
    DeleteTableBucketMetricsConfiguration, DeleteTableBucketReplication, DeleteTableReplication,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketStorageClass,
    GetTableMaintenanceJobStatus, GetTableReplication, GetTableReplicationStatus, PutTableBucketEncryption,
    PutTableBucketMaintenanceConfiguration, PutTableBucketMetricsConfiguration, PutTableBucketStorageClass
  - repro: CreateRole(no policy) -> AssumeRole -> 15 ops; 2x (PutRolePolicy Allow GetTableBucket, poll 1/s;
    DeleteRolePolicy, poll 1/s)
  - measurements: allow_visible_s=[86.1, 95.2], revoke_visible_s=[95.4, 95.3], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-012](#s3t-service-012) · evidence: service/creative/reverify-set-e
  - notes: Second-opinion reproduction of [S3T-SERVICE-012](#s3t-service-012) (which had n=1: 86.1 s / 100.5 s). Same assumed-role
    session reused across cycles, so the measured delay includes any per-session authorization caching.
  - full notes: [details/S3T-SERVICE-028.md](details/S3T-SERVICE-028.md)

### Quotas and rate limits

- <a id="s3t-service-004"></a>**S3T-SERVICE-004** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **TooManyRequestsException 429 carries NO Retry-After or rate-limit header; CreateNamespace burst: 16 x 200 before the first 429 at 0.5 s**
  Back-to-back CreateNamespace into one bucket: codes {'TooManyRequestsException': 8, '200': 16}. First 429:
  HTTP 429 message 'Too Many Requests'; response headers ['content-length', 'content-type', 'date',
  'x-amzn-errortype', 'x-amzn-requestid']; error body {'Message': 'Too Many Requests', 'Code':
  'TooManyRequestsException'}.
  - ACK: requeue · ops: CreateNamespace
  - repro: CreateNamespace x24 back-to-back (no SDK retries) into one table bucket; inspect
    ResponseMetadata.HTTPHeaders of the first 429
  - measurements: creates_before_429=16, first_429_at_s=0.51, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-008](#s3t-service-008), [S3T-SERVICE-018](#s3t-service-018), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018) · evidence: service/creative/raw-wire-codes
  - notes: Creative round, seed error-messages: the 429 is RETRY (no state to wait for) but the server gives
    no Retry-After hint, so the controller must pick its own backoff ([S3T-NAMESPACE-035](namespace.md#s3t-namespace-035) measured a throttled
    create succeeding again after ~0.5 s).

- <a id="s3t-service-008"></a>**S3T-SERVICE-008** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateNamespace/DeleteNamespace 429 budget is shared across buckets (~16 per burst, account/principal-wide): A 12/12 OK, B 429 from 5th call**
  12 CreateNamespace into bucket A then 12 into bucket B back-to-back (wall 0.6 s): A codes ['200', '200',
  '200', '200', '200', '200', '200', '200', '200', '200', '200', '200'] (first 429 at index None); B codes
  ['200', '200', '200', '200', 'TooManyRequestsException', 'TooManyRequestsException',
  'TooManyRequestsException', 'TooManyRequestsException', '200', 'TooManyRequestsException',
  'TooManyRequestsException', 'TooManyRequestsException'] (first 429 at index 4). ListNamespaces afterwards: A
  has 12, B has 5; names whose create got 429 but exist: []; names whose create got 200 but are missing: [].
  Delete burst (wall 0.4 s): A ['204', '204', '204', '204', '204', '204', '204', '204', '204', '204', '204',
  '204'], B ['204', '204', '204', '204', 'TooManyRequestsException']; namespaces whose delete got 429 but are
  gone: [].
  - ACK: requeue, one-per-reconcile · ops: CreateNamespace, DeleteNamespace
  - repro: Two fresh table buckets; 12 CreateNamespace into A then 12 into B without pause; ListNamespaces
    both; repeat with DeleteNamespace
  - measurements: creates_ok_A=12, creates_ok_B=5, first_429_index_A=null, first_429_index_B=4, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-004](#s3t-service-004), [S3T-SERVICE-018](#s3t-service-018), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018) · evidence: service/creative/throttle-scope
  - notes: Creative round, seed error-messages: fixes what the TooManyRequestsException RETRY row is keyed on:
    NOT the bucket - bucket B was throttled once the combined count reached 16 (A 12 + B 4), and the delete
    burst behaved the same (A 12 + B 4 then 429). A 429'd create/delete had no side effect (0 ghost...
  - full notes: [details/S3T-SERVICE-008.md](details/S3T-SERVICE-008.md)

### Appendix: low-impact and duplicate service-wide findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-service-001"></a>**S3T-SERVICE-001** | error-code | low | confirmed | Raw endpoint: unknown paths/wrong methods -> HTTP 404 <UnknownOperationException/> (HEAD: empty body); no modeled error shape, SDK code 404 | [S3T-TABLEBUCKET-039](tablebucket.md#s3t-tablebucket-039), [S3T-SERVICE-023](#s3t-service-023) | - |
| <a id="s3t-service-005"></a>**S3T-SERVICE-005** | error-code | low | confirmed | SigV4 clock skew on GET /buckets: -4 min and 0 accepted; -14/-16/-20/-60 min and +16/+20 min -> InvalidSignatureException 403 | [S3T-SERVICE-003](#s3t-service-003), [S3T-SERVICE-013](#s3t-service-013), [S3T-SERVICE-006](#s3t-service-006), [S3T-SERVICE-025](#s3t-service-025), [S3T-SERVICE-023](#s3t-service-023) | [S3T-SERVICE-013](#s3t-service-013) |
| <a id="s3t-service-006"></a>**S3T-SERVICE-006** | error-code | low | confirmed | Garbled STS token: tail edits (last char flipped, last 4 -> AAAA, 1 appended) still 200; head/middle/truncated/empty -> UnrecognizedClient | [S3T-SERVICE-003](#s3t-service-003), [S3T-SERVICE-005](#s3t-service-005), [S3T-SERVICE-013](#s3t-service-013), [S3T-SERVICE-025](#s3t-service-025), [S3T-SERVICE-023](#s3t-service-023) | - |
| <a id="s3t-service-009"></a>**S3T-SERVICE-009** | other | low | confirmed | Doc claim C040 TRUE: maxTables/maxNamespaces/maxBuckets=0 -> BadRequestException 400 'limit ... positive number' (botocore rejects first) | [S3T-TABLE-168](table.md#s3t-table-168), [S3T-TABLE-092](table.md#s3t-table-092), [S3T-NAMESPACE-032](namespace.md#s3t-namespace-032), [S3T-TABLEBUCKET-069](tablebucket.md#s3t-tablebucket-069), [S3T-NAMESPACE-038](namespace.md#s3t-namespace-038), [S3T-TABLE-293](table.md#s3t-table-293), [S3T-TABLE-091](table.md#s3t-table-091) | - |
| <a id="s3t-service-013"></a>**S3T-SERVICE-013** | error-code | low | confirmed | SigV4 clock-skew tolerance is +/-5 min: -4.5/+4.5 min accepted; -5.5/-6/+5.5/+6 min -> InvalidSignatureException 403 (message states 5 min) | [S3T-SERVICE-005](#s3t-service-005), [S3T-SERVICE-003](#s3t-service-003), [S3T-SERVICE-006](#s3t-service-006), [S3T-SERVICE-025](#s3t-service-025), [S3T-SERVICE-023](#s3t-service-023) | - |
| <a id="s3t-service-024"></a>**S3T-SERVICE-024** | error-code | high | confirmed | s3tables sees an IAM identity-policy change ~90 s late: Allow visible 86 s after PutRolePolicy, revoke 100 s after DeleteRolePolicy (n=1) | [S3T-SERVICE-012](#s3t-service-012), [S3T-TABLE-234](table-maintenance.md#s3t-table-234), [S3T-SERVICE-027](#s3t-service-027), [S3T-SERVICE-011](#s3t-service-011) | [S3T-SERVICE-027](#s3t-service-027) |
| <a id="s3t-service-026"></a>**S3T-SERVICE-026** | error-code | low | confirmed | Put*Replication with configuration={} or rules=[] -> 400 with the GetTable identifier text 'Must provide either table arn alone OR all of' | [S3T-SERVICE-007](#s3t-service-007), [S3T-TABLEREPLICATION-011](table-replication.md#s3t-tablereplication-011), [S3T-TABLEBUCKETREPLICATION-003](tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-015](#s3t-service-015), [S3T-SERVICE-021](#s3t-service-021), [S3T-TABLEBUCKET-021](tablebucket-replication.md#s3t-tablebucket-021), [S3T-SERVICE-022](#s3t-service-022), [S3T-TABLEBUCKET-098](tablebucket-replication.md#s3t-tablebucket-098) | - |

## Tag API

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## Codegen notes

- [S3T-TABLE-030](table-encryption-storage.md#s3t-table-030) - Model: GetTable input 'tableArn' vs output 'tableARN'; DeleteTable requires bucket
  ARN+namespace+name; no UpdateTable operation ([table-encryption-storage.md](table-encryption-storage.md))
- [S3T-TABLE-211](table-maintenance.md#s3t-table-211) - Maintenance model hazards: enum value "z-order" (hyphen), required path member named "type",
  settings is a tagged union keyed by type ([table-maintenance.md](table-maintenance.md))
- [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096) - Model marks the 4 bucket sub-resource Deletes idempotent; on an absent config
  metrics/policy -> 204 but replication/encryption -> 404
  ([tablebucket-replication.md](tablebucket-replication.md))
- [S3T-TABLEBUCKET-104](tablebucket-encryption-storage.md#s3t-tablebucket-104) - PutTableBucketEncryption sseAlgorithm is the literal 'aws:kms': 'AWS_KMS', 'aws_kms',
  'AWS:KMS', 'Aes256' -> BadRequestException 400
  ([tablebucket-encryption-storage.md](tablebucket-encryption-storage.md))

## Out of scope or deferred

| finding | resource | verdict | title | doc |
| --- | --- | --- | --- | --- |
| [S3T-TABLE-147](table-record-expiration.md#s3t-table-147) | Table | defer | PutTableRecordExpirationConfiguration on customer tables -> 405 for days 1/7/365, with/without Iceberg schema, in us-west-2 and us-east-1 | [table-record-expiration.md](table-record-expiration.md) |
| [S3T-TABLE-221](table-record-expiration.md#s3t-table-221) | Table | defer | AWS-managed S3 Metadata journal table (type=aws): record-expiration Get, Put and GetJobStatus all -> MethodNotAllowedException 405 | [table-record-expiration.md](table-record-expiration.md) |
| [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](table-record-expiration.md#s3t-tablerecordexpirationconfiguration-001) | TableRecordExpirationConfiguration | defer | Customer-created tables: every well-formed record-expiration Put/Get/JobStatus call -> MethodNotAllowedException 405 'not allowed' | [table-record-expiration.md](table-record-expiration.md) |

## Open experiments

None listed in manifest.yaml (`open_experiments`).

## Handling gaps summary

Every suspected controller bug (handling suspect-bug, or notes confirming one) and every partially handled
finding across all documents; each document's own `Handling gaps (bugs to file)` section carries the excerpts.

| finding | label | title | doc |
| --- | --- | --- | --- |
| [S3T-NAMESPACE-024](namespace.md#s3t-namespace-024) | suspected bug | DeleteNamespace with one table inside -> BadRequestException 400 'not empty' on 6/6 retries over 10 s; table and namespace untouched | [namespace.md](namespace.md) |
| [S3T-SERVICE-010](#s3t-service-010) | suspected bug | Fresh IAM role as policy Principal: Put*Policy -> BadRequestException 'Invalid principal in policy' 7.2/7.9/7.8 s after CreateRole, then OK | [service.md](service.md) |
| [S3T-SERVICE-015](#s3t-service-015) | suspected bug | Error catalogue BadRequestException (400): 57 message classes - 45 PERMANENT-SPEC, 5 WAIT-FOR-STATE, 4 AMBIGUOUS, 1 AUTHZ, 1 QUOTA, 1 RETRY | [service.md](service.md) |
| [S3T-SERVICE-021](#s3t-service-021) | suspected bug | Error catalogue ValidationException (HTTP 400): 17 message classes - 1 PERMANENT-QUOTA, 16 PERMANENT-SPEC | [service.md](service.md) |
| [S3T-SERVICE-022](#s3t-service-022) | suspected bug | Error catalogue ForbiddenException (HTTP 403): 4 message classes - 3 AUTHZ, 1 PERMANENT-SPEC | [service.md](service.md) |
| [S3T-TABLE-006](table-metadata-location.md#s3t-table-006) | suspected bug | RenameTable(newNamespaceName only) moves the table synchronously (HTTP 204); namespaceId follows the destination, tableARN unchanged | [table-metadata-location.md](table-metadata-location.md) |
| [S3T-TABLE-035](table.md#s3t-table-035) | partially handled | Raw CreateTable with an extra 'type':'aws' member -> HTTP 400 BadRequestException "Creating tables of this type are not allowed."... | [table.md](table.md) |
| [S3T-TABLE-098](table-metadata-location.md#s3t-table-098) | suspected bug | DeleteTable with the create-time versionToken after one UpdateTableMetadataLocation (location moved) -> 409; current token -> 204 | [table-metadata-location.md](table-metadata-location.md) |
| [S3T-TABLE-108](table-metadata-location.md#s3t-table-108) | suspected bug | Rollback: UpdateTableMetadataLocation back to the original 00000-*.metadata.json -> 200; response versionToken == the CreateTable token | [table-metadata-location.md](table-metadata-location.md) |
| [S3T-TABLE-119](table.md#s3t-table-119) | suspected bug | Table tag bounds: 129-char key / 257-char value -> ValidationException 400 (unmodeled); '\|', '#', emoji key or '☃' value -> BadRequest 400 | [table.md](table.md) |
| [S3T-TABLE-120](table.md#s3t-table-120) | suspected bug | TagResource over-long key/value and empty key return wire code ValidationException (HTTP 400), which the s3tables model does not declare | [table.md](table.md) |
| [S3T-TABLE-133](table-schema.md#s3t-table-133) | suspected bug | CreateTable schema={} (fields omitted on the wire) -> 500 in 4/4, no table created; partitionSpec={} and writeOrder={orderId:1} -> 200 | [table-schema.md](table-schema.md) |
| [S3T-TABLE-231](table.md#s3t-table-231) | partially handled | S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s; plain s3tables DeleteTable removes it (204) | [table.md](table.md) |
| [S3T-TABLE-312](table-schema.md#s3t-table-312) | partially handled | Schema 'required' round-trips into metadata.json: required=true -> True, required=false -> False, omitted -> False | [table-schema.md](table-schema.md) |
| [S3T-TABLE-313](table-schema.md#s3t-table-313) | partially handled | Table 'type' is 'customer' in GetTable and ListTables (agree); schema field types live only in metadata.json, in no API output | [table-schema.md](table-schema.md) |
| [S3T-TABLE-314](table-schema.md#s3t-table-314) | partially handled | CreateTable ignores explicit schema field ids (renumbered 1..n) and normalizes 'decimal(10,2)' to 'decimal(10, 2)' in metadata.json | [table-schema.md](table-schema.md) |
| [S3T-TABLEBUCKET-004](tablebucket.md#s3t-tablebucket-004) | partially handled | CreateTableBucket is synchronous: GetTableBucket succeeds on the first poll ~0 ms after Create, no status/state field exists | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKET-019](tablebucket.md#s3t-tablebucket-019) | suspected bug | DeleteTableBucket on a bucket holding a table -> BadRequestException (HTTP 400); with only an empty namespace -> same code | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKET-030](tablebucket.md#s3t-tablebucket-030) | suspected bug | TagResource length violations return the NON-modeled ValidationException (HTTP 400), unlike CreateTableBucket's BadRequestException | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKET-062](tablebucket-policy.md#s3t-tablebucket-062) | suspected bug | DeleteTableBucketPolicy is idempotent: HTTP 204 when no policy exists and on a second Delete (not NotFoundException) | [tablebucket-policy.md](tablebucket-policy.md) |
| [S3T-TABLEBUCKET-075](tablebucket-maintenance-metrics.md#s3t-tablebucket-075) | suspected bug | Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and 2147483647 round-trip; 2147483648 -> 204, read back clamped | [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) |
| [S3T-TABLEBUCKET-076](tablebucket-maintenance-metrics.md#s3t-tablebucket-076) | partially handled | Server-side: settings {} (empty union) -> HTTP 204 no-op; request body without value -> ValidationException 400 'Member must not be null' | [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) |
| [S3T-TABLEBUCKET-087](tablebucket-encryption-storage.md#s3t-tablebucket-087) | suspected bug | PutTableBucketStorageClass enum/null violations return wire code ValidationException (HTTP 400), a shape NOT in the API model | [tablebucket-encryption-storage.md](tablebucket-encryption-storage.md) |
| [S3T-TABLEBUCKET-122](tablebucket.md#s3t-tablebucket-122) | partially handled | ListTableBuckets summary at +0 ms equals GetTableBucket field by field (5/5; same 6 fields, raw createdAt strings identical) | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKET-130](tablebucket.md#s3t-tablebucket-130) | partially handled | ListTableBuckets(type=aws) is empty in us-west-2 and us-east-1; own bucket type=customer in Get and List; aws-type bucket not reachable | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKET-159](tablebucket.md#s3t-tablebucket-159) | suspected bug | CreateTableBucket ConflictException has two meanings (exists vs post-delete transitional); only the message and a Get tell them apart | [tablebucket.md](tablebucket.md) |
| [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](tablebucket-maintenance-metrics.md#s3t-tablebucketmaintenanceconfiguration-001) | partially handled | PutTableBucketMaintenanceConfiguration silently clamps nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204 | [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) |
| [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](tablebucket-maintenance-metrics.md#s3t-tablebucketmaintenanceconfiguration-002) | partially handled | Maintenance days edges: 2147483647 stored; 0 and -1 -> ValidationException 400; 2^63 and 2^64 -> 204 clamped; 1.5 -> 204 stored as 1 | [tablebucket-maintenance-metrics.md](tablebucket-maintenance-metrics.md) |
| [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001) | suspected bug | PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400... | [tablebucket-policy.md](tablebucket-policy.md) |
| [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) | suspected bug | GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar... | [tablebucket-policy.md](tablebucket-policy.md) |

<a id="ground-truth"></a>
## Ground truth catalog

AWS behaviors the controller already handles or works around, as catalogued before probing
(`services/s3tables/ground-truth.yaml`, 44 entries; the behavior column is the catalog text on one line, the
mechanism column is how the controller copes). `GT-*` ids cited in the overviews link here. File:line
references are relative to controller commit 1eda357.

| id | resource | category | behavior | mechanism |
| --- | --- | --- | --- | --- |
| <a id="gt-s3t-001"></a>**GT-S3T-001** | TableBucket | identity | GetTableBucket and DeleteTableBucket are keyed by the bucket ARN only; there is no lookup by name. CreateTableBucket returns only the ARN. The ARN is deterministic (arn:{partition}:s3tables:{region}:{account}:bucket/{name}), so it can be derived from the name. | is_arn_primary_key: true. Status.ACKResourceMetadata.ARN is the identifier; the generated requiredFieldsMissingFromReadOneInput returns NotFound while the status ARN is unset (so the first reconcile goes to Create). Adoption uses adoption-fields {"arn": ...} (SetIdentifiers / PopulateResourceFromAnnotation write the ARN into status, not spec.name). |
| <a id="gt-s3t-002"></a>**GT-S3T-002** | TableBucket | immutable-field | There is no UpdateTableBucket operation; the bucket name cannot be changed after creation. Only sub-resources (encryption, storage class, policy, maintenance, tags) are mutable. | Name is_immutable (CEL self == oldSelf). update_operation.custom_method_name customUpdateTableBucket dispatches only to the sub-resource Put/Delete/Tag APIs. |
| <a id="gt-s3t-003"></a>**GT-S3T-003** | TableBucket | read-gap | GetTableBucket returns only arn, name, ownerAccountId, createdAt, tableBucketId and type. The encryption configuration, storage class, tags, resource policy and maintenance configuration are each readable only through their own Get API (GetTableBucketEncryption, GetTableBucketStorageClass, ListTagsForResource, GetTableBucketPolicy, GetTableBucketMaintenanceConfiguration), i.e. 6 read calls to observe one bucket. | sdk_read_one_post_set_output hook calls setBucketConfigurations, which fans out the five extra Get calls on every reconcile and writes the results into Spec so the delta against desired is accurate (otherwise every reconcile would re-Put). Fields are declared via `from:` so they are generated into the CRD. |
| <a id="gt-s3t-004"></a>**GT-S3T-004** | TableBucket | sub-resource-api | Bucket encryption is accepted inline by CreateTableBucket, but afterwards it is managed only through the dedicated PutTableBucketEncryption / GetTableBucketEncryption / DeleteTableBucketEncryption pair; it is not part of GetTableBucket. | EncryptionConfiguration kept in Spec; read back via GetTableBucketEncryption in setBucketConfigurations; delta on Spec.EncryptionConfiguration drives syncEncryption -> PutTableBucketEncryption in customUpdateTableBucket. |
| <a id="gt-s3t-005"></a>**GT-S3T-005** | TableBucket | sub-resource-api | The bucket default storage class is accepted inline by CreateTableBucket but afterwards is managed only through PutTableBucketStorageClass / GetTableBucketStorageClass; it is not part of GetTableBucket and there is no Delete operation for it. | StorageClassConfiguration kept in Spec; read back via GetTableBucketStorageClass; delta drives syncStorageClass -> PutTableBucketStorageClass. |
| <a id="gt-s3t-006"></a>**GT-S3T-006** | TableBucket | server-default | When encryptionConfiguration / storageClassConfiguration are omitted at create, the service applies defaults (SSE-S3 AES256 and STANDARD) and GetTableBucketEncryption / GetTableBucketStorageClass return those defaults as if they had been set (no NotFound), so a bucket created with neither reads back fully populated. | late_initialize with skip_incomplete_check on EncryptionConfiguration and StorageClassConfiguration: the observed defaults are copied into Spec once (a nil spec field adopts the AWS value) and later out-of-band drift is still detected. Comment: "GetTableBucket does not return encryption/storage class config; S3 Tables applies server-side defaults". |
| <a id="gt-s3t-007"></a>**GT-S3T-007** | TableBucket | unsettable-field | Storage class has no Delete API: once set it can only be changed to another enum value, never removed. Encryption does have DeleteTableBucketEncryption (reverts to the service default), but reverting requires that extra operation rather than omitting the field. | syncEncryption and syncStorageClass are no-ops when the desired field is nil; late-init then refills Spec from the observed value, so dropping either field from the CR changes nothing. DeleteTableBucketEncryption is never called; reverting encryption to SSE-S3 requires setting sseAlgorithm: AES256 explicitly. |
| <a id="gt-s3t-008"></a>**GT-S3T-008** | TableBucket | prerequisite | SSE-KMS (sseAlgorithm aws:kms) requires a customer-managed KMS key ARN in kmsKeyArn, and the S3 Tables maintenance principal must be granted access to that key; otherwise Create/PutTableBucketEncryption (and table creation in the bucket) fail. | EncryptionConfiguration.KMSKeyARN is a cross-service reference to kms Key (Status.ACKResourceMetadata.ARN). No validation in the controller; e2e deliberately tests only AES256 because the bootstrap provisions no KMS key. |
| <a id="gt-s3t-009"></a>**GT-S3T-009** | TableBucket, Table | tag-semantics | Tags are not returned by GetTableBucket or GetTable. Tagging uses the shared ListTagsForResource / TagResource / UntagResource trio keyed by resource ARN (ResourceArn pattern covers buckets and tables only). TagResource is an upsert ("applies ... or updates existing tags"), UntagResource takes tag keys as a query string, and the limit is 50 tags per resource (ACK system tags count against it). Create accepts tags inline. | Read hooks call ListTagsForResource and write Spec.Tags (nil when empty) so the generated tag delta is accurate. syncTags computes added/removed via ackcompare.GetTagsDifference and treats a key present in both as a value change kept only in TagResource (no Untag). Untag runs before Tag. Tests assert with assert_equal_without_ack_tags. |
| <a id="gt-s3t-010"></a>**GT-S3T-010** | TableBucket | sub-resource-api | The bucket resource policy is not part of CreateTableBucket or GetTableBucket; it is managed exclusively via PutTableBucketPolicy (replaces the whole document) / GetTableBucketPolicy / DeleteTableBucketPolicy. | Policy spec field with `from: GetTableBucketPolicy.ResourcePolicy`; read in setBucketConfigurations; syncPolicy issues PutTableBucketPolicy when set and DeleteTableBucketPolicy when nil/empty. sdk_create_post_set_output nudges Synced=False when Spec.Policy is set so the update reconcile applies it immediately instead of waiting for the periodic resync (post-create-nudge). |
| <a id="gt-s3t-011"></a>**GT-S3T-011** | TableBucket | error-code | GetTableBucketPolicy returns NotFoundException (HTTP 404) when the bucket exists but has no policy attached - the same code used for a missing bucket. | setBucketConfigurations treats errors.As(*types.NotFoundException) from GetTableBucketPolicy as "no policy" (Spec.Policy = nil) instead of failing the read; any other error aborts ReadOne. The generated 404 mapping only covers GetTableBucket, so this has to be hand-written. |
| <a id="gt-s3t-012"></a>**GT-S3T-012** | TableBucket | first-sync-destructive | Because the policy lives behind its own Put/Delete pair and nothing in CreateTableBucket or GetTableBucket carries it, a controller that treats an unset desired policy as "no policy" will delete a policy attached out-of-band (or present on an adopted bucket) on its first update reconcile. | Policy is deliberately NOT late_initialize'd (comment: "update runs before late-init and treats a nil Policy as delete, so back-filling spec from a live policy would let an unset spec delete it"). syncPolicy calls DeleteTableBucketPolicy whenever desired Policy is nil or "" and a delta exists. Mirrors the S3 controller's Bucket.Policy semantics. |
| <a id="gt-s3t-013"></a>**GT-S3T-013** | TableBucket | normalization | The resource policy is an IAM-style JSON document; the service may return it from GetTableBucketPolicy in a canonicalized/reformatted form (whitespace, key order, single-element arrays), so byte comparison against the submitted document would churn. | Policy is marked is_iam_policy: true, so the generated delta uses ackcompare.IAMPolicyDocumentEqual (semantic comparison) instead of string equality. |
| <a id="gt-s3t-014"></a>**GT-S3T-014** | TableBucket | update-granularity | Bucket maintenance configuration is not part of CreateTableBucket or GetTableBucket. GetTableBucketMaintenanceConfiguration returns a map keyed by maintenance type (currently only icebergUnreferencedFileRemoval), whereas PutTableBucketMaintenanceConfiguration writes ONE type per call (type is a URI parameter, value = {status, settings}). | MaintenanceConfiguration spec field `from: GetTableBucketMaintenanceConfiguration.Configuration` (so the CRD mirrors the map-by-type shape, yielding the double-nested maintenanceConfiguration.icebergUnreferencedFileRemoval.settings.icebergUnreferencedFileRemoval YAML). syncMaintenance loops over the desired map issuing one Put per declared type. Post-create nudge (Synced=False) when the field is set so the Put happens right after create. |
| <a id="gt-s3t-015"></a>**GT-S3T-015** | TableBucket | server-default | A new table bucket already has a maintenance configuration: icebergUnreferencedFileRemoval is status=enabled with service-default unreferencedDays/nonCurrentDays. GetTableBucketMaintenanceConfiguration always returns a configuration, never NotFound. | late_initialize (skip_incomplete_check) copies the observed config into Spec when unset, so an absent field preserves the AWS default instead of disabling it; the controller deliberately hard-codes no default so a service-side change is reflected. maintenanceNeedsSync treats a nil desired sub-field as "adopt whatever AWS defaulted". Test asserts clearing the field keeps status=enabled. |
| <a id="gt-s3t-016"></a>**GT-S3T-016** | TableBucket | unsettable-field | There is no DeleteTableBucketMaintenanceConfiguration; the configuration cannot be removed, only turned off by Put with status=disabled. Put accepts status alone (settings optional). | syncMaintenance never deletes; a maintenance type absent from the desired spec is left alone (late-initialized), so disabling requires an explicit status: disabled in the spec. e2e disables with {"status": "disabled"} and no settings. |
| <a id="gt-s3t-017"></a>**GT-S3T-017** | TableBucket | response-fidelity | GetTableBucketMaintenanceConfiguration returns the fully defaulted value (status plus settings.unreferencedDays and nonCurrentDays) even when the caller only set status, so a partial desired configuration never DeepEquals the observed map. | compare.is_ignored on MaintenanceConfiguration removes the generated whole-map DeepEqual (which would re-Put every reconcile); delta_post_compare -> customDeltaPostCompare / maintenanceNeedsSync compares only the sub-fields the user declared and adds Spec.MaintenanceConfiguration to the delta only on a genuine difference. |
| <a id="gt-s3t-018"></a>**GT-S3T-018** | TableBucket | codegen-artifact | TableBucketMaintenanceSettings (and TableMetadata on Table) are Smithy tagged unions. The ACK code generator cannot map union members in output shapes and only handles input unions with a one-member guard, so union-typed read paths need hand-written code. | maintenanceValueFromSDK type-asserts *types.TableBucketMaintenanceSettingsMemberIcebergUnreferencedFileRemoval to unwrap, and syncMaintenance wraps the settings back into the member struct. Unknown/future maintenance types are skipped (logged) so a half-mapped value never lands in Spec. The generated CreateTable code handles the TableMetadata input union with an isInterfaceSet guard ("can only set one of the members for Iceberg"). |
| <a id="gt-s3t-019"></a>**GT-S3T-019** | TableBucket | codegen-artifact | The API models day counts and Iceberg ids as int32 (PositiveInteger 1..2147483647) while ACK CRDs model every integer as int64, so values must be narrowed on the way out and could wrap negative without a guard. | Hand-written maintenanceDaysToInt32 returns ackerr.NewTerminalError on values outside 0..MaxInt32 for unreferencedDays/nonCurrentDays; the generator emits math.MaxInt32 range checks for schema/partition/sort field ids on CreateTable. |
| <a id="gt-s3t-020"></a>**GT-S3T-020** | TableBucket | quota-limit | Table buckets are capped by a low per-account (per-Region) quota; stranded buckets quickly exhaust it and make CreateTableBucket fail for unrelated callers. | none in the controller - documented in tests. The e2e adopted_bucket_arns fixture registers out-of-band buckets immediately and reaps survivors by exact ARN at teardown. |
| <a id="gt-s3t-021"></a>**GT-S3T-021** | Namespace, Table | request-validation | Naming rules differ per level: table bucket names match [0-9a-z-]{3,63} (hyphens allowed, no underscores) while namespace and table names match [0-9a-z_]{1,255} (underscores allowed, no hyphens). Kubernetes object names cannot contain underscores, so namespace/table CR names cannot equal the AWS names when underscores are used. | none - documented in tests: e2e derives AWS names with .replace("-", "_") and CR names with .replace("_", "-"). CRD field docs carry the regex but no pattern validation is enforced. |
| <a id="gt-s3t-022"></a>**GT-S3T-022** | TableBucket, Namespace | delete-semantics | A table bucket cannot be deleted while it still contains namespaces/tables, and a namespace cannot be deleted while it contains tables; children must be removed first. | none - documented in tests: fixtures tear down Table -> Namespace -> TableBucket and the Table test's finally-block deletes the table before the fixture deletes its parents. ACK references do not sequence deletion. |
| <a id="gt-s3t-023"></a>**GT-S3T-023** | Namespace | identity | The namespace identifier is a LIST in CreateNamespace input/output and GetNamespace output (namespace: [name]) but a SCALAR path parameter in GetNamespace/DeleteNamespace (and in CreateTable/GetTable/DeleteTable inputs). Only single-element namespaces exist. | ignore.field_paths drops CreateNamespaceInput.Namespace, CreateNamespaceOutput.Namespace and GetNamespaceOutput.Namespace; a synthetic scalar `Name` (type: "*string", is_primary_key, is_immutable) is declared; renames map the GetNamespace/DeleteNamespace input member Namespace -> Name so ACK fills the scalar automatically; sdk_create_post_build_request wraps Spec.Name into []string{name}. |
| <a id="gt-s3t-024"></a>**GT-S3T-024** | Namespace | identity | Namespaces have no ARN: no operation returns one, and GetNamespace/DeleteNamespace are keyed by the composite (tableBucketARN, namespace). GetNamespace returns tableBucketId but not the bucket ARN. | Composite identity from Spec.Name + Spec.TableBucketARN (requiredFieldsMissingFromReadOneInput needs both); Status.ACKResourceMetadata.ARN stays nil; adoption fields are {name, tableBucketARN}. Table's NamespaceRef therefore resolves via `path: Spec.Name` rather than an ARN. |
| <a id="gt-s3t-025"></a>**GT-S3T-025** | Namespace | tag-semantics | Namespaces cannot be tagged: CreateNamespace has no tags parameter and the Tag APIs need a resource ARN (buckets and tables only), which namespaces do not have. | tags.ignore: true - no Spec.Tags field is generated and EnsureTags is a no-op for Namespace. |
| <a id="gt-s3t-026"></a>**GT-S3T-026** | Namespace | immutable-field | Namespaces have no update or rename API (RenameTable's newNamespaceName moves a table between namespaces; it does not rename a namespace). Every namespace attribute is create-only. | Name and TableBucketARN is_immutable; no update_operation, so the generated sdkUpdate returns a Terminal NotImplemented error. e2e has no update phase ("Namespace has no mutable fields"). |
| <a id="gt-s3t-027"></a>**GT-S3T-027** | Table | identity | The table name is the user-facing key but is MUTABLE via RenameTable (newName). The table ARN (bucket/{bucket}/table/{tableId}) is UUID-based and stable across renames, so GetTable by ARN returns the current name. | Name is_primary_key without is_immutable. The read path keys off the stable ARN (see [GT-S3T-028](#gt-s3t-028) (controller hooks catalog entry)) so GetTable returns the CURRENT AWS name; a spec.name change then shows as a Spec.Name delta and customUpdateTable -> renameTable issues RenameTable targeting the latest (bucket, namespace, name) with NewName = desired and the latest versionToken. |
| <a id="gt-s3t-028"></a>**GT-S3T-028** | Table | identity | GetTable accepts either tableArn OR the name triplet (tableBucketARN + namespace + name) as query parameters, but not both together; all four members are optional in the model. | sdk_read_one_post_build_request: when Status.ACKResourceMetadata.ARN is known, set input.TableArn and nil out TableBucketARN/Namespace/Name (the generator populates both); otherwise nil TableArn so the first reconcile and adoption read by triplet. Adoption fields are {name, namespace, tableBucketARN}. |
| <a id="gt-s3t-029"></a>**GT-S3T-029** | Table | shape-mismatch | GetTable (and UpdateTableMetadataLocation / ListTables outputs) return namespace as a LIST, while CreateTable, GetTable, DeleteTable, RenameTable and UpdateTableMetadataLocation inputs take it as a scalar. | ignore.field_paths GetTableOutput.Namespace (the list would otherwise collide with the scalar spec field); sdk_read_one_post_set_output bridges resp.Namespace[0] back into Spec.Namespace so observed equals desired. |
| <a id="gt-s3t-030"></a>**GT-S3T-030** | Table | read-gap | GetTable does not return the Iceberg metadata/schema, encryptionConfiguration, storageClassConfiguration or tags that CreateTable accepted. The schema has no S3 Tables read API at all (only the Iceberg REST catalog exposes it); encryption and storage class have separate GetTableEncryption / GetTableStorageClass operations. | Metadata, EncryptionConfiguration and StorageClassConfiguration are is_immutable, so no drift detection is attempted; the generated sdkFind starts from a DeepCopy of the desired CR so these fields carry the desired value and never delta. GetTableEncryption/GetTableStorageClass are not called. Tags are fetched via ListTagsForResource. The Required-field PR verified the created schema through the Iceberg REST catalog because "GetTable doesn't return the schema, so there's no e2e assertion". |
| <a id="gt-s3t-031"></a>**GT-S3T-031** | Table | idempotency | Tables carry an optimistic-concurrency versionToken: CreateTable and GetTable return it, DeleteTable and RenameTable accept it optionally, UpdateTableMetadataLocation requires it, and every commit (including engine writes) changes it; a stale token is rejected. | VersionToken is a read-only status field `from: GetTable`; the generated DeleteTable payload passes Status.VersionToken as the querystring token; renameTable and updateMetadataLocation pass latest.Status.VersionToken (fresh from this reconcile's GetTable). |
| <a id="gt-s3t-032"></a>**GT-S3T-032** | Table | sub-resource-api | metadataLocation is not a CreateTable input; it is set only through UpdateTableMetadataLocation (and implicitly by query engines committing Iceberg snapshots) and read from GetTable / GetTableMetadataLocation. | MetadataLocation modeled as a Spec field sourced `from: GetTable.MetadataLocation`; sdk_create_post_set_output sets Synced=False and requeues after 1s when desired metadataLocation is set so the next reconcile sees the delta; customUpdateTable -> updateMetadataLocation calls UpdateTableMetadataLocation with the latest triplet and versionToken (no-op when desired is nil). |
| <a id="gt-s3t-033"></a>**GT-S3T-033** | Table | request-validation | UpdateTableMetadataLocation requires an S3 URI that begins with the table's AWS-assigned warehouseLocation and ends with .metadata.json (or .metadata.json.gz); other values are rejected with BadRequestException. | none - no client-side validation; BadRequestException is in terminal_codes so a bad location marks the Table Terminal. warehouseLocation is surfaced read-only from GetTable so users can construct a valid value. |
| <a id="gt-s3t-034"></a>**GT-S3T-034** | Table | immutable-field | Apart from the name (RenameTable), metadataLocation and tags, table attributes are create-only: format, metadata (schema/partition spec/sort order/properties), encryptionConfiguration and storageClassConfiguration have no update path. RenameTable can additionally move a table to another namespace (newNamespaceName). | Format, Metadata, EncryptionConfiguration, StorageClassConfiguration, Namespace and TableBucketARN are is_immutable (CEL); customUpdateTable handles only Name, MetadataLocation and Tags and never sends newNamespaceName, so cross-namespace moves are not supported. |
| <a id="gt-s3t-035"></a>**GT-S3T-035** | Table | request-validation | format only accepts ICEBERG, and creating an ICEBERG table requires metadata.iceberg.schema with at least one field. | none - documented in tests ("ICEBERG requires a schema with at least one field"); the e2e resource always supplies a two-field schema. A violating CR gets BadRequestException -> Terminal. |
| <a id="gt-s3t-036"></a>**GT-S3T-036** | Table | server-default | S3 Tables assigns Iceberg schema field ids itself and defaults required to false when the caller omits them; because the schema is never read back, the assigned ids are invisible to the API caller. | none - SchemaField id/required are optional in the CRD and never reconciled. Documented in the (since removed) generator comment "S3 Tables defaults required=false and assigns field IDs itself". |
| <a id="gt-s3t-037"></a>**GT-S3T-037** | Table | codegen-artifact | SchemaField.required is a non-pointer bool in the Go SDK (Smithy primitive boolean), which the ACK code generator could not round-trip (it generated *bool and a broken setter) until code-generator v0.60.0. | Originally dropped via ignore.field_paths SchemaField.Required; now generated as `elem.Required = *iter.Required` so non-null Iceberg columns can be created. |
| <a id="gt-s3t-038"></a>**GT-S3T-038** | Table | server-default | A table created without encryptionConfiguration / storageClassConfiguration inherits the bucket-level configuration; specifying them at create overrides the bucket default for that table only. | none - both fields optional and is_immutable; the effective values are not read back (see [GT-S3T-030](#gt-s3t-030) (controller hooks catalog entry)). |
| <a id="gt-s3t-039"></a>**GT-S3T-039** | TableBucket, Namespace, Table | error-code | S3 Tables uses a small generic REST error set on every operation: NotFoundException (404) for missing resources and absent sub-resources, BadRequestException (400) for validation, ConflictException (409), ForbiddenException and AccessDeniedException (403), TooManyRequestsException (429), InternalServerErrorException (500). There are no resource-specific *NotFound / ValidationException codes. | exceptions.errors.404.code: NotFoundException and terminal_codes: [BadRequestException] on all three resources; every other code is treated as recoverable and retried. |
| <a id="gt-s3t-040"></a>**GT-S3T-040** | TableBucket, Namespace, Table | codegen-artifact | AWS-assigned identifiers are returned as namespaceId / tableBucketId (lower-case d). ACK renames them with the `ID` initialism (NamespaceID, TableBucketID), and the code generator's output-mapping loop matches SDK member names against resource field keys, so the members do not match and are silently dropped from the generated SetResource. | sdk_read_one_post_set_output hooks on all three resources assign Status.NamespaceID / Status.TableBucketID from resp.NamespaceId / resp.TableBucketId by hand. (Unrelated to the namespace list field being ignored - TableBucket hits the same skip with no ignore rule.) |
| <a id="gt-s3t-041"></a>**GT-S3T-041** | TableBucket, Table | codegen-artifact | Several shapes have a member literally named `type` (TableBucketType/TableType enum customer\|aws, SchemaField.type). `type` is a Go keyword, so the code generator emits the JSON tag `type_` for such members (visible on the untouched TableBucketSummary/TableSummary shapes). | go_tag: json:"type,omitempty" overrides on Table.Type, TableBucket.Type and Metadata.Iceberg.Schema.Fields.Type so the CRD exposes `type`. |
| <a id="gt-s3t-042"></a>**GT-S3T-042** | TableBucket, Namespace, Table | codegen-artifact | Every Get output carries ownerAccountId (and GetNamespace/GetTable also createdBy/modifiedBy account ids). ACK does not surface ownerAccountId as a status field; the generated code instead stamps the controller's own account into Status.ACKResourceMetadata.OwnerAccountID. | Generated setStatusDefaults fills ACKResourceMetadata.OwnerAccountID from rm.awsAccountID; no `from:` field is declared for ownerAccountId. e2e asserts status.ackResourceMetadata.ownerAccountID ("the ownerAccountID of an ACK-managed resource lives under the common ackResourceMetadata block"). |
| <a id="gt-s3t-043"></a>**GT-S3T-043** | TableBucket, Namespace, Table | shape-mismatch | Create responses are minimal and differ from the Get responses: CreateTableBucket returns only arn; CreateNamespace echoes tableBucketARN + namespace list; CreateTable returns only tableARN + versionToken. Server-assigned attributes (createdAt, createdBy, modifiedAt, tableBucketId, namespaceId, type, warehouseLocation, metadataLocation) exist only in the Get outputs. | All such status fields are declared `from: GetTableBucket\|GetNamespace\|GetTable` (is_read_only) so they populate on the post-create read; CreateNamespace's echoed tableBucketARN is written back into Spec by the generated code. |
| <a id="gt-s3t-044"></a>**GT-S3T-044** | Table | scope | Tables have their own sub-resource APIs - Put/Get/DeleteTablePolicy, Put/GetTableMaintenanceConfiguration (icebergCompaction, icebergSnapshotManagement), GetTableMaintenanceJobStatus, Put/Get/DeleteTableReplication, Put/GetTableRecordExpirationConfiguration, GetTableEncryption/GetTableStorageClass - and table buckets have replication and metrics configuration APIs. AWS-managed (type=aws, managedByService) tables also exist and cannot be created or deleted by customers. | none - not modeled (scope:defer). The generator still emits the referenced shapes (IcebergCompactionSettings, IcebergSnapshotManagementSettings, TableMaintenanceConfigurationValue, TableRecordExpirationSettings, ReplicationDestination, ...) into apis/v1alpha1/types.go even though no resource field uses them. |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `service.md` (s3tables): unverified, single-run or untestable items worth a future probe.
One line each, with the finding id that raised it.

- [S3T-SERVICE-003](#s3t-service-003) vs [S3T-SERVICE-013](#s3t-service-013): -003 records "clock skew -20 min -> HTTP 200" while -013 rejects X-Amz-Date
  shifted by -5.5/-6 min; the two runs do not agree on what a -20 min shift does - re-probe with the shift applied to
  the signed X-Amz-Date and to the Date header separately to settle which header -003 shifted.
- [S3T-SERVICE-027](#s3t-service-027): "IAM-wide propagation, not an s3tables cache" rests on s3 and s3tables flipping in the same second
  in two cycles (n=2, confidence medium); more cycles and a third service would firm it up.
- [S3T-SERVICE-012](#s3t-service-012) / -028 / -027: the allow window spread 86.1-100.8 s and the revoke window 95.3-102.0 s over n=5 with
  no identified cause for the ~15 s spread between runs; whether a fresh STS session vs an existing one matters was not
  isolated.
- [S3T-SERVICE-025](#s3t-service-025): the AssumeRole flap (6 x OK, then AccessDenied x 2-3, then OK) and the one-off
  UnrecognizedClientException on credentials issued ~2 s after the first OK are n=1 each and were seen for a role
  re-created under a just-deleted name; not reproduced for a brand-new name.
- [S3T-SERVICE-004](#s3t-service-004) / -008: ~16 accepted calls per burst is from one burst each; the refill rate, and whether
  CreateTable/Tag*/Put* writes draw from the same account-wide budget as CreateNamespace/DeleteNamespace, are
  unmeasured; [S3T-SERVICE-018](#s3t-service-018) saw 429 on only 6 operations (reads never, except one metrics Get).
- [S3T-SERVICE-006](#s3t-service-006): session-token tail edits accepted (one request per variant); untested with the Go SDK's token
  handling and with more than one appended character.
- [S3T-SERVICE-020](#s3t-service-020): "3 retries is enough for cause 1" comes from bursts of 10-20 concurrent writers; no 500 was ever
  observed under serialized writes, so the recurrence rate of a 500 in a quiet state is unknown.
- [S3T-SERVICE-022](#s3t-service-022): "Access Denied" for CreateTable WITH metadata without s3tables:PutTableData is n=2 and
  "Cross-account pass role is not allowed." is n=1 plus one live variant - small samples for two terminal-vs-retry
  decisions.
- [S3T-SERVICE-014](#s3t-service-014): the metrics-configuration 404 after Delete was observed on the first read in only 2/3 repetitions
  (one stale 200, per the row citing [S3T-TABLEBUCKET-095](tablebucket-maintenance-metrics.md#s3t-tablebucket-095)); the only eventual-consistency hint in the whole 404
  catalogue and worth a dedicated poll.
- [S3T-SERVICE-017](#s3t-service-017): the 405 was returned by every creatable table; the table type that passes the gate was never
  produced, so the "scope signal" reading is inferred from the check order, not from a positive case.
- [S3T-SERVICE-001](#s3t-service-001)..028: every service-wide finding carries regions: [us-west-2]; none of the catalogues, the IAM windows
  or the throttling budget were re-run in us-east-1.
- [S3T-SERVICE-009](#s3t-service-009): values above the modeled maximum (1001) for maxTables/maxNamespaces/maxBuckets are accepted with
  200; the page size actually returned for such values was not recorded.
<!-- preserved:end -->

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
