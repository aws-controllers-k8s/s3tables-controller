<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-136: Replica table: GetTable has managedByService='replication.s3tables.amazonaws.com' and replicationInformation.sourceTableARN...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-136"></a>**S3T-TABLE-136** `read-gap` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replica table: GetTable has managedByService='replication.s3tables.amazonaws.com' and replicationInformation.sourceTableARN...**
  The replica created by the service in the destination bucket has the same namespace and table name as the
  source (yes); GetTable(replica) -> HTTP 200 with type='customer',
  managedByService='replication.s3tables.amazonaws.com',
  managedTableInformation.replicationInformation.sourceTableARN equals the source table ARN; metadataLocation
  is under the replica's own warehouseLocation with the same file name as the source.
  GetTableStorageClass(replica)=INTELLIGENT_TIERING while the source table is STANDARD and the destination
  bucket default is INTELLIGENT_TIERING; GetTableEncryption(replica)=AES256;
  GetTableMaintenanceConfiguration(replica) -> HTTP 200. On the DESTINATION bucket GetTableBucketReplication
  -> NotFoundException (HTTP 404) 'The specified replication configuration does not exist.';
  GetTableReplication(replica) -> BadRequestException (HTTP 400) 'Replication operations are not allowed on
  service managed tables.'; GetTableReplicationStatus(replica) -> BadRequestException (HTTP 400) 'Replication
  operations are not allowed on service managed tables.'.
  - ACK: is_read_only, late_initialize, custom_find · ops: GetTable, GetTableStorageClass, GetTableEncryption,
    GetTableBucketReplication, GetTableReplication, GetTableReplicationStatus · fields: managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, type,
    storageClassConfiguration.storageClass
  - repro: PutTableReplication -> wait completed -> GetTable(destinationTableArn) and sub-resource reads in
    the destination region
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](../table-replication.md#s3t-table-177), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299),
    [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041),
    [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030), [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019),
    [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028) · hypotheses: H-S-117, H-S-059 · evidence: table/round-trip/repl-put-get

## Notes

H-S-117 confirmed: the bucket-level read on the destination is NotFoundException; the replica itself exposes
its source through managedTableInformation.replicationInformation. H-S-059 qualified (no KMS tested):
cross-region replication works and the replica took storage class INTELLIGENT_TIERING (destination default
INTELLIGENT_TIERING vs source STANDARD).

Full original title: Replica table: GetTable has managedByService='replication.s3tables.amazonaws.com' and
replicationInformation.sourceTableARN; storage class INTELLIGENT_TIERING (dest default), encryption AES256;
bucket Get on dest -> NotFoundException
