<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-032: ListNamespaces maxNamespaces 0/-1 -> BadRequestException, 1001 -> 200 (not rejected); =1 pages 3 namespaces in 3 pages, last page no token
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-032"></a>**S3T-NAMESPACE-032** `request-validation` · impact medium · handled · verified 2026-10-13
  **ListNamespaces maxNamespaces 0/-1 -> BadRequestException, 1001 -> 200 (not rejected); =1 pages 3 namespaces in 3 pages, last page no token**
  Bucket with exactly 3 namespaces, client validation off: maxNamespaces=0 -> BadRequestException (HTTP 400)
  'The specified limit is not valid. A limit should be a positive number'; maxNamespaces=1 -> HTTP 200, 1
  items, token present; maxNamespaces=2 -> HTTP 200, 2 items, token present; maxNamespaces=3 -> HTTP 200, 3
  items, token absent; maxNamespaces=1000 -> HTTP 200, 3 items, token absent; maxNamespaces=1001 -> HTTP 200,
  3 items, token absent; maxNamespaces=-1 -> BadRequestException (HTTP 400) 'The specified limit is not valid.
  A limit should be a positive number'. Following the token at maxNamespaces=1 took 3 pages, returned
  ['ackq_f745ff_nsa', 'ackq_f745ff_nsb', 'ackq_f745ff_nsc'] (0 duplicates); the final page omits
  continuationToken.
  - ACK: custom_find · ops: ListNamespaces · fields: maxNamespaces, continuationToken
  - repro: CreateTableBucket; 3 x CreateNamespace; ListNamespaces(maxNamespaces=0|1|2|3|1000|1001|-1); follow
    tokens at 1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-038](../namespace.md#s3t-namespace-038), [S3T-TABLE-293](../table.md#s3t-table-293), [S3T-SERVICE-009](../service.md#s3t-service-009), [S3T-TABLE-168](../table.md#s3t-table-168), [S3T-TABLE-092](../table.md#s3t-table-092),
    [S3T-TABLEBUCKET-069](../tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](../table.md#s3t-table-091), [S3T-NAMESPACE-008](../namespace.md#s3t-namespace-008), [S3T-NAMESPACE-009](../namespace.md#s3t-namespace-009), [S3T-NAMESPACE-010](../namespace.md#s3t-namespace-010),
    [S3T-NAMESPACE-005](../namespace.md#s3t-namespace-005) · hypotheses: H-N-031 · evidence: namespace/limits/pagination-and-bursts

## Notes

H-N-031 qualified: maxNamespaces=0 and -1 -> BadRequestException 400 'The specified limit is not valid. A
limit should be a positive number' as predicted, =1 returns one entry plus a token, the last page omits the
token (key absent), BUT maxNamespaces=1001 is accepted with HTTP 200 (not rejected, so either clamped or
unenforced above 1000). The 'at most 1000 per page when omitted' clause needs > 1000 namespaces and was not
tested.
