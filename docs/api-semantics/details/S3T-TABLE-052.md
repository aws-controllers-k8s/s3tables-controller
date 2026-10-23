<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-052: RenameTable.newName: 'aws_' prefix, 255 chars accepted; upper, '-', leading '_', 256 chars, '.', non-ASCII, '' -> 400 'name is not valid'
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-052"></a>**S3T-TABLE-052** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **RenameTable.newName: 'aws_' prefix, 255 chars accepted; upper, '-', leading '_', 256 chars, '.', non-ASCII, '' -> 400 'name is not valid'**
  Per newName candidate (SDK pattern/max not enforced client-side, so the server answered): uppercase ->
  BadRequestException (HTTP 400) 'The specified table name is not valid.'; hyphen -> BadRequestException (HTTP 400)
  'The specified table name is not valid.'; leading_underscore -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; aws_prefix -> HTTP 204; len_256 -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; len_255 -> HTTP 204; dot -> BadRequestException (HTTP 400) 'The
  specified table name is not valid.'; unicode -> BadRequestException (HTTP 400) 'The specified table name is
  not valid.'. Empty newName via a raw signed request -> HTTP 400 BadRequestException {"message":"A table name
  is not specified"}. Final GetTable(ns1,a) -> HTTP 200 (name restored: True).
  - ACK: terminal_codes, custom_update · ops: RenameTable · fields: newName
  - repro: RenameTable(ns1,a,newName=<candidate>) for uppercase, hyphen, leading underscore, aws_ prefix, 256
    chars, 255 chars, dot, unicode; raw PUT .../rename {newName:''}
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-078](../table.md#s3t-table-078), [S3T-TABLE-027](../table.md#s3t-table-027) · hypotheses: H-T-065 · evidence: table/error-taxonomy/rename-errors

## Notes

H-T-065 qualified: uppercase, hyphen, leading underscore, 256 chars, '.', non-ASCII and empty are rejected
with BadRequestException 400, but every rejection carries the same message 'The specified table name is not
valid.' (empty: 'A table name is not specified') - it does not name the rule; an 'aws_' prefix is ACCEPTED (no
reserved prefix for table names); a 255-char [a-z0-9_] name is accepted.

Full original title: RenameTable.newName server-side rules: accepted aws_prefix, len_255; rejected uppercase,
hyphen, leading_underscore, len_256, dot, unicode; empty newName -> HTTP 400
