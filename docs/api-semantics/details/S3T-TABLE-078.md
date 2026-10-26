<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-078: CreateTable name: 255 chars, 'a__b', trailing '_', leading digit, 'aws_*' ok; 256, leading '_', '__', upper, '-', '.', space, non-ASCII 400
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-078"></a>**S3T-TABLE-078** `request-validation` · impact high · handled · verified 2026-10-13
  **CreateTable name: 255 chars, 'a__b', trailing '_', leading digit, 'aws_*' ok; 256, leading '_', '__', upper, '-', '.', space, non-ASCII 400**
  CreateTable(format=ICEBERG) per name cell: len255 -> HTTP 200; len256 -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; non_ascii -> BadRequestException (HTTP 400) 'The specified table name
  is not valid.'; hyphen -> BadRequestException (HTTP 400) 'The specified table name is not valid.'; dot ->
  BadRequestException (HTTP 400) 'The specified table name is not valid.'; space -> BadRequestException (HTTP 400)
  'The specified table name is not valid.'; uppercase -> BadRequestException (HTTP 400) 'The specified table
  name is not valid.'; leading_underscore -> BadRequestException (HTTP 400) 'The specified table name is not
  valid.'; trailing_underscore -> HTTP 200; leading_digit -> HTTP 200; double_underscore_inside -> HTTP 200;
  only_underscores -> BadRequestException (HTTP 400) 'The specified table name is not valid.'; aws_prefix ->
  HTTP 200; aws_literal -> HTTP 200; aws_s3_metadata_prefix -> HTTP 200. Server messages seen: ['The specified
  table name is not valid.']. Raw (non-boto3) request with name='' -> HTTP 400
  BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/ "{\"message\":\"A table name
  is not specified\"}"; without the name member -> HTTP 400 "{\"message\":\"A table name is not specified\"}".
  After the uppercase attempt, GetTable(lower-cased name) -> NotFoundException (HTTP 404) 'The specified table
  does not exist.' and GetTable(uppercase name) -> NotFoundException (HTTP 404) 'The specified table does not
  exist.'; GetTable(hyphen name) -> NotFoundException (HTTP 404) 'The specified table does not exist.'.
  - ACK: terminal_codes · ops: CreateTable, GetTable · fields: name
  - repro: CreateTable with each name cell (255/256 chars, non-ASCII, hyphen, dot, space, uppercase,
    leading/trailing underscore, leading digit, '__', 'aws_*', 'aws')
  - handling: handled via `test/e2e/tests/test_namespace.py:55-57; test/e2e/tests/test_table.py:63-64; generator.yaml:99-104; generator.yaml:170-175; generator.yaml:45-46; generator.yaml:60-65`
  - related: [S3T-TABLE-047](../table.md#s3t-table-047), [S3T-TABLE-079](../table.md#s3t-table-079), [S3T-TABLE-035](../table.md#s3t-table-035), [S3T-TABLE-080](../table.md#s3t-table-080), [S3T-TABLE-052](../table.md#s3t-table-052), [S3T-TABLE-027](../table.md#s3t-table-027) ·
    hypotheses: H-T-013, H-T-014, H-T-015, H-T-016, H-T-017 · evidence: table/error-taxonomy/create-validation

## Notes

H-T-013 confirmed (255 -> HTTP 200; 256 -> BadRequestException (HTTP 400) 'The specified table name is not
valid.'; empty -> client-side ParamValidationError in boto3, server 400; non-ASCII -> BadRequestException
(HTTP 400) 'The specified table name is not valid.'). H-T-014 confirmed (hyphen -> BadRequestException (HTTP 400)
'The specified table name is not valid.'; dot -> BadRequestException (HTTP 400) 'The specified table name is
not valid.'). H-T-015 confirmed (uppercase -> BadRequestException (HTTP 400) 'The specified table name is not
valid.'; no lower-cased table exists afterwards: True). H-T-016 confirmed (leading underscore ->
BadRequestException (HTTP 400) 'The specified table name is not valid.'; trailing underscore -> HTTP 200;
leading digit -> HTTP 200; '__' -> BadRequestException (HTTP 400) 'The specified table name is not valid.').
H-T-017 confirmed (aws_ prefix -> HTTP 200 type=customer; 'aws' -> HTTP 200 type=customer).

Full original title: CreateTable name: 255 chars, trailing '_', leading digit, 'aws', 'aws_*' accepted; 256
chars, '_lead', '__', upper, '-', '.', space, non-ASCII -> 400
