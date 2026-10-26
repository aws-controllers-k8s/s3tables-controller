<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-299: Replica copies the source's snapshot-management config but forces compaction disabled (512/auto); no tags, no aws: tags, no policy
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-299"></a>**S3T-TABLE-299** `read-gap` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Replica copies the source's snapshot-management config but forces compaction disabled (512/auto); no tags, no aws: tags, no policy**
  Source table (empty, explicit INTELLIGENT_TIERING) primed before PutTableReplication with tags {"team":
  "data", "env": "lab", "ackq-run": "97f95f"}, policy Sid SrcAllowGet (Resource = source ARN) and maintenance
  {"icebergCompaction": {"status": "enabled", "settings": {"icebergCompaction": {"targetFileSizeMB": 256,
  "strategy": "binpack"}}}, "icebergSnapshotManagement": {"status": "enabled", "settings":
  {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5, "maxSnapshotAgeHours": 48}}}}. Replica (same
  namespace/name in the destination bucket, status 'completed' at +66s): ListTagsForResource -> {} (aws: keys:
  []); GetTablePolicy -> NotFoundException (HTTP 404) 'The specified table policy does not exist.';
  GetTableMaintenanceConfiguration -> {"icebergCompaction": {"status": "disabled", "settings":
  {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}}, "icebergSnapshotManagement":
  {"status": "enabled", "settings": {"icebergSnapshotManagement": {"minSnapshotsToKeep": 5,
  "maxSnapshotAgeHours": 48}}}} (equal to source: False); GetTableStorageClass -> {"storageClass": "STANDARD"}
  (destination bucket default STANDARD, source IT); GetTableEncryption -> {"sseAlgorithm": "AES256"}. Replica
  GetTable: createdBy == source createdBy, modifiedBy=None,
  managedByService=replication.s3tables.amazonaws.com. Destination namespace GetNamespace: createdBy == the
  source namespace's, keys ['createdAt', 'createdBy', 'namespace', 'namespaceId', 'ownerAccountId',
  'tableBucketId'].
  - ACK: tags.custom-sync, scope:skip, hook:sdk_read_one_post_set_output · ops: PutTableReplication,
    ListTagsForResource, GetTablePolicy, GetTableMaintenanceConfiguration, GetTable, GetNamespace · fields:
    tags, resourcePolicy, configuration.icebergCompaction, configuration.icebergSnapshotManagement, createdBy,
    managedByService
  - repro: CreateTable -> TagResource + PutTablePolicy + PutTableMaintenanceConfiguration ->
    PutTableReplication(-> D) -> wait completed -> read the replica's tags/policy/maintenance
  - measurements: completed_at_s=65.68, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-177](../table-replication.md#s3t-table-177), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-300](../table-replication-status.md#s3t-table-300),
    [S3T-TABLE-306](../table-replication-status.md#s3t-table-306) · evidence: table/creative/replica-settings-fidelity

## Notes

Cross-service analogy (DynamoDB restore drops tags/policy/TTL; keeps SSE/billing): the replica is a partial
copy. Corrects/extends [S3T-TABLE-223](../table-replication-status.md#s3t-table-223) (whose note says the maintenance config was 'echoed': only
icebergSnapshotManagement is copied; icebergCompaction is forced to status=disabled with 512/auto regardless
of the source's enabled 256/binpack) with the tag, policy, namespace-identity and system-tag cells. Empty
table, same-region destination, n=1.

Contradiction with [S3T-TABLE-223](../table-replication-status.md#s3t-table-223): 223's note says the source's maintenance configuration was 'echoed' on the
replica; 299 primed the source with compaction enabled 256/binpack + snapshots 5/48 and read the replica as
compaction DISABLED 512/auto + snapshots 5/48 - only icebergSnapshotManagement is copied. 223's own replica
data (compaction disabled 512/auto while a fresh table defaults to enabled 512/auto per 309) already fits 299;
the disagreement is 223's interpretation, not its data Resolution: keep both; [S3T-TABLE-299](../table-replication-status.md#s3t-table-299) is canonical for
what the replica copies (snapshot settings yes, compaction forced disabled, tags/policy not copied)
