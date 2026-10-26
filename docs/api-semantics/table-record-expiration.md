<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table record expiration configuration (scope investigation)
_Put/GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus: MethodNotAllowedException 405 on every creatable table, the check order before the 405 gate, ValidationException on days < 1; scope verdict (not implementable today)._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
<!-- Overview seed for table-record-expiration.md (s3tables). Hand-written; preserved across renders. Findings are canonical. -->

### Orientation

This document covers the three record-expiration operations on a Table - PutTableRecordExpirationConfiguration,
GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus - and is the scope investigation for a
would-be TableRecordExpirationConfiguration CRD. Scope verdict: not implementable today (`scope_verdict: defer`). On
every table reachable from the lab account - customer ICEBERG tables with and without an Iceberg schema in us-west-2 and
us-east-1, a replication replica (`managedByService=replication.s3tables.amazonaws.com`) and the AWS-managed S3 Metadata
journal table (`type=aws`) - every well-formed Put (days 1/7/365, status enabled/disabled; five value shapes on customer
tables) and Get
answers MethodNotAllowedException 405 'Table record expiration configuration API calls to this table are not allowed',
and GetJobStatus answers the same 405 on customer tables and the journal table, synchronously and from 0.09 s after
CreateTable ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-147](#s3t-table-147), [S3T-TABLE-146](#s3t-table-146),
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009)). There is no state machine to model: no Put was ever accepted, so round-trip,
replace-vs-merge and the job lifecycle (NotYetRun/Successful/Disabled) are unobservable, and GetTable carries no
record-expiration member to read back ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004),
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLE-148](#s3t-table-148)). What IS
observable is the check order in front of the 405 gate: ARN shape (400) -> IAM authorization (403) -> value validation
(ValidationException 400) / table existence (404) -> table eligibility (405) ([S3T-TABLE-190](#s3t-table-190),
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002)). For the CRD to become implementable the service would have to accept these
calls on a table a controller can create; per the findings' notes the documented round trip exists only on S3 Storage
Lens / SageMaker Catalog managed tables, which this account cannot create and which were not probed
([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLE-147](#s3t-table-147)).

### Rules

- Treat MethodNotAllowedException 405 'Table record expiration configuration API calls to this table are not allowed' as
  terminal and never requeue on it: it is the answer to every well-formed Put/Get/GetJobStatus on every customer table
  probed (two regions, with/without schema, days 1/7/365, enabled/disabled, five value shapes), it is already there
  0.09 s after CreateTable and it does not change after rejected Puts ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001),
  [S3T-TABLE-147](#s3t-table-147), [S3T-TABLE-146](#s3t-table-146)).
- Do not poll GetTableRecordExpirationJobStatus for a transition: 90/90 calls (30 sequential, 2 x 30 concurrent,
  repeated after 20 s) on a fresh customer table -> 405 with no 429; 10/10 Gets over 10 s -> 405, neither HTTP 200 with
  status=disabled nor NotFoundException ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-146](#s3t-table-146),
  [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007)).
- Do not use `managedByService` (or `type`) to decide eligibility: the identical 405 code and message come back for
  customer tables (member absent), for a replication replica (also 2 min after DeleteTableReplication) and for the
  `type=aws` journal table ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007)).
- Tell 'gone' from 'not applicable': a never-existing table UUID under a real bucket and a table right after DeleteTable
  (HTTP 204) answer NotFoundException 404 'The specified table does not exist.' for Put, Get and GetJobStatus; the 405 is
  only returned for an existing table ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-170](#s3t-table-170)).
- With own credentials the order is value validation -> existence -> eligibility: days=0/-1 and status 'Disabled',
  'ENABLED' or '' get ValidationException 400 on an existing customer table (instead of 405) and on a ghost UUID (instead
  of 404); status 'Enabled' -> the same 400 on a customer table ([S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-170](#s3t-table-170),
  [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002)).
- Under an explicit IAM Deny the order is ARN shape -> authorization -> validation/lookup -> 405: a denied caller gets
  AccessDeniedException 403 (undeclared; the model lists ForbiddenException) on the existing table, on a ghost UUID and
  with days=0, while the bucket ARN as tableArn -> BadRequestException 400 'The specified table ARN is not valid.' for
  both callers; a 403 therefore says nothing about table existence or eligibility ([S3T-TABLE-189](#s3t-table-189), [S3T-TABLE-190](#s3t-table-190)).
- Send `status` exactly as `enabled`/`disabled`: 'Disabled', 'ENABLED', 'Enabled' or '' -> ValidationException 400
  'Member must satisfy enum value set: [enabled, disabled]', a wire code the model does not declare (it lists
  BadRequestException); days as a JSON string "7" -> BadRequestException 400 'Invalid input...' ([S3T-TABLE-169](#s3t-table-169),
  [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)).
- Do not rely on the service to validate the value shape: {status: enabled} without settings, {status: enabled,
  settings: {}}, {} (empty value), a value without status, and days 36500 / 2147483647 / 2147483648 / 1.5 all pass the
  pre-table validation and reach the 405 gate (12 of 18 variants); a 405 proves only 'not rejected before the table
  check', not acceptance ([S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)).
- A tableArn from the other region -> BadRequestException 400 'The provided Region does not match the Region of the
  resource.': for GetTableRecordExpirationConfiguration in both directions, never 404 and never 200 even when a
  same-named table exists in the endpoint's own region ([S3T-TABLE-243](#s3t-table-243)); the same 400 for a us-west-2 table ARN sent to
  the us-east-1 endpoint in the Put/Get/GetJobStatus probe ([S3T-TABLE-147](#s3t-table-147)).
- KMS key state is not the cause of the 405: with the key grant revoked, the key Disabled or PendingDeletion,
  GetTableRecordExpirationJobStatus -> 405 exactly as at the fully-granted baseline, while only CreateTable in that
  bucket -> BadRequestException 400 ([S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-240](#s3t-table-240), [S3T-TABLE-241](#s3t-table-241)).
- A rejected Put has no side effect on the parent: GetTable versionToken/modifiedAt/modifiedBy are identical at 0 s and
  5 s after three rejected Puts on a customer table and after one on the journal table; the GetTable body has no
  record-expiration member ([S3T-TABLE-148](#s3t-table-148), [S3T-TABLE-222](#s3t-table-222)).
- Rapid repeats are deterministic: 10 alternating Put(enabled days=3)/Put(disabled) in 0.12 s -> 10 x 405, no
  ConflictException, no TooManyRequestsException ([S3T-TABLE-171](#s3t-table-171)).

### Timing ranges

This stem has no state-machine durations: no Put was ever accepted, so there is no CREATING-style wait, no job run and no
read-after-write window to time ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004),
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008)). What was measured is the latency and stability of the rejection:

- Rejected Put/Get/GetJobStatus latency 8-50 ms, n=15 across three customer tables
  ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001)); rejected Put latency 10-12 ms, n=9 ([S3T-TABLE-147](#s3t-table-147)).
- GetTableRecordExpirationJobStatus sequential latency 10-13 ms (30 calls); 30 concurrent calls complete in 0.05-0.13 s
  wall, n=90 total ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003)).
- 10 back-to-back Puts: latency 10-17 ms each, 0.12 s total, n=10 ([S3T-TABLE-171](#s3t-table-171)).
- No wait is ever rewarded: the 405 is present from the first Get 0.09 s after CreateTable through 10 s of 1 s polling,
  n=10 ([S3T-TABLE-146](#s3t-table-146), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007)), after a 20 s pause, n=90
  ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003)), and for GetJobStatus both at the fully-granted KMS baseline and right
  after (15 s) each of three key-state changes ([S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-240](#s3t-table-240), [S3T-TABLE-241](#s3t-table-241)).

### Handling gaps

No finding in this document is marked HANDLING-GAP (`suspect-bug`); all 23 are `handling: unhandled`. The controller
implements no record-expiration resource or field today: generator.yaml defines Table, Namespace and TableBucket only
and names no record-expiration operation. What the findings imply for a future implementation:

- Both scope findings carry `scope_verdict: defer`: a TableRecordExpirationConfiguration resource for customer tables
  would be permanently terminal-errored ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001)), and a controller exposing the
  field must treat 405 as terminal and document it as AWS-managed-tables-only ([S3T-TABLE-147](#s3t-table-147)).
- `terminal_codes` would have to include MethodNotAllowedException (declared) plus two undeclared wire codes:
  ValidationException 400 for days < 1 / non-enum status (model: BadRequestException) and AccessDeniedException 403 for a
  denied principal (model: ForbiddenException) ([S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-189](#s3t-table-189), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)).
- `exceptions.404`/find logic must map NotFoundException 404 to 'resource gone' and 405 to 'not applicable'; they are
  distinct answers for the same ARN shape ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLE-146](#s3t-table-146)).
- No parent read-back exists: GetTable has no record-expiration member and a rejected Put leaves versionToken unchanged,
  so a Table-level read cannot detect the configuration ([S3T-TABLE-148](#s3t-table-148), [S3T-TABLE-222](#s3t-table-222)); the only read is
  GetTableRecordExpirationConfiguration itself, which is 405 today ([S3T-TABLE-146](#s3t-table-146)).
- A CRD schema should pin `status` to the lower-case enum and `settings.days >= 1` (the service enforces both, but with
  the undeclared ValidationException 400) and decide itself about enabled-without-days, an empty value, a missing status
  and days > 2^31-1, which pass the service's pre-table validation; what an eligible table accepts is unknown
  ([S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)).
- A tableArn from the other region is BadRequestException 400 'provided Region does not match', not NotFoundException:
  a find must map that 400 neither to 'gone' nor to 'not applicable' ([S3T-TABLE-243](#s3t-table-243), [S3T-TABLE-147](#s3t-table-147)).

### Where next

- Sibling documents (render.yaml): `table.md` (Table lifecycle, lookup by ARN or triple, the 404 message shapes the
  record-expiration 404s share); `table-maintenance.md` (the sibling GetTableMaintenanceJobStatus answers HTTP 200 on the
  same tables where record expiration answers 405 - [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001); the KMS
  maintenance-principal prerequisite and Failed-job soaks - [S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-240](#s3t-table-240), [S3T-TABLE-241](#s3t-table-241));
  `table-policy.md` (IAM action requirements, AccessDeniedException vs ForbiddenException - [S3T-TABLE-189](#s3t-table-189),
  [S3T-TABLE-190](#s3t-table-190)); `table-encryption-storage.md` (KMS key-state effects on CreateTable from the same probes);
  `table-replication.md` (`managedByService`, replica tables - [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009));
  `table-metadata-schema.md` (versionToken - [S3T-TABLE-148](#s3t-table-148), [S3T-TABLE-222](#s3t-table-222)); `tablebucket-maintenance-metrics.md`
  (the bucket-level ValidationException-on-days < 1 analogue).
- `details/<id>.md` pages (entries expected to exceed the 300-char inline excerpt): the 18-variant value matrix
  ([S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-172](#s3t-table-172)), the check-order catalogues ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002),
  [S3T-TABLE-190](#s3t-table-190)), the 11-operation KMS phase tables ([S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-240](#s3t-table-240), [S3T-TABLE-241](#s3t-table-241)), the 12-read
  cross-region table ([S3T-TABLE-243](#s3t-table-243)).
- Open experiments: STATUS.md (Account state, Phase 8 result) lists none for this stem. The findings leave these
  untestable here, all needing an AWS-managed S3 Storage Lens / SageMaker Catalog table: the enabled/disabled round trip
  and replace semantics ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005)), the
  job-status lifecycle ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001)), whether
  an eligible table rejects enabled-without-days or days > 2^31-1 ([S3T-TABLE-172](#s3t-table-172),
  [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)), whether an accepted Put bumps the table versionToken ([S3T-TABLE-148](#s3t-table-148)).
  Full list: `table-record-expiration.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 24 (high 10 / medium 7 / low 7); duplicates folded into the appendix: 0
- handling: handled 0 · partial 0 · tracked 0 · unhandled 24 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 0 · last_verified: 2026-10-13..2026-10-14 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: error-code 9, other 4, prerequisite 3, scope 3, read-gap 2, cross-region 1, idempotency 1,
  request-validation 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetNamespace | read | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMaintenanceConfiguration | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMaintenanceJobStatus | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTablePolicy | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableRecordExpirationConfiguration | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableRecordExpirationJobStatus | read | tableArn | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableMaintenanceConfiguration | create | tableBucketARN, namespace, name, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableRecordExpirationConfiguration | create | tableArn, value | InternalServerErrorException, MethodNotAllowedException, ForbiddenException, NotFoundException, TooManyRequestsException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **JobStatus**: Not_Yet_Run, Successful, Failed, Disabled
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)
- **TableRecordExpirationJobStatus**: NotYetRun, Successful, Failed, Disabled
- **TableRecordExpirationStatus**: enabled, disabled

No high/medium state-machine findings.

## Errors

- <a id="s3t-table-146"></a>**S3T-TABLE-146** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Customer table: GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405, not 200/404**
  Plain ICEBERG table (CreateTable without metadata, type=customer, GetTable has no managedByService member),
  never configured: GetTableRecordExpirationConfiguration called 0.09 s after CreateTable and then every 1 s
  for 10 s -> 10/10 x MethodNotAllowedException/405; message ['Table record expiration configuration API calls
  to this table are not allowed']. GetTableRecordExpirationJobStatus on the same fresh table ->
  MethodNotAllowedException (HTTP 405) 'Table record expiration configuration API calls to this table are not
  allowed'. After three rejected Put attempts, GetTableRecordExpirationJobStatus -> MethodNotAllowedException
  (HTTP 405) 'Table record expiration configuration API calls to this table are not allowed'.
  - ACK: terminal_codes, exceptions.404, custom_find · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus · fields: configuration.status, configuration.settings.days, status
  - repro: CreateTable(format=ICEBERG) -> GetTableRecordExpirationConfiguration(tableArn) immediately and
    every 1 s for 10 s; GetTableRecordExpirationJobStatus
  - measurements: first_get_after_create_s=0.09, window_s=10, n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-221](#s3t-table-221), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLE-173](table-replication.md#s3t-table-173),
    [S3T-TABLE-174](table-replication.md#s3t-table-174), [S3T-TABLE-176](table-replication.md#s3t-table-176), [S3T-TABLE-147](#s3t-table-147) · hypotheses: H-U-101, H-U-102, H-U-121, H-U-122 · evidence:
    table/sub-resources/expiration-absent-roundtrip-jobstatus
  - notes: H-U-101 refuted (no 200/default struct), H-U-102 refuted (the code is MethodNotAllowedException,
    not NotFoundException), H-U-121 refuted (no status=Disabled body), H-U-122 refuted (not 404 either): on a
    customer-created table both reads are rejected with 405 'Table record expiration configuration...
  - full notes: [details/S3T-TABLE-146.md](details/S3T-TABLE-146.md)

- <a id="s3t-table-169"></a>**S3T-TABLE-169** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration: days<1 or status not in [enabled, disabled] -> undeclared ValidationException 400 before table lookup**
  Through a client without botocore parameter validation: value.settings.days=0 and days=-1 ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'value.settings.days' failed to
  satisfy constraint: Member must have value greater than or equal to 1'; value.status='Disabled', 'ENABLED'
  and '' -> ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to
  satisfy constraint: Member must satisfy enum value set: [enabled, disabled]'. The same values sent with a
  never-existing table UUID get the same ValidationException (not 404), and on an existing customer table they
  get it instead of the table's MethodNotAllowedException 405, so this validation runs before the table is
  resolved. days as a JSON string "7" -> BadRequestException (HTTP 400) 'Invalid input. Consult the S3 Tables
  and/or Iceberg REST Catalog API reference for constraints on request parameters.' The other 12 of 18
  variants reached the 405 gate.
  - ACK: terminal_codes, custom_update · ops: PutTableRecordExpirationConfiguration · fields:
    value.settings.days, value.status
  - repro: CreateTable (customer) -> Put each value variant through a client with parameter_validation=False
    -> record code/HTTP/message
  - measurements: variants_rejected_400=6, variants_total=18, n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLE-172](#s3t-table-172), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002) · hypotheses: H-U-112, H-U-128 ·
    evidence: table/mutation-matrix/expiration-status-days-variants
  - notes: H-U-112 qualified: days=0 is rejected with HTTP 400 and a message stating the minimum (1), but the
    wire code is ValidationException, which the model does not declare (it declares BadRequestException); the
    'stored configuration unchanged' half is unobservable on customer tables (Get -> 405). H-U-128...
  - full notes: [details/S3T-TABLE-169.md](details/S3T-TABLE-169.md)

- <a id="s3t-table-170"></a>**S3T-TABLE-170** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration on a never-existing table UUID -> NotFoundException 404 for well-formed/{} values; days=0 -> 400 first**
  tableArn = <real bucket ARN>/table/<random uuid> (never existed), region us-west-2: enabled-days-1 ->
  NotFoundException/404 'The specified table does not exist.'; enabled-days-0 -> ValidationException/400 '1
  validation error detected: Value at 'value.settings.days' failed to satisfy constraint: Member must have
  value greater than or equal to 1'; status-Disabled-pascal -> ValidationException/400 '1 validation error
  detected: Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set:
  [enabled, disabled]'; enabled-no-settings -> NotFoundException/404 'The specified table does not exist.';
  value-empty -> NotFoundException/404 'The specified table does not exist.'; status-missing ->
  NotFoundException/404 'The specified table does not exist.'.
  - ACK: exceptions.404, terminal_codes · ops: PutTableRecordExpirationConfiguration · fields: tableArn, value
  - repro: Put(value) with tableArn=<bucket ARN>/table/<uuid4> for a well-formed and several malformed values
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-189](#s3t-table-189),
    [S3T-TABLE-190](#s3t-table-190) · hypotheses: H-U-109 · evidence: table/mutation-matrix/expiration-status-days-variants
  - notes: H-U-109 (Put side) confirmed: NotFoundException (HTTP 404) 'The specified table does not exist.'
    for a well-formed value, also for {} and for a value without status; days=0 and status='Disabled' get
    ValidationException 400 instead, so value validation precedes the existence lookup....
  - full notes: [details/S3T-TABLE-170.md](details/S3T-TABLE-170.md)

- <a id="s3t-table-189"></a>**S3T-TABLE-189** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM explicit Deny on the record-expiration actions -> AccessDeniedException/403 for Put/Get/GetJobStatus, ahead of the table's 405 gate**
  Assumed role with Allow s3tables:* plus an explicit Deny on s3tables:PutTableRecordExpirationConfiguration,
  s3tables:GetTableRecordExpirationConfiguration and s3tables:GetTableRecordExpirationJobStatus. On an
  existing customer table, where own (admin) credentials get MethodNotAllowedException 405 at the same moment:
  GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus and
  PutTableRecordExpirationConfiguration(enabled, days=1) each -> AccessDeniedException (HTTP 403) with message
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<role>/<session> is not authorized to perform: s3tables:<Action>
  on resource: <table ARN>' (the action name and the full table ARN are in the message). Control calls through
  the same role: GetTable -> HTTP 200, GetTableMaintenanceConfiguration -> HTTP 200. The Deny was effective on
  the first call made after AssumeRole succeeded (AssumeRole itself needed 4 attempts over 9.1 s for the new
  role to propagate).
  - ACK: terminal_codes, requeue · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus
  - repro: CreateRole(trust: account) + inline policy Allow s3tables:* / Deny the three record-expiration
    actions -> AssumeRole -> call the three ops on an existing table
  - measurements: assume_role_attempts=4, assume_role_wait_s=9.1, deny_effective_on_first_call=true, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-190](#s3t-table-190), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002) · hypotheses: H-U-107 ·
    evidence: table/error-taxonomy/expiration-iam-deny
  - notes: H-U-107 confirmed: AccessDeniedException (HTTP 403), not the declared ForbiddenException (the model
    declares only ForbiddenException for these three operations; AccessDeniedException is undeclared for
    them). Authorization is evaluated before the table-eligibility gate: the same table, same moment,...
  - full notes: [details/S3T-TABLE-189.md](details/S3T-TABLE-189.md)

- <a id="s3t-table-190"></a>**S3T-TABLE-190** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Under an explicit Deny: ghost-UUID ARN and days=0 -> AccessDeniedException 403 (authz before lookup/validation); bucket ARN -> 400 first**
  Same denied role: Put(enabled, days=1) on <bucketARN>/table/<random uuid> -> AccessDeniedException (HTTP
  403); Get on that ghost ARN -> AccessDeniedException 403 (own credentials: NotFoundException 404 'The
  specified table does not exist.'). Put with value.settings.days=0 on the existing table ->
  AccessDeniedException 403 (own credentials: ValidationException 400 about the minimum). Get with the table
  BUCKET ARN as tableArn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.' under the
  denied role too (own credentials: the same 400), so the ARN-shape check runs before authorization while the
  existence lookup and the days range validation run after it.
  - ACK: terminal_codes · ops: PutTableRecordExpirationConfiguration, GetTableRecordExpirationConfiguration ·
    fields: tableArn, value.settings.days
  - repro: Denied role -> Put/Get with a ghost UUID ARN, with days=0 (parameter_validation=False client), with
    the bucket ARN
  - measurements: n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-189](#s3t-table-189), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-243](#s3t-table-243),
    [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179) · hypotheses: H-U-107, H-U-109, H-U-112 · evidence:
    table/error-taxonomy/expiration-iam-deny
  - notes: Ordering observed: ARN-shape validation (400) -> IAM authorization (403) -> value validation
    (ValidationException 400) / table lookup (404) -> table eligibility gate (405). Compare with the
    own-credentials codes in table/mutation-matrix/expiration-status-days-variants and...
  - full notes: [details/S3T-TABLE-190.md](details/S3T-TABLE-190.md)

- <a id="s3t-tablerecordexpirationconfiguration-002"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-002** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Record-expiration check order (customer table): enum typo -> ValidationException 400, ghost/deleted table -> 404, bad ARN -> 400, then 405**
  On a customer table (where every valid call is MethodNotAllowedException 405): Put days=0 and days=-1 are
  rejected CLIENT-SIDE by botocore (ParamValidationError 'valid min value: 1', the model's minimum) and never
  reach the service; status='Enabled' (capitalized) -> ValidationException 400 '1 validation error detected:
  Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set: [enabled,
  disabled]'. Nonexistent table id under the same bucket: GetTableRecordExpirationJobStatus ->
  NotFoundException 404 'The specified table does not exist.'; GetTableRecordExpirationConfiguration ->
  NotFoundException 404; Put -> NotFoundException 404. Malformed table ARN (non-uuid id) ->
  BadRequestException 400 'The specified table ARN is not valid.'. Right after DeleteTable (HTTP 204): job
  status -> NotFoundException 404 'The specified table does not exist.'; Put -> NotFoundException 404; config
  -> NotFoundException 404.
  - ACK: exceptions.404, terminal_codes · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus, GetTableRecordExpirationConfiguration, DeleteTable · fields: tableArn,
    value.settings.days, value.status
  - repro: customer table: Put days=0/-1/'Enabled'; random table uuid; malformed ARN; DeleteTable then Get/Put
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLE-172](#s3t-table-172), [S3T-TABLE-189](#s3t-table-189), [S3T-TABLE-190](#s3t-table-190), [S3T-TABLE-243](#s3t-table-243),
    [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179) · hypotheses: H-U-127 · evidence:
    tablerecordexpirationconfiguration/state-machine/job-status-soak
  - notes: Check order observed: input enum validation (ValidationException 400, a wire code that is NOT in
    the operation's modeled error list) -> table existence (NotFoundException 404) -> 'allowed for this table'
    (MethodNotAllowedException 405). A controller can therefore tell 'gone' (404) from 'not...
  - full notes: [details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-002.md](details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-002.md)

- <a id="s3t-tablerecordexpirationconfiguration-003"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-003** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableRecordExpirationJobStatus on a fresh customer table -> MethodNotAllowedException 405 in 90/90 calls (30 seq + 2x30 conc.), no 429**
  On a fresh table with no record-expiration configuration: 30 sequential
  GetTableRecordExpirationJobStatus(tableArn) with no pause -> {"MethodNotAllowedException/405": 30} (latency
  [10, 13] ms); 30 concurrent (10 threads) -> {"MethodNotAllowedException/405": 30} (wall 0.13s), repeated
  after 20 s -> {"MethodNotAllowedException/405": 30} (wall 0.05s). Messages: {"MethodNotAllowedException":
  "Table record expiration configuration API calls to this table are not allowed"}.
  - ACK: terminal_codes, requeue · ops: GetTableRecordExpirationJobStatus · fields: tableArn
  - repro: 30 x GetTableRecordExpirationJobStatus(tableArn) back-to-back; 10 threads x 30; repeat after 20 s
  - measurements: sequential_latency_ms_min_max=[10, 13], burst_wall_s=[0.13, 0.05], n=90
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-147](#s3t-table-147),
    [S3T-TABLE-171](#s3t-table-171), [S3T-TABLE-222](#s3t-table-222), [S3T-TABLE-148](#s3t-table-148), [S3T-TABLE-187](table-maintenance.md#s3t-table-187) · hypotheses: H-U-130 · evidence:
    table/limits/default-page-120-tables
  - notes: H-U-130 untested for the 429 question: on a freshly created ICEBERG table with no record-expiration
    configuration every call (30 sequential, 2 x 30 concurrent) was rejected with MethodNotAllowedException
    (HTTP 405) 'Table record expiration configuration API calls to this table are not allowed' - a...
  - full notes: [details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-003.md](details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-003.md)

- <a id="s3t-tablerecordexpirationconfiguration-006"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-006** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C101 PARTLY: Put days=0/-1 -> ValidationException 400 (not BadRequestException), pre-lookup; enabled w/o days reaches the 405 gate**
  days=0 and days=-1 -> ValidationException (HTTP 400; undeclared - the model lists BadRequestException) '1
  validation error detected: Value at 'value.settings.days' failed to satisfy constraint: Member must have
  value greater than or equal to 1', returned before the table lookup (same answer for a never-existing table
  UUID). {status: enabled} without settings and {status: enabled, settings: {}} PASS that validation and reach
  the table gate (MethodNotAllowedException 405 on customer tables), so 'enabled without a day count' is not a
  bad request; whether an eligible AWS-managed table would reject it is unobservable. botocore rejects days<1
  client-side (ParamValidationError).
  - ACK: docs-only · ops: PutTableRecordExpirationConfiguration · fields: value.settings.days, value.status
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005) · evidence:
    service/static/doc-claims-3
  - notes: VERDICT: PARTLY - days=0 is rejected (but as ValidationException, not BadRequestException);
    enabled-without-days is not rejected by request validation.

- <a id="s3t-tablerecordexpirationconfiguration-007"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-007** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C102 FALSE: GetTableRecordExpirationConfiguration on a never-configured table -> MethodNotAllowedException 405 (n=10), not 200/404**
  On a fresh customer ICEBERG table (never configured) GetTableRecordExpirationConfiguration ->
  MethodNotAllowedException (HTTP 405) 'Table record expiration configuration API calls to this table are not
  allowed' in 10/10 calls over 10 s starting 0.09 s after CreateTable - neither HTTP 200 with
  configuration.status=disabled nor NotFoundException. The AWS-managed S3 Metadata journal table (type=aws)
  answers the same 405. 'Record which': MethodNotAllowedException 405.
  - ACK: docs-only, terminal_codes · ops: GetTableRecordExpirationConfiguration · fields: configuration.status
  - repro: Settled from the related findings (their probes are the evidence); no new API call in this probe.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-146](#s3t-table-146), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-221](#s3t-table-221),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-147](#s3t-table-147), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225) · evidence:
    service/static/doc-claims-3
  - notes: VERDICT: FALSE - Neither of the two documented answers: the read is 405 on every reachable table.

## Request validation

- <a id="s3t-table-172"></a>**S3T-TABLE-172** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration pre-table validation passes enabled-w/o-settings, {} value, no status, days=2^31, 1.5 to the 405 gate**
  On a customer ICEBERG table (where any well-formed Put -> MethodNotAllowedException 405 'Table record
  expiration configuration API calls to this table are not allowed') these 12 of 18 value variants were
  answered by the 405 table gate, i.e. passed the service's request validation: {status: enabled, settings:
  {days: 1|36500|2147483647|2147483648|1.5}}, {status: enabled} (no settings), {status: enabled, settings:
  {}}, {status: disabled, settings: {days: 7}}, {status: disabled}, {status: disabled, settings: {}},
  {settings: {days: 3}} (no status) and {} (empty value). The only variants rejected ahead of the gate were
  days<1, a non-enum status and a string-typed days (HTTP 400, see the ValidationException finding). The Get
  after all variants -> MethodNotAllowedException 405.
  - ACK: custom_update, none · ops: PutTableRecordExpirationConfiguration · fields: value, value.status,
    value.settings, value.settings.days
  - repro: CreateTable (customer) -> Put each value variant through a client with parameter_validation=False
    -> record code/HTTP/message
  - measurements: variants_passed_to_gate=12, variants_total=18, n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](#s3t-table-169), [S3T-TABLE-147](#s3t-table-147), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002) ·
    hypotheses: H-U-114, H-U-113, H-U-115, H-U-116, H-U-117 · evidence:
    table/mutation-matrix/expiration-status-days-variants
  - notes: H-U-114 qualified: 'enabled without settings' is NOT rejected by the request validation that runs
    before the table check (it reaches the 405 gate); whether an eligible (AWS-managed) table would reject it
    is unknown. H-U-113 qualified: days=2147483647 and 2147483648 both pass request validation (no...
  - full notes: [details/S3T-TABLE-172.md](details/S3T-TABLE-172.md)

## Update granularity and ordering

- <a id="s3t-table-239"></a>**S3T-TABLE-239** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Key grant revoked after creation: only CreateTable -> 400; existing-table reads/Put config/Tag 200/204; 3 maintenance jobs Failed in 8 min**
  Phase 'revoked' (DescribeKey.KeyState 'Enabled'). Right after the key change (15 s): GetTable -> HTTP 200;
  GetTableMetadataLocation -> HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> HTTP 204;
  GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405
  'Table record expiration configuration API calls to this table are not allowed'; ListTables -> HTTP 200;
  CreateTable -> BadRequestException 400 'Insufficient access to perform table maintenance. Ensure that the
  key used to encrypt the table still exists, is active, and has a resource policy granting acce'; TagResource
  -> HTTP 200. Baseline (key fully granted) for comparison: GetTable -> HTTP 200; GetTableMetadataLocation ->
  HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> TooManyRequestsException
  429; GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus ->
  MethodNotAllowedException 405; ListTables -> HTTP 200; CreateTable -> HTTP 200; TagResource -> HTTP 200.
  Maintenance job status sampled every 60 s for 20 min: start
  'icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Not_Yet_Run',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions [[302,
  "icebergCompaction=Not_Yet_Run;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Failed+ts+msg"],
  [362,
  "icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Not_Yet_Run;icebergUnreferencedFileRemoval=Failed+ts+msg"],
  [483,
  "icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg"]].
  Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> PutKeyPolicy(root only) -> the read/write set ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: soak_window_min=[20.0], n=21
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-240](#s3t-table-240), [S3T-TABLE-241](#s3t-table-241), [S3T-TABLE-253](table-maintenance.md#s3t-table-253), [S3T-TABLE-301](table-maintenance.md#s3t-table-301), [S3T-TABLE-254](table-maintenance.md#s3t-table-254), [S3T-TABLE-246](table-maintenance.md#s3t-table-246),
    [S3T-TABLE-248](table-encryption-storage.md#s3t-table-248), [S3T-TABLE-054](table-maintenance.md#s3t-table-054) · hypotheses: H-U-126 · evidence: table/dependencies/kms-key-revoked-jobs
  - notes: H-U-126 qualified: a table cannot be created on a key without the maintenance grant (CreateTable
    rejects, [S3T-TABLE-054](table-maintenance.md#s3t-table-054)), so the state was reached by revoking the grant AFTER creation. Existing-table
    operations (GetTable, GetTableMetadataLocation, GetTableEncryption, GetTableBucketEncryption,...
  - full notes: [details/S3T-TABLE-239.md](details/S3T-TABLE-239.md)

- <a id="s3t-table-240"></a>**S3T-TABLE-240** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **KMS key Disabled: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 15 min**
  Phase 'disabled' (DescribeKey.KeyState 'Disabled'). Right after the key change (15 s): GetTable -> HTTP 200;
  GetTableMetadataLocation -> HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> HTTP 204;
  GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405
  'Table record expiration configuration API calls to this table are not allowed'; ListTables -> HTTP 200;
  CreateTable -> BadRequestException 400 'Insufficient access to perform table maintenance. Ensure that the
  key used to encrypt the table still exists, is active, and has a resource policy granting acce'; TagResource
  -> HTTP 200. Baseline (key fully granted) for comparison: GetTable -> HTTP 200; GetTableMetadataLocation ->
  HTTP 200; GetTableEncryption -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  GetTableMaintenanceConfiguration -> HTTP 200; PutTableMaintenanceConfiguration -> TooManyRequestsException
  429; GetTableMaintenanceJobStatus -> HTTP 200; GetTableRecordExpirationJobStatus ->
  MethodNotAllowedException 405; ListTables -> HTTP 200; CreateTable -> HTTP 200; TagResource -> HTTP 200.
  Maintenance job status sampled every 60 s for 15 min: start
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions []. Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> DisableKey -> the read/write set -> GetTableMaintenanceJobStatus every 60
    s
  - measurements: soak_window_min=[15.0], n=17
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-241](#s3t-table-241) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: H-U-126 qualified (key disabled variant). DisableKey changed nothing on the control plane compared
    with the revoked-grant phase: the same 11 operations gave the same codes, CreateTable's 400 message is the
    same 'Insufficient access to perform table maintenance ... key ... still exists, is active...
  - full notes: [details/S3T-TABLE-240.md](details/S3T-TABLE-240.md)

- <a id="s3t-table-241"></a>**S3T-TABLE-241** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **KMS key PendingDeletion: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 10 min**
  Phase 'pending-deletion' (DescribeKey.KeyState 'PendingDeletion'). Right after the key change (15 s):
  GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; GetTableEncryption -> HTTP 200;
  GetTableBucketEncryption -> HTTP 200; GetTableMaintenanceConfiguration -> HTTP 200;
  PutTableMaintenanceConfiguration -> HTTP 204; GetTableMaintenanceJobStatus -> HTTP 200;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405 'Table record expiration configuration
  API calls to this table are not allowed'; ListTables -> HTTP 200; CreateTable -> BadRequestException 400
  'Insufficient access to perform table maintenance. Ensure that the key used to encrypt the table still
  exists, is active, and has a resource policy granting acce'; TagResource -> HTTP 200. Baseline (key fully
  granted) for comparison: GetTable -> HTTP 200; GetTableMetadataLocation -> HTTP 200; GetTableEncryption ->
  HTTP 200; GetTableBucketEncryption -> HTTP 200; GetTableMaintenanceConfiguration -> HTTP 200;
  PutTableMaintenanceConfiguration -> TooManyRequestsException 429; GetTableMaintenanceJobStatus -> HTTP 200;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405; ListTables -> HTTP 200; CreateTable ->
  HTTP 200; TagResource -> HTTP 200. Maintenance job status sampled every 60 s for 10 min: start
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  end
  'icebergCompaction=Failed+ts+msg;icebergSnapshotManagement=Failed+ts+msg;icebergUnreferencedFileRemoval=Failed+ts+msg',
  transitions []. Errors 0. No re-run (lastRunTimestamp never advanced after the first run).
  - ACK: terminal_codes, requeue, references · ops: CreateTable, GetTable, GetTableBucketEncryption,
    GetTableEncryption, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetTableMetadataLocation, GetTableRecordExpirationJobStatus, ListTables, PutTableMaintenanceConfiguration,
    TagResource · fields: encryptionConfiguration.kmsKeyArn, status.*.status, status.*.failureMessage
  - repro: SSE-KMS bucket + table -> ScheduleKeyDeletion(7) -> the read/write set ->
    GetTableMaintenanceJobStatus every 60 s
  - measurements: soak_window_min=[10.0], n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-239](#s3t-table-239), [S3T-TABLE-240](#s3t-table-240) · hypotheses: H-U-126 · evidence:
    table/dependencies/kms-key-revoked-jobs
  - notes: H-U-126 qualified (key pending deletion variant). ScheduleKeyDeletion(7 days) changed nothing on
    the control plane versus Disabled: same codes for the 11 operations, CreateTable 400 with the same
    message. The Failed entries did not change and no job re-ran in 10 min. Total time the Failed entries...
  - full notes: [details/S3T-TABLE-241.md](details/S3T-TABLE-241.md)

## Cross-region

- <a id="s3t-table-243"></a>**S3T-TABLE-243** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **All 12 table/namespace reads with the other region's bucket ARN or tableArn -> 400 'provided Region does not match', uniform, both ways**
  A same-named bucket/namespace/table exists in BOTH us-west-2 and us-east-1. us-west-2 endpoint with the
  us-east-1 identifiers: GetTable(triple), GetTableMetadataLocation, GetTableStorageClass, GetTableEncryption,
  GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetNamespace, ListTables,
  GetTable(tableArn), GetTableRecordExpirationConfiguration, ListTagsForResource -> BadRequestException (HTTP 400)
  'The provided Region does not match the Region of the resource.'. us-east-1 endpoint with the us-west-2
  identifiers: GetTable(triple), GetTableMetadataLocation, GetTableStorageClass, GetTableEncryption,
  GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetNamespace, ListTables,
  GetTable(tableArn), GetTableRecordExpirationConfiguration, ListTagsForResource -> BadRequestException (HTTP 400)
  'The provided Region does not match the Region of the resource.'. No read returned 200 (neither the other
  region's table nor this region's same-named table) and none returned NotFoundException.
  - ACK: terminal_codes, is_arn_primary_key · ops: GetTable, GetTableMetadataLocation, GetTableStorageClass,
    GetTableEncryption, GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetNamespace, ListTables, GetTableRecordExpirationConfiguration, ListTagsForResource · fields:
    tableBucketARN, tableArn, resourceArn
  - repro: Create bucket+ns+table with identical names in us-west-2 and us-east-1; from each endpoint call
    every table read with the OTHER region's ARNs
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-190](#s3t-table-190) ·
    hypotheses: H-T-213, H-T-214, H-T-215, H-U-069, H-U-106, H-T-075 · evidence:
    table/cross-region/foreign-region-arn-reads-writes
  - notes: H-T-213 confirmed (GetTableMetadataLocation -> BadRequestException, not 404); H-T-214 confirmed
    (GetTableStorageClass/Encryption/MetadataLocation/GetTable share the exact wire code and message); H-T-215
    (contrarian: read-only ops skip the region check -> 404 or region-A body) refuted: no 200 and no...
  - full notes: [details/S3T-TABLE-243.md](details/S3T-TABLE-243.md)

## Scope

- <a id="s3t-table-147"></a>**S3T-TABLE-147** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration on customer tables -> 405 for days 1/7/365, with/without Iceberg schema, in us-west-2 and us-east-1**
  **Scope verdict: defer**
  Put(enabled, days=1/7/365) on a plain customer ICEBERG table in us-west-2 ->
  ['MethodNotAllowedException/405'], message ['Table record expiration configuration API calls to this table
  are not allowed']; the Get right after each Put -> ['MethodNotAllowedException/405']. Table created with
  metadata.iceberg.schema (int id + two timestamp columns named report_time/snapshot_time; GetTable has no
  managedByService member): GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405);
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  PutTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405). Plain table in us-east-1:
  GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405);
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  PutTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405). A us-west-2 table ARN sent to
  the us-east-1 endpoint -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.'.
  - ACK: scope:defer, terminal_codes, docs-only · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: value.status,
    value.settings.days
  - repro: CreateTable (with and without iceberg schema; in us-west-2 and us-east-1) -> Put(enabled, days=1) /
    Get / GetJobStatus by tableArn
  - measurements: put_latency_ms=[12, 12, 10], n=9
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-173](table-replication.md#s3t-table-173), [S3T-TABLE-174](table-replication.md#s3t-table-174), [S3T-TABLE-176](table-replication.md#s3t-table-176), [S3T-TABLE-146](#s3t-table-146),
    [S3T-TABLE-221](#s3t-table-221), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225) · hypotheses: H-U-111, H-U-108 · evidence:
    table/sub-resources/expiration-absent-roundtrip-jobstatus
  - notes: H-U-111 refuted on customer tables: the Put never returns 200/204, so no round-trip is observable;
    the 405 is table-type gated, not region- or schema-gated (3 tables, 2 regions). H-U-108 qualified: 405 is
    not specific to service-managed tables - it is the answer for every customer table too...
  - full notes: [details/S3T-TABLE-147.md](details/S3T-TABLE-147.md)

- <a id="s3t-table-221"></a>**S3T-TABLE-221** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **AWS-managed S3 Metadata journal table (type=aws): record-expiration Get, Put and GetJobStatus all -> MethodNotAllowedException 405**
  **Scope verdict: defer**
  CreateBucketMetadataConfiguration(JournalTableConfiguration.RecordExpiration ENABLED, Days=7) on a
  general-purpose bucket created a journal table in the account's 'aws-s3' table bucket (namespace 'b_<bucket
  name>', table 'journal'); TableStatus CREATING -> ACTIVE in 15.2 s. GetTable(tableArn): type=aws,
  managedByService='metadata.s3.amazonaws.com' (a string naming the owning service, absent on customer
  tables), no managedTableInformation, format=ICEBERG. GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus and PutTableRecordExpirationConfiguration(enabled, days=1) each ->
  MethodNotAllowedException (HTTP 405) 'Table record expiration configuration API calls to this table are not
  allowed' - the identical code and message a customer table gets. The journal table's own RecordExpiration
  stays visible only through s3 GetBucketMetadataConfiguration ({Expiration: ENABLED, Days: 7}, unchanged by
  the rejected Put). GetTableMaintenanceConfiguration on the journal table -> HTTP 200.
  - ACK: scope:defer, terminal_codes, docs-only · ops: GetTableRecordExpirationConfiguration,
    PutTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus, GetTable · fields:
    configuration.status, configuration.settings.days, managedByService, type, managedTableInformation
  - repro: s3 CreateBucket -> CreateBucketMetadataConfiguration(JournalTable RecordExpiration ENABLED 7) ->
    wait ACTIVE -> s3tables GetTable/Get*/Put on the journal TableArn
  - measurements: journal_table_creating_s=[15.2], get_after_put_s=0.01, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](#s3t-table-146), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLE-031](table-metadata-location.md#s3t-table-031),
    [S3T-TABLE-317](table.md#s3t-table-317), [S3T-TABLE-136](table-replication-status.md#s3t-table-136), [S3T-TABLE-223](table-replication-status.md#s3t-table-223), [S3T-TABLE-177](table-replication.md#s3t-table-177), [S3T-TABLE-147](#s3t-table-147) · hypotheses: H-U-108, H-U-101,
    H-U-111, H-U-121, H-U-122 · evidence: table/creative/expiration-aws-managed-journal-table
  - notes: H-U-108 refuted for this service-managed table: Put AND Get are both 405 (the hypothesis expected
    Get -> 200 on a managedByService table). Combined with the customer-table result the s3tables
    record-expiration API answered 405 on every table reachable from a lab account (customer plain,
    customer...
  - full notes: [details/S3T-TABLE-221.md](details/S3T-TABLE-221.md)

- <a id="s3t-tablerecordexpirationconfiguration-001"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-001** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Customer-created tables: every well-formed record-expiration Put/Get/JobStatus call -> MethodNotAllowedException 405 'not allowed'**
  **Scope verdict: defer**
  On three fresh ICEBERG tables created by the caller (with schema, without metadata; GetTable shows
  type=customer, managedByService=None): GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration({enabled, days 1}), ({disabled,
  days 1}) and ({disabled} without settings) -> MethodNotAllowedException/405 with message ["Table record
  expiration configuration API calls to this table are not allowed"]; raw error body "{\"message\":\"Table
  record expiration configuration API calls to this table are not allowed\"}". The sibling
  GetTableMaintenanceJobStatus on the same tables -> HTTP 200, HTTP 200. Per the S3 user guide
  (s3-tables-record-expiration) record expiration is available only for AWS managed tables (S3 Storage Lens,
  SageMaker Catalog), 'not for S3 tables that you create'.
  - ACK: scope:defer, terminal_codes, docs-only · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: tableArn, value.status,
    value.settings.days, status, metrics
  - repro: CreateTable(ICEBERG) -> any of the three record-expiration operations with the table ARN
  - measurements: latency_ms=[50, 14, 10, 10, 10, 33, 10, 9, 10, 10, 32, 12, 8, 10, 10], n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-147](#s3t-table-147),
    [S3T-TABLE-171](#s3t-table-171), [S3T-TABLE-222](#s3t-table-222), [S3T-TABLE-148](#s3t-table-148) · hypotheses: H-U-123, H-U-124, H-U-125, H-U-127 · evidence:
    tablerecordexpirationconfiguration/state-machine/job-status-soak
  - notes: H-U-123, H-U-124, H-U-125 and H-U-127 are UNTESTABLE on tables a controller can create: the premise
    (a customer table with {enabled, days 1}) is rejected synchronously, so there is no
    NotYetRun/Successful/Disabled lifecycle to soak. They can only be probed on an AWS managed table (Storage
    Lens or...
  - full notes: [details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-001.md](details/S3T-TABLERECORDEXPIRATIONCONFIGURATION-001.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-146](#s3t-table-146) | Customer table: GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405, not 200/404 | first_get_after_create_s=0.09, window_s=10 | 10 |
| [S3T-TABLE-147](#s3t-table-147) | PutTableRecordExpirationConfiguration on customer tables -> 405 for days 1/7/365, with/without Iceberg schema, in us-west-2 and us-east-1 | put_latency_ms=[12, 12, 10] | 9 |
| [S3T-TABLE-148](#s3t-table-148) | Rejected PutTableRecordExpirationConfiguration (405) leaves GetTable versionToken/modifiedAt/modifiedBy unchanged (0 s and 5 s after) | get_table_samples=3 | 3 |
| [S3T-TABLE-169](#s3t-table-169) | PutTableRecordExpirationConfiguration: days<1 or status not in [enabled, disabled] -> undeclared ValidationException 400 before table lookup | variants_rejected_400=6, variants_total=18 | 18 |
| [S3T-TABLE-171](#s3t-table-171) | 10 back-to-back PutTableRecordExpirationConfiguration calls in 0.12 s on a customer table: all MethodNotAllowedException 405, no throttle | total_s=0.121, latency_ms_min=10, latency_ms_max=17 | 10 |
| [S3T-TABLE-172](#s3t-table-172) | PutTableRecordExpirationConfiguration pre-table validation passes enabled-w/o-settings, {} value, no status, days=2^31, 1.5 to the 405 gate | variants_passed_to_gate=12, variants_total=18 | 18 |
| [S3T-TABLE-189](#s3t-table-189) | IAM explicit Deny on the record-expiration actions -> AccessDeniedException/403 for Put/Get/GetJobStatus, ahead of the table's 405 gate | assume_role_attempts=4, assume_role_wait_s=9.1, deny_effective_on_first_call=true | 3 |
| [S3T-TABLE-221](#s3t-table-221) | AWS-managed S3 Metadata journal table (type=aws): record-expiration Get, Put and GetJobStatus all -> MethodNotAllowedException 405 | journal_table_creating_s=[15.2], get_after_put_s=0.01 | 1 |
| [S3T-TABLE-239](#s3t-table-239) | Key grant revoked after creation: only CreateTable -> 400; existing-table reads/Put config/Tag 200/204; 3 maintenance jobs Failed in 8 min | soak_window_min=[20.0] | 21 |
| [S3T-TABLE-240](#s3t-table-240) | KMS key Disabled: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 15 min | soak_window_min=[15.0] | 17 |
| [S3T-TABLE-241](#s3t-table-241) | KMS key PendingDeletion: control plane unchanged (reads/Put config/Tag 200, CreateTable 400 same message); Failed job entries frozen 10 min | soak_window_min=[10.0] | 10 |
| [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001) | Customer-created tables: every well-formed record-expiration Put/Get/JobStatus call -> MethodNotAllowedException 405 'not allowed' | latency_ms=[50, 14, 10, 10, 10, 33, 10, 9, 10, 10, 32, 12, 8, 10, 10] | 15 |
| [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003) | GetTableRecordExpirationJobStatus on a fresh customer table -> MethodNotAllowedException 405 in 90/90 calls (30 seq + 2x30 conc.), no 429 | sequential_latency_ms_min_max=[10, 13], burst_wall_s=[0.13, 0.05] | 90 |

## Open questions

<!-- preserved:start id=open-questions -->
<!-- Open questions for table-record-expiration.md (s3tables). One line each, with the finding id. Hand-written. -->

- Does any table a controller can create ever answer other than 405? Every table reachable here (customer, replica,
  type=aws journal) answered 405; Storage Lens / SageMaker Catalog managed tables were not probed
  ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLE-147](#s3t-table-147)).
- Put round trip and read-back of enabled/days on an eligible table ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004)).
- Replace-vs-merge semantics of a second Put ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005)).
- GetTableRecordExpirationJobStatus lifecycle (NotYetRun/Successful/Disabled, lastRunTimestamp) on an eligible table
  ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001)).
- Whether an eligible table rejects {status: enabled} without settings, {} or days 2147483648 / 1.5; here they only
  reached the 405 gate ([S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006)).
- Whether an accepted Put changes GetTable versionToken/modifiedAt; only the rejected path was observable
  ([S3T-TABLE-148](#s3t-table-148), [S3T-TABLE-222](#s3t-table-222)).
- No finding in this document exercises a Delete operation for record expiration, although the render.yaml description
  names Put/Get/Delete; unprobed.
- Idempotency and ConflictException on accepted Puts (H-U-118/H-U-110) untestable; only 405 repeats observed
  ([S3T-TABLE-171](#s3t-table-171)).
- Throttling (429) of the record-expiration operations untested: 90 rapid calls were all 405
  ([S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003)).
- IAM propagation before a new role's Deny takes effect was observed once (4 AssumeRole attempts), no range
  ([S3T-TABLE-189](#s3t-table-189)).
- Several findings are single-run or n=1 per cell: [S3T-TABLE-169](#s3t-table-169) (n=1 per variant), [S3T-TABLE-170](#s3t-table-170), [S3T-TABLE-189](#s3t-table-189),
  [S3T-TABLE-190](#s3t-table-190), [S3T-TABLE-222](#s3t-table-222) (n=1), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](#s3t-tablerecordexpirationconfiguration-002) (single run).
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-148"></a>**S3T-TABLE-148** | read-gap | low | confirmed | Rejected PutTableRecordExpirationConfiguration (405) leaves GetTable versionToken/modifiedAt/modifiedBy unchanged (0 s and 5 s after) | [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-171](#s3t-table-171), [S3T-TABLE-222](#s3t-table-222) | - |
| <a id="s3t-table-171"></a>**S3T-TABLE-171** | idempotency | low | confirmed | 10 back-to-back PutTableRecordExpirationConfiguration calls in 0.12 s on a customer table: all MethodNotAllowedException 405, no throttle | [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-222](#s3t-table-222), [S3T-TABLE-148](#s3t-table-148) | - |
| <a id="s3t-table-222"></a>**S3T-TABLE-222** | read-gap | low | confirmed | Journal table: a rejected PutTableRecordExpirationConfiguration (405) leaves GetTable versionToken/modifiedAt/modifiedBy unchanged | [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-171](#s3t-table-171), [S3T-TABLE-148](#s3t-table-148) | - |
| <a id="s3t-tablerecordexpirationconfiguration-004"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-004** | other | low | confirmed | Doc claim C099 UNTESTABLE: PutTableRecordExpirationConfiguration(enabled, days 30) succeeds and Get returns enabled/30 | [S3T-TABLE-147](#s3t-table-147), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-221](#s3t-table-221), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003) | - |
| <a id="s3t-tablerecordexpirationconfiguration-005"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-005** | other | low | confirmed | Doc claim C100 UNTESTABLE: A second Put(disabled, days 60) replaces the first; Get returns disabled/60 only | [S3T-TABLE-147](#s3t-table-147), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-172](#s3t-table-172), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](#s3t-tablerecordexpirationconfiguration-006), [S3T-TABLE-169](#s3t-table-169) | - |
| <a id="s3t-tablerecordexpirationconfiguration-008"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-008** | other | low | confirmed | Doc claim C103 UNTESTABLE: GetTableRecordExpirationJobStatus on a table enabled minutes ago -> 200 with status, no lastRunTimestamp | [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-146](#s3t-table-146), [S3T-TABLE-221](#s3t-table-221), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-147](#s3t-table-147), [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225) | - |
| <a id="s3t-tablerecordexpirationconfiguration-009"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-009** | other | low | confirmed | Doc claim C104 TRUE: PutTableRecordExpirationConfiguration on a service-managed replica -> 405, but every reachable table gets the same 405 | [S3T-TABLE-224](table-replication-status.md#s3t-table-224), [S3T-TABLE-225](table-replication-status.md#s3t-table-225), [S3T-TABLE-147](#s3t-table-147), [S3T-TABLE-221](#s3t-table-221), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-146](#s3t-table-146) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
