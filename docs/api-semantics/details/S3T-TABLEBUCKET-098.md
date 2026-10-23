<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-098: Sub-resource Deletes on a never-existing bucket -> ['NotFoundException/404']...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-098"></a>**S3T-TABLEBUCKET-098** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Sub-resource Deletes on a never-existing bucket -> ['NotFoundException/404']...**
  Never-existing bucket ARN: policy -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  metrics -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; replication ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; encryption -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'. Fresh bucket, nothing configured: policy -> HTTP 204; metrics ->
  HTTP 204; replication -> NotFoundException (HTTP 404) 'The specified replication configuration does not
  exist.'; encryption -> NotFoundException (HTTP 404) 'The specified encryption configuration does not
  exist.'.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketPolicy, DeleteTableBucketMetricsConfiguration,
    DeleteTableBucketReplication, DeleteTableBucketEncryption · fields: tableBucketARN
  - repro: four Delete* sub-resource calls on arn:...:bucket/<never-existed-name>, then on a fresh bucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-021](../tablebucket-replication.md#s3t-tablebucket-021), [S3T-SERVICE-022](../service.md#s3t-service-022), [S3T-SERVICE-026](../service.md#s3t-service-026), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-096](../tablebucket-replication.md#s3t-tablebucket-096),
    [S3T-SERVICE-014](../service.md#s3t-service-014) · hypotheses: H-S-088 · evidence: tablebucket/error-taxonomy/subresource-arn-variants

## Notes

H-S-088 qualified: policy and metrics Deletes are idempotent 204 on an absent configuration, replication and
encryption Deletes are 404 with a noun-specific message; on a missing bucket all four are
['NotFoundException/404'] with ['The specified bucket does not exist.']. So 404 on Delete policy/metrics does
mean 'bucket missing', on Delete replication/encryption it is ambiguous without the message.

Full original title: Sub-resource Deletes on a never-existing bucket -> ['NotFoundException/404']; on a fresh
bucket with nothing configured -> policy HTTP 204, metrics HTTP 204, replication NotFoundException (HTTP 404),
encryption NotFoundException (HTTP 404)
