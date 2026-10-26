<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-025: ListTables(namespace=<missing>) -> NotFoundException; ListTables(namespace=<existing, empty>) -> HTTP 200 {"tables": []}...
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-025"></a>**S3T-NAMESPACE-025** `error-code` · impact medium · handled · verified 2026-10-13
  **ListTables(namespace=<missing>) -> NotFoundException; ListTables(namespace=<existing, empty>) -> HTTP 200 {"tables": []}...**
  Bucket with tables ns1/t1 and ns_b/t2 plus an empty namespace: ListTables(namespace=ns1) -> HTTP 200 with 1
  table(s) [('ackq_54dbb7_with_tbl', 't1')] (summary members ['createdAt', 'modifiedAt', 'name', 'namespace',
  'tableARN', 'type']; namespace is a one-element list). ListTables() -> 2 tables across both namespaces.
  ListTables(namespace=<does_not_exist>) -> NotFoundException (HTTP 404) 'The specified namespace does not
  exist.'. ListTables(namespace=<empty namespace>) -> HTTP 200 with raw body '{"tables":[]}' (member present,
  empty). ListTables(prefix=<lab>_t) -> HTTP 200 2 tables (prefix matches the table name across namespaces).
  - ACK: custom_find, exceptions.404, pre-delete-cleanup · ops: ListTables · fields: namespace, tables, prefix
  - repro: CreateNamespace x3, CreateTable in two of them -> ListTables(namespace=...) for an existing, an
    empty and a missing namespace; ListTables() ; ListTables(prefix)
  - handling: handled via `generator.yaml:4-9; generator.yaml:123-131`
  - related: [S3T-NAMESPACE-002](../namespace.md#s3t-namespace-002), [S3T-NAMESPACE-041](../namespace.md#s3t-namespace-041), [S3T-NAMESPACE-005](../namespace.md#s3t-namespace-005), [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003), [S3T-NAMESPACE-004](../namespace.md#s3t-namespace-004),
    [S3T-NAMESPACE-001](../namespace.md#s3t-namespace-001), [S3T-NAMESPACE-026](../namespace.md#s3t-namespace-026), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029) · hypotheses: H-N-035, H-N-036
    · evidence: namespace/dependencies/delete-with-tables

## Notes

H-N-035 confirmed; H-N-036 (contrarian: 200 []) refuted. An 'is the namespace empty' precheck gets a 404 for a
missing namespace and 200 [] for an empty one.

Full original title: ListTables(namespace=<missing>) -> NotFoundException; ListTables(namespace=<existing,
empty>) -> HTTP 200 {"tables": []}; filter is strict, summaries echo namespace as a list
