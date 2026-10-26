<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-127: GetTableBucket ARN variants -> other-region or bogus region 400 region mismatch before lookup; same-region nonexistent 404; empty region 400
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-127"></a>**S3T-TABLEBUCKET-127** `error-code` · impact medium · handled · verified 2026-10-14
  **GetTableBucket ARN variants -> other-region or bogus region 400 region mismatch before lookup; same-region nonexistent 404; empty region 400**
  On the us-west-2 endpoint: ARN with region us-east-1 and a bucket name that exists nowhere ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.'; same-region
  nonexistent name -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; ARN region
  'xx-fake-9' -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.'; ARN with an empty region segment -> BadRequestException (HTTP 400) 'The specified table bucket
  ARN is not valid.'. The region check runs before the existence lookup: a wrong-region ARN never yields
  NotFoundException.
  - ACK: terminal_codes, exceptions.404 · ops: GetTableBucket · fields: tableBucketARN
  - repro: GetTableBucket with the four ARN variants on the us-west-2 endpoint
  - handling: handled via `generator.yaml:191-192; pkg/resource/table_bucket/sdk.go:133-155; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-014](../tablebucket.md#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015), [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115),
    [S3T-TABLEBUCKET-016](../tablebucket.md#s3t-tablebucket-016), [S3T-TABLEBUCKET-159](../tablebucket.md#s3t-tablebucket-159), [S3T-TABLEBUCKET-022](../tablebucket.md#s3t-tablebucket-022), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-TABLEBUCKET-033](../tablebucket.md#s3t-tablebucket-033),
    [S3T-TABLEBUCKET-003](../tablebucket.md#s3t-tablebucket-003), [S3T-TABLEBUCKET-125](../tablebucket.md#s3t-tablebucket-125), [S3T-TABLEBUCKET-162](../tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](../tablebucket-policy.md#s3t-tablebucket-163), [S3T-TABLEBUCKET-164](../tablebucket.md#s3t-tablebucket-164),
    [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029), [S3T-SERVICE-011](../service.md#s3t-service-011) · hypotheses: H-B-062,
    H-S-070 · evidence: tablebucket/cross-region/same-name-two-regions-arn-routing

## Notes

Supports H-B-062/H-S-070: the 400 is a pure ARN-region-vs-endpoint check, independent of whether the bucket
exists in the other region.

Full original title: GetTableBucket ARN variants: other-region nonexistent -> BadRequestException; same-region
nonexistent -> NotFoundException; bogus region -> BadRequestException; empty region -> BadRequestException
