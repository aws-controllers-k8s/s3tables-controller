<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-224: Replica writes: Rename -> 2xx; PutMaintenance -> BadRequestException; PutExpiration -> MethodNotAllowedException; Tag -> 2xx...
_Full entry and notes of one finding; its summary entry is in
[table-replication-status.md](../table-replication-status.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-224"></a>**S3T-TABLE-224** `immutable-field` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Replica writes: Rename -> 2xx; PutMaintenance -> BadRequestException; PutExpiration -> MethodNotAllowedException; Tag -> 2xx...**
  On the service-created replica while the source configuration is active. By bucket+namespace+name:
  PutTableMaintenanceConfiguration(icebergCompaction) -> BadRequestException (HTTP 400) 'The specified
  maintenance type is not valid for service managed tables.'; PutTablePolicy -> HTTP 200;
  UpdateTableMetadataLocation -> BadRequestException (HTTP 400) 'Updates to service managed tables are not
  allowed.'; RenameTable -> HTTP 204 - the replica's GetTable.name became 'ackq_f81240_t_renamed'
  (managedByService still 'replication.s3tables.amazonaws.com'), ListTables(D) then ['ackq_f81240_t_renamed'],
  GetTableReplicationStatus(source) {"destinations": [{"destinationTableArn":
  "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest/table/93cf68ab-77f5-416b-b185-18cbf19578e8",
  "destinationTableBucketArn": "arn:aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-b73e55-repl-dest",
  "replicationStatus": "completed"}], "sourceTableArn":
  "arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-f81240-lc-src/table/46c786d8-2add-407d-a37c-66608a1da538"};
  RenameTable back -> HTTP 204 (name 'ackq_f81240_t'). By table ARN:
  PutTableRecordExpirationConfiguration(enabled, 5 days) -> MethodNotAllowedException (HTTP 405) 'Table record
  expiration configuration API calls to this table are not allowed'; TagResource -> HTTP 200
  (ListTagsForResource then {"replkey": "replval"}); UntagResource -> HTTP 204; PutTableReplication(replica ->
  S) -> BadRequestException (HTTP 400) 'Replication operations are not allowed on service managed tables.';
  GetTableReplication -> BadRequestException (HTTP 400) 'Replication operations are not allowed on service
  managed tables.'; GetTableReplicationStatus -> BadRequestException (HTTP 400) 'Replication operations are
  not allowed on service managed tables.'; GetTableRecordExpirationConfiguration -> MethodNotAllowedException
  (HTTP 405) 'Table record expiration configuration API calls to this table are not allowed' null;
  GetTableRecordExpirationJobStatus -> MethodNotAllowedException (HTTP 405) 'Table record expiration
  configuration API calls to this table are not allowed' null. ListTagsForResource(replica) before any tag
  write -> {} (source tags: {"ackq-run": "f81240", "srckey": "srcval"}).
  - ACK: is_read_only, terminal_codes, tags.custom-sync · ops: RenameTable, PutTableMaintenanceConfiguration,
    PutTableRecordExpirationConfiguration, TagResource, UntagResource, UpdateTableMetadataLocation,
    PutTableReplication, PutTablePolicy, GetTableReplication, GetTableReplicationStatus,
    GetTableRecordExpirationConfiguration, GetTableRecordExpirationJobStatus · fields: managedByService, tags,
    configuration.status, metadataLocation, name
  - repro: replica present -> each write op once (by-name ops first, rename last and renamed back); reads on
    the replica ARN
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-136](../table-replication-status.md#s3t-table-136), [S3T-TABLE-137](../table-replication-status.md#s3t-table-137), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-007](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-007),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-008](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-008), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-009](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-009), [S3T-TABLE-146](../table-record-expiration.md#s3t-table-146),
    [S3T-TABLE-221](../table-record-expiration.md#s3t-table-221), [S3T-TABLE-225](../table-replication-status.md#s3t-table-225), [S3T-TABLE-299](../table-replication-status.md#s3t-table-299), [S3T-TABLE-223](../table-replication-status.md#s3t-table-223), [S3T-TABLE-307](../table-replication-status.md#s3t-table-307), [S3T-TABLE-300](../table-replication-status.md#s3t-table-300), [S3T-TABLE-306](../table-replication-status.md#s3t-table-306),
    [S3T-TABLE-196](../table-replication-status.md#s3t-table-196), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147), [S3T-TABLEBUCKETREPLICATION-031](../tablebucket-replication.md#s3t-tablebucketreplication-031), [S3T-TABLEBUCKETREPLICATION-032](../tablebucket-replication.md#s3t-tablebucketreplication-032),
    [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-041](../table-replication-status.md#s3t-tablereplication-041), [S3T-TABLEREPLICATION-025](../table-replication-status.md#s3t-tablereplication-025), [S3T-TABLEREPLICATION-030](../table-replication-status.md#s3t-tablereplication-030),
    [S3T-TABLEREPLICATION-031](../table-replication-status.md#s3t-tablereplication-031), [S3T-TABLEREPLICATION-019](../table-replication-status.md#s3t-tablereplication-019), [S3T-TABLEREPLICATION-028](../table-replication-status.md#s3t-tablereplication-028) · hypotheses: H-U-165,
    H-U-194, H-U-199, H-U-158, H-U-159, H-U-178 · evidence: table/state-machine/repl-status-lifecycle

## Notes

H-U-165 REFUTED in part (RenameTable on a replica succeeds): codes per op {'rename': '2xx', 'maintenance':
'BadRequestException', 'expiration': 'MethodNotAllowedException', 'commit': 'BadRequestException', 'policy':
'2xx'}. H-U-194 confirmed (tags writable on the replica, source tags NOT copied). H-U-199 confirmed.
H-U-158/H-U-159 both refuted: replica-ARN replication reads -> BadRequestException. H-U-178 confirmed as
BadRequestException (not 405).

Full original title: Replica writes: Rename -> 2xx; PutMaintenance -> BadRequestException; PutExpiration ->
MethodNotAllowedException; Tag -> 2xx; Untag -> 2xx; commit -> BadRequestException; chained Put ->
BadRequestException; PutPolicy -> 2xx
