<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-058: Table encryption is a creation-time snapshot: inheriting KMS table unchanged 66 polls/120 s after DeleteTableBucketEncryption...
_Full entry and notes of one finding; its summary entry is in
[table-encryption-storage.md](../table-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-058"></a>**S3T-TABLE-058** `immutable-field` · impact high · handled · verified 2026-10-13
  **Table encryption is a creation-time snapshot: inheriting KMS table unchanged 66 polls/120 s after DeleteTableBucketEncryption...**
  t_inherit (created without config under bucket default aws:kms K) read
  'aws:kms|arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a'.
  DeleteTableBucketEncryption -> HTTP 204; GetTableBucketEncryption right after -> HTTP 200
  {"encryptionConfiguration": {"sseAlgorithm": "AES256"}}. GetTableEncryption(t_inherit) polled 1 s x 60 then
  10 s to 120 s: values ['aws:kms|arn:aws:kms:us-west-2:<ACCOUNT>:key/44faa909-a32b-42d8-baed-eec5099b107a'].
  A no-config table created after the delete reads {"sseAlgorithm": "AES256"}.
  PutTableBucketEncryption(aws:kms K) again -> HTTP 200; that AES256 no-config table polled 1 s x 60:
  ['AES256|<absent>']; the explicit-AES256 table sampled every 10 s: ['AES256|<absent>'].
  - ACK: is_immutable, compare.is_ignored+delta_pre_compare, late_initialize · ops:
    DeleteTableBucketEncryption, PutTableBucketEncryption, GetTableEncryption, GetTableBucketEncryption ·
    fields: encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: Bucket default aws:kms K -> CreateTable(no config) -> DeleteTableBucketEncryption -> poll
    GetTableEncryption 120 s -> CreateTable(no config) -> PutTableBucketEncryption(K) -> poll 60 s
  - measurements: poll_window_after_delete_s=120, poll_window_after_put_s=60, polls_inherit=66,
    polls_plain=60, n=126
  - handling: handled via `generator.yaml:24-49; pkg/resource/table/hooks.go:69-75; generator.yaml:47-49; apis/v1alpha1/table.go:54-59`
  - related: [S3T-TABLE-021](../table-encryption-storage.md#s3t-table-021), [S3T-TABLE-057](../table-encryption-storage.md#s3t-table-057) · hypotheses: H-T-153, H-T-154 · evidence:
    table/dependencies/kms-maintenance-principal

## Notes

H-T-153 (snapshot) confirmed; H-T-154 (live inheritance within seconds) refuted. Only one KMS key is available
in this shard, so the 'switch to K2' cell was replaced by delete-default and re-put-default.

Full original title: Table encryption is a creation-time snapshot: inheriting KMS table unchanged 66 polls/120
s after DeleteTableBucketEncryption; AES256 table unchanged 60 s after bucket re-keyed
