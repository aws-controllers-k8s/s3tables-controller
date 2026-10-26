<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-221: AWS-managed S3 Metadata journal table (type=aws): record-expiration Get, Put and GetJobStatus all -> MethodNotAllowedException 405
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-221"></a>**S3T-TABLE-221** `scope` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **AWS-managed S3 Metadata journal table (type=aws): record-expiration Get, Put and GetJobStatus all -> MethodNotAllowedException 405**
  **Scope verdict: defer**
  CreateBucketMetadataConfiguration(JournalTableConfiguration.RecordExpiration ENABLED, Days=7) on a
  general-purpose bucket created a journal table in the account's 'aws-s3' table bucket (namespace 'b_<bucket
  name>', table 'journal'); TableStatus CREATING -> ACTIVE in 15.2 s. GetTable(tableArn): type=aws,
  managedByService='metadata.s3.amazonaws.com' (a string naming the owning service, absent on customer
  tables), no managedTableInformation, format=ICEBERG. GetTableRecordExpirationConfiguration,
  GetTableRecordExpirationJobStatus and PutTableRecordExpirationConfiguration(enabled, days=1) each ->
  MethodNotAllowedException (HTTP 405) 'Table record expiration configuration API calls to this table are not
  allowed' - the identical code and message a customer table gets. The journal table's own RecordExpiration
  stays visible only through s3 GetBucketMetadataConfiguration ({Expiration: ENABLED, Days: 7}, unchanged by
  the rejected Put). GetTableMaintenanceConfiguration on the journal table -> HTTP 200.
  - ACK: scope:defer, terminal_codes, docs-only · ops: GetTableRecordExpirationConfiguration,
    PutTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus, GetTable · fields:
    configuration.status, configuration.settings.days, managedByService, type, managedTableInformation
  - repro: s3 CreateBucket -> CreateBucketMetadataConfiguration(JournalTable RecordExpiration ENABLED 7) ->
    wait ACTIVE -> s3tables GetTable/Get*/Put on the journal TableArn
  - measurements: journal_table_creating_s=[15.2], get_after_put_s=0.01, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146), [S3T-TABLE-224](../table-replication-status.md#s3t-table-224), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-031](../table-metadata-location.md#s3t-table-031),
    [S3T-TABLE-317](../table.md#s3t-table-317), [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-177](../table-replication.md#s3t-table-177), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147) · hypotheses: H-U-108, H-U-101,
    H-U-111, H-U-121, H-U-122 · evidence: table/creative/expiration-aws-managed-journal-table

## Notes

H-U-108 refuted for this service-managed table: Put AND Get are both 405 (the hypothesis expected Get -> 200
on a managedByService table). Combined with the customer-table result the s3tables record-expiration API
answered 405 on every table reachable from a lab account (customer plain, customer with schema, us-west-2 and
us-east-1, S3 Metadata journal); the only documented targets are S3 Storage Lens and SageMaker Catalog tables,
which were not reachable. H-U-101/H-U-121 on a managed table: refuted the same way (no 200 body). Also:
GetTable.managedByService is a string (service principal), not a boolean. One table, one region.
