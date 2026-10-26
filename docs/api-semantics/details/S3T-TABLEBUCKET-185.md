<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-185: Disabled or PendingDeletion KMS key: PutTableBucketEncryption and CreateTableBucket(aws:kms) -> 400 'is disabled.' at +5/+30/+120 s (30/30)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-185"></a>**S3T-TABLEBUCKET-185** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Disabled or PendingDeletion KMS key: PutTableBucketEncryption and CreateTableBucket(aws:kms) -> 400 'is disabled.' at +5/+30/+120 s (30/30)**
  Enabled-key control (n=3): Put aws:kms K -> ['HTTP 200'], CreateTableBucket(aws:kms K) -> ['HTTP 200'].
  DisableKey -> HTTP 200; then per offset {KMS KeyState, Put outcomes, Create outcomes}: {"5": {"at_s": 6.4,
  "create": ["BadRequestException (HTTP 400)"], "key_state": "Disabled", "put": ["BadRequestException (HTTP
  400)"]}, "30": {"at_s": 31.5, "create": ["BadRequestException (HTTP 400)"], "key_state": "Disabled", "put":
  ["BadRequestException (HTTP 400)"]}, "120": {"at_s": 121.5, "create": ["BadRequestException (HTTP 400)"],
  "key_state": "Disabled", "put": ["BadRequestException (HTTP 400)"]}}. ScheduleKeyDeletion(7 d) -> HTTP 200;
  per offset: {"5": {"at_s": 6.4, "create": ["BadRequestException (HTTP 400)"], "key_state":
  "PendingDeletion", "put": ["BadRequestException (HTTP 400)"]}, "60": {"at_s": 61.6, "create":
  ["BadRequestException (HTTP 400)"], "key_state": "PendingDeletion", "put": ["BadRequestException (HTTP
  400)"]}}. Rejection messages: ["BadRequestException (HTTP 400) 'The specified KMS key ARN is disabled.'"].
  GetTableBucketEncryption after a rejected Put kept the AES256 reset value: True.
  - ACK: terminal_codes, references · ops: PutTableBucketEncryption, CreateTableBucket,
    GetTableBucketEncryption · fields: encryptionConfiguration.kmsKeyArn
  - repro: CreateKey K -> bucket B (AES256) -> DisableKey -> at +5/+30/+120 s: 3 x [Put AES256; Put aws:kms K;
    Get] and 3 x CreateTableBucket(aws:kms K) -> ScheduleKeyDeletion -> +5/+60 s same
  - measurements: disabled_offsets_s=[5, 30, 120], pending_offsets_s=[5, 60], trials_per_offset=3,
    put_cells=15, create_cells=15, n=30
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035) · evidence:
    tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny

## Notes

FU2-01: 107 holds. At +5 (actual 6.4), +30 and +120 s after DisableKey and at +5 and +60 s after
ScheduleKeyDeletion(7 d), every PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K) was
rejected with BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put cells, 15/15 Create
cells; the enabled-key control right before DisableKey was 200 3/3), and the Get after a rejected Put still
showed the AES256 reset value (Put validates before storing). 037's HTTP 200 'for a key disabled ~5 s earlier'
did not reproduce; if a KMS key-state propagation window exists it is shorter than ~6 s (offsets below 5 s
were not tested). The PendingDeletion message is the same 'is disabled.' text as 107 recorded.

Contradiction with [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107): FU2-01 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and +120 s
after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every PutTableBucketEncryption(aws:kms K)
and every CreateTableBucket(aws:kms K) was rejected with BadRequestException 400 'The specified KMS key ARN is
disabled.' (15/15 Put cells, 15/15 Create cells; the enabled-key control right before DisableKey was 200 3/3),
and the Get a Resolution: keep all; [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule

Contradiction with [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035): FU2-01 settled by follow-up
probe tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and
+120 s after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every
PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K) was rejected with
BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put cells, 15/15 Create cells; the
enabled-key control right before DisableKey was 200 3/3), and the Get a (also supersedes -035's 'lazy
validation' remark: both Create and Put with a disabled key return 400 in 30/30.) Resolution: keep all;
[S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule
