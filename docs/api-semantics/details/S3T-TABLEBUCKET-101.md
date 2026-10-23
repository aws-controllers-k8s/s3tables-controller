<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-101: DeleteTableBucket -> HTTP 204; GetTableBucket polled at 200 ms: NotFoundException/404 from the first read (0.009 s)...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-replication.md](../tablebucket-replication.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-101"></a>**S3T-TABLEBUCKET-101** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucket -> HTTP 204; GetTableBucket polled at 200 ms: NotFoundException/404 from the first read (0.009 s)...**
  After deleting the table and namespace, DeleteTableBucket -> HTTP 204. GetTableBucket polled every 200 ms:
  codes in order ['NotFoundException/404', 'NotFoundException/404', 'NotFoundException/404',
  'NotFoundException/404', 'NotFoundException/404'] (first poll at +0.009 s, message 'The specified bucket
  does not exist.'); no ConflictException and no DELETING-style intermediate. Then on the deleted ARN:
  encryption -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; policy ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'; metrics -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'; replication -> NotFoundException (HTTP 404) 'The specified bucket
  does not exist.'; maintenance -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  storage_class -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; PutTableBucketPolicy
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'; PutTableBucketMetricsConfiguration
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: exceptions.404, deletable.when · ops: DeleteTableBucket, GetTableBucket, PutTableBucketPolicy,
    PutTableBucketMetricsConfiguration, GetTableBucketEncryption, GetTableBucketPolicy,
    GetTableBucketMetricsConfiguration, GetTableBucketReplication, GetTableBucketMaintenanceConfiguration,
    GetTableBucketStorageClass · fields: tableBucketARN
  - repro: DeleteTable -> DeleteNamespace -> DeleteTableBucket -> GetTableBucket every 200 ms -> six
    sub-resource Gets + Put policy + Put metrics on the ARN
  - measurements: first_404_after_delete_s=0.009, polls=5, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-161](../tablebucket.md#s3t-tablebucket-161), [S3T-TABLEBUCKET-162](../tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-014](../tablebucket.md#s3t-tablebucket-014), [S3T-TABLEBUCKET-015](../tablebucket.md#s3t-tablebucket-015),
    [S3T-TABLEBUCKET-097](../tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116), [S3T-TABLEBUCKET-092](../tablebucket-replication.md#s3t-tablebucket-092) · hypotheses: H-S-071, H-S-098, H-S-002 ·
    evidence: tablebucket/error-taxonomy/subresource-arn-variants

## Notes

H-S-071 (no DELETING state, 404 on first read, sub-resource Put on the deleted ARN -> 404 not 409) confirmed.
H-S-098 cell (b) (deleted bucket) and H-S-002: the deleted-bucket messages are ['The specified bucket does not
exist.'] - identical to the never-existing-bucket message(s).

Full original title: DeleteTableBucket -> HTTP 204; GetTableBucket polled at 200 ms: NotFoundException/404
from the first read (0.009 s); sub-resource Put/Get on the ARN -> ['NotFoundException/404']
