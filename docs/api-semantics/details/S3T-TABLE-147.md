<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-147: PutTableRecordExpirationConfiguration on customer tables -> 405 for days 1/7/365, with/without Iceberg schema, in us-west-2 and us-east-1
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-147"></a>**S3T-TABLE-147** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration on customer tables -> 405 for days 1/7/365, with/without Iceberg schema, in us-west-2 and us-east-1**
  **Scope verdict: defer**
  Put(enabled, days=1/7/365) on a plain customer ICEBERG table in us-west-2 ->
  ['MethodNotAllowedException/405'], message ['Table record expiration configuration API calls to this table
  are not allowed']; the Get right after each Put -> ['MethodNotAllowedException/405']. Table created with
  metadata.iceberg.schema (int id + two timestamp columns named report_time/snapshot_time; GetTable has no
  managedByService member): GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405);
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  PutTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405). Plain table in us-east-1:
  GetTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405);
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405);
  PutTableRecordExpirationConfiguration -> MethodNotAllowedException (HTTP 405). A us-west-2 table ARN sent to
  the us-east-1 endpoint -> BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.'.
  - ACK: scope:defer, terminal_codes, docs-only · ops: PutTableRecordExpirationConfiguration,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: value.status,
    value.settings.days
  - repro: CreateTable (with and without iceberg schema; in us-west-2 and us-east-1) -> Put(enabled, days=1) /
    Get / GetJobStatus by tableArn
  - measurements: put_latency_ms=[12, 12, 10], n=9
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-004](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-004), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-005),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-001](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-001),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-003](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-003), [S3T-TABLE-173](../table-replication.md#s3t-table-173), [S3T-TABLE-174](../table-replication.md#s3t-table-174), [S3T-TABLE-176](../table-replication.md#s3t-table-176), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225) · hypotheses: H-U-111, H-U-108 · evidence:
    table/sub-resources/expiration-absent-roundtrip-jobstatus

## Notes

H-U-111 refuted on customer tables: the Put never returns 200/204, so no round-trip is observable; the 405 is
table-type gated, not region- or schema-gated (3 tables, 2 regions). H-U-108 qualified: 405 is not specific to
service-managed tables - it is the answer for every customer table too (declared MethodNotAllowedException:
'resource managed by a service or has restrictions'). A controller exposing recordExpiration on Table must
treat 405 as terminal (never requeue) and should document the field as AWS-managed-tables-only.
Service-managed tables themselves were not probed here (replication / Storage Lens / SageMaker not in this
shard).
