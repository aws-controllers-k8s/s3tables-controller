<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-069: ListTableBuckets maxBuckets=0/-1 -> BadRequestException 400, maxBuckets=1001 -> 200 (not rejected); last page omits continuationToken
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-069"></a>**S3T-TABLEBUCKET-069** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTableBuckets maxBuckets=0/-1 -> BadRequestException 400, maxBuckets=1001 -> 200 (not rejected); last page omits continuationToken**
  ListTableBuckets(maxBuckets=0) -> BadRequestException (HTTP 400) 'The specified limit is not valid. A limit
  should be a positive number'; (maxBuckets=1001) -> HTTP 200, 11 items; (maxBuckets=-1) ->
  BadRequestException (HTTP 400) 'The specified limit is not valid. A limit should be a positive number';
  (maxBuckets=1000) -> HTTP 200, 11 items. With 9 matching buckets, maxBuckets=1 + prefix returned 1 item per
  page with continuationToken present on 8 of 9 pages (token length [608]), the final page OMITS
  continuationToken entirely (key absent, not an empty string), 0 duplicate(s) across pages. The default page
  (maxBuckets omitted) returned 11 buckets with no continuationToken.
  - ACK: custom_find, list_operation.match_fields · ops: ListTableBuckets · fields: maxBuckets,
    continuationToken, prefix
  - repro: ListTableBuckets(maxBuckets=0|1001|-1) with client validation off; ListTableBuckets(prefix,
    maxBuckets=1) following tokens
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-040](../tablebucket.md#s3t-tablebucket-040), [S3T-TABLEBUCKET-041](../tablebucket.md#s3t-tablebucket-041), [S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130), [S3T-TABLEBUCKET-125](../tablebucket.md#s3t-tablebucket-125),
    [S3T-NAMESPACE-038](../namespace.md#s3t-namespace-038), [S3T-TABLE-293](../table.md#s3t-table-293), [S3T-SERVICE-009](../service.md#s3t-service-009), [S3T-NAMESPACE-032](../namespace.md#s3t-namespace-032), [S3T-TABLE-168](../table.md#s3t-table-168), [S3T-TABLE-092](../table.md#s3t-table-092),
    [S3T-TABLE-091](../table.md#s3t-table-091), [S3T-TABLEBUCKET-168](../tablebucket.md#s3t-tablebucket-168), [S3T-TABLEBUCKET-169](../tablebucket.md#s3t-tablebucket-169) · hypotheses: H-B-050 · evidence:
    tablebucket/limits/bucket-quota-create-burst

## Notes

H-B-050 qualified: 0 rejected as predicted (400 'The specified limit is not valid. A limit should be a
positive number'), token present/absent semantics as predicted (absent on the last page, never an empty
string), BUT maxBuckets=1001 was accepted with HTTP 200 (not a 400) - the documented max of 1000 is not
enforced server-side at 1001 with 11 buckets in the account. Default page: all 11 buckets in one page, no
token.
