<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-095: ListTables continuationToken survives DeleteTable, is multi-use (maxTables may change); tampered, foreign-bucket or filter-changed -> 400
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-095"></a>**S3T-TABLE-095** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables continuationToken survives DeleteTable, is multi-use (maxTables may change); tampered, foreign-bucket or filter-changed -> 400**
  ListTables(maxTables=2) -> page ['ns2/a', 'ns1/a'] + token (720 chars). DeleteTable(ns1/abc) (3rd in order)
  -> HTTP 204. Token reused -> ['ns1/abd', 'ns2/abz']; reused again -> ['ns1/abd', 'ns2/abz']; reused with
  maxTables=5 -> ['ns1/abd', 'ns2/abz', 'ns2/b', 'ns1/xyz', 'ns1/z']; reused with an added namespace filter ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'. Token+'x' ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'; 'not-a-token' ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'; token from another bucket ->
  BadRequestException (HTTP 400) 'The continuation token is no longer valid.'.
  - ACK: custom_find, requeue · ops: ListTables, DeleteTable · fields: continuationToken, maxTables
  - repro: ListTables(maxTables=2) -> token; DeleteTable(3rd table); ListTables(token) x2;
    ListTables(token+'x'); ListTables(token from bucket2)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-091](../table.md#s3t-table-091), [S3T-TABLE-305](../table.md#s3t-table-305), [S3T-TABLE-168](../table.md#s3t-table-168), [S3T-TABLE-217](../table.md#s3t-table-217) · hypotheses: H-T-074 · evidence:
    table/weird-inputs/list-tables-bounds

## Notes

H-T-074 confirmed: the token stayed valid after a DeleteTable of the next table in order (deleted table
absent, no duplicates with page 1), was reused twice with identical results and with a different maxTables;
token+'x', 'not-a-token' and a token from another bucket all -> BadRequestException 400 'The continuation
token is no longer valid.' (one message for every token problem). Extra: reusing the token with a namespace
filter that was not on the original request also -> the same 400.

Full original title: ListTables continuationToken survives DeleteTable and is multi-use; tampered,
foreign-bucket or filter-changed token -> BadRequestException 400
