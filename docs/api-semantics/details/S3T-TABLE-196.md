<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-196: After DeleteTable(source) the replicas persist (+0/+90 s) as read-only service-managed tables: commit -> 400; DeleteTable(replica) -> 204
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-196"></a>**S3T-TABLE-196** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTable(source) the replicas persist (+0/+90 s) as read-only service-managed tables: commit -> 400; DeleteTable(replica) -> 204**
  ListTables(D, ns) right after the last source delete: ['tc', 'td']; 90 s later: ['tc', 'td'] (ta = deleted
  at no-destinations, tb = config deleted at +2 s, tc = deleted at 'pending', td = deleted at completed, te =
  nonexistent destination). Per leftover replica: {"tc": {"commit": "BadRequestException (HTTP 400) 'Updates
  to service managed tables are not allowed.'", "delete": "HTTP 204", "managedByService":
  "replication.s3tables.amazonaws.com", "mti_names_deleted_source": true}, "td": {"commit":
  "BadRequestException (HTTP 400) 'Updates to service managed tables are not allowed.'", "delete": "HTTP 204",
  "managedByService": "replication.s3tables.amazonaws.com", "mti_names_deleted_source": true}}. ListTables
  after deleting them: [].
  - ACK: pre-delete-cleanup, is_read_only · ops: ListTables, GetTable, UpdateTableMetadataLocation,
    DeleteTable · fields: managedByService, managedTableInformation
  - repro: after the deletes: ListTables(D) at +0 and +90 s ->
    GetTable/UpdateTableMetadataLocation/DeleteTable on each leftover
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](../table-replication-status.md#s3t-table-139), [S3T-TABLEBUCKETREPLICATION-016](../tablebucket-replication.md#s3t-tablebucketreplication-016), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225),
    [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-195](../table-replication-status.md#s3t-table-195), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLEBUCKET-118](../tablebucket-replication.md#s3t-tablebucket-118), [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012),
    [S3T-TABLEBUCKETREPLICATION-015](../tablebucket-replication.md#s3t-tablebucketreplication-015), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEBUCKETREPLICATION-005](../tablebucket-replication.md#s3t-tablebucketreplication-005), [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) · hypotheses: H-U-169, H-U-170, H-U-198 · evidence:
    table/dependencies/repl-source-delete

## Notes

H-U-169 orphan half confirmed: the replica of a deleted source remains service-managed and read-only but IS
deletable (HTTP 204). H-U-198 confirmed: DeleteTableReplication 2 s after the Put -> HTTP 204, status ->
NotFoundException, replica of tb never appeared. H-U-170 (replica deletable after the source config is gone)
confirmed.

Full original title: After the sources were deleted: replicas in D at +0 ['tc', 'td'], at +90 s ['tc', 'td'];
leftovers managedByService=['replication.s3tables.amazonaws.com'], commit -> ['BadRequestException'],
DeleteTable -> ['204']

Contradiction with [S3T-TABLE-329](../table-replication-status.md#s3t-table-329), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307): FU2-03 settled by follow-up probe
tablereplication/creative/fu2-03-05-orphan-body-dead-destination: 225 holds and 307's 'replicationInformation
present=False' is a key-path artifact, not a different behavior: in the GetTable body replicationInformation
is NESTED under managedTableInformation (there is no top-level replicationInformation key, so a top-level
check is always False). The orphan left by DeleteTableReplication keeps type=customer,
managedByService=replication.s3tables.amazonaw Resolution: keep all; [S3T-TABLE-329](../table-replication-status.md#s3t-table-329) is canonical for the
settled rule
