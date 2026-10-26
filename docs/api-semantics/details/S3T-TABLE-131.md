<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-131: CreateTable WITH metadata by a role lacking s3tables:PutTableData -> ForbiddenException 403...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-131"></a>**S3T-TABLE-131** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **CreateTable WITH metadata by a role lacking s3tables:PutTableData -> ForbiddenException 403...**
  Role policy: Allow s3tables:CreateTable, GetTable, DeleteTable on the table bucket and bucket/*; nothing
  else. CreateTable(format=ICEBERG, no metadata) by the role -> HTTP 200 (attempts while IAM propagated:
  ['HTTP 200']). CreateTable with metadata.iceberg.schema -> ForbiddenException (HTTP 403) 'Access Denied'.
  Admin GetTable on that name -> NotFoundException 404; admin ListTables contains it: False. Immediate retry
  by the role -> ForbiddenException 403. Admin CreateTable of the same name with the same metadata -> HTTP 200.
  Role CreateTable with metadata={iceberg:{properties:{k:v}}} (no schema) -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTable, ListTables · fields: metadata,
    metadata.iceberg.schema
  - repro: IAM role with s3tables:CreateTable+GetTable+DeleteTable only -> AssumeRole ->
    CreateTable(metadata.iceberg.schema) -> admin GetTable/ListTables -> admin CreateTable same name
  - measurements: assume_role_wait_s=10.1, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](../tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](../tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-132](../table-policy.md#s3t-table-132), [S3T-TABLE-140](../table-policy.md#s3t-table-140), [S3T-TABLE-142](../table-policy.md#s3t-table-142), [S3T-TABLEBUCKET-129](../tablebucket-policy.md#s3t-tablebucket-129) ·
    hypotheses: H-T-105, H-T-106 · evidence: table/error-taxonomy/iam-putdata-denied

## Notes

H-T-105 confirmed: ForbiddenException (HTTP 403) with the terse message 'Access Denied' (not the IAM-style
'User ... is not authorized' text that identity-policy denials produce for other operations); admin GetTable
-> NotFoundException 404 and ListTables does not list the name. H-T-106 REFUTED: nothing is left behind - the
immediate retry by the role gives the same 403 (no ConflictException) and the admin's CreateTable of the same
name with the same metadata -> 200. Also: with metadata={iceberg:{properties:...}} (no schema) the role got
BadRequestException 400 'The specified metadata is not valid' - metadata validation runs before the
PutTableData authorization. n=1 per cell; AssumeRole became possible 10 s after CreateRole; the policy was
effective on the first CreateTable.

Full original title: CreateTable WITH metadata by a role lacking s3tables:PutTableData -> ForbiddenException
403; no table left behind (admin GetTable NotFoundException 404); admin CreateTable of the same name
afterwards -> HTTP 200
