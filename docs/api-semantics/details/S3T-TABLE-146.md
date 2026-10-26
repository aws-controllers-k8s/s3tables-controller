<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-146: Customer table: GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405, not 200/404
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-146"></a>**S3T-TABLE-146** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Customer table: GetTableRecordExpirationConfiguration and GetTableRecordExpirationJobStatus -> MethodNotAllowedException 405, not 200/404**
  Plain ICEBERG table (CreateTable without metadata, type=customer, GetTable has no managedByService member),
  never configured: GetTableRecordExpirationConfiguration called 0.09 s after CreateTable and then every 1 s
  for 10 s -> 10/10 x MethodNotAllowedException/405; message ['Table record expiration configuration API calls
  to this table are not allowed']. GetTableRecordExpirationJobStatus on the same fresh table ->
  MethodNotAllowedException (HTTP 405) 'Table record expiration configuration API calls to this table are not
  allowed'. After three rejected Put attempts, GetTableRecordExpirationJobStatus -> MethodNotAllowedException
  (HTTP 405) 'Table record expiration configuration API calls to this table are not allowed'.
  - ACK: terminal_codes, exceptions.404, custom_find · ops: GetTableRecordExpirationConfiguration,
    GetTableRecordExpirationJobStatus · fields: configuration.status, configuration.settings.days, status
  - repro: CreateTable(format=ICEBERG) -> GetTableRecordExpirationConfiguration(tableArn) immediately and
    every 1 s for 10 s; GetTableRecordExpirationJobStatus
  - measurements: first_get_after_create_s=0.09, window_s=10, n=10
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-173](../table-replication.md#s3t-table-173),
    [S3T-TABLE-174](../table-replication.md#s3t-table-174), [S3T-TABLE-176](../table-replication.md#s3t-table-176), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147) · hypotheses: H-U-101, H-U-102, H-U-121, H-U-122 · evidence:
    table/sub-resources/expiration-absent-roundtrip-jobstatus

## Notes

H-U-101 refuted (no 200/default struct), H-U-102 refuted (the code is MethodNotAllowedException, not
NotFoundException), H-U-121 refuted (no status=Disabled body), H-U-122 refuted (not 404 either): on a
customer-created table both reads are rejected with 405 'Table record expiration configuration API calls to
this table are not allowed'. The 405 is immediate (first call < 1 s after create) and stable over 10 s.
Consistent with the public S3 Tables user guide: record expiration is only offered for specific AWS-managed
tables (S3 Storage Lens, SageMaker Catalog), 'not available for S3 tables that you create'.
