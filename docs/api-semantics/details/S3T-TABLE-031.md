<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-031: Fresh table without metadata: raw GetTable JSON has metadataLocation absent, managedByService absent, managedTableInformation absent...
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-031"></a>**S3T-TABLE-031** `response-fidelity` · impact high · handled · verified 2026-10-13
  **Fresh table without metadata: raw GetTable JSON has metadataLocation absent, managedByService absent, managedTableInformation absent...**
  Raw GetTable(tableArn) body for a table created with format=ICEBERG and no metadata (HTTP 200): keys
  ['createdAt', 'createdBy', 'format', 'modifiedAt', 'name', 'namespace', 'namespaceId', 'ownerAccountId',
  'tableARN', 'tableBucketId', 'type', 'versionToken', 'warehouseLocation']. metadataLocation: absent;
  warehouseLocation: present(str) (s3:// URI: True); managedByService: absent value="<absent>";
  managedTableInformation: absent; type=customer; namespace raw=["<ns>"]. GetTableMetadataLocation -> HTTP 200
  with metadataLocation absent, warehouseLocation present(str), versionToken equal to the create token: True.
  - ACK: compare.nil_equals_zero_value, is_read_only, late_initialize · ops: GetTable,
    GetTableMetadataLocation · fields: metadataLocation, warehouseLocation, managedByService,
    managedTableInformation, type, namespace
  - repro: CreateTable(format=ICEBERG) without metadata; GetTable(tableArn) and inspect the raw HTTP body
    (botocore drops JSON nulls); GetTableMetadataLocation
  - handling: handled via `generator.yaml:10-13; generator.yaml:31-39; generator.yaml:76-83; generator.yaml:105-110; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-014](../table-metadata-location.md#s3t-table-014), [S3T-TABLE-033](../table-schema.md#s3t-table-033), [S3T-TABLE-309](../table-maintenance.md#s3t-table-309), [S3T-TABLE-315](../table-metadata-location.md#s3t-table-315), [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-001](../table-metadata-location.md#s3t-table-001),
    [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-177](../table-replication.md#s3t-table-177) · hypotheses: H-T-029, H-T-033, H-U-189,
    H-T-012 · evidence: table/round-trip/fresh-table-fields

## Notes

H-T-029 confirmed (metadataLocation absent on both reads; warehouseLocation always present). H-U-189 REFUTED:
managedByService is not returned as false, it is absent, so both managed* members are omitted symmetrically
(managedTableInformation absent). H-T-033 confirmed (type=customer, managed* keys as stated; the raw type=aws
request is a separate finding). H-T-012 confirmed again (namespace is a 1-element list).

Full original title: Fresh table without metadata: raw GetTable JSON has metadataLocation absent,
managedByService absent, managedTableInformation absent, type=customer
