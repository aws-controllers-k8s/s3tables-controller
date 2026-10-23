<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-094: ListTables(namespace = missing | prefix of existing | just-deleted | upper-cased) -> NotFoundException 404; existing empty namespace -> 200
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-094"></a>**S3T-TABLE-094** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables(namespace = missing | prefix of existing | just-deleted | upper-cased) -> NotFoundException 404; existing empty namespace -> 200**
  ListTables(namespace=<never existed>) -> NotFoundException (HTTP 404) 'The specified namespace does not
  exist.'. ListTables(namespace=<ns1 minus its last char>) -> NotFoundException (HTTP 404) 'The specified
  namespace does not exist.'. ListTables(namespace=<existing empty ns3>) -> HTTP 200 with 0 tables. After
  DeleteNamespace(ns3) (-> HTTP 204): ListTables(namespace=ns3) -> NotFoundException (HTTP 404) 'The specified
  namespace does not exist.'. Upper-cased namespace -> NotFoundException (HTTP 404) 'The specified namespace
  does not exist.' (no case folding).
  - ACK: exceptions.404, pre-delete-cleanup · ops: ListTables, DeleteNamespace · fields: namespace
  - repro: ListTables(namespace=nope); ListTables(namespace=<prefix of ns1>); CreateNamespace ns3;
    ListTables(namespace=ns3); DeleteNamespace ns3; ListTables(namespace=ns3)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-080](../table.md#s3t-table-080), [S3T-TABLE-049](../table.md#s3t-table-049), [S3T-TABLE-053](../table.md#s3t-table-053), [S3T-TABLE-082](../table.md#s3t-table-082), [S3T-TABLE-093](../table.md#s3t-table-093), [S3T-TABLE-097](../table.md#s3t-table-097) ·
    hypotheses: H-T-073 · evidence: table/weird-inputs/list-tables-bounds

## Notes

H-T-073 confirmed (predicted NotFoundException 404 for missing and just-deleted namespaces and exact-match
semantics so a prefix also yields 404). Message: 'The specified namespace does not exist.'.

Full original title: ListTables(namespace=<missing>) -> NotFoundException; prefix of an existing namespace ->
NotFoundException; just-deleted namespace -> NotFoundException; empty existing namespace -> HTTP 200
