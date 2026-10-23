<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-047: CreateTable format other than exact 'ICEBERG' ('iceberg', DELTA, HUDI, '' or omitted) -> BadRequestException 400 'format is not valid'
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-047"></a>**S3T-TABLE-047** `request-validation` · impact medium · handled · verified 2026-10-13
  **CreateTable format other than exact 'ICEBERG' ('iceberg', DELTA, HUDI, '' or omitted) -> BadRequestException 400 'format is not valid'**
  With SDK validation disabled, no metadata: format='iceberg' -> BadRequestException (HTTP 400) 'The specified
  open table format is not valid'. 'DELTA' -> BadRequestException (HTTP 400) 'The specified open table format
  is not valid'. 'HUDI' -> BadRequestException (HTTP 400) 'The specified open table format is not valid'. ''
  -> BadRequestException (HTTP 400) 'The specified open table format is not valid'. format omitted ->
  BadRequestException (HTTP 400) 'The specified open table format is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: format
  - repro: CreateTable with each format value (parameter_validation=False); GetTable afterwards
  - handling: handled via `test/e2e/tests/test_table.py:175-193; test/e2e/resources/table.yaml; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-079](../table.md#s3t-table-079), [S3T-TABLE-078](../table.md#s3t-table-078), [S3T-TABLE-035](../table.md#s3t-table-035), [S3T-TABLE-080](../table.md#s3t-table-080) · hypotheses: H-T-127, H-T-128 ·
    evidence: table/weird-inputs/schema-variants

## Notes

H-T-127 confirmed; H-T-128 refuted (omitted format -> BadRequestException 400). One attempt each.

Full original title: CreateTable format: 'iceberg' -> BadRequestException 400; DELTA -> BadRequestException
400; HUDI -> BadRequestException 400; '' -> BadRequestException 400; omitted -> BadRequestException 400
