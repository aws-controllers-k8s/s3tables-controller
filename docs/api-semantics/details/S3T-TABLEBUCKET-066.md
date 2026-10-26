<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-066: tableBucketARN forms: bare name -> BadRequestException 400, nonexistent bucket -> NotFoundException 404...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-066"></a>**S3T-TABLEBUCKET-066** `identity` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **tableBucketARN forms: bare name -> BadRequestException 400, nonexistent bucket -> NotFoundException 404...**
  With a policy set on the bucket, 7 identifier forms were sent as tableBucketARN to three Get operations:
  bare_name -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; nonexistent_bucket -> GetTableBucket NotFoundException 404
  / GetTableBucketPolicy NotFoundException 404 / GetTableBucketEncryption NotFoundException 404;
  uppercase_name -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; trailing_slash -> GetTableBucket BadRequestException 400 /
  GetTableBucketPolicy BadRequestException 400 / GetTableBucketEncryption BadRequestException 400;
  uppercase_prefix -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400; table_style_suffix -> GetTableBucket BadRequestException
  400 / GetTableBucketPolicy BadRequestException 400 / GetTableBucketEncryption BadRequestException 400;
  wrong_region -> GetTableBucket BadRequestException 400 / GetTableBucketPolicy BadRequestException 400 /
  GetTableBucketEncryption BadRequestException 400. Messages: bare name 'The specified table bucket ARN is not
  valid.'; upper-case name 'The specified bucket name isn't valid. Specify a different bucket name, and then
  try again.'; trailing slash 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; wrong region 'The provided Region does not match the Region of the resource.'.
  - ACK: is_arn_primary_key, exceptions.404, terminal_codes · ops: GetTableBucket, GetTableBucketPolicy,
    GetTableBucketEncryption · fields: tableBucketARN
  - repro: Send each identifier form as tableBucketARN to GetTableBucket, GetTableBucketPolicy,
    GetTableBucketEncryption; record code/HTTP/message
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-061](../tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-178](../tablebucket-policy.md#s3t-tablebucket-178), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065),
    [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLE-081](../table.md#s3t-table-081), [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-162](../tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](../tablebucket-policy.md#s3t-tablebucket-163),
    [S3T-TABLEBUCKET-164](../tablebucket.md#s3t-tablebucket-164), [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029) · hypotheses: H-S-069 · evidence:
    tablebucket/sub-resources/policy-get-put-delete

## Notes

H-S-069 confirmed: bare name BadRequestException 400, nonexistent NotFoundException 404, upper-case name
BadRequestException 400, trailing slash BadRequestException 400, upper-case prefix BadRequestException 400
(expected 400/404/400/400/400).

Full original title: tableBucketARN forms: bare name -> BadRequestException 400, nonexistent bucket ->
NotFoundException 404, upper-case name -> BadRequestException 400, trailing '/' -> BadRequestException 400,
'ARN:AWS:' prefix -> BadRequestException 400
