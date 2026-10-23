<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-037: PutTableBucketEncryption switches AES256<->aws:kms both ways (visible 0.01 s); unchanged re-send 200; Delete on aws:kms -> 204, reads AES256
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-037"></a>**S3T-TABLEBUCKET-037** `sub-resource-api` · impact high · handled · verified 2026-10-13
  **PutTableBucketEncryption switches AES256<->aws:kms both ways (visible 0.01 s); unchanged re-send 200; Delete on aws:kms -> 204, reads AES256**
  Bucket created with AES256: Put AES256 (unchanged) -> HTTP 200; Put aws:kms -> HTTP 200; Put aws:kms
  (unchanged) -> HTTP 200; Put aws:kms without kmsKeyArn -> BadRequestException HTTP 400 "The specified KMS
  key ARN is invalid. Only key ARN format 'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is
  supported."; Put AES256 (from aws:kms) -> HTTP 200; Put aws:kms again -> HTTP 200;
  DeleteTableBucketEncryption on aws:kms bucket -> HTTP 204; Get after delete -> {"sseAlgorithm": "AES256"};
  DeleteTableBucketEncryption again -> NotFoundException HTTP 404 'The specified encryption configuration does
  not exist.'; Put AES256 after delete -> HTTP 200; DeleteTableBucketEncryption on explicit AES256 -> HTTP
  204; Put aws:kms with DISABLED key -> HTTP 200. GetTableBucketEncryption after Put aws:kms returned
  {"kmsKeyArn": "arn:aws:kms:us-west-2:<ACCOUNT>:key/12aa39a2-b0c3-475c-aa11-bb7fd4033450", "sseAlgorithm":
  "aws:kms"} (kmsKeyArn echoed byte-identical).
  - ACK: custom_update, late_initialize, exceptions.404 · ops: PutTableBucketEncryption,
    GetTableBucketEncryption, DeleteTableBucketEncryption · fields: encryptionConfiguration.sseAlgorithm,
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(AES256) -> Put AES256 -> Put aws:kms -> Get -> Put AES256 -> Put aws:kms ->
    DeleteTableBucketEncryption -> Get -> Delete again -> Put AES256 -> Delete
  - measurements: put_to_visible_s=[0.01, 0.01], n=2
  - handling: handled via `generator.yaml:194-195; generator.yaml:284-285; pkg/resource/table_bucket/hooks.go:69-88; pkg/resource/table_bucket/hooks.go:473-506; pkg/resource/table_bucket/hooks.go:484-486; pkg/resource/table_bucket/hooks.go:519-522; generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-005](../tablebucket-encryption-storage.md#s3t-tablebucket-005), [S3T-TABLEBUCKET-049](../tablebucket-encryption-storage.md#s3t-tablebucket-049), [S3T-TABLEBUCKET-051](../tablebucket-encryption-storage.md#s3t-tablebucket-051), [S3T-TABLEBUCKET-052](../tablebucket-encryption-storage.md#s3t-tablebucket-052),
    [S3T-TABLEBUCKET-072](../tablebucket-encryption-storage.md#s3t-tablebucket-072), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLE-249](../table-encryption-storage.md#s3t-table-249), [S3T-TABLEBUCKET-123](../tablebucket-encryption-storage.md#s3t-tablebucket-123), [S3T-TABLEBUCKET-073](../tablebucket-encryption-storage.md#s3t-tablebucket-073),
    [S3T-TABLEBUCKET-008](../tablebucket-encryption-storage.md#s3t-tablebucket-008), [S3T-TABLEBUCKET-009](../tablebucket-encryption-storage.md#s3t-tablebucket-009), [S3T-TABLE-061](../table-encryption-storage.md#s3t-table-061), [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035) ·
    hypotheses: H-B-071, H-B-007, H-B-014 · evidence: tablebucket/sub-resources/kms-encryption-forms

## Notes

H-B-071: confirmed for every cell listed. DeleteTableBucketEncryption is the only 'unset' path and the read
afterwards shows the AES256 default (never 404 for the setting itself), but a second Delete ->
NotFoundException HTTP 404 'The specified encryption configuration does not exist.'. H-B-014 (b) disabled key:
Put -> HTTP 200; Create -> HTTP 200. H-B-014 (b) REFUTED: a key disabled 5 s earlier was accepted by both
PutTableBucketEncryption and CreateTableBucket (HTTP 200); key state is not validated at bucket create/put.

Full original title: PutTableBucketEncryption switches AES256<->aws:kms in either direction (visible after
0.01/0.01 s); unchanged re-send HTTP 200; DeleteTableBucketEncryption on aws:kms -> HTTP 204 then reads AES256

Contradiction with [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107): Disabled KMS key at PutTableBucketEncryption(aws:kms): 037 got HTTP
200 for a key disabled ~5 s earlier (and CreateTableBucket with it -> 200; its notes conclude 'key state is
not validated at bucket create/put'); 107 got BadRequestException 400 'The specified KMS key ARN is disabled.'
for a Disabled key and the same 400 for a PendingDeletion key. Both n=1; 107 does not record the delay between
DisableKey and the Put. The exact server message in 107 shows key state is checked at least sometimes; KMS
key-state propagation lag at +5 s would reconcile the two but is untested, and 107 is a refuted-hypothesis
record so cannot be canonical Resolution: re-probe: weird-inputs - PutTableBucketEncryption(aws:kms) and
CreateTableBucket(aws:kms) at +5/+30/+120 s after DisableKey, n>=3 per offset; record 200 vs 400 'The
specified KMS key ARN is disabled.' and the Get afterwards; repeat once after ScheduleKeyDeletion

Contradiction with [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107): FU2-01 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and +120 s
after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every PutTableBucketEncryption(aws:kms K)
and every CreateTableBucket(aws:kms K) was rejected with BadRequestException 400 'The specified KMS key ARN is
disabled.' (15/15 Put cells, 15/15 Create cells; the enabled-key control right before DisableKey was 200 3/3),
and the Get a Resolution: keep all; [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule

Consolidation note: Follow-up re-probe outcome: the HTTP 200 on a disabled key (037) did not reproduce in
30/30 at +5..+120 s; [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) settles it as 400 'is disabled.' Resolution: keep as evidence;
superseded reading - [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) settles it as 400 'is disabled.'

Contradiction with [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035): FU2-01 settled by follow-up
probe tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and
+120 s after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every
PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K) was rejected with
BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put cells, 15/15 Create cells; the
enabled-key control right before DisableKey was 200 3/3), and the Get a (also supersedes -035's 'lazy
validation' remark: both Create and Put with a disabled key return 400 in 30/30.) Resolution: keep all;
[S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule
