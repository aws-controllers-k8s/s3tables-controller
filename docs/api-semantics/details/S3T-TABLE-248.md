<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-248: CreateTable(aws:kms K) with K Disabled or PendingDeletion -> BadRequestException 400 'KMS key ARN is disabled.' (same text both)
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-248"></a>**S3T-TABLE-248** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **CreateTable(aws:kms K) with K Disabled or PendingDeletion -> BadRequestException 400 'KMS key ARN is disabled.' (same text both)**
  K1 = SYMMETRIC_DEFAULT with root + maintenance-principal policy; CreateTable(aws:kms K1) while Enabled ->
  HTTP 200. DisableKey -> Disabled visible after 0.0 s, +5 s: CreateTable(aws:kms K1) -> BadRequestException
  (HTTP 400) 'The specified KMS key ARN is disabled.' (x-amzn-ErrorType
  BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/; second attempt 10 s later:
  same code). PutTableBucketEncryption(aws:kms K1) while Disabled -> BadRequestException (HTTP 400) 'The
  specified KMS key ARN is disabled.'. EnableKey -> Enabled after 0.0 s, +5 s: CreateTable -> HTTP 200
  (attempts 1). ScheduleKeyDeletion(7 d) -> PendingDeletion after 0.0 s, +5 s: CreateTable(aws:kms K1) ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is disabled.' (second attempt: same code). Message
  texts: Disabled='The specified KMS key ARN is disabled.'; PendingDeletion='The specified KMS key ARN is
  disabled.'.
  - ACK: terminal_codes, requeue · ops: CreateTable, PutTableBucketEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateKey(symmetric, maintenance principal) -> DisableKey -> CreateTable(aws:kms K); EnableKey ->
    CreateTable; ScheduleKeyDeletion(7) -> CreateTable
  - measurements: kms_state_visible_in_describe_key_s.disabled=0.01,
    kms_state_visible_in_describe_key_s.enabled=0.02,
    kms_state_visible_in_describe_key_s.pending_deletion=0.01, create_table_latency_ms.enabled=652,
    create_table_latency_ms.disabled=[30, 30], create_table_latency_ms.pending_deletion=[18, 36],
    create_table_latency_ms.reenabled=531, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-253](../table-maintenance.md#s3t-table-253), [S3T-TABLE-301](../table-maintenance.md#s3t-table-301), [S3T-TABLE-254](../table-maintenance.md#s3t-table-254), [S3T-TABLE-246](../table-maintenance.md#s3t-table-246), [S3T-TABLE-054](../table-maintenance.md#s3t-table-054), [S3T-TABLE-239](../table-record-expiration.md#s3t-table-239),
    [S3T-TABLE-055](../table-maintenance.md#s3t-table-055), [S3T-TABLEBUCKET-119](../tablebucket-encryption-storage.md#s3t-tablebucket-119), [S3T-TABLE-242](../table-encryption-storage.md#s3t-table-242) · hypotheses: H-T-167 · evidence:
    table/error-taxonomy/kms-key-states-types

## Notes

H-T-167 confirmed: BadRequestException 400 (not NotFoundException) for both Disabled and PendingDeletion;
qualified in that the message does not embed the KMS exception name and is identical for both states ('The
specified KMS key ARN is disabled.'), so PendingDeletion is indistinguishable from Disabled; a missing key
says 'The specified KMS key ARN does not exist.' and a key ARN in another region says 'Cross region KMS keys
are not supported.' (see the companion finding). Key-state changes were visible to S3 Tables on the first
CreateTable 5 s after the KMS call; no propagation lag observed (n=1 per state).

Full original title: CreateTable(aws:kms K) with K Disabled or PendingDeletion -> BadRequestException 400 'The
specified KMS key ARN is disabled.' (same text for both states)
