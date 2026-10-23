<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-016: Error catalogue ConflictException (HTTP 409): 8 message classes - 1 AMBIGUOUS, 2 PERMANENT-SPEC, 2 RETRY, 3 WAIT-FOR-STATE
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

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
  - related: [S3T-SERVICE-014](../service.md#s3t-service-014), [S3T-SERVICE-015](../service.md#s3t-service-015), [S3T-SERVICE-019](../service.md#s3t-service-019), [S3T-SERVICE-020](../service.md#s3t-service-020), [S3T-SERVICE-021](../service.md#s3t-service-021),
    [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-TABLEBUCKETREPLICATION-001](../tablebucket-replication.md#s3t-tablebucketreplication-001), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEBUCKETREPLICATION-004](../tablebucket-replication.md#s3t-tablebucketreplication-004), [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012) · evidence: service/creative/raw-wire-codes,
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

\## ConflictException

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| The bucket is in a transitional state because of a previous deletion attempt. Try again later. (n=472) |
CreateTableBucket | 409 | WAIT-FOR-STATE | Name cooldown after DeleteTableBucket: ~600 s (604.7 s measured,
[S3T-TABLEBUCKET-047](../tablebucket.md#s3t-tablebucket-047)/016/115); meanwhile every read of the old ARN is 404. Keyed on the bucket name in the
region. A controller re-creating a deleted bucket (or re-applying a deleted CR with the same name) must
requeue with a long backoff; nothing accelerates it. 472 occurrences in the seed = probes polling this state.
|
| The bucket that you tried to create already exists, and you own it. (n=11) | CreateTableBucket | 409 |
PERMANENT-SPEC | Duplicate create regardless of configuration ([S3T-TABLEBUCKET-013](../tablebucket-encryption-storage.md#s3t-tablebucket-013)/017). The ARN is
deterministic ([S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003)) -> adopt: GetTableBucket and continue. Two concurrent creates: exactly one
200 ([S3T-TABLEBUCKET-017](../tablebucket.md#s3t-tablebucket-017)). |
| A namespace with an identical name already exists in the bucket. (n=33) | CreateNamespace | 409 |
PERMANENT-SPEC | Duplicate create ([S3T-NAMESPACE-021](../namespace.md#s3t-namespace-021)/034); adopt by GetNamespace. Same name immediately after
DeleteNamespace is 200 (no cooldown, unlike buckets, [S3T-NAMESPACE-023](../namespace.md#s3t-namespace-023)). |
| A table with an identical name already exists in the namespace. (n=13) | CreateTable, RenameTable | 409 |
AMBIGUOUS | (a) CreateTable duplicate ([S3T-TABLE-064](../table-metadata-location.md#s3t-table-064)) -> adopt via GetTable; (b) RenameTable into an occupied
name ([S3T-TABLE-048](../table.md#s3t-table-048)) -> spec; (c) RenameTable to the CURRENT name/namespace is NOT a no-op but this 409
([S3T-TABLE-050](../table.md#s3t-table-050)) -> the controller must skip RenameTable when nothing changes. No cooldown after DeleteTable
([S3T-TABLE-067](../table-metadata-location.md#s3t-table-067)). |
| Provided version token does not match the table version token. (n=34) | UpdateTableMetadataLocation,
DeleteTable, RenameTable, DeleteTableReplication(given the TABLE token) | 409 | WAIT-FOR-STATE | The caller's
versionToken is stale or garbage (any malformed token is 409, never 400, [S3T-TABLE-065](../table-metadata-location.md#s3t-table-065)/103/112) -> re-read
GetTable.versionToken and retry in the same reconcile. The token changes only on metadata commits, not on
tag/policy/maintenance/rename ([S3T-TABLE-110](../table-maintenance.md#s3t-table-110)/134). DeleteTableReplication wants the REPLICATION token, not the
table's ([S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004)/007). |
| The specified version token does not match the configuration. (n=15) | PutTableReplication,
PutTableBucketReplication, DeleteTableBucketReplication | 409 | WAIT-FOR-STATE | Stale replication token ->
re-read Get*Replication.versionToken (rotates on every Put, [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002)) and retry;
concurrent Puts with the same token: exactly one wins ([S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016)). When NO config exists the
same stale token is a 404, see NotFoundException. |
| The resource was modified by a concurrent request. Please retry your request. (n=17) |
PutTableMaintenanceConfiguration, PutTablePolicy | 409 | RETRY | Optimistic-concurrency collision between
writers of one table's policy/maintenance; sequential retries succeed ([S3T-TABLE-191](../table-maintenance.md#s3t-table-191)). Tags and bucket policy
collisions are reported as 500 instead (see InternalServerErrorException). |
| The table was modified concurrently. Try again. (n=4) | RenameTable | 409 | RETRY | 10 concurrent renames: 1
x 204, 4 x 409, 1 x 500, 4 x 404 ([S3T-TABLE-193](../table-metadata-location.md#s3t-table-193)); retry after re-reading the current name. |

Classification counts: {'AMBIGUOUS': 1, 'PERMANENT-SPEC': 2, 'RETRY': 2, 'WAIT-FOR-STATE': 3}.
