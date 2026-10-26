<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLERECORDEXPIRATIONCONFIGURATION-001: Customer-created tables: every well-formed record-expiration Put/Get/JobStatus call -> MethodNotAllowedException 405 'not allowed'
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablerecordexpirationconfiguration-001"></a>**S3T-TABLERECORDEXPIRATIONCONFIGURATION-001** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Customer-created tables: every well-formed record-expiration Put/Get/JobStatus call -> MethodNotAllowedException 405 'not allowed'**
  **Scope verdict: defer**
  On three fresh ICEBERG tables created by the caller (with schema, without metadata; GetTable shows
  type=customer, managedByService=None): GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus, PutTableRecordExpirationConfiguration({enabled, days 1}), ({disabled,
  days 1}) and ({disabled} without settings) -> MethodNotAllowedException/405 with message ["Table record
  expiration configuration API calls to this table are not allowed"]; raw error body "{\"message\":\"Table
  record expiration configuration API calls to this table are not allowed\"}". The sibling
  GetTableMaintenanceJobStatus on the same tables -> HTTP 200, HTTP 200. Per the S3 user guide
  (s3-tables-record-expiration) record expiration is available only for AWS managed tables (S3 Storage Lens,
  SageMaker Catalog), 'not for S3 tables that you create'.
  - ACK: scope:defer, terminal_codes, docs-only · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: tableArn, value.status,
    value.settings.days, status, metrics
  - repro: CreateTable(ICEBERG) -> any of the three record-expiration operations with the table ARN
  - measurements: latency_ms=[50, 14, 10, 10, 10, 33, 10, 9, 10, 10, 32, 12, 8, 10, 10], n=15
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147),
    [S3T-TABLE-171](../table-record-expiration.md#s3t-table-171), [S3T-TABLE-222](../table-record-expiration.md#s3t-table-222), [S3T-TABLE-148](../table-record-expiration.md#s3t-table-148) · hypotheses: H-U-123, H-U-124, H-U-125, H-U-127 · evidence:
    tablerecordexpirationconfiguration/state-machine/job-status-soak

## Notes

H-U-123, H-U-124, H-U-125 and H-U-127 are UNTESTABLE on tables a controller can create: the premise (a
customer table with {enabled, days 1}) is rejected synchronously, so there is no NotYetRun/Successful/Disabled
lifecycle to soak. They can only be probed on an AWS managed table (Storage Lens or SageMaker Catalog export),
which this shard did not set up. MethodNotAllowedException IS in the modeled error list of all three
operations. A TableRecordExpirationConfiguration resource for customer tables would be permanently
terminal-errored.
