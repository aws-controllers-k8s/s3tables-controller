<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-118: DeleteTableBucket(destination) with a replica inside -> BadRequestException; once emptied -> 204...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-118"></a>**S3T-TABLEBUCKET-118** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **DeleteTableBucket(destination) with a replica inside -> BadRequestException; once emptied -> 204...**
  With t replicated ('completed' at +62.22s) into the destination bucket D2: DeleteTableBucket(D2) ->
  BadRequestException (HTTP 400) 'The bucket that you tried to delete is not empty.'. Emptying D2 (the
  service-created replica and its namespace): 2 deletes, codes ['204']. DeleteTableBucket(D2) -> HTTP 204;
  GetTableBucket(D2) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'. Source side over
  the next 60 s (no new Put): GetTableReplicationStatus [(0.0, 'completed')]; GetTableReplication -> HTTP 200
  (configuration still stored). PutTableReplication(identical, current token) with the destination bucket gone
  -> HTTP 200; status transitions [(0.05, 'completed'), (86.81, 'pending'), (87.84, 'failed:The specified
  destination bucket does not exist.')]. DeleteTableReplication -> HTTP 204.
  - ACK: pre-delete-cleanup, synced.when, requeue · ops: DeleteTableBucket, DeleteTable, DeleteNamespace,
    GetTableReplicationStatus, GetTableReplication, PutTableReplication, DeleteTableReplication · fields:
    destinations[].replicationStatus, destinations[].failureMessage
  - repro: replica completed -> DeleteTableBucket(D) -> empty D -> DeleteTableBucket(D) -> poll source status
    60 s -> re-Put (token) -> poll -> DeleteTableReplication
  - measurements: put_to_terminal_s=[62.22], re_put_to_failed_s=[87.84], delete_nonempty_400_n=[2],
    delete_empty_204_n=[2], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-139](../table-replication-status.md#s3t-table-139), [S3T-TABLEREPLICATION-005](../table-replication-status.md#s3t-tablereplication-005), [S3T-TABLE-138](../table-replication-status.md#s3t-table-138), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLEBUCKET-012](../tablebucket-replication.md#s3t-tablebucket-012),
    [S3T-TABLE-306](../table-replication-status.md#s3t-table-306), [S3T-TABLEREPLICATION-040](../table-replication-status.md#s3t-tablereplication-040), [S3T-TABLEREPLICATION-042](../table-replication-status.md#s3t-tablereplication-042), [S3T-TABLEREPLICATION-012](../table-replication.md#s3t-tablereplication-012),
    [S3T-TABLEREPLICATION-023](../table-replication.md#s3t-tablereplication-023) · hypotheses: H-U-169, H-U-170 · evidence:
    tablebucket/dependencies/repl-dest-teardown

## Notes

Destination-bucket deletion cell (n=2 for the two DeleteTableBucket codes: run 1 on the shard's destination
bucket gave the same 400 'not empty' then 204): the destination owner can delete replicas (['204']) and then
the bucket; the source keeps its configuration and its last status until a new Put. Run 1 side effect: a
replication that was in flight toward the deleted bucket ended 'failed: The specified destination bucket does
not exist.' at +62 s (observed by table/cross-region/repl-kms-parity run 1). H-U-169/H-U-170: replicas are
deletable by the destination owner, so orphans do not block DeleteTableBucket forever.

Full original title: DeleteTableBucket(destination) with a replica inside -> BadRequestException; once emptied
-> 204; source status then 'completed' for 60 s; re-Put -> 200 then 'failed:The specified destination bucket
does not exist.'
