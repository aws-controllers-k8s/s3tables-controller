<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-075: Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and 2147483647 round-trip; 2147483648 -> 204, read back clamped
_Full entry and notes of one finding; its summary entry is in
[tablebucket-maintenance-metrics.md](../tablebucket-maintenance-metrics.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-tablebucket-075"></a>**S3T-TABLEBUCKET-075** `request-validation` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and 2147483647 round-trip; 2147483648 -> 204, read back clamped**
  PutTableBucketMaintenanceConfiguration (botocore parameter validation disabled), status enabled,
  (unreferencedDays, nonCurrentDays): (0,10) and (-1,10) -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'value.settings.icebergUnreferencedFileRemoval.unreferencedDays' failed to satisfy
  constraint: Member must have value greater than or equal to 1'; (3,0) -> the same with 'nonCurrentDays';
  (36500,36500), (30,1), (36501,1), (1,36501), (1,1) and (2147483647,2147483647) -> HTTP 204 and Get returns
  exactly the sent integers (unreferencedDays > nonCurrentDays is accepted, no undocumented cap).
  (2147483648,1) -> HTTP 204 and Get returns unreferencedDays 2147483647: the out-of-int32 value was clamped,
  not rejected.
  - ACK: terminal_codes, none · ops: PutTableBucketMaintenanceConfiguration · fields:
    value.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: Put enabled with each (unreferencedDays, nonCurrentDays) pair above using a client with
    parameter_validation=False -> Get after accepted ones
  - measurements: accepted_put_latency_ms_n=7, n=10
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-074](../tablebucket-maintenance-metrics.md#s3t-tablebucket-074), [S3T-TABLEBUCKET-076](../tablebucket-maintenance-metrics.md#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](../tablebucket-maintenance-metrics.md#s3t-tablebucket-077), [S3T-TABLEBUCKET-080](../tablebucket-maintenance-metrics.md#s3t-tablebucket-080),
    [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](../tablebucket-maintenance-metrics.md#s3t-tablebucketmaintenanceconfiguration-001), [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](../tablebucket-maintenance-metrics.md#s3t-tablebucketmaintenanceconfiguration-002),
    [S3T-TABLEBUCKET-170](../tablebucket-maintenance-metrics.md#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](../tablebucket-maintenance-metrics.md#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](../tablebucket-maintenance-metrics.md#s3t-tablebucket-172), [S3T-TABLEBUCKET-054](../tablebucket-maintenance-metrics.md#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](../tablebucket-maintenance-metrics.md#s3t-tablebucket-055),
    [S3T-TABLEBUCKET-053](../tablebucket-maintenance-metrics.md#s3t-tablebucket-053) · hypotheses: H-S-029, H-S-111 · evidence:
    tablebucket/weird-inputs/maintenance-validation

## Notes

H-S-029 qualified: minimum 1 enforced with HTTP 400 naming the field and no upper bound or ordering constraint
(as hypothesized), but the wire code is ValidationException, not BadRequestException. H-S-111 confirmed:
2147483647 round-trips exactly, 36501 accepted. Surprise: 2147483648 is accepted and silently clamped to
2147483647 (one cell, n=1). With default botocore settings 0 and -1 never leave the client
(ParamValidationError, PositiveInteger min 1).

Full original title: Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and 2147483647
round-trip exactly; 2147483648 -> 204 but reads back clamped to 2147483647

Suspected controller bug confirmed by evidence: not flagged; evidence-driven. unreferencedDays/nonCurrentDays =
0 passes both the CRD (no Minimum marker) and the hand-written int32 guard (lower bound 0), reaches
PutTableBucketMaintenanceConfiguration, comes back as ValidationException 400 which is not a TableBucket
terminal code, and the resource requeues forever - raise the guard's lower bound to 1 (or add a kubebuilder
Minimum=1 marker) and/or add ValidationException to the TableBucket terminal codes. code:
config/crd/bases/s3tables.services.k8s.aws_tablebuckets.yaml:83-88 nonCurrentDays/unreferencedDays are bare
'format: int64, type: integer' (apis/v1alpha1/types.go:106-107 has no Minimum marker);
pkg/resource/table_bucket/hooks.go:35-40 maintenanceDaysToInt32 rejects only v < 0 || v > MaxInt32, so 0 is
forwarded by hooks.go:375-388 into the Put at hooks.go:393-404 and the error is returned raw;
pkg/resource/table_bucket/sdk.go:400-405 terminalAWSError matches only BadRequestException, so the
ValidationException is Recoverable and requeued; hooks.go:279-311 maintenanceNeedsSync keeps the 0 != observed
delta alive each cycle. The 2147483648 clamp is NOT reachable: the guard turns > MaxInt32 into a TerminalError
before any API call (hooks.go:376-386). finding: [S3T-TABLEBUCKET-075](../tablebucket-maintenance-metrics.md#s3t-tablebucket-075) (0 / -1 -> ValidationException 400
'Member must have value greater than or equal to 1'; 2147483648 clamped), [S3T-TABLEBUCKET-080](../tablebucket-maintenance-metrics.md#s3t-tablebucket-080)
(ValidationException is the wire code for every input validation failure on this operation and is absent from
the model).
