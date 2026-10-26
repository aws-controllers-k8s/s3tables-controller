<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-307: Orphaned replica after DeleteTableReplication: still managedByService for 874 s (30 samples); commit/maintenance 400...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-307"></a>**S3T-TABLE-307** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Orphaned replica after DeleteTableReplication: still managedByService for 874 s (30 samples); commit/maintenance 400...**
  DeleteTableReplication(current token) -> HTTP 204. GetTable(replica ARN) every 30 s for 874 s:
  managedByService=replication.s3tables.amazonaws.com, replicationInformation present=False, type=customer on
  every sample. At the end: UpdateTableMetadataLocation(replica) -> BadRequestException/400 'Updates to
  service managed tables are not allowed.'; PutTableMaintenanceConfiguration -> BadRequestException/400 'The
  specified maintenance type is not valid for service managed tables.'; PutTableReplication with the replica
  as source -> BadRequestException/400 'Replication operations are not allowed on service managed tables.'.
  - ACK: is_read_only, custom_delete, pre-delete-cleanup · ops: DeleteTableReplication, GetTable,
    UpdateTableMetadataLocation, PutTableMaintenanceConfiguration, PutTableReplication · fields:
    managedByService, replicationInformation
  - repro: table replication completed -> DeleteTableReplication -> GetTable(replica) every 30 s for 15 min ->
    write attempts on the replica
  - measurements: watched_s=874, samples=30, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224),
    [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLE-195](../table-replication-status.md#s3t-table-195), [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) · evidence:
    tablereplication/creative/controller-loop

## Notes

extends [S3T-TABLE-225](../table-replication-status.md#s3t-table-225) (>= 2 min) to a 15 min watch. Decision rule for a Table CR adopting a replica: treat
managedByService != '' as read-only for metadata/maintenance; the only converging path to a writable table is
DeleteTable + CreateTable (new ARN). Single run.

Full original title: Orphaned replica after DeleteTableReplication: still managedByService for 874 s (30
samples); commit/maintenance 400; PutTableReplication on it 400

Contradiction with [S3T-TABLE-225](../table-replication-status.md#s3t-table-225): orphan after DeleteTableReplication: 225 records managedTableInformation
present ('mti') on every 10 s sample for 2 min; 307 records replicationInformation present=False on every 30 s
sample for 874 s (same managedByService in both). 196 shows managedTableInformation still naming the deleted
SOURCE when the source table (not the config) is deleted. May reconcile if the parent key stays as an empty
object, but neither record shows the raw body; 225 was cross-region, 307 same-region Resolution: re-probe:
state-machine - completed replica -> DeleteTableReplication -> raw GetTable(replica) at 0/10/60/300 s
recording managedByService, managedTableInformation and replicationInformation presence; one same-region and
one cross-region destination; compare with the DeleteTable(source) orphan

Contradiction with [S3T-TABLE-329](../table-replication-status.md#s3t-table-329), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196): FU2-03 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: 225 holds and 307's 'replicationInformation
present=False' is a key-path artifact, not a different behavior: in the GetTable body replicationInformation
is NESTED under managedTableInformation (there is no top-level replicationInformation key, so a top-level
check is always False). The orphan left by DeleteTableReplication keeps type=customer,
managedByService=replication.s3tables.amazonaw Resolution: keep all; [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) is canonical for the
settled rule

Consolidation note: Follow-up re-probe outcome: 'replicationInformation absent' was a key-path artifact: it is
nested under managedTableInformation and present 0-479 s; [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) settles it Resolution: keep as
evidence; superseded reading - [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) settles it
