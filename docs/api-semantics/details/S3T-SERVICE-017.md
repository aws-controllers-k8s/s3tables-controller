<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-017: Error catalogue MethodNotAllowedException (HTTP 405): 1 message classes - 1 PERMANENT-SPEC
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-018](../service.md#s3t-service-018), [S3T-SERVICE-023](../service.md#s3t-service-023) · evidence: service/creative/raw-wire-codes,
    service/creative/sig-skew-and-empty-500, service/creative/iam-authz-variants,
    service/creative/throttle-scope, service/creative/error-catalogue

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

\## MethodNotAllowedException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| Table record expiration configuration API calls to this table are not allowed (n=337) |
GetTableRecordExpirationConfiguration, PutTableRecordExpirationConfiguration,
GetTableRecordExpirationJobStatus | 405 | PERMANENT-SPEC | Scope: every customer table AND the S3 Metadata
journal table (type=aws) answer 405 regardless of days/schema/region ([S3T-TABLE-146](../table-record-expiration.md#s3t-table-146)/147/221,
[S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-001)); no throttle on repeats ([S3T-TABLE-171](../table-record-expiration.md#s3t-table-171)). Check order:
ValidationException (shape) -> AccessDeniedException (authz) -> NotFoundException (ghost UUID) -> 405
([S3T-TABLE-169](../table-record-expiration.md#s3t-table-169)/170/189, [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002)). A TableRecordExpirationConfiguration
resource on such a table is terminal; the only modeled 405 in the API. |

Classification counts: {'PERMANENT-SPEC': 1}.
