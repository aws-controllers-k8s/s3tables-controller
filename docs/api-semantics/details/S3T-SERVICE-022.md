<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-022: Error catalogue ForbiddenException (HTTP 403): 4 message classes - 3 AUTHZ, 1 PERMANENT-SPEC
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-022"></a>**S3T-SERVICE-022** `error-code` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **Error catalogue ForbiddenException (HTTP 403): 4 message classes - 3 AUTHZ, 1 PERMANENT-SPEC**
  ForbiddenException: declared on 49/49; produced by 26+ ops (cross-account ARNs) + 3 special texts. 4 message
  classes; classification counts {'AUTHZ': 3, 'PERMANENT-SPEC': 1}. Cross-account 'Unauthorized' plus three
  pass-through texts (KMS key policy, S3 'Access Denied', cross-account PassRole); one is really a spec error.
  - ACK: terminal_codes, docs-only · ops: CreateTable, CreateTableBucket, PutTableReplication
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-016](../service.md#s3t-service-016), [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-020](../service.md#s3t-service-020),
    [S3T-SERVICE-021](../service.md#s3t-service-021), [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021), [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-TABLEBUCKET-098](../tablebucket-replication.md#s3t-tablebucket-098) · evidence:
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

\## ForbiddenException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| Unauthorized (n=51) | 26+ ops given a well-formed ARN whose account id is not the caller's (existing or
random account) | 403 | AUTHZ | Cross-account ARNs are never 404: existence is not disclosed
([S3T-TABLE-026](../table-policy.md#s3t-table-026)/179, [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021)/099). For a controller an account-id typo in a referenced ARN looks
like a permission problem; terminal. ListTableBuckets and CreateTableBucket have no ARN input and cannot
produce it. |
| User: <ARN> is not authorized to perform: kms:DescribeKey on this resource because the resource does not
exist in this Region, no resource-based policies allow access, or a resource-based policy explicitly denies
access (n=3) | CreateTable, CreateTableBucket with aws:kms | 403 | AUTHZ | KMS's own error passed through as
ForbiddenException: the KEY POLICY does not let the caller use the key. Contrast: an IDENTITY-policy deny on
kms gives AccessDeniedException naming s3tables:PutTableEncryption ([S3T-TABLE-141](../table-policy.md#s3t-table-141)), and a key lacking the
maintenance principal gives BadRequestException 'Insufficient access...' - three codes for three KMS
permission shapes. |
| Access Denied (n=2) | CreateTable WITH metadata by a caller lacking s3tables:PutTableData | 403 | AUTHZ |
S3-style terse text; CreateTable without metadata by the same role is 200 ([S3T-TABLE-131](../table-policy.md#s3t-table-131)). Nothing is created.
The IAM recommendation for a Table reconciler that writes a schema must include s3tables:PutTableData (and
GetTableData for sort/z-order maintenance). |
| Cross-account pass role is not allowed. (n=1 + live) | PutTableReplication (configuration.role in another
account) | 403 | PERMANENT-SPEC | A spec error behind a 403: the role must be in the caller's account. Checked
BEFORE the rules validation (a body with only a foreign role and no rules gets this 403, [S3T-SERVICE-007](../service.md#s3t-service-007)
variants). |

Classification counts: {'AUTHZ': 3, 'PERMANENT-SPEC': 1}.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. ForbiddenException never heals
by retrying unchanged - two of its four classes are spec errors ('Unauthorized' for any ARN whose account id
is not the caller's, i.e. an account-id typo in a referenced ARN; 'Cross-account pass role is not allowed.' on
Put*Replication) and two need a KMS key policy / s3tables:PutTableData fix - yet the mechanism requeues all of
them indefinitely with no condition naming the cause; treat 'Unauthorized' and the PassRole text as terminal
and surface the others as a Recoverable condition (AccessDeniedException must stay retryable: transient
~86-100 s after IAM changes, [S3T-SERVICE-019](../service.md#s3t-service-019)).
