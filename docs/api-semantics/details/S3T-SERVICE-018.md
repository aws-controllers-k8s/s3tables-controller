<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-018: Error catalogue TooManyRequestsException (HTTP 429): 1 message classes - 1 RETRY
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-004](../service.md#s3t-service-004), [S3T-SERVICE-008](../service.md#s3t-service-008), [S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018), [S3T-SERVICE-017](../service.md#s3t-service-017), [S3T-SERVICE-023](../service.md#s3t-service-023) ·
    evidence: service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
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

\## TooManyRequestsException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| Too Many Requests (n=321 + live) | CreateTableBucket, CreateNamespace, DeleteNamespace, DeleteTable,
PutTableMaintenanceConfiguration, GetTableBucketMetricsConfiguration (reads otherwise never throttled at 30
concurrent) | 429 | RETRY | No Retry-After and no rate-limit headers ([S3T-SERVICE-004](../service.md#s3t-service-004)) -> the controller picks
its own backoff. Budgets observed: namespace create+delete share ONE bucket of ~16 per burst keyed on the
account/principal, not the table bucket ([S3T-SERVICE-008](../service.md#s3t-service-008); [S3T-NAMESPACE-033](../namespace.md#s3t-namespace-033)/035, refill ~0.5 s); DeleteTable
~67 then 429 ([S3T-TABLE-218](../table.md#s3t-table-218)); CreateTableBucket throttled at ~8/s ([S3T-TABLEBUCKET-018](../tablebucket.md#s3t-tablebucket-018)). A 429'd create/delete
has NO side effect (0 ghost namespaces, 0 applied deletes, [S3T-SERVICE-008](../service.md#s3t-service-008)) so retrying as-is is safe. With
default SDK retries the 429s are invisible and show up as latency. |

Classification counts: {'RETRY': 1}.
