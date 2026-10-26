<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-020: Error catalogue InternalServerErrorException (HTTP 500): 1 message classes - 1 AMBIGUOUS
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-002](../service.md#s3t-service-002), [S3T-SERVICE-007](../service.md#s3t-service-007), [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-016](../service.md#s3t-service-016),
    [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-021](../service.md#s3t-service-021), [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-TABLE-300](../table-replication-status.md#s3t-table-300), [S3T-TABLE-304](../table-replication.md#s3t-table-304) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
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

\## InternalServerErrorException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| An internal error occurred. Try again. (n=70 + live) | TagResource, UntagResource, CreateTable,
PutTableBucketPolicy, PutTablePolicy, PutTableBucketMaintenanceConfiguration, RenameTable,
PutTableReplication, PutTableBucketReplication | 500 | AMBIGUOUS | ONE text, four causes: (1) RETRY - write
collisions on one resource: 10 concurrent TagResource -> 8-9 x 500 ([S3T-TABLE-186](../table.md#s3t-table-186), [S3T-TABLEBUCKET-070](../tablebucket.md#s3t-tablebucket-070)), 10
concurrent PutTableBucketPolicy -> 3 x 500 ([S3T-TABLEBUCKET-082](../tablebucket-encryption-storage.md#s3t-tablebucket-082)), 20 concurrent maintenance Puts -> 1 x 500
([S3T-TABLEBUCKET-078](../tablebucket-maintenance-metrics.md#s3t-tablebucket-078)), concurrent PutTablePolicy/RenameTable ([S3T-TABLE-191](../table-maintenance.md#s3t-table-191)/193); the losing writes are NOT
applied and sequential retries succeed; (2) PERMANENT (should be 404) - UntagResource on a deleted or
never-existing table ARN returns 500 deterministically while TagResource/ListTagsForResource on the same ARN
return 404 ([S3T-TABLE-130](../table.md#s3t-table-130)/143, [S3T-TABLEBUCKET-033](../tablebucket.md#s3t-tablebucket-033)/043); (3) PERMANENT-SPEC - CreateTable with
metadata.iceberg.schema={} (fields omitted) 4/4 x 500, no table created ([S3T-TABLE-036](../table-schema.md#s3t-table-036)/133); (4)
PERMANENT-SPEC - PutTableReplication / PutTableBucketReplication with the 'configuration' member absent or
null (6/6 and variants: with a ghost table UUID, with only a versionToken) -> 500, no config created
([S3T-SERVICE-007](../service.md#s3t-service-007)); SDKs normally reject this client-side, so it is reachable only through raw requests or a
nil pointer that serializes as absent. Controller rule: retry a 500 with backoff a bounded number of times (3
is enough for cause 1); if it repeats in a stable state, surface it (it will not heal); for a finalizer
untagging a resource that reads 404 elsewhere, treat the 500 as 'gone'. |

Classification counts: {'AMBIGUOUS': 1}.

Which causes are reachable from a generated controller: cause 1 whenever two reconcilers (or a reconciler and
a user) write tags/policy of one resource; cause 2 whenever a Table is deleted out-of-band between the
controller's read and its UntagResource; cause 3/4 only with SDK validation bypassed (empty struct pointers
that serialize as absent).
