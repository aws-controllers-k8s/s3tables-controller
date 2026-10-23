<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-014: Error catalogue NotFoundException (HTTP 404): 9 message classes - 7 AMBIGUOUS, 1 PERMANENT-SPEC, 1 WAIT-FOR-STATE
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-016](../service.md#s3t-service-016), [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-021](../service.md#s3t-service-021),
    [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-096](../tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-098](../tablebucket-replication.md#s3t-tablebucket-098) · evidence:
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

\## NotFoundException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| The specified table does not exist. (n=648) | GetTable, DeleteTable, RenameTable,
UpdateTableMetadataLocation, GetTableMetadataLocation, Get/Put/DeleteTablePolicy,
Get/PutTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetTableEncryption, GetTableStorageClass,
ListTagsForResource, TagResource, Put/GetTableRecordExpiration*, Get/Put/DeleteTableReplication(+Status) | 404
| AMBIGUOUS | Causes: (a) never existed / already deleted (the 404 is immediate after DeleteTable,
[S3T-TABLE-122](../table-metadata-location.md#s3t-table-122)) -> on read: resource gone, on delete: treat as success; (b) renamed out-of-band: the OLD
(bucket, namespace, name) triple is 404 while GetTable(tableArn) is still 200 with the new name
([S3T-TABLE-002](../table-metadata-location.md#s3t-table-002)/009/175) -> when the ARN is known, confirm with GetTable(tableArn) before recreating; (c) name
case mismatch ([S3T-TABLE-027](../table.md#s3t-table-027)) and raw paths with an empty namespace/name segment ([S3T-TABLE-184](../table-policy.md#s3t-table-184)) -> spec
problem, not absence; (d) a replayed RenameTable after it already succeeded ([S3T-TABLE-003](../table-metadata-location.md#s3t-table-003)) -> idempotent
success; (e) the message is also what every table SUB-RESOURCE op returns when the parent table is gone -> for
TablePolicy/Maintenance/Replication reconcilers it means 'parent gone', not 'sub-resource unset'.
UntagResource on a deleted table returns 500 instead of this 404 (see InternalServerErrorException). |
| The specified bucket does not exist. (n=185) | GetTableBucket, DeleteTableBucket, ListTagsForResource,
Tag/UntagResource, CreateNamespace, GetNamespace, DeleteNamespace, ListNamespaces, ListTables, CreateTable,
Get/Put/Delete every bucket sub-resource (policy, encryption, maintenance, metrics, storage class,
replication) +22 more | 404 | AMBIGUOUS | Causes: (a) never existed / deleted -> gone; (b) deleted within the
last ~10 min: every read/write on the ARN is 404 but CreateTableBucket(same name) is ConflictException
'transitional state' for ~600 s ([S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015)/047/115) -> a controller that recreates after a 404 must
treat that 409 as WAIT-FOR-STATE, not as 'already exists'; (c) for namespace/table/sub-resource ops it means
the PARENT bucket is gone (references: requeue until the TableBucket resource exists). Reads are immediately
consistent ([S3T-TABLEBUCKET-101](../tablebucket-replication.md#s3t-tablebucket-101)), so no 404 read-after-create window needs masking. |
| The specified namespace does not exist. (n=166) | ListTables, GetNamespace, CreateTable, DeleteNamespace,
DeleteTable, GetTable, RenameTable | 404 | AMBIGUOUS | (a) GetNamespace/DeleteNamespace: gone (delete is
idempotent-by-404, [S3T-NAMESPACE-022](../namespace.md#s3t-namespace-022)) -> success on delete; (b) CreateTable/ListTables/DeleteTable: the parent
namespace is missing and is NOT auto-created ([S3T-NAMESPACE-026](../namespace.md#s3t-namespace-026), [S3T-TABLE-080](../table.md#s3t-table-080)) -> WAIT-FOR-STATE on the
Namespace resource (references), terminal if the user spec names a namespace nobody creates. ListTables with a
prefix of an existing namespace is also 404 (exact match, [S3T-TABLE-094](../table.md#s3t-table-094)). |
| The specified replication configuration does not exist. (n=114) | GetTableReplication,
GetTableReplicationStatus, GetTableBucketReplication, DeleteTableBucketReplication, DeleteTableReplication,
PutTableBucketReplication, PutTableReplication | 404 | AMBIGUOUS | (a) Get on a table/bucket that never had a
config: the normal 'unset' state ([S3T-TABLEREPLICATION-001](../table-replication.md#s3t-tablereplication-001)) -> absent, not an error; (b) Delete when absent or
a second Delete: non-idempotent 404 ([S3T-TABLEBUCKETREPLICATION-008](../tablebucket-replication.md#s3t-tablebucketreplication-008), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018)) -> finalizer
treats as success; (c) Put WITH a versionToken when no config exists (first Put with a garbage token, or Put
with the pre-delete token after a Delete) -> 404 instead of 409 ([S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010),
[S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005)) -> drop the token and re-Put; (d) GetTableReplication on a table covered only
by a BUCKET-level config -> 404 while GetTableReplicationStatus is 200 ([S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021)). |
| The specified table policy does not exist. (n=101) | GetTablePolicy | 404 | AMBIGUOUS | For a TablePolicy
resource this IS 'not found' (create via PutTablePolicy); DeleteTablePolicy on an absent policy is 204
(idempotent, [S3T-TABLE-150](../table-policy.md#s3t-table-150)). The other 404 on the same op ('table does not exist') means the parent is gone ->
the two messages need different handling (exceptions.404 must look at the message or GetTable first). The
message differs from the missing-table one ([S3T-TABLE-149](../table-policy.md#s3t-table-149)). |
| The specified bucket policy does not exist. (n=85) | GetTableBucketPolicy | 404 | AMBIGUOUS | Same pattern
as the table policy: absent policy (create it) vs parent bucket gone ('bucket does not exist');
DeleteTableBucketPolicy is idempotent 204 ([S3T-TABLEBUCKET-062](../tablebucket-policy.md#s3t-tablebucket-062)); GetTableBucketPolicy on a nonexistent bucket
is also 404 ([S3T-TABLEBUCKET-061](../tablebucket-policy.md#s3t-tablebucket-061)). |
| No such metric configuration exists (n=72) | GetTableBucketMetricsConfiguration | 404 | AMBIGUOUS | Absent
config (fresh bucket, after Delete) -> not an error; Delete is idempotent 204 ([S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092)). Note the
first Get after Delete was 404 in only 2/3 reps ([S3T-TABLEBUCKET-095](../tablebucket-maintenance-metrics.md#s3t-tablebucket-095)): a reconciler polling for 'gone' must
tolerate one stale 200. Different wording (no 'The specified ...' prefix, no period). |
| The specified encryption configuration does not exist. (n=29) | DeleteTableBucketEncryption | 404 |
PERMANENT-SPEC | Non-idempotent delete: 404 when encryption was never set explicitly and on a second Delete
([S3T-TABLEBUCKET-049](../tablebucket-encryption-storage.md#s3t-tablebucket-049)/096) although GetTableBucketEncryption always returns 200 with the AES256 default -> a
reconciler that 'unsets' encryption must swallow this 404 as success (desired state already holds). |
| The specified destination namespace does not exist. (n=7) | RenameTable(newNamespaceName) | 404 |
WAIT-FOR-STATE | Prerequisite: the destination namespace must exist ([S3T-TABLE-049](../table.md#s3t-table-049)); the table is untouched.
Requeue until the Namespace resource exists (references); terminal if nobody creates it. Distinct text from
the source-namespace 404. |

Classification counts: {'AMBIGUOUS': 7, 'PERMANENT-SPEC': 1, 'WAIT-FOR-STATE': 1}.
