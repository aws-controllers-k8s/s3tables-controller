<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-001: Fresh bucket/table: the 6 encryption/maintenance/storage-class/metadata Gets return 200 server defaults, none 404 (metrics/policy not read)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-001"></a>**S3T-TABLEBUCKET-001** `server-default` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Fresh bucket/table: the 6 encryption/maintenance/storage-class/metadata Gets return 200 server defaults, none 404 (metrics/policy not read)**
  On a table bucket and an ICEBERG table created without encryption, storage-class or maintenance settings:
  GetTableBucketEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm": "AES256"}};
  GetTableBucketMaintenanceConfiguration -> HTTP 200 {"configuration": {"icebergUnreferencedFileRemoval":
  {"status": "enabled", "settings": {"icebergUnreferencedFileRemoval": {"unreferencedDays": 3,
  "nonCurrentDays": 10}}}}}; GetTableEncryption -> HTTP 200 {"encryptionConfiguration": {"sseAlgorithm":
  "AES256"}}; GetTableStorageClass -> HTTP 200 {"storageClassConfiguration": {"storageClass": "STANDARD"}};
  GetTableMetadataLocation -> HTTP 200 {"versionToken": "<token>", "warehouseLocation": "<s3-uri>"};
  GetTableMaintenanceConfiguration -> HTTP 200 {"configuration": {"icebergCompaction": {"status": "enabled",
  "settings": {"icebergCompaction": {"targetFileSizeMB": 512, "strategy": "auto"}}},
  "icebergSnapshotManagement": {"status": "enabled", "settings": {"icebergSnapshotManagement":
  {"minSnapshotsToKeep": 1, "maxSnapshotAgeHours": 120}}}}}. No sub-resource read returned NotFoundException
  for an unset setting.
  - ACK: late_initialize, hook:sdk_read_one_post_set_output · ops: GetTableBucketEncryption,
    GetTableBucketMaintenanceConfiguration, GetTableEncryption, GetTableStorageClass,
    GetTableMetadataLocation, GetTableMaintenanceConfiguration · fields: encryptionConfiguration.sseAlgorithm,
    storageClassConfiguration.storageClass, configuration, metadataLocation
  - repro: CreateTableBucket(name only) -> CreateNamespace -> CreateTable(format=ICEBERG) -> each Get*
    sub-resource read
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](../tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-173](../tablebucket-maintenance-metrics.md#s3t-tablebucket-173), [S3T-TABLEBUCKET-171](../tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-056](../tablebucket-maintenance-metrics.md#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](../tablebucket-encryption-storage.md#s3t-tablebucket-059) ·
    evidence: tablebucket/round-trip/smoke

## Notes

Through the API a controller cannot tell 'never set by the user' from 'set to the default' for these settings;
a nil spec field compared against the Get* value would show a permanent delta unless late-initialized or
ignored. GetTableMetadataLocation on a fresh table has versionToken+warehouseLocation but no metadataLocation.
Single run.

Contradiction with [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093): 001 ends 'No sub-resource read returned NotFoundException for an unset
setting' (6 Get* reads: bucket encryption/maintenance, table encryption/storage class/metadata
location/maintenance); 093 and 100 show GetTableBucketMetricsConfiguration on a bucket with no metrics
configuration -> NotFoundException 404 'No such metric configuration exists', and 158 shows
GetTableBucketPolicy 404 on a fresh re-created bucket - the 'none 404' statement holds only for the six nouns
001 read Resolution: keep both; [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093) is canonical for the metrics noun; 001 over-generalizes in
its closing sentence, see title_fixes
