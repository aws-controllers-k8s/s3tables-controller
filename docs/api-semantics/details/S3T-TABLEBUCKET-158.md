<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-158: Delete->404->CreateTableBucket every 15 s: name blocked ~609 s whether the bucket was empty, populated or not retried until +420 s
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-158"></a>**S3T-TABLEBUCKET-158** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Delete->404->CreateTableBucket every 15 s: name blocked ~609 s whether the bucket was empty, populated or not retried until +420 s**
  Controller loop (DeleteTableBucket 204, GetTableBucket 404 at once, CreateTableBucket(same name) every 15
  s): first 200 at +609.4 s for a bucket created and deleted within one second, +608.2 s for a bucket that had
  held a namespace, 2 tables, tags, policy, maintenance, storage-class and encryption settings (children
  deleted just before), and +608.8 s for a name whose first retry was only sent at +420 s (so 409 retries do
  not extend the window and content does not lengthen it). Every rejected retry was ConflictException HTTP 409
  with the single message ["The bucket is in a transitional state because of a previous deletion attempt. Try
  again later."] (n_409 per name: {"empty": 40, "full": 40, "late": 12}). After the window the ARN is
  identical and tableBucketId is new (changed); the re-created 'full' name reads fresh defaults (policy
  NotFoundException, maintenance {"icebergUnreferencedFileRemoval": {"status": "enabled", "settings":
  {"icebergUnreferencedFileRemoval": {"unreferencedDays": 3, "nonCurrentDays": 10}}}}, tags {"ackq-run":
  "5d7947"}).
  - ACK: requeue, terminal_codes, custom_create · ops: DeleteTableBucket, CreateTableBucket, GetTableBucket ·
    fields: name, tableBucketId
  - repro: CreateTableBucket x3 -> (populate one) -> DeleteTableBucket x3 -> loop: CreateTableBucket every 15
    s until 200 (one name waits 420 s first)
  - measurements: recreate_ok_at_s.empty=609.4, recreate_ok_at_s.full=608.2,
    recreate_ok_at_s.late_first_retry_420s=608.8, retry_interval_s=15, n_409_per_name.empty=40,
    n_409_per_name.full=40, n_409_per_name.late=12, n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-047](../tablebucket.md#s3t-tablebucket-047), [S3T-TABLEBUCKET-115](../tablebucket.md#s3t-tablebucket-115), [S3T-TABLEBUCKET-116](../tablebucket-replication.md#s3t-tablebucket-116), [S3T-TABLEBUCKET-001](../tablebucket-encryption-storage.md#s3t-tablebucket-001),
    [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](../tablebucket-maintenance-metrics.md#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](../tablebucket-maintenance-metrics.md#s3t-tablebucket-100), [S3T-TABLEBUCKET-020](../tablebucket-encryption-storage.md#s3t-tablebucket-020),
    [S3T-TABLEBUCKET-079](../tablebucket-maintenance-metrics.md#s3t-tablebucket-079) · evidence: tablebucket/creative/delete-recreate-window

## Notes

extends [S3T-TABLEBUCKET-047](../tablebucket.md#s3t-tablebucket-047)/-115 with the controller-shaped loop and the three window variants (content,
no-retry, retry); extends -116 with the fresh sub-resource reads after re-creation. Decision rule: a create
that returns 409 while GetTableBucket is 404 is 'wait', not 'exists' and not terminal; requeue with a >= 15 s
backoff for up to ~11 min; retrying costs nothing. Measured once per variant (n=3 names); the ~10 min length
agrees with the two earlier single measurements.
