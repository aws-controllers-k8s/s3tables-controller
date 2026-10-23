<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-082: 10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (no 409), last writer wins; 10 maintenance Puts 204 but ~6.5 s
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-082"></a>**S3T-TABLEBUCKET-082** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (no 409), last writer wins; 10 maintenance Puts 204 but ~6.5 s**
  Fired concurrently on one table bucket: PutTableBucketPolicy x10 -> {"200": 7,
  "InternalServerErrorException/500": 3} (wall 6.29s); PutTableBucketMaintenanceConfiguration x10 -> {"204":
  10} (wall 6.46s); Put/DeleteTableBucketEncryption 5+5 -> {"200": 5, "NotFoundException/404": 2, "204": 3}
  (wall 0.18s); PutTableBucketStorageClass(STANDARD) x20 -> {"200": 20} (wall 0.24s). Error messages:
  {"PutTableBucketPolicy x10": {"InternalServerErrorException": "An internal error occurred. Try again."},
  "Put/DeleteTableBucketEncryption 5+5": {"NotFoundException": "The specified encryption configuration does
  not exist."}}. After the policy race GetTableBucketPolicy showed Sid P07 (which is one of the 7 successful
  puts); rejected puts retried sequentially -> [('P03', 'InternalServerErrorException', 'HTTP 200'), ('P06',
  'InternalServerErrorException', 'HTTP 200'), ('P09', 'InternalServerErrorException', 'HTTP 200')]. After the
  maintenance race unreferencedDays=6 (which is one of the 10 successful puts). After the encryption
  put/delete race GetTableBucketEncryption -> HTTP 200 {"sseAlgorithm": "AES256"}.
  - ACK: requeue, one-per-reconcile · ops: PutTableBucketPolicy, GetTableBucketPolicy,
    PutTableBucketMaintenanceConfiguration, PutTableBucketEncryption, DeleteTableBucketEncryption,
    PutTableBucketStorageClass · fields: resourcePolicy, configuration, encryptionConfiguration,
    storageClassConfiguration
  - repro: 10 threads x PutTableBucketPolicy(distinct Sid); 10 x
    PutTableBucketMaintenanceConfiguration(unreferencedDays 1..10); 5 Put + 5 Delete TableBucketEncryption; 20
    x PutTableBucketStorageClass(STANDARD); Get* after each
  - measurements: wall_s.policy=6.29, wall_s.maintenance=6.46, wall_s.encryption=0.18,
    wall_s.storageclass=0.24, n=50, latency_ms_min_max.policy=[118, 6281],
    latency_ms_min_max.maintenance=[134, 6452], latency_ms_min_max.encryption=[10, 178],
    latency_ms_min_max.storageclass=[16, 226]
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-078](../tablebucket-maintenance-metrics.md#s3t-tablebucket-078), [S3T-TABLEBUCKET-093](../tablebucket-maintenance-metrics.md#s3t-tablebucket-093) · hypotheses: H-S-074, H-S-073 · evidence:
    tablebucket/limits/read-write-bursts-policy-size

## Notes

H-S-074 refuted on the code: concurrent PutTableBucketPolicy produced InternalServerErrorException 500 'An
internal error occurred. Try again.' (3/10), never ConflictException; the 7 successful puts and the stored
policy (Sid P07, one of the successes) are consistent with last-writer-wins and the 3 rejected puts succeeded
on sequential retry. 10 concurrent PutTableBucketMaintenanceConfiguration all returned 204 but the burst took
6.46 s wall (individual latencies 134-6452 ms) - the service appears to serialize writes to one bucket's
maintenance/policy sub-resources (policy burst wall 6.29 s) while 20 concurrent PutTableBucketStorageClass
completed in 0.24 s. Interleaved Put/DeleteTableBucketEncryption: 5 puts 200, 3 deletes 204, 2 deletes
NotFoundException 404 'The specified encryption configuration does not exist.' (a delete that lands when no
configuration exists is 404, not 204). H-S-073's PutTableBucketStorageClass cell at 20 concurrent: all 200.
Single run per race; 2 other probes running.

Full original title: 10 concurrent PutTableBucketPolicy -> 3/10 InternalServerErrorException 500 (no 409),
last writer wins; maintenance x10 all 204 but serialized to 6.5 s
