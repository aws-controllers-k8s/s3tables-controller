<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-225: After DeleteTableReplication the replica stays managedByService='replication.s3tables.amazonaws.com' for 2 min...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-225"></a>**S3T-TABLE-225** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableReplication the replica stays managedByService='replication.s3tables.amazonaws.com' for 2 min...**
  DeleteTableReplication(current token) -> HTTP 204; GetTableReplicationStatus(source) right after ->
  NotFoundException (HTTP 404) 'The specified replication configuration does not exist.'. GetTable(replica)
  every 10 s for 2 min: [(0.1, 'replication.s3tables.amazonaws.com|mti')]. Then
  UpdateTableMetadataLocation(replica) -> BadRequestException (HTTP 400) 'Updates to service managed tables
  are not allowed.'; PutTableRecordExpirationConfiguration(replica) -> MethodNotAllowedException (HTTP 405)
  'Table record expiration configuration API calls to this table are not allowed'.
  - ACK: pre-delete-cleanup, is_read_only, docs-only · ops: DeleteTableReplication, GetTable,
    UpdateTableMetadataLocation, PutTableRecordExpirationConfiguration, GetTableReplicationStatus · fields:
    managedByService, managedTableInformation
  - repro: completed replication -> DeleteTableReplication -> GetTable(replica) @10 s x 2 min -> write ops on
    the replica
  - measurements: replica_managed_observed_s=[0.1], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-024](../tablebucket-replication.md#s3t-tablebucketreplication-024), [S3T-TABLEBUCKETREPLICATION-025](../tablebucket-replication.md#s3t-tablebucketreplication-025),
    [S3T-TABLEREPLICATION-035](../table-replication-status.md#s3t-tablereplication-035), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLE-195](../table-replication-status.md#s3t-table-195), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147),
    [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) · hypotheses: H-U-170, H-U-171, H-U-172 · evidence:
    table/state-machine/repl-status-lifecycle

## Notes

H-U-170 confirmed at table level (replica remains, still service-managed, writes rejected;
DeleteTable(replica) succeeded later in this probe). H-U-171 REFUTED (no promotion within 2 min). H-U-172
qualified: status -> 404 immediately; the no-propagation half was shown at bucket level
([S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015)) and not repeated here.

Full original title: After DeleteTableReplication the replica stays
managedByService='replication.s3tables.amazonaws.com' for 2 min; commit -> BadRequestException; expiration Put
-> MethodNotAllowedException; status(source) -> NotFoundException

Contradiction with [S3T-TABLE-307](../table-replication-status.md#s3t-table-307): orphan after DeleteTableReplication: 225 records managedTableInformation
present ('mti') on every 10 s sample for 2 min; 307 records replicationInformation present=False on every 30 s
sample for 874 s (same managedByService in both). 196 shows managedTableInformation still naming the deleted
SOURCE when the source table (not the config) is deleted. May reconcile if the parent key stays as an empty
object, but neither record shows the raw body; 225 was cross-region, 307 same-region Resolution: re-probe:
state-machine - completed replica -> DeleteTableReplication -> raw GetTable(replica) at 0/10/60/300 s
recording managedByService, managedTableInformation and replicationInformation presence; one same-region and
one cross-region destination; compare with the DeleteTable(source) orphan

Contradiction with [S3T-TABLE-329](../table-replication-status.md#s3t-table-329), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196): FU2-03 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: 225 holds and 307's 'replicationInformation
present=False' is a key-path artifact, not a different behavior: in the GetTable body replicationInformation
is NESTED under managedTableInformation (there is no top-level replicationInformation key, so a top-level
check is always False). The orphan left by DeleteTableReplication keeps type=customer,
managedByService=replication.s3tables.amazonaw Resolution: keep all; [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) is canonical for the
settled rule
