<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-038: v1 SchemaField.type: 14 Iceberg primitives + 'INT'/'String' accepted (lower-cased; decimal(10,2) re-spaced); integer/varchar/bigint/'' 400
_Full entry and notes of one finding; its summary entry is in [table-schema.md](../table-schema.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-table-038"></a>**S3T-TABLE-038** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **v1 SchemaField.type: 14 Iceberg primitives + 'INT'/'String' accepted (lower-cased; decimal(10,2) re-spaced); integer/varchar/bigint/'' 400**
  CreateTable with a one-field v1 schema per type string. Accepted (and written to metadata.json as): {"int":
  [["c1", "int"]], "long": [["c1", "long"]], "string": [["c1", "string"]], "boolean": [["c1", "boolean"]],
  "timestamp": [["c1", "timestamp"]], "timestamptz": [["c1", "timestamptz"]], "decimal(10,2)": [["c1",
  "decimal(10, 2)"]], "fixed[16]": [["c1", "fixed[16]"]], "date": [["c1", "date"]], "time": [["c1", "time"]],
  "uuid": [["c1", "uuid"]], "binary": [["c1", "binary"]], "float": [["c1", "float"]], "double": [["c1",
  "double"]], "INT": [["c1", "int"]], "String": [["c1", "string"]]}. Rejected: {"integer":
  "BadRequestException 400 'The specified metadata is not valid'", "varchar": "BadRequestException 400 'The
  specified metadata is not valid'", "": "BadRequestException 400 'The specified metadata is not valid'", "
  int ": "BadRequestException 400 'The specified metadata is not valid'", "decimal": "BadRequestException 400
  'The specified metadata is not valid'", "fixed": "BadRequestException 400 'The specified metadata is not
  valid'", "bigint": "BadRequestException 400 'The specified metadata is not valid'"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: metadata.iceberg.schema.fields.type
  - repro: Loop CreateTable over the type strings above (one field each); GetObject metadata.json of accepted
    ones; DeleteTable
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-037](../table-schema.md#s3t-table-037), [S3T-TABLE-039](../table-schema.md#s3t-table-039), [S3T-TABLE-040](../table-schema.md#s3t-table-040), [S3T-TABLE-314](../table-schema.md#s3t-table-314), [S3T-TABLE-313](../table-schema.md#s3t-table-313), [S3T-TABLE-073](../table-schema.md#s3t-table-073),
    [S3T-TABLE-076](../table-schema.md#s3t-table-076), [S3T-TABLE-042](../table-schema.md#s3t-table-042) · hypotheses: H-T-109 · evidence: table/weird-inputs/schema-variants

## Notes

H-T-109 QUALIFIED: all 14 Iceberg primitives accepted; 'integer','varchar','bigint','' and ' int ' (whitespace
not trimmed), bare 'decimal'/'fixed' rejected with BadRequestException 400; but upper/mixed case 'INT' and
'String' are ACCEPTED and written lower-cased ('int','string'), and 'decimal(10,2)' is written as 'decimal(10,
2)' (space inserted). The 400 message is the generic 'The specified metadata is not valid' and does NOT name
the offending type. One attempt per type.

Full original title: v1 SchemaField.type: accepted ['int', 'long', 'string', 'boolean', 'timestamp',
'timestamptz', 'decimal(10,2)', 'fixed[16]', 'date', 'time', 'uuid', 'binary', 'float', 'double', 'INT',
'String']; rejected ['integer', 'varchar', '', ' int ', 'decimal', 'fixed', 'bigint'] (['BadRequestException
400'])
