<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-107: PutTableBucketEncryption validates the KMS key at Put time: nonexistent, Disabled and PendingDeletion keys -> BadRequestException 400
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-107"></a>**S3T-TABLEBUCKET-107** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketEncryption validates the KMS key at Put time: nonexistent, Disabled and PendingDeletion keys -> BadRequestException 400** (hypothesis refuted; behavior confirmed)
  Well-formed ARN of a key that does not exist (same region) -> BadRequestException 400 ('The specified KMS
  key ARN does not exist.'); same in us-east-1 -> BadRequestException 400 ('Cross region KMS keys are not
  supported.'). Key disabled (DisableKey -> HTTP 200): Put -> BadRequestException 400 ('The specified KMS key
  ARN is disabled.'), Get -> AES256. Key scheduled for deletion (ScheduleKeyDeletion -> HTTP 200): Put ->
  BadRequestException 400 ('The specified KMS key ARN is disabled.'), Get -> AES256.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketEncryption, GetTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Put with a random key UUID ARN; DisableKey -> Put; ScheduleKeyDeletion(7) -> Put; Get after each
  - measurements: n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-105](../tablebucket-encryption-storage.md#s3t-tablebucket-105), [S3T-TABLEBUCKET-128](../tablebucket-encryption-storage.md#s3t-tablebucket-128), [S3T-TABLEBUCKET-131](../tablebucket.md#s3t-tablebucket-131), [S3T-TABLE-085](../table-encryption-storage.md#s3t-table-085),
    [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLE-249](../table-encryption-storage.md#s3t-table-249), [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035) · hypotheses: H-S-008 ·
    evidence: tablebucket/weird-inputs/encryption-put-validation

## Notes

H-S-008 REFUTED (hypothesis: no key-state validation at Put): a nonexistent key -> 400 'The specified KMS key
ARN does not exist.'; a Disabled key -> 400 'The specified KMS key ARN is disabled.'; a PendingDeletion key ->
400 with the SAME 'is disabled' message (KMS reports KeyState=PendingDeletion). Put rejects before storing:
Get still returned the previous AES256 value. 'A key the caller cannot use' (other account) was not tested.
The key was then left in PendingDeletion as the shard's cleanup.

Contradiction with [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037): Disabled KMS key at PutTableBucketEncryption(aws:kms): 037 got HTTP
200 for a key disabled ~5 s earlier (and CreateTableBucket with it -> 200; its notes conclude 'key state is
not validated at bucket create/put'); 107 got BadRequestException 400 'The specified KMS key ARN is disabled.'
for a Disabled key and the same 400 for a PendingDeletion key. Both n=1; 107 does not record the delay between
DisableKey and the Put. The exact server message in 107 shows key state is checked at least sometimes; KMS
key-state propagation lag at +5 s would reconcile the two but is untested, and 107 is a refuted-hypothesis
record so cannot be canonical Resolution: re-probe: weird-inputs - PutTableBucketEncryption(aws:kms) and
CreateTableBucket(aws:kms) at +5/+30/+120 s after DisableKey, n>=3 per offset; record 200 vs 400 'The
specified KMS key ARN is disabled.' and the Get afterwards; repeat once after ScheduleKeyDeletion

Contradiction with [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037): FU2-01 settled by follow-up probe
tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and +120 s
after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every PutTableBucketEncryption(aws:kms K)
and every CreateTableBucket(aws:kms K) was rejected with BadRequestException 400 'The specified KMS key ARN is
disabled.' (15/15 Put cells, 15/15 Create cells; the enabled-key control right before DisableKey was 200 3/3),
and the Get a Resolution: keep all; [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule

Contradiction with [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-035](../tablebucket-encryption-storage.md#s3t-tablebucket-035): FU2-01 settled by follow-up
probe tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and
+120 s after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every
PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K) was rejected with
BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put cells, 15/15 Create cells; the
enabled-key control right before DisableKey was 200 3/3), and the Get a (also supersedes -035's 'lazy
validation' remark: both Create and Put with a disabled key return 400 in 30/30.) Resolution: keep all;
[S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule
