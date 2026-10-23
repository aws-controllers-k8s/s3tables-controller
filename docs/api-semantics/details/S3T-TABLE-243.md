<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-243: All 12 table/namespace reads with the other region's bucket ARN or tableArn -> 400 'provided Region does not match', uniform, both ways
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-243"></a>**S3T-TABLE-243** `cross-region` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **All 12 table/namespace reads with the other region's bucket ARN or tableArn -> 400 'provided Region does not match', uniform, both ways**
  A same-named bucket/namespace/table exists in BOTH us-west-2 and us-east-1. us-west-2 endpoint with the
  us-east-1 identifiers: GetTable(triple), GetTableMetadataLocation, GetTableStorageClass, GetTableEncryption,
  GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetNamespace, ListTables,
  GetTable(tableArn), GetTableRecordExpirationConfiguration, ListTagsForResource -> BadRequestException (HTTP 400)
  'The provided Region does not match the Region of the resource.'. us-east-1 endpoint with the us-west-2
  identifiers: GetTable(triple), GetTableMetadataLocation, GetTableStorageClass, GetTableEncryption,
  GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus, GetNamespace, ListTables,
  GetTable(tableArn), GetTableRecordExpirationConfiguration, ListTagsForResource -> BadRequestException (HTTP 400)
  'The provided Region does not match the Region of the resource.'. No read returned 200 (neither the other
  region's table nor this region's same-named table) and none returned NotFoundException.
  - ACK: terminal_codes, is_arn_primary_key · ops: GetTable, GetTableMetadataLocation, GetTableStorageClass,
    GetTableEncryption, GetTablePolicy, GetTableMaintenanceConfiguration, GetTableMaintenanceJobStatus,
    GetNamespace, ListTables, GetTableRecordExpirationConfiguration, ListTagsForResource · fields:
    tableBucketARN, tableArn, resourceArn
  - repro: Create bucket+ns+table with identical names in us-west-2 and us-east-1; from each endpoint call
    every table read with the OTHER region's ARNs
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-TABLE-179](../table-maintenance.md#s3t-table-179), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002), [S3T-TABLE-190](../table-record-expiration.md#s3t-table-190) ·
    hypotheses: H-T-213, H-T-214, H-T-215, H-U-069, H-U-106, H-T-075 · evidence:
    table/cross-region/foreign-region-arn-reads-writes

## Notes

H-T-213 confirmed (GetTableMetadataLocation -> BadRequestException, not 404); H-T-214 confirmed
(GetTableStorageClass/Encryption/MetadataLocation/GetTable share the exact wire code and message); H-T-215
(contrarian: read-only ops skip the region check -> 404 or region-A body) refuted: no 200 and no 404 even
though a same-named table exists in the endpoint's own region; H-U-069 confirmed (us-east-1 endpoint,
us-west-2 ARN: GetTablePolicy/GetTableMaintenanceConfiguration -> 400); H-U-106 confirmed (tableArn of the
other region -> 400, both directions); H-T-075 read part (ListTagsForResource, GetTable(tableArn)) confirmed.

Full original title: All 12 table/namespace reads with the other region's bucket ARN or tableArn ->
BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.' (uniform, both
directions)
