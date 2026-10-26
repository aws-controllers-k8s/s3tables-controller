<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-087: PutTableBucketStorageClass enum/null violations return wire code ValidationException (HTTP 400), a shape NOT in the API model
_Full entry and notes of one finding; its summary entry is in
[tablebucket-encryption-storage.md](../tablebucket-encryption-storage.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-087"></a>**S3T-TABLEBUCKET-087** `error-code` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **PutTableBucketStorageClass enum/null violations return wire code ValidationException (HTTP 400), a shape NOT in the API model**
  Seven invalid storageClassConfiguration values ('standard', 'Standard', 'STANDARD_IA', 'GLACIER',
  'INTELLIGENT-TIERING', '' and {}) were all rejected with HTTP 400 and error code 'ValidationException',
  message "1 validation error detected: Value at 'storageClassConfiguration.storageClass' failed to satisfy
  constraint: Member must satisfy enum value set: [INTELLIGENT_TIERING, STANDARD]" (for {}: '... Member must
  not be null'). The model declares only BadRequestException for this operation; a malformed ARN on the same
  operation family returns BadRequestException 400 instead.
  - ACK: terminal_codes · ops: PutTableBucketStorageClass · fields: storageClassConfiguration.storageClass
  - repro: PutTableBucketStorageClass with storageClass 'standard' (SDK parameter validation disabled) -> read
    the error code
  - measurements: n=7
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-086](../tablebucket-encryption-storage.md#s3t-tablebucket-086), [S3T-TABLEBUCKET-026](../tablebucket.md#s3t-tablebucket-026), [S3T-TABLEBUCKET-023](../tablebucket-encryption-storage.md#s3t-tablebucket-023), [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024),
    [S3T-TABLEBUCKET-104](../tablebucket-encryption-storage.md#s3t-tablebucket-104), [S3T-TABLE-012](../table-encryption-storage.md#s3t-table-012), [S3T-TABLE-013](../table-encryption-storage.md#s3t-table-013), [S3T-TABLE-020](../table-encryption-storage.md#s3t-table-020) · hypotheses: H-S-040, H-S-072 ·
    evidence: tablebucket/sub-resources/storage-class-default-transitions

## Notes

H-S-072 qualified: the wire code for enum validation failures is ValidationException, not one of the seven
declared shapes; terminal-code lists keyed on BadRequestException would miss it. Seen 7/7 times on one bucket.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. A misspelled
spec.storageClassConfiguration.storageClass on an existing TableBucket is never rejected by the CRD and is
re-Put on every reconcile forever (ValidationException is not in terminal_codes), while the same value at
create goes Terminal via BadRequestException - add an enum marker on storageClass or add ValidationException
to the TableBucket terminal codes. code: config/crd/bases/s3tables.services.k8s.aws_tablebuckets.yaml:116-117
storageClass is a plain 'type: string' (no enum; apis/v1alpha1/types.go:164 carries no kubebuilder marker), so
'standard' passes admission; pkg/resource/table_bucket/hooks.go:519-536 syncStorageClass sends the raw string
to PutTableBucketStorageClass and returns the error unchanged; pkg/resource/table_bucket/sdk.go:400-405
terminalAWSError matches only BadRequestException, so the ValidationException is Recoverable and the
reconciler requeues; hooks.go:98-105 reads STANDARD back into latest so delta.go keeps the 'standard' !=
'STANDARD' delta alive each cycle; create path goes Terminal (generator.yaml:267-268). finding:
[S3T-TABLEBUCKET-087](../tablebucket-encryption-storage.md#s3t-tablebucket-087) (ValidationException 400 on PutTableBucketStorageClass, 7/7), [S3T-TABLEBUCKET-024](../tablebucket-encryption-storage.md#s3t-tablebucket-024)
(BadRequestException 400 for the same value on CreateTableBucket).
