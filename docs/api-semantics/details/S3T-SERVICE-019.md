<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-019: Error catalogue AccessDeniedException (HTTP 403): 9 message classes - 9 AUTHZ
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-019"></a>**S3T-SERVICE-019** `error-code` · impact high · handled · verified 2026-10-14
  **Error catalogue AccessDeniedException (HTTP 403): 9 message classes - 9 AUTHZ**
  AccessDeniedException: declared on only 15/49 operations but produced by 49/49 (seed ~35 + 15 live); NOT in
  terminal lists generated from the model. 9 message classes; classification counts {'AUTHZ': 9}. AUTHZ in
  every case, one message template with 5 reason suffixes; the denied action may differ from the operation;
  visible ~86-100 s after an IAM allow/revoke ([S3T-SERVICE-012](../service.md#s3t-service-012)) and from a fresh role's first call.
  - ACK: terminal_codes, requeue · ops: CreateNamespace, CreateTable, CreateTableBucket, DeleteNamespace,
    GetNamespace, GetTable, GetTableBucket, GetTableMaintenanceJobStatus, ListNamespaces, ListTableBuckets,
    ListTables, PutTableBucketReplication, PutTableMaintenanceConfiguration, PutTableReplication
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-016](../service.md#s3t-service-016), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-021](../service.md#s3t-service-021),
    [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-SERVICE-012](../service.md#s3t-service-012) · evidence: service/creative/raw-wire-codes,
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

\## AccessDeniedException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| User: <ARN> is not authorized to perform: <action> on resource: <ARN> with an explicit deny in an
identity-based policy (n=~140, 30 ops) | every op (seed: 30 ops; live: +15) | 403 | AUTHZ | IAM identity
policy Deny. See the rules below. |
| User: <ARN> is not authorized to perform: <action> on resource: <ARN> because no identity-based policy
allows the <action> action (n=~25) | every op (seed: 13 ops; live: 15 more incl. all bucket sub-resource
Put/Delete, replication Gets/Deletes, GetTableMaintenanceJobStatus) | 403 | AUTHZ | Implicit deny (no Allow).
[S3T-SERVICE-012](../service.md#s3t-service-012) covers the 15 ops the seed never saw denied: 15/15 AccessDeniedException. |
| User: <ARN> is not authorized to perform: <action> on resource: <ARN> with an explicit deny in a
resource-based policy (n=~35) | namespace/table/bucket ops under a table-bucket or table policy Deny | 403 |
AUTHZ | A table bucket / table policy can deny the controller's own principal, including DeleteTable and
Put/DeleteTablePolicy (lockout, [S3T-TABLE-213](../table-maintenance.md#s3t-table-213)/214, [S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091)) - enforced ~30 ms after PutTablePolicy
and lifted 10-40 ms after Delete ([S3T-TABLE-216](../table-policy.md#s3t-table-216), [S3T-TABLEBUCKETPOLICY-005](../tablebucket-policy.md#s3t-tablebucketpolicy-005)). |
| User: <ARN> is not authorized to perform: <action> on resource: <ARN> with an explicit deny in a session
policy (n=6) | CreateNamespace, DeleteNamespace, GetNamespace, GetTableBucket, ListNamespaces, ListTables |
403 | AUTHZ | Session policy Deny ([S3T-NAMESPACE-031](../namespace.md#s3t-namespace-031)). |
| User: <ARN> is not authorized to perform: <action> on resource: <ARN> because no permissions boundary allows
the <action> action (n=live) | GetTableBucket, ListTableBuckets (resource bucket/*), GetTable, CreateNamespace
| 403 | AUTHZ | New variant from a role with a PermissionsBoundary that lacks s3tables ([S3T-SERVICE-011](../service.md#s3t-service-011)). |
| User: <ARN> is not authorized to perform: s3tables:PutTableEncryption / s3tables:PutTableBucketEncryption /
s3tables:PutTableBucketStorageClass / s3tables:TagResource on resource: <ARN> ... (n=19) | CreateTable (with
encryptionConfiguration, even AES256), CreateTableBucket (with encryptionConfiguration /
storageClassConfiguration / tags) | 403 | AUTHZ | The denied ACTION is not the operation: Create* needs the
sub-resource permissions for the optional members it carries ([S3T-TABLE-140](../table-policy.md#s3t-table-140)/141, [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129)). Nothing
is created. The controller's IAM policy recommendations must include them. |
| User: <ARN> is not authorized to perform: iam:PassRole on resource: <ARN> with an explicit deny in an
identity-based policy (n=2) | PutTableBucketReplication, PutTableReplication | 403 | AUTHZ | PassRole on
configuration.role ([S3T-TABLEBUCKETREPLICATION-012](../tablebucket-replication.md#s3t-tablebucketreplication-012)); AccessDeniedException although the model declares only
Forbidden for these ops. |
| Insufficient access to update the table maintenance configuration (n=4, [S3T-TABLE-234](../table-maintenance.md#s3t-table-234)) |
PutTableMaintenanceConfiguration(strategy sort/z-order) without s3tables:GetTableData | 403 | AUTHZ |
Service-worded (no 'User:' prefix) AccessDeniedException; binpack/size Puts by the same role succeed. Observed
to persist 87.9 s after the role policy was fixed ([S3T-TABLE-234](../table-maintenance.md#s3t-table-234)) - consistent with the ~86-100 s authz
propagation in [S3T-SERVICE-012](../service.md#s3t-service-012). |
| User: <ARN> is not authorized to perform: kms:DescribeKey on resource: <ARN> with an explicit deny in an
identity-based policy (n=1, HTTP 400) | kms:DescribeKey called directly by a probe | 400 | AUTHZ | Sibling
service: KMS returns AccessDeniedException with HTTP 400. S3 Tables itself never needs the caller to
DescribeKey for reads ([S3T-TABLEBUCKET-106](../tablebucket-policy.md#s3t-tablebucket-106)). |

Classification counts: {'AUTHZ': 9}.

Rules for the whole code: (1) the model declares it on 15 operations; every operation can return it -> the
generated terminal-code list must include it everywhere; (2) check order on GetTableBucket under deny: ARN
syntax / region / partition 400 first, then authz 403, then lookup (a ghost ARN under deny is 403,
[S3T-SERVICE-011](../service.md#s3t-service-011); record-expiration ops likewise, [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190)); (3) an identity-policy Allow became visible
to s3tables 86.1 s after PutRolePolicy and a revoke 100.5 s after DeleteRolePolicy (n=1, [S3T-SERVICE-012](../service.md#s3t-service-012);
TABLE-234 saw 87.9 s) - the controller must requeue AccessDeniedException with backoff for at least ~2 min
after any IAM change instead of marking the resource terminal, and must expect its own access to linger ~100 s
after a revoke; (4) a table-bucket/table policy Deny on the controller's principal is enforced within ~30 ms
and can lock the controller out of DeleteTable/Put*Policy - finalizers must surface the 403 as a condition.
