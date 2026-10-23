<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLERECORDEXPIRATIONCONFIGURATION-003: GetTableRecordExpirationJobStatus on a fresh customer table -> MethodNotAllowedException 405 in 90/90 calls (30 seq + 2x30 conc.), no 429
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablerecordexpirationconfiguration-003"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-003** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableRecordExpirationJobStatus on a fresh customer table -> MethodNotAllowedException 405 in 90/90 calls (30 seq + 2x30 conc.), no 429**
  On a fresh table with no record-expiration configuration: 30 sequential
  GetTableRecordExpirationJobStatus(tableArn) with no pause -> {"MethodNotAllowedException/405": 30} (latency
  [10, 13] ms); 30 concurrent (10 threads) -> {"MethodNotAllowedException/405": 30} (wall 0.13s), repeated
  after 20 s -> {"MethodNotAllowedException/405": 30} (wall 0.05s). Messages: {"MethodNotAllowedException":
  "Table record expiration configuration API calls to this table are not allowed"}.
  - ACK: terminal_codes, requeue · ops: GetTableRecordExpirationJobStatus · fields: tableArn
  - repro: 30 x GetTableRecordExpirationJobStatus(tableArn) back-to-back; 10 threads x 30; repeat after 20 s
  - measurements: sequential_latency_ms_min_max=[10, 13], burst_wall_s=[0.13, 0.05], n=90
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-001), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147),
    [S3T-TABLE-171](../table-record-expiration.md#s3t-table-171), [S3T-TABLE-222](../table-record-expiration.md#s3t-table-222), [S3T-TABLE-148](../table-record-expiration.md#s3t-table-148), [S3T-TABLE-187](../table-maintenance.md#s3t-table-187) · hypotheses: H-U-130 · evidence:
    table/limits/default-page-120-tables

## Notes

H-U-130 untested for the 429 question: on a freshly created ICEBERG table with no record-expiration
configuration every call (30 sequential, 2 x 30 concurrent) was rejected with MethodNotAllowedException (HTTP 405)
'Table record expiration configuration API calls to this table are not allowed' - a code that is in the model
only for the table sub-resource operations and that a controller would otherwise treat as terminal. None of
the 90 rapid calls returned TooManyRequestsException. Which table state makes this API 'allowed' was not
investigated here (another runner covers record expiration).

Full original title: GetTableRecordExpirationJobStatus on a fresh table -> MethodNotAllowedException 405
'Table record expiration configuration API calls to this table are not allowed' (90/90, no 429)
