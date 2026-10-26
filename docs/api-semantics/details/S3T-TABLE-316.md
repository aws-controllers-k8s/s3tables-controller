<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-316: modifiedBy is absent on a fresh table and still absent after UpdateTableMetadataLocation; RenameTable populates it (== caller)
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-316"></a>**S3T-TABLE-316** `server-default` · impact medium · handled · verified 2026-10-14
  **modifiedBy is absent on a fresh table and still absent after UpdateTableMetadataLocation; RenameTable populates it (== caller)**
  GetNamespace: ownerAccountId present True (== caller account True), createdBy present True (== caller True),
  no modifiedBy/modifiedAt member. GetTable on both fresh tables: ownerAccountId present/== caller True/True,
  createdBy present/== caller True/True, modifiedBy present False, modifiedAt == createdAt True; still absent
  at +3 s: True. UpdateTableMetadataLocation(versionToken, <warehouse>/metadata/00001-<uuid>.metadata.json) ->
  HTTP 200 (wire keys ['metadataLocation', 'name', 'namespace', 'tableARN', 'versionToken']): afterwards
  modifiedBy present False (== caller False), modifiedAt changed True, versionToken changed True (Update
  response versionToken == GetTable's: True). RenameTable(newName only, same namespace) -> HTTP 204:
  afterwards modifiedBy present True (== caller True), modifiedAt changed True, versionToken changed False.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare, late_initialize · ops: GetNamespace, GetTable,
    UpdateTableMetadataLocation, RenameTable · fields: ownerAccountId, createdBy, modifiedBy, modifiedAt,
    versionToken
  - repro: CreateNamespace+GetNamespace; CreateTable x2 + GetTable at 0 s and +3 s;
    UpdateTableMetadataLocation on one table, RenameTable(newName) on the other; GetTable after each
  - measurements: n_tables=2, n=2
  - handling: handled via `test/e2e/tests/test_table_bucket.py:115-118; pkg/resource/table_bucket/sdk.go:303-304; generator.yaml:57-98; generator.yaml:137-157`
  - related: [S3T-TABLE-015](../table-schema.md#s3t-table-015), [S3T-TABLE-034](../table-metadata-location.md#s3t-table-034), [S3T-TABLE-019](../table-metadata-location.md#s3t-table-019), [S3T-TABLE-123](../table-metadata-location.md#s3t-table-123), [S3T-TABLE-311](../table-schema.md#s3t-table-311), [S3T-TABLE-005](../table-metadata-location.md#s3t-table-005),
    [S3T-TABLE-007](../table-metadata-location.md#s3t-table-007), [S3T-TABLE-112](../table-metadata-location.md#s3t-table-112), [S3T-TABLE-193](../table-metadata-location.md#s3t-table-193), [S3T-TABLE-320](../table-metadata-location.md#s3t-table-320), [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009), [S3T-TABLE-101](../table-metadata-location.md#s3t-table-101), [S3T-TABLE-104](../table-metadata-location.md#s3t-table-104),
    [S3T-TABLE-124](../table-metadata-location.md#s3t-table-124) · evidence: table/round-trip/fu-12-create-vs-get-members

## Notes

FU-11. Single-principal lab: createdBy/modifiedBy/ownerAccountId all equal the caller's 12-digit account id,
so a modifiedBy change to a different principal is not observable (only presence and equality were recorded).
UpdateTableMetadataLocation bumped modifiedAt and versionToken but did NOT populate modifiedBy (n=1, this
probe); RenameTable populated modifiedBy and bumped modifiedAt but left versionToken unchanged (n=2: this
probe and table/sub-resources/fu-07-rename-kms-table, where the update's versionToken bump was also seen,
n=2). A controller that copies modifiedBy into Status sees nil until a rename.
