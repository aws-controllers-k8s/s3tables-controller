<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-223: GetTable(replica) vs GetTable(source): type='customer', managedByService='replication.s3tables.amazonaws.com'...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-223"></a>**S3T-TABLE-223** `response-fidelity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTable(replica) vs GetTable(source): type='customer', managedByService='replication.s3tables.amazonaws.com'...**
  Replica GetTable: type='customer', managedByService='replication.s3tables.amazonaws.com',
  managedTableInformation={"replicationInformation": {"sourceTableARN":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"}},
  format='ICEBERG'; fields equal to the source: ['createdBy', 'format', 'name', 'namespace', 'ownerAccountId',
  'type']; fields that differ: ['createdAt', 'managedByService', 'managedTableInformation', 'modifiedAt',
  'namespaceId', 'tableARN', 'tableBucketId', 'versionToken', 'warehouseLocation']. Source GetTable has
  managedByService=None and managedTableInformation=None. metadataLocation: source None, replica None.
  ListTables(D) entry for the replica: {"createdAt": "2026-10-13 23:41:41.473835+00:00", "managedByService":
  "replication.s3tables.amazonaws.com", "modifiedAt": "2026-10-13 23:41:41.669341+00:00", "name":
  "ackq_f81240_t", "namespace": ["ackq_f81240_ns"], "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "type": "customer"}. Replica sub-resource reads: storage class {"storageClassConfiguration":
  {"storageClass": "STANDARD"}}; encryption {"encryptionConfiguration": {"sseAlgorithm": "AES256"}};
  maintenance {"configuration": {"icebergCompaction": {"settings": {"icebergCompaction": {"strategy": "auto",
  "targetFileSizeMB": 512}}, "status": "disabled"}, "icebergSnapshotManagement": {"settings":
  {"icebergSnapshotManagement": {"maxSnapshotAgeHours": 48, "minSnapshotsToKeep": 2}}, "status": "enabled"}},
  "tableARN":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8"};
  metadata location {"versionToken": "c52faa2fa3f8cd457bbe", "warehouseLocation":
  "s3://93cf68ab-77f5-416b-zgdhmbc3kh1cb9n9eq6a3aenwqidouse1b--table-s3"}.
  - ACK: is_read_only, custom_find, late_initialize · ops: GetTable, ListTables, GetTableStorageClass,
    GetTableEncryption, GetTableMaintenanceConfiguration · fields: type, managedByService,
    managedTableInformation.replicationInformation.sourceTableARN, metadataLocation, warehouseLocation,
    versionToken
  - repro: replication terminal -> GetTable(source) and GetTable(destinationTableArn); diff all members
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031), [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](../table-replication.md#s3t-table-177), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299),
    [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307) · hypotheses: H-U-163, H-U-164, H-U-193 · evidence:
    table/state-machine/repl-status-lifecycle

## Notes

H-U-163 qualified: managedByService is the STRING 'replication.s3tables.amazonaws.com' (not a boolean) and
managedTableInformation.replicationInformation.sourceTableARN names the source; type stays 'customer'. H-U-164
refuted (type is in the enum). H-U-193 confirmed: own tableARN/warehouseLocation/versionToken; on an EMPTY
source both metadataLocation values are absent. The source's maintenance config (snapshot management enabled)
was echoed on the replica.

Full original title: GetTable(replica) vs GetTable(source): type='customer',
managedByService='replication.s3tables.amazonaws.com', managedTableInformation present; identical fields
['createdBy', 'format', 'name', 'namespace', 'ownerAccountId', 'type']

Contradiction with [S3T-TABLE-299](../table-replication-status.md#s3t-table-299): 223's note says the source's maintenance configuration was 'echoed' on the
replica; 299 primed the source with compaction enabled 256/binpack + snapshots 5/48 and read the replica as
compaction DISABLED 512/auto + snapshots 5/48 - only icebergSnapshotManagement is copied. 223's own replica
data (compaction disabled 512/auto while a fresh table defaults to enabled 512/auto per 309) already fits 299;
the disagreement is 223's interpretation, not its data Resolution: keep both; [S3T-TABLE-299](../table-replication-status.md#s3t-table-299) is canonical for
what the replica copies (snapshot settings yes, compaction forced disabled, tags/policy not copied)
