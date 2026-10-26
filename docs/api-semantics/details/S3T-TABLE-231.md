<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-231: S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s; plain s3tables DeleteTable removes it (204)
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-231"></a>**S3T-TABLE-231** `delete-semantics` · impact medium · partially handled · verified 2026-10-13
  **S3 Metadata journal table (type=aws) outlives DeleteBucketMetadataConfiguration by >= 440 s; plain s3tables DeleteTable removes it (204)**
  s3 DeleteBucketMetadataConfiguration -> HTTP 204, but GetTable(journal tableArn) kept returning 200 for the
  whole 300 s watch (5 s polls) and again 440 s after the configuration was created (ListTables in the
  'b_<bucket>' namespace of the aws-s3 table bucket still listed 'journal'). Customer credentials could then
  remove the AWS-managed table and its container directly: DeleteTable(aws-s3 bucket ARN, namespace,
  'journal') -> HTTP 204 without versionToken, DeleteNamespace -> 204, DeleteTableBucket(aws-s3) -> 204;
  ListTableBuckets(type=aws) afterwards -> empty.
  - ACK: pre-delete-cleanup, docs-only · ops: DeleteTable, DeleteNamespace, DeleteTableBucket, GetTable,
    ListTables · fields: type, managedByService
  - repro: s3 CreateBucketMetadataConfiguration(journal) -> wait ACTIVE -> DeleteBucketMetadataConfiguration
    -> GetTable every 5 s for 300 s -> DeleteTable/DeleteNamespace/DeleteTableBucket on aws-s3
  - measurements: journal_still_present_after_s=440, watch_window_s=300, n=1
  - handling: partially handled via `generator.yaml:45-46; generator.yaml:60-65; apis/v1alpha1/types.go:40; apis/v1alpha1/types.go:83` - see Handling gaps
  - related: [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-089](../table-metadata-location.md#s3t-table-089) · evidence: table/creative/expiration-aws-managed-journal-table

## Notes

Side observation from the record-expiration creative probe, one table. Whether the service would have
garbage-collected the journal table later is unknown (the lab deleted it). Relevant to any controller that
creates metadata configurations or counts tables in the account: AWS-managed tables are not removed
synchronously and the aws-s3 bucket appears in ListTableBuckets(type=aws).

Partially handled in controller via: [GT-S3T-041](../service.md#gt-s3t-041) (controller hooks catalog entry): go_tag: json:"type,omitempty" overrides on Table.Type,
TableBucket.Type and Metadata.Iceberg.Schema.Fields.Type so the CRD exposes `type`.; [GT-S3T-044](../service.md#gt-s3t-044) (controller hooks catalog entry): none - not
modeled (scope:defer). The generator still emits the referenced shapes (IcebergCompactionSettings,
IcebergSnapshotManagementSettings, TableMaintenanc (scorer verdict: partial match)
