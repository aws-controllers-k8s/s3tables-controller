<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-168: ListTables maxTables=0 or -1 -> BadRequestException 400 'limit should be a positive number'; 1000 and 1001 -> 200; maxTables=1 walk: 3 pages
_Full entry and notes of one finding; its summary entry is in [table.md](../table.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-table-168"></a>**S3T-TABLE-168** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTables maxTables=0 or -1 -> BadRequestException 400 'limit should be a positive number'; 1000 and 1001 -> 200; maxTables=1 walk: 3 pages**
  Namespace holding exactly 3 tables, client validation off: maxTables=0 -> BadRequestException (HTTP 400)
  'The specified limit is not valid. A limit should be a positive number'; maxTables=1000 -> HTTP 200, 3
  items, token absent; maxTables=1001 -> HTTP 200, 3 items, token absent; maxTables=-1 -> BadRequestException
  (HTTP 400) 'The specified limit is not valid. A limit should be a positive number'. maxTables=1 + token
  following: 3 pages, items ['ackq_f745ff_t1', 'ackq_f745ff_t2', 'ackq_f745ff_t3'] (0 duplicates),
  continuationToken lengths [640, 640, 0], final page omits continuationToken.
  - ACK: custom_find, pre-delete-cleanup · ops: ListTables · fields: maxTables, continuationToken, namespace
  - repro: 3 x CreateTable in one namespace; ListTables(maxTables=0|1000|1001|-1); ListTables(maxTables=1)
    following tokens
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-NAMESPACE-038](../namespace.md#s3t-namespace-038), [S3T-TABLE-293](../table.md#s3t-table-293), [S3T-SERVICE-009](../service.md#s3t-service-009), [S3T-NAMESPACE-032](../namespace.md#s3t-namespace-032), [S3T-TABLE-092](../table.md#s3t-table-092),
    [S3T-TABLEBUCKET-069](../tablebucket.md#s3t-tablebucket-069), [S3T-TABLE-091](../table.md#s3t-table-091), [S3T-TABLE-305](../table.md#s3t-table-305), [S3T-TABLE-095](../table.md#s3t-table-095), [S3T-TABLE-217](../table.md#s3t-table-217) · hypotheses: H-N-032,
    H-T-070 · evidence: namespace/limits/pagination-and-bursts

## Notes

H-N-032 confirmed for pagination (1 item per page, 3 pages, no duplicates, 640-char tokens, token absent on
the last page) and for maxTables=0 -> BadRequestException 400; H-T-070 qualified: maxTables=1000 -> 200 and -1
-> 400 as predicted, but maxTables=1001 -> HTTP 200 (accepted, not a 400), and the 400 message ('The specified
limit is not valid. A limit should be a positive number') does not state the 1-1000 range.
