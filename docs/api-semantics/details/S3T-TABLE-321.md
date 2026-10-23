<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-321: RenameTable keeps the table versionToken; DeleteTable(current triple) with a bogus token -> ConflictException...
_Full entry and notes of one finding; its summary entry is in
[table-metadata-location.md](../table-metadata-location.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-321"></a>**S3T-TABLE-321** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-15
  **RenameTable keeps the table versionToken; DeleteTable(current triple) with a bogus token -> ConflictException...**
  Fresh ICEBERG table (no metadata): RenameTable(newName, no versionToken) -> HTTP 204 (4 renames); afterwards
  GetTable(tableArn).versionToken does NOT change (create token 20 chars, post-rename token 20 chars;
  GetTable.modifiedAt strictly increased across the 3 renames: True). DeleteTable(renamed triple,
  versionToken='ackqbogus0000000000000') -> ConflictException (HTTP 409) 'Provided version token does not
  match the table version token.'. DeleteTable(renamed triple, versionToken=<CreateTable token>) -> HTTP 204
  (second table) and -> HTTP 204 (first table, after 3 renames). DeleteTable(old pre-rename triple) ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: is_arn_primary_key, custom_delete · ops: RenameTable, DeleteTable, GetTable · fields: versionToken,
    name, namespace
  - repro: CreateTable -> RenameTable(newName) -> GetTable(tableArn) compare versionToken ->
    DeleteTable(versionToken=bogus) -> DeleteTable(versionToken=create token)
  - measurements: tables=2, renames=4, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009) · evidence: table/creative/reverify-set-a

## Notes

Re-verification precision gain for [S3T-TABLE-009](../table-metadata-location.md#s3t-table-009): states whether the 'pre-rename token still deletes' clause
is because rename does not rotate the token or because DeleteTable ignores/loosely checks versionToken. n=2
tables.

Full original title: RenameTable keeps the table versionToken; DeleteTable(current triple) with a bogus token
-> ConflictException, with the CreateTable token -> HTTP 204
