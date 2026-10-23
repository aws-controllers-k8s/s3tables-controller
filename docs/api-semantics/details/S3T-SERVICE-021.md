<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-021: Error catalogue ValidationException (HTTP 400): 17 message classes - 1 PERMANENT-QUOTA, 16 PERMANENT-SPEC
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-SERVICE-007](../service.md#s3t-service-007), [S3T-TABLEBUCKETREPLICATION-003](../tablebucket-replication.md#s3t-tablebucketreplication-003), [S3T-SERVICE-015](../service.md#s3t-service-015),
    [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-016](../service.md#s3t-service-016), [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011),
    [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021) · evidence: service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue

## Notes

Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the request is
right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion cooldown, or
the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC = the request is wrong or the
desired state already holds (non-idempotent deletes), terminal until the spec changes; PERMANENT-QUOTA = a
limit; AUTHZ = permissions (IAM/resource/boundary/session policy, KMS key policy, cross-account); AMBIGUOUS =
the same code+message has different causes with different remediations (named in the rule). Sources:
seed-error-messages.md (105 structured probes, 391 rows; finding ids cited per row) plus the creative-round
probes service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
service/creative/iam-authz-variants, service/creative/throttle-scope ([S3T-SERVICE-001](../service.md#s3t-service-001)..013). Messages are
quoted as recorded; <ARN>, <N>, <T> are normalised. Controllers should branch on the wire code and HTTP status
and treat the message as diagnostic text only - several texts echo caller input or differ by a curly
apostrophe.

\## ValidationException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| 1 validation error detected: Value at 'value.status' failed to satisfy constraint: Member must satisfy enum
value set: [enabled, disabled] (n=13) | PutTableRecordExpirationConfiguration,
PutTableBucketMaintenanceConfiguration | 400 | PERMANENT-SPEC | Bucket-level/record-expiration status enum
([S3T-TABLEBUCKET-074](../tablebucket-maintenance-metrics.md#s3t-tablebucket-074), [S3T-TABLE-169](../table-record-expiration.md#s3t-table-169)); the TABLE maintenance status enum is a BadRequestException instead. |
| 1 validation error detected: Value at 'storageClassConfiguration.storageClass' failed to satisfy constraint:
Member must satisfy enum value set: [INTELLIGENT_TIERING, STANDARD] (n=8) | PutTableBucketStorageClass | 400 |
PERMANENT-SPEC | [S3T-TABLEBUCKET-086](../tablebucket-encryption-storage.md#s3t-tablebucket-086)/087; the same typo at Create* is a BadRequestException. |
| 1 validation error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value
set: [icebergUnreferencedFileRemoval] (n=5) | PutTableBucketMaintenanceConfiguration | 400 | PERMANENT-SPEC |
Bucket-level type path enum ([S3T-TABLEBUCKET-077](../tablebucket-maintenance-metrics.md#s3t-tablebucket-077)); the table-level type is a BadRequestException 'Invalid
table maintenance type'. |
| 1 validation error detected: Value at 'value.settings.days' failed to satisfy constraint: Member must have
value greater than or equal to 1 (n=5) | PutTableRecordExpirationConfiguration | 400 | PERMANENT-SPEC |
Validated BEFORE the table lookup and the 405 gate ([S3T-TABLE-169](../table-record-expiration.md#s3t-table-169)). |
| 1 validation error detected: Value at 'value.settings.icebergUnreferencedFileRemoval.unreferencedDays' /
'.nonCurrentDays' failed to satisfy constraint: Member must have value greater than or equal to 1 (n=3) |
PutTableBucketMaintenanceConfiguration | 400 | PERMANENT-SPEC | [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075). |
| 1 validation error detected: Value at 'value' failed to satisfy constraint: Member must not be null (n=1 +
live) | PutTableBucketMaintenanceConfiguration, PutTableRecordExpirationConfiguration (body {}) | 400 |
PERMANENT-SPEC | [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-SERVICE-002](../service.md#s3t-service-002). |
| 1 validation error detected: Value at 'storageClassConfiguration' / 'storageClassConfiguration.storageClass'
failed to satisfy constraint: Member must not be null (n=1 + live) | PutTableBucketStorageClass | 400 |
PERMANENT-SPEC | [S3T-TABLEBUCKET-087](../tablebucket-encryption-storage.md#s3t-tablebucket-087), [S3T-SERVICE-002](../service.md#s3t-service-002). |
| 1 validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
less than or equal to <N> (n=5) | PutTablePolicy, PutTableBucketPolicy | 400 | PERMANENT-QUOTA | 20480 bytes
accepted, 20481 rejected ([S3T-TABLEPOLICY-001](../table-policy.md#s3t-tablepolicy-001), [S3T-TABLEBUCKETPOLICY-002](../tablebucket-policy.md#s3t-tablebucketpolicy-002)). |
| 1 validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
greater than or equal to 1 / Member must not be null (n=1 + live) | PutTableBucketPolicy, PutTablePolicy
(resourcePolicy='' or body {}) | 400 | PERMANENT-SPEC | [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090), [S3T-SERVICE-002](../service.md#s3t-service-002); whitespace-only
is a BadRequestException 'malformed' instead. |
| 1 validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy constraint:
[Member must have length less than or equal to <N>, Member must have length greater than or equal to 1] (n=3)
| TagResource | 400 | PERMANENT-SPEC | Empty or 129-char key ([S3T-TABLE-117](../table.md#s3t-table-117)/120, [S3T-TABLEBUCKET-030](../tablebucket.md#s3t-tablebucket-030)). |
| 1 validation error detected: Value at 'tags' failed to satisfy constraint: Map value must satisfy
constraint: [Member must have length less than or equal to <N>, Member must have length greater than or equal
to 0] (n=2) | TagResource | 400 | PERMANENT-SPEC | 257-char value ([S3T-TABLE-120](../table.md#s3t-table-120)). |
| 1 validation error detected: Value at 'tagKeys' failed to satisfy constraint: Member must satisfy
constraint: [Member must have length less than or equal to <N>, Member must have length greater than or equal
to 1] (n=2) | UntagResource | 400 | PERMANENT-SPEC | One over-long key anywhere in tagKeys rejects the whole
call, nothing removed ([S3T-TABLE-126](../table.md#s3t-table-126)). |
| 1 validation error detected: Value at 'versionToken' failed to satisfy constraint: Member must have length
greater than or equal to 1 (n=2); 2 validation errors ... 'versionToken' and 'metadataLocation' must not be
null (live) | UpdateTableMetadataLocation | 400 | PERMANENT-SPEC | versionToken='' ([S3T-TABLE-103](../table-metadata-location.md#s3t-table-103)); body {}
([S3T-SERVICE-002](../service.md#s3t-service-002)). |
| 1 validation error detected: Value at 'configuration.rules' failed to satisfy constraint: Member must have
length less than or equal to 1 / greater than or equal to 1 / must not be null (n=5) |
PutTableBucketReplication, PutTableReplication | 400 | PERMANENT-SPEC | Exactly one rule
([S3T-TABLEREPLICATION-011](../table-replication.md#s3t-tablereplication-011), [S3T-TABLEBUCKETREPLICATION-017](../tablebucket-replication.md#s3t-tablebucketreplication-017)). configuration ABSENT is a 500 instead
([S3T-SERVICE-007](../service.md#s3t-service-007)). |
| 1 validation error detected: Value at 'configuration.rules.1.member.destinations' failed to satisfy
constraint: Member must have length less than or equal to 5 / greater than or equal to 1 (n=5) |
PutTableBucketReplication, PutTableReplication | 400 | PERMANENT-SPEC | 1..5 destinations
([S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012)). |
| 1 validation error detected: Value at
'configuration.rules.1.member.destinations.1.member.destinationTableBucketARN' failed to satisfy constraint:
Member must satisfy regular expression pattern:
(arn:aws[-a-z0-9]*:[a-z0-9]+:[-a-z0-9]*:[0-9]{12}:bucket/[a-z0-9_-]{3,63}) (n=1) | PutTableBucketReplication |
400 | PERMANENT-SPEC | Only ARN syntax is checked; a nonexistent destination/role is accepted
([S3T-TABLEBUCKETREPLICATION-009](../tablebucket-replication.md#s3t-tablebucketreplication-009), [S3T-TABLEBUCKET-011](../tablebucket-replication.md#s3t-tablebucket-011)). |
| 1 validation error detected: Value at 'configuration.role' failed to satisfy constraint: Member must satisfy
regular expression pattern: arn:.+:iam::[0-9]{12}:role/<PRINCIPAL> (n=1) | PutTableReplication | 400 |
PERMANENT-SPEC | IAM user ARN rejected, aws-cn role ARN accepted by the pattern ([S3T-TABLEREPLICATION-013](../table-replication.md#s3t-tablereplication-013)). |

Classification counts: {'PERMANENT-QUOTA': 1, 'PERMANENT-SPEC': 16}.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. ValidationException (HTTP 400,
wire code absent from the s3tables model) is returned by 10 operations for required/enum/length/pattern
violations (resourcePolicy empty or > 20480 bytes on PutTableBucketPolicy/ PutTablePolicy, tag key > 128 chars
on TagResource/UntagResource, bucket maintenance enum/min values, storage class enum, Put*Replication
rules/destinations bounds) and is not in terminal_codes, so a permanent spec error is requeued forever instead
of setting ACK.Terminal; add ValidationException to terminal_codes on every resource.
