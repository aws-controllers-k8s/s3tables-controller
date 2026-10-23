<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-035: CreateTableBucket with aws:kms and a nonexistent key ARN -> BadRequestException 400 (not 404); other-region key -> 400...
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-035"></a>**S3T-TABLEBUCKET-035** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTableBucket with aws:kms and a nonexistent key ARN -> BadRequestException 400 (not 404); other-region key -> 400...**
  Nonexistent key in the caller account/region -> BadRequestException HTTP 400 'The specified KMS key ARN does
  not exist.', GetTableBucket on the constructed bucket ARN -> NotFoundException HTTP 404 'The specified
  bucket does not exist.'; key ARN in us-east-1 -> BadRequestException HTTP 400 'Cross region KMS keys are not
  supported.'; key ARN of another account -> ForbiddenException HTTP 403 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on this
  resource because the resource does not exist in this Region, no resource-based policies allow access, or a
  resource-based policy explicitly denies access'. No bucket was created in any case.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTableBucket, GetTableBucket · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTableBucket(aws:kms,
    kmsKeyArn=arn:aws:kms:<region>:<account>:key/00000000-0000-4000-8000-000000000000) and variants
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-034](../tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLEBUCKET-181](../tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLEBUCKET-106](../tablebucket-policy.md#s3t-tablebucket-106), [S3T-TABLE-086](../table-policy.md#s3t-table-086),
    [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107) · hypotheses: H-B-014 · evidence:
    tablebucket/sub-resources/kms-encryption-forms

## Notes

REFUTES H-B-014 (a): a well-formed nonexistent key ARN yields BadRequestException 'The specified KMS key ARN
does not exist.', not NotFoundException; so 404 from CreateTableBucket was not reachable through KMS.
Other-region key: BadRequestException 'Cross region KMS keys are not supported.'. Another account's key ARN:
ForbiddenException 403 whose message is the KMS authorization failure for kms:DescribeKey (the service
resolves the key as the caller). H-B-014 (b) disabled key is in the mutation finding: Create and Put both
succeeded with a disabled key (lazy validation), refuting the 400 prediction.

Full original title: CreateTableBucket with aws:kms and a nonexistent key ARN -> BadRequestException 400 (not
404); other-region key -> 400; other-account key -> ForbiddenException 403

Contradiction with [S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185), [S3T-TABLEBUCKET-037](../tablebucket-encryption-storage.md#s3t-tablebucket-037), [S3T-TABLEBUCKET-107](../tablebucket-encryption-storage.md#s3t-tablebucket-107): FU2-01 settled by follow-up
probe tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny: 107 holds. At +5 (actual 6.4), +30 and
+120 s after DisableKey and at +5 and +60 s after ScheduleKeyDeletion(7 d), every
PutTableBucketEncryption(aws:kms K) and every CreateTableBucket(aws:kms K) was rejected with
BadRequestException 400 'The specified KMS key ARN is disabled.' (15/15 Put cells, 15/15 Create cells; the
enabled-key control right before DisableKey was 200 3/3), and the Get a (also supersedes -035's 'lazy
validation' remark: both Create and Put with a disabled key return 400 in 30/30.) Resolution: keep all;
[S3T-TABLEBUCKET-185](../tablebucket-encryption-storage.md#s3t-tablebucket-185) is canonical for the settled rule
