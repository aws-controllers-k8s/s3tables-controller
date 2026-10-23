<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-182: Disabled KMS key (used seconds before): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; EnableKey recovers
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-182"></a>**S3T-TABLEBUCKET-182** `error-code` · impact high · handled · verified 2026-10-14
  **Disabled KMS key (used seconds before): PutTableBucketEncryption/CreateTable BadRequestException 400 from +3 s to +300 s; EnableKey recovers**
  Key used by CreateTable(schema) seconds before (-> HTTP 200), then kms DisableKey (KeyState Disabled).
  Series: +3 s: Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200;
  +30 s: Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +120 s:
  Put BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200; +300 s: Put
  BadRequestException (HTTP 400), CreateTable BadRequestException (HTTP 400), Get HTTP 200. Error messages
  seen: ["BadRequestException: 'Insufficient access to perform table maintenance. Ensure that the key used to
  encrypt the table still exists, is active, and has a resource policy granting access to the S3 service
  principal 'maintenance.s3tables.amazonaws.com'.'", "BadRequestException: 'The specified KMS key ARN is
  disabled.'"]. After EnableKey, PutTableBucketEncryption succeeded after 0.04 s (polled every 2 s; codes seen
  ['OK']); CreateTable(schema) right after -> HTTP 200.
  - ACK: terminal_codes, requeue · ops: PutTableBucketEncryption, CreateTable, GetTableBucketEncryption ·
    fields: encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(schema) in an SSE-KMS bucket -> kms DisableKey -> PutTableBucketEncryption(aws:kms,
    key)/CreateTable(schema)/GetTableBucketEncryption at +3, +30, +120, +300 s -> EnableKey -> poll Put
  - measurements: offsets_s=[3, 30, 120, 300], recovery_after_enable_s=0.04, n=1
  - handling: handled via `generator.yaml:220-225; test/e2e/tests/test_table_bucket.py:158-162`
  - related: [S3T-TABLEBUCKET-180](../tablebucket-encryption-storage.md#s3t-tablebucket-180), [S3T-TABLEBUCKET-183](../tablebucket-encryption-storage.md#s3t-tablebucket-183) · evidence:
    tablebucket/sub-resources/fu-01-kms-key-policy-disable

## Notes

FU-01 part 3 (recently used key). One series; whether the service caches key state would show as a success at
+3 s turning into a failure later - see the series for what was observed. Messages differ per operation:
PutTableBucketEncryption -> 'The specified KMS key ARN is disabled.' (specific), while CreateTable ->
'Insufficient access to perform table maintenance. Ensure that the key ... still exists, is active, and has a
resource policy granting access to ... maintenance.s3tables.amazonaws.com' - the same generic message as a
missing key-policy grant, so a controller cannot tell a disabled key from a missing grant from the CreateTable
error alone. No caching was observed: the failure was visible at +3 s and the recovery on the first call after
EnableKey (0.04 s); GetTableBucketEncryption kept returning 200 with the disabled key ARN throughout.
