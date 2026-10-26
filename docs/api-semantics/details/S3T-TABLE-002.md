<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-002: RenameTable(newName) keeps tableARN/ids/locations; GetTable by old ARN -> 200 new name...
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-002"></a>**S3T-TABLE-002** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable(newName) keeps tableARN/ids/locations; GetTable by old ARN -> 200 new name...**
  On an ICEBERG table with schema, tags and a policy: RenameTable(newName) -> HTTP 204 (body keys: []). Then
  GetTable(tableArn=<pre-rename ARN>) -> HTTP 200 with name=<newName>; GetTable(bucket, ns, newName) -> HTTP
  200 and its payload is identical to the ARN read; GetTable(old triple) -> NotFoundException (HTTP 404) 'The
  specified table does not exist.'. Fields unchanged: tableARN, tableBucketId, namespaceId, warehouseLocation,
  metadataLocation, createdAt, createdBy, format, type, ownerAccountId. Fields changed: modifiedAt,
  modifiedBy, name.
  - ACK: is_arn_primary_key, custom_find · ops: RenameTable, GetTable · fields: tableARN, tableBucketId,
    namespaceId, warehouseLocation, versionToken, name
  - repro: CreateTable(ns1,t1,schema,tags) -> PutTablePolicy -> RenameTable(newName=t2) -> GetTable(tableArn),
    GetTable(ns1,t2), GetTable(ns1,t1)
  - measurements: rename_latency_ms=[149], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-004](../table-policy.md#s3t-table-004), [S3T-TABLE-032](../table.md#s3t-table-032) · hypotheses: H-T-050, H-T-055 · evidence:
    table/identity/rename-identity

## Notes

H-T-050 confirmed (ARN byte-identical: True; tableBucketId/namespaceId unchanged: True). H-T-055: rename
response is HTTP 204 with empty body; modifiedAt changed: True; modifiedBy changed: True; createdAt changed:
False; createdBy changed: False.

Full original title: RenameTable(newName) keeps tableARN/ids/locations; GetTable by old ARN -> 200 new name,
old triple -> NotFoundException (HTTP 404) 'The specified table does not exist.'
