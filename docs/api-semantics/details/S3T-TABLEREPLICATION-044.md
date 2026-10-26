<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEREPLICATION-044: Existing table-level config: Put token omitted/'' -> BadRequestException; table or stale token -> ConflictException; current -> 200, rotates
_Full entry and notes of one finding; its summary entry is in [table-replication.md](../table-replication.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-tablereplication-044"></a>**S3T-TABLEREPLICATION-044** `idempotency` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Existing table-level config: Put token omitted/'' -> BadRequestException; table or stale token -> ConflictException; current -> 200, rotates**
  With a 'completed' table-level replication config: PutTableReplication(same document) with versionToken
  omitted -> BadRequestException/400 'A version token is not specified.'; '' -> BadRequestException/400 'A
  version token is not specified.'; GetTable.versionToken -> ConflictException/409 'The specified version
  token does not match the configuration.'; the token from the first Put (stale) -> ConflictException/409 'The
  specified version token does not match the configuration.'; the current GetTableReplication token -> HTTP
  200 (new token in the response: True). Status polled 1/s for 90 s after that identical re-Put:
  completed@0.0s. GetTableReplication versionToken sampled after Put/at failed/at completed/after re-Put: 3
  distinct values over 4 samples (['after-put-0', 'at-failed', 'at-completed', 'after-identical-reput']), i.e.
  status transitions do not rotate it; only Puts do.
  - ACK: custom_update, is_read_only, compare.is_ignored+delta_pre_compare · ops: PutTableReplication,
    GetTableReplication, GetTableReplicationStatus · fields: versionToken
  - repro: completed config -> Put(same doc) with each token source -> Put(current) -> poll status 90 s;
    GetTableReplication token sampled across phases
  - measurements: tokens_distinct=3, tokens_sampled=4, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010), [S3T-TABLEREPLICATION-004](../table-replication.md#s3t-tablereplication-004), [S3T-TABLEBUCKETREPLICATION-002](../tablebucket-replication.md#s3t-tablebucketreplication-002),
    [S3T-TABLEREPLICATION-015](../table-replication.md#s3t-tablereplication-015), [S3T-TABLEREPLICATION-016](../table-replication.md#s3t-tablereplication-016), [S3T-TABLEREPLICATION-018](../table-replication.md#s3t-tablereplication-018), [S3T-TABLEREPLICATION-007](../table-replication.md#s3t-tablereplication-007),
    [S3T-TABLEREPLICATION-020](../table-replication-status.md#s3t-tablereplication-020), [S3T-TABLEREPLICATION-021](../table-replication-status.md#s3t-tablereplication-021), [S3T-TABLEREPLICATION-038](../table-replication.md#s3t-tablereplication-038), [S3T-TABLEREPLICATION-039](../table-replication.md#s3t-tablereplication-039),
    [S3T-TABLEREPLICATION-014](../table-replication.md#s3t-tablereplication-014), [S3T-TABLEREPLICATION-029](../table-replication-status.md#s3t-tablereplication-029), [S3T-TABLEREPLICATION-003](../table-replication.md#s3t-tablereplication-003) · evidence:
    tablereplication/creative/controller-loop

## Notes

extends [S3T-TABLEREPLICATION-010](../table-replication.md#s3t-tablereplication-010) (first-Put token rules) to the update path a controller runs on an existing
config, and -004 with the token's stability across status transitions. Decision rule: on update always read
GetTableReplication first and forward its token; never forward the table's token; an unchanged re-Put is
harmless for the status but rotates the token (so a cached token from the previous reconcile is immediately
stale) so a controller must avoid re-Putting on no-op diffs. Single run.

Full original title: Existing table replication config: Put token omitted -> BadRequestException, '' ->
BadRequestException, table token -> ConflictException, stale -> ConflictException; identical re-Put with
current token -> 200 + token rotates, status stays completed
