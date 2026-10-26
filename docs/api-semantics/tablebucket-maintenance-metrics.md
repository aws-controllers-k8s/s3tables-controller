<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# TableBucket maintenance configuration and metrics configuration
_PutTableBucketMaintenanceConfiguration (merge-patch semantics, int32 clamping, ValidationException on days < 1, server defaults enabled/3/10) and the metrics configuration pair; scope verdict for TableBucketMaintenanceConfiguration._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `tablebucket-maintenance-metrics.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the two bucket-scoped sub-resources that are neither policy, encryption nor replication: the
maintenance configuration - PutTableBucketMaintenanceConfiguration / GetTableBucketMaintenanceConfiguration on a bucket
ARN, one type (`icebergUnreferencedFileRemoval`), one value {status, settings.icebergUnreferencedFileRemoval
{unreferencedDays, nonCurrentDays}} - and the metrics configuration triple Put/Get/DeleteTableBucketMetricsConfiguration.
There is no state machine and no asynchrony on either: every maintenance Put returned 204 and the Get issued immediately
after it carried the new value (12/12 Put/Get pairs; status flips visible on the first Get in 4/4 and unchanged at +1 s),
and the metrics Delete was 404 on the first poll 3-13 ms later with no later 200 ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054), -056, -095).
The maintenance value is read back only through GetTableBucketMaintenanceConfiguration, as a `configuration` map keyed by
the type string with value keys ['settings', 'status'] plus the ARN echoed byte-identical; Put takes (type, value)
([S3T-TABLEBUCKET-057](#s3t-tablebucket-057), -053). Scope verdict for a would-be TableBucketMaintenanceConfiguration CRD: the findings show a
Put/Get pair with no Delete operation in the SDK model ([S3T-TABLEBUCKET-056](#s3t-tablebucket-056)), a merge-patch Put (status-only keeps the
days, one day keeps the other, settings-only keeps the status, value {} is a 204 no-op, and the defaults never reappear
once days were set: [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), -055, -172) and a server default materialized on every fresh bucket (Get -> 200
enabled/3/10 at +0 s and +5 s, never 404 on a live bucket: [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), -173, -079) - so there is no absent
state for such a resource to create or to delete into, and "spec removed" can only be expressed as a Put of an explicit
value (-056 records the implication `scope:field-on-parent`). The controller implements no such CRD today and reconciles
the value as `spec.maintenanceConfiguration` on TableBucket (render.yaml partition note); the metrics triple has no
controller surface at all and its findings render as unhandled ([S3T-TABLEBUCKET-093](#s3t-tablebucket-093), -094, -095, -100, -177).

### Rules

1. PutTableBucketMaintenanceConfiguration is a merge-patch into the stored value, not a whole-value replace: after
   enabled 7/20, {status: disabled} without settings read back disabled/7/20, {status: enabled} without settings
   enabled/7/20, {enabled, unreferencedDays 5 only} enabled/5/20, {enabled, nonCurrentDays 20 only} enabled/5/20 and
   {settings 5/15, no status} kept the stored status (disabled/5/15); omitting a field never resets it to the 3/10
   default (doc claim C050 is FALSE), so compare only the sub-fields the spec declares and never expect reset-by-omission
   ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054), -172, -170, -171).
2. There is no DeleteTableBucketMaintenanceConfiguration in the SDK model; a spec field removed has no API counterpart,
   so a controller must Put an explicit value (the fresh-bucket default is enabled/3/10) and turning maintenance off
   requires an explicit `status: disabled` ([S3T-TABLEBUCKET-056](#s3t-tablebucket-056), -053).
3. A Put that carries nothing changes nothing and cannot act as a reset: value {} -> 204 with the stored enabled/9/9
   kept, settings {icebergUnreferencedFileRemoval: {}} -> 204 no-op, server-side settings {} (empty union, client
   validation off) -> 204 no-op; only a body with no `value` member at all is rejected, ValidationException 400 'Member
   must not be null' ([S3T-TABLEBUCKET-055](#s3t-tablebucket-055), -076).
4. A never-Put bucket already carries enabled/3/10 - GetTableBucketMaintenanceConfiguration is 200 with identical bodies
   at +0 s and +5 s, not 404 (doc claim C051 TRUE), and Put {status: enabled} without settings keeps 3/10 - so there is
   no "unset" state to detect; a spec that omits the field must take the observed value as its default rather than
   assume disabled ([S3T-TABLEBUCKET-053](#s3t-tablebucket-053), -173).
5. Every server-side input-validation rejection on this Put (status outside [enabled, disabled], type other than
   icebergUnreferencedFileRemoval, days < 1, missing value) came back as wire code ValidationException, HTTP 400, '1
   validation error detected: Value at '<path>' failed to satisfy constraint: ...' - 11 of 11 rejections - a code the
   SDK model does not list for the operation (it lists BadRequestException, seen only for an invalid bucket name/region
   in the ARN and for the table-level Put of the bucket-only type); a terminal-code list keyed on BadRequestException
   misses all of them ([S3T-TABLEBUCKET-080](#s3t-tablebucket-080), -074, -077, -075).
6. The only server-side bound on the day counts is >= 1: 0 and -1 -> ValidationException 400 'Value at
   'value.settings.icebergUnreferencedFileRemoval.<field>' ... Member must have value greater than or equal to 1' with
   the stored value unchanged; 1, 30, 36500, 36501 and 2147483647 round-trip exactly, and unreferencedDays >
   nonCurrentDays is accepted (no undocumented cap, no ordering constraint) ([S3T-TABLEBUCKET-075](#s3t-tablebucket-075),
   [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002), -170).
7. Values above int32 are not rejected but silently clamped: 2147483648, 4294967296 and 9223372036854775807 (n=3 per
   cell, both fields), 2^63 and 2^64 (n=1) -> Put 204 and Get shows 2147483647, no message; 1.5 -> 204 stored as 1; the
   sent value is never read back, so a spec above 2147483647 can never match its observation
   ([S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001), -002, [S3T-TABLEBUCKET-075](#s3t-tablebucket-075)).
8. Enum literals are exact and lowercase: status reads back as 'enabled'; Enabled/ENABLED/Disabled/DISABLED/on ->
   ValidationException 400 'Member must satisfy enum value set: [enabled, disabled]' (not case-folded, stored value
   unchanged); type IcebergUnreferencedFileRemoval, icebergCompaction, icebergSnapshotManagement, the all-caps form and
   'foo' -> ValidationException 400 'Member must satisfy enum value set: [icebergUnreferencedFileRemoval]', configuration
   unchanged ([S3T-TABLEBUCKET-074](#s3t-tablebucket-074), -077).
9. NotFoundException 404 'The specified bucket does not exist.' on Get or Put maintenance is an unambiguous bucket-gone
   signal: it appeared immediately after DeleteTableBucket returned 204 and on a never-existing ARN, while the live
   bucket's Get was 200 every time, fresh and under concurrent Puts ([S3T-TABLEBUCKET-079](#s3t-tablebucket-079)).
10. Metrics has no materialized default: GetTableBucketMetricsConfiguration before any Put -> NotFoundException 404 'No
    such metric configuration exists' (doc claim C059 TRUE); Put (input only the ARN) -> 204 and Get returns exactly
    {id, tableBucketARN} with id the constant literal 'EntireBucket' - identical across three Gets, a second Put (204, no
    ConflictException), Delete+Put and on a second bucket, so it is not a per-bucket identifier ([S3T-TABLEBUCKET-093](#s3t-tablebucket-093),
    -174, -175, -176).
11. DeleteTableBucketMetricsConfiguration is idempotent: 204 after a successful Delete and 204 on a never-configured
    bucket (doc claim C060 only PARTLY true - the Get half is 404 immediately, the "second Delete 404" half is false);
    a Put right after the Delete is 204 and the id is 'EntireBucket' again ([S3T-TABLEBUCKET-177](#s3t-tablebucket-177), -093, -095).
12. Only the canonical lowercase ARN resolves a sub-resource call: on the metrics Put an upper-case name, trailing slash,
    table ARN, empty region, service `s3`, another region, partition `aws-cn` and a bare name -> BadRequestException
    400 with four distinct messages ('bucket name isn't valid', 'table bucket ARN is not valid', 'Region does not match',
    'Partition does not match'), while an underscore name passes the parser and is looked up -> NotFoundException 404
    'The specified bucket does not exist.'; no variant created a configuration; the maintenance Get behaves the same for
    an upper-cased name or region ([S3T-TABLEBUCKET-100](#s3t-tablebucket-100), -057).
13. Concurrent writers to one bucket's maintenance configuration serialize and one in twenty failed: 20 Puts from 5
    threads -> 19 x 204 and 1 InternalServerErrorException 500 'An internal error occurred. Try again.' after 6168 ms,
    Put latencies 28-6459 ms (5 above 2 s), Gets during the burst all 200 but the first blocked ~4.1 s, no
    ConflictException on either side - treat the 500 on this Put as transient ([S3T-TABLEBUCKET-078](#s3t-tablebucket-078)).
14. Metrics reads throttle before writes do: 50 back-to-back calls in 1.31 s (10 x Put, Put, Get, Delete, Get) gave
    TooManyRequestsException 429 on the 20th Get within 1.29 s and the 3 Gets of the next 300 ms, while the Puts and the
    Delete of the same iteration stayed 204; a 429 can carry over into the first poll of a following step; no
    ConflictException and no stale read in the burst ([S3T-TABLEBUCKET-094](#s3t-tablebucket-094), -175, -095).

### Timing ranges

- Maintenance read-after-write: no wait exists - the Get immediately after each Put reflected it in 12/12 pairs
  ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054)), status flips were visible on the first Get in 4/4 and matched at +1 s, only the literals
  'enabled'/'disabled' ever appeared, no transitional value ([S3T-TABLEBUCKET-056](#s3t-tablebucket-056)).
- PutTableBucketMaintenanceConfiguration latency, single writer: 28-519 ms, n=12 ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054)); 308-339 ms, n=4
  ([S3T-TABLEBUCKET-056](#s3t-tablebucket-056)). Under 5-thread concurrency on one bucket: 28-6459 ms, n=20, with 15 calls <= 338 ms and 5 at
  2388-6459 ms, the one 500 at 6168 ms ([S3T-TABLEBUCKET-078](#s3t-tablebucket-078)).
- GetTableBucketMaintenanceConfiguration latency: 42-49 ms, n=2 (re-verified 20-81 ms, n=3) ([S3T-TABLEBUCKET-053](#s3t-tablebucket-053));
  25-314 ms, n=4 ([S3T-TABLEBUCKET-056](#s3t-tablebucket-056)). The ~4.1 s first read during the concurrent Put burst is a single value (n=1,
  [S3T-TABLEBUCKET-078](#s3t-tablebucket-078)) - a ceiling seen once, not a range.
- Server default visibility: enabled/3/10 at +0 s and +5 s after CreateTableBucket, identical bodies, n=2 reads on one
  bucket ([S3T-TABLEBUCKET-053](#s3t-tablebucket-053)) - there is no fill-in window to wait for.
- Metrics Put latency 44-70 ms, n=4; Get 12-13 ms, n=6 ([S3T-TABLEBUCKET-093](#s3t-tablebucket-093)); mixed Put/Get/Delete burst 3-50 ms per
  call, n=50 ([S3T-TABLEBUCKET-094](#s3t-tablebucket-094)).
- Metrics Delete -> Get 404: first poll at 0.003-0.013 s after the Delete's 204 was NotFoundException in 2 of 3
  repetitions (the third a 429 carried over from the preceding burst), and 0 of 57 polls at 200 ms over 3 x 4 s
  returned 200 ([S3T-TABLEBUCKET-095](#s3t-tablebucket-095)).
- Metrics read throttle: first 429 on the 20th Get within 1.29 s with 37 calls on the bucket in the preceding second (15
  Put, 7 Delete, 15 Get), 3 further 429 within 300 ms, the Get 0.6 s later succeeded - one burst, n=50 calls, so the
  limit is a single observation, not a range ([S3T-TABLEBUCKET-094](#s3t-tablebucket-094)).

### Handling gaps

- [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) (suspect-bug, not yet handled): today `maintenanceDaysToInt32` (pkg/resource/table_bucket/hooks.go:35-40)
  rejects only `v < 0 || v > MaxInt32`, so a spec `unreferencedDays`/`nonCurrentDays` of 0 passes the guard and the CRD
  (the finding's notes record no `Minimum` marker on the int64 fields), `syncMaintenance` forwards it into
  PutTableBucketMaintenanceConfiguration and returns the error raw (hooks.go:375-404), `terminal_codes` lists only
  BadRequestException (generator.yaml:267-268), and `maintenanceNeedsSync` (hooks.go:279-311) re-detects `0 != observed`
  every cycle. Evidence: 0 and -1 -> ValidationException 400 'Member must have value greater than or equal to 1' with the
  stored value unchanged (-075, [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002)) and ValidationException is the wire code
  for 11/11 validation rejections and absent from the model (-080), so the resource requeues forever on a permanent user
  error instead of going Terminal. The finding proposes raising the guard's lower bound to 1 (or a `Minimum=1` marker)
  and/or adding ValidationException to the TableBucket terminal codes; the 2147483648 clamp cell of -075 is unreachable
  from the controller (the guard is Terminal before any call).
- [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001) (partial, [GT-S3T-019](service.md#gt-s3t-019) (controller hooks catalog entry) scorer verdict "partial match", not yet fully
  handled): today the same guard turns any day count above 2147483647 into `ackerr.NewTerminalError` before the API is
  called (hooks.go:376-386), so the controller never sends the values that clamp. Evidence: 2147483648, 4294967296 and
  9223372036854775807 in either field -> Put 204, Get 2147483647, no message (n=3 per cell, 18 cells); the API gives no
  signal a reconciler could map, so the client-side Terminal is the only handling available; what remains open is the
  guard's lower bound of 0 (the -075 gap above) - the clamp is not part of the modeled range 1..2147483647.
- [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002) and [S3T-TABLEBUCKET-076](#s3t-tablebucket-076) (partial, not yet fully handled): -002's extra
  cells ran once each (n=1): 2^63 and 2^64 -> 204 clamped (stopped by the guard), 1.5 -> 204 stored as 1 (unreachable
  from an integer CRD field), 0/-1 -> 400 (the -075 gap). -076: the server takes settings {} as a 204 no-op and rejects
  only a body without `value` ('Member must not be null', ValidationException 400); `syncMaintenance` always sends a
  `value` (hooks.go:368-399), so a spec entry with neither status nor settings goes out as value {} - a silent no-op on
  the wire (-055, -076): nothing destroyed, nothing enforced.
- [S3T-TABLEBUCKET-080](#s3t-tablebucket-080) (unhandled, medium): ValidationException is absent from `terminal_codes` (generator.yaml:267-268);
  it is the root cause of the -075 loop and also the code for status/type typos (-074, -077), which `syncMaintenance`
  casts to the SDK enums without a client-side check (hooks.go:370, 397), so a misspelt literal reaches the API.
- Metrics configuration ([S3T-TABLEBUCKET-093](#s3t-tablebucket-093), -094, -095, -100, -177, unhandled): generator.yaml and hooks.go carry no
  MetricsConfiguration member or Put/Get/Delete call; these findings are scope input for a future field, not defects.

### Where next

- `tablebucket.md` carries the bucket lifecycle behind rule 9 (Create/Get/DeleteTableBucket, ARN construction, the
  identifier-form 404/400 split that -100 and -057 mirror, the ~10 min post-delete name reservation, throttling, limits).
- `tablebucket-policy.md` owns the authorization side of both sub-resources: the metrics triple from a foreign account id
  (ForbiddenException 'Unauthorized', [S3T-TABLEBUCKET-099](#s3t-tablebucket-099)) and without any CloudWatch permission or service-linked role
  ([S3T-TABLEBUCKET-113](#s3t-tablebucket-113)) render there, as does the resource-policy Deny whose probe target was
  GetTableBucketMaintenanceConfiguration ([S3T-TABLEBUCKETPOLICY-005](tablebucket-policy.md#s3t-tablebucketpolicy-005)).
- `table-maintenance.md` is the table-level twin (merge-patch Put with its own defaults, `modifiedAt` bump on every Put);
  -080 records that the bucket-only type sent to the table Put is the one place BadRequestException 400 ('Invalid table
  maintenance type: icebergUnreferencedFileRemoval') appeared instead of ValidationException.
  `tablebucket-encryption-storage.md` and `tablebucket-replication.md` cover the other bucket sub-resource pairs
  (encryption/storage class have materialized defaults like maintenance; metrics does not, [S3T-TABLEBUCKET-176](#s3t-tablebucket-176) via -097).
- Long catalogues belong in `details/<id>.md` rather than inline: the 12-step Put/Get merge sequence ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054)),
  the 10 day-count cells (-075) and the 18 + 12 overflow cells ([S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001), -002), the
  five status variants with the full enum message (-074) and five type variants (-077), the nine ARN variants with their
  four messages (-100), and the 20 sorted concurrent-Put latencies (-078).
- Open experiments: STATUS.md ("Account state", "Phase 8 result") lists none for this stem; -053 was independently
  re-verified in the Phase 8 sample (tag `reverified`). Single-run cells worth a future probe are listed in
  `tablebucket-maintenance-metrics.open-questions.md`: the -078 burst (n=20, one run), the -094 throttle burst and its 429
  carry-over into -095, the n=1 edge cells of -002, and the 'missing value' cell boto3 blocks client-side (-055, -076).
<!-- preserved:end -->

## At a glance

- canonical findings: 28 (high 2 / medium 16 / low 10); duplicates folded into the appendix: 0
- handling: handled 8 · partial 3 · tracked 0 · unhandled 16 · suspect-bug 1 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 1 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: other 6, request-validation 6, error-code 4, identity 2, server-default 2, unsettable-field 2,
  eventual-consistency 1, idempotency 1, prerequisite 1, quota-limit 1, shape-mismatch 1, update-granularity 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| DeleteTableBucketMetricsConfiguration | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketMaintenanceConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketMetricsConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketMaintenanceConfiguration | create | tableBucketARN, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketMetricsConfiguration | create | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableMaintenanceConfiguration | create | tableBucketARN, namespace, name, type, value | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **MaintenanceStatus**: enabled, disabled
- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-tablebucket-093"></a>**S3T-TABLEBUCKET-093** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketMetricsConfiguration.id is the constant string 'EntireBucket' across repeated Puts, Delete+Put and on a second bucket**
  PutTableBucketMetricsConfiguration (no body) -> HTTP 204; three Gets returned id ['EntireBucket',
  'EntireBucket', 'EntireBucket'] (format len12:EaaaaaBaaaaa). Second Put -> HTTP 204; Get id -> EntireBucket.
  Delete -> HTTP 204; Get -> NotFoundException/404; Put immediately after Delete -> HTTP 204; Get id ->
  EntireBucket. On a second bucket Put -> HTTP 204, Get id -> EntireBucket. Delete -> HTTP 204, Delete again
  -> HTTP 204. Raw Get JSON keys: ['id', 'tableBucketARN']; tableBucketARN equals the request ARN and
  GetTableBucket.arn: True.
  - ACK: is_read_only, compare.is_ignored+delta_pre_compare · ops: PutTableBucketMetricsConfiguration,
    GetTableBucketMetricsConfiguration, DeleteTableBucketMetricsConfiguration · fields: id, tableBucketARN
  - repro: Put -> Get x3 -> Put -> Get -> Delete -> Get -> Put -> Get; Put+Get on a second bucket; Delete
    twice
  - measurements: put_latency_ms=[45, 44, 48, 70], get_latency_ms=[12, 13, 12, 12, 13, 12], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](#s3t-tablebucket-099), [S3T-TABLEBUCKET-113](#s3t-tablebucket-113),
    [S3T-TABLEBUCKET-082](tablebucket-encryption-storage.md#s3t-tablebucket-082), [S3T-TABLEBUCKET-174](#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-177](#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096) · hypotheses: H-S-044, H-S-092, H-S-097,
    H-S-101 · evidence: tablebucket/sub-resources/metrics-lifecycle
  - notes: H-S-044 confirmed (constant, 'EntireBucket'-style literal; Put returns HTTP 204, not 200). H-S-092
    refuted: the id is unchanged after Delete -> Put. H-S-097 qualified: id is present and non-empty after
    every Put and cannot be set, but it is NOT unique per bucket (identical 'EntireBucket' on both...
  - full notes: [details/S3T-TABLEBUCKET-093.md](details/S3T-TABLEBUCKET-093.md)

- <a id="s3t-tablebucket-100"></a>**S3T-TABLEBUCKET-100** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics Put ARN variants: upper-case, trailing slash, table ARN, empty region, s3, other region, aws-cn, bare name -> 400; underscore -> 404**
  On an existing bucket with one table: uppercase-name
  (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ACKQ-28A1D7-EARN) -> BadRequestException (HTTP 400) 'The specified
  bucket name isn't valid. Specify a different bucket name, and then try again.'; underscore-name
  (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq_28a1d7_earn) -> NotFoundException (HTTP 404) 'The specified
  bucket does not exist.'; trailing-slash (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn/) ->
  BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; table-arn (aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn/) ->
  BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and
  then try again.'; empty-region (aws:s3tables::<ACCOUNT>:bucket/ackq-28a1d7-earn) -> BadRequestException
  (HTTP 400) 'The specified table bucket ARN is not valid.'; service-s3
  (aws:s3:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-earn) -> BadRequestException (HTTP 400) 'The specified table
  bucket ARN is not valid.'; other-region (aws:s3tables:us-east-1:<ACCOUNT>:bucket/ackq-28a1d7-earn) ->
  BadRequestException (HTTP 400) 'The provided Region does not match the Region of the resource.';
  partition-aws-cn (aws-cn:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-28a1d7-ea) -> BadRequestException (HTTP 400)
  'The provided Partition does not match the Partition of the resource.'; name-only (ackq-28a1d7-earn) ->
  BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'.
  GetTableBucketMetricsConfiguration with the canonical ARN afterwards -> NotFoundException (HTTP 404) 'No
  such metric configuration exists' (no configuration was created by any variant).
  - ACK: is_arn_primary_key, terminal_codes · ops: PutTableBucketMetricsConfiguration,
    GetTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: Put metrics with each ARN variant (upper-case name, underscore name, trailing slash, table ARN,
    empty region, service s3, other region, aws-cn partition, bare name) -> canonical Get
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](#s3t-tablebucket-093),
    [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](#s3t-tablebucket-099), [S3T-TABLEBUCKET-113](#s3t-tablebucket-113),
    [S3T-TABLEBUCKET-126](tablebucket-encryption-storage.md#s3t-tablebucket-126), [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLE-143](table.md#s3t-table-143), [S3T-TABLE-267](table.md#s3t-table-267), [S3T-TABLE-027](table.md#s3t-table-027) ·
    hypotheses: H-S-091, H-S-096 · evidence: tablebucket/error-taxonomy/subresource-arn-variants
  - notes: H-S-091 qualified: upper-case name and trailing slash -> BadRequestException 400 'The specified
    bucket name isn't valid...' (confirmed, no lower-casing), but an underscore name passes the ARN parser and
    is looked up: NotFoundException 404 'The specified bucket does not exist.' (refuted for that...
  - full notes: [details/S3T-TABLEBUCKET-100.md](details/S3T-TABLEBUCKET-100.md)

## Idempotency

- <a id="s3t-tablebucket-177"></a>**S3T-TABLEBUCKET-177** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C060 PARTLY: After DeleteTableBucketMetricsConfiguration Get -> 404, but a second Delete -> HTTP 204 (idempotent), not 404**
  Delete -> HTTP 204; Get -> NotFoundException/404 'No such metric configuration exists'; Delete again -> HTTP
  204 ([S3T-TABLEBUCKET-093](#s3t-tablebucket-093)); a second Delete right after a successful one -> HTTP 204 and Delete on a
  never-configured bucket -> HTTP 204 ([S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096)/092). Get polled at 200 ms after Delete: 404 on the
  first poll in 3/3 reps (one first poll was a 429 throttle), no 200 ever returned ([S3T-TABLEBUCKET-095](#s3t-tablebucket-095)).
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketMetricsConfiguration,
    GetTableBucketMetricsConfiguration · fields: id
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095),
    [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092); no new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026),
    [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034) · evidence: service/static/doc-claims-2
  - notes: VERDICT: PARTLY - the Get half is true (404 immediately); the Delete half is false -
    DeleteTableBucketMetricsConfiguration is idempotent and returns 204 when nothing exists.

## Errors

- <a id="s3t-tablebucket-079"></a>**S3T-TABLEBUCKET-079** `error-code` · impact medium · handled · verified 2026-10-13
  **Get/Put maintenance configuration on a just-deleted or never-existing bucket -> NotFoundException 404 'The specified bucket does not exist.'**
  Immediately after DeleteTableBucket returned 204 (bucket emptied first):
  GetTableBucketMaintenanceConfiguration -> NotFoundException (HTTP 404) 'The specified bucket does not
  exist.'; PutTableBucketMaintenanceConfiguration {enabled, 3/10} -> NotFoundException (HTTP 404) 'The
  specified bucket does not exist.'. Same Get on an ARN whose bucket never existed -> NotFoundException (HTTP 404)
  'The specified bucket does not exist.'. On the live bucket this Get always returned HTTP 200 (fresh and
  under concurrent Puts).
  - ACK: exceptions.404 · ops: GetTableBucketMaintenanceConfiguration, PutTableBucketMaintenanceConfiguration
    · fields: tableBucketARN
  - repro: DeleteTable -> DeleteNamespace -> DeleteTableBucket -> Get/Put maintenance on the same ARN; Get on
    a never-existing ARN
  - handling: handled via `generator.yaml:243-249; pkg/resource/table_bucket/hooks.go:268-278`
  - related: [S3T-TABLEBUCKET-020](tablebucket-encryption-storage.md#s3t-tablebucket-020), [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158) · hypotheses: H-S-107 · evidence:
    tablebucket/weird-inputs/maintenance-validation
  - notes: H-S-107 part 1 (404 on this Get is an unambiguous bucket-gone signal because a live bucket never
    404s) confirmed.
  - full notes: [details/S3T-TABLEBUCKET-079.md](details/S3T-TABLEBUCKET-079.md)

- <a id="s3t-tablebucket-080"></a>**S3T-TABLEBUCKET-080** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration input validation returns wire code ValidationException (HTTP 400)...**
  Every server-side input validation failure seen on PutTableBucketMaintenanceConfiguration (status not in
  [enabled, disabled], type not icebergUnreferencedFileRemoval, unreferencedDays/nonCurrentDays < 1, missing
  value) came back as code 'ValidationException', HTTP 400, message '1 validation error detected: Value at
  '<path>' failed to satisfy constraint: ...'. The SDK model lists BadRequestException for this operation and
  does not model ValidationException; BadRequestException (HTTP 400) was seen only for an invalid bucket name
  / region in the ARN and for the table-level Put of the bucket-only type ('Invalid table maintenance type:
  icebergUnreferencedFileRemoval').
  - ACK: terminal_codes · ops: PutTableBucketMaintenanceConfiguration, PutTableMaintenanceConfiguration ·
    fields: value.status, type, value.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: Put with status 'Enabled' / type 'foo' / unreferencedDays 0 (parameter_validation=False) and
    inspect the error code
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-074](#s3t-tablebucket-074), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](#s3t-tablebucket-077) · hypotheses:
    H-S-106, H-S-029, H-S-032, H-S-030 · evidence: tablebucket/weird-inputs/maintenance-validation,
    tablebucket/sub-resources/maintenance-defaults-fill
  - notes: Qualifies the '400 BadRequestException' expectation in H-S-106, H-S-029, H-S-032 and H-S-030: the
    HTTP status matches but the code string does not. A terminal-code list keyed on BadRequestException would
    miss these user errors; match on ValidationException too (11 of 11 validation rejections in this...
  - full notes: [details/S3T-TABLEBUCKET-080.md](details/S3T-TABLEBUCKET-080.md)

- <a id="s3t-tablebucket-099"></a>**S3T-TABLEBUCKET-099** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics Put/Get/Delete with a foreign account id in the bucket ARN -> ['ForbiddenException/403']...**
  <ACCOUNT>/put -> ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/get -> ForbiddenException (HTTP 403)
  'Unauthorized'; <ACCOUNT>/delete -> ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/put ->
  ForbiddenException (HTTP 403) 'Unauthorized'; <ACCOUNT>/get -> ForbiddenException (HTTP 403) 'Unauthorized';
  <ACCOUNT>/delete -> ForbiddenException (HTTP 403) 'Unauthorized'. Contrast:
  PutTableBucketMetricsConfiguration on the caller's account with a never-existing bucket name ->
  NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: terminal_codes · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: Put/Get/Delete metrics with arn:aws:s3tables:<region>:<ACCOUNT>:bucket/<existing-name> and
    ...:<ACCOUNT>:...; Put on own-account never-existing name
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](#s3t-tablebucket-113) · hypotheses: H-S-090 · evidence: tablebucket/error-taxonomy/subresource-arn-variants
  - notes: H-S-090 confirmed: all six calls (two foreign account ids x Put/Get/Delete) -> ForbiddenException
    (HTTP 403) 'Unauthorized'; the own-account never-existing name -> NotFoundException 404. Whether a bucket
    of that name exists in the foreign accounts is unknown.
  - full notes: [details/S3T-TABLEBUCKET-099.md](details/S3T-TABLEBUCKET-099.md)

## Request validation

- <a id="s3t-tablebucket-074"></a>**S3T-TABLEBUCKET-074** `request-validation` · impact medium · handled · verified 2026-10-13
  **MaintenanceStatus reads back as lowercase 'enabled'; Put with Enabled/ENABLED/Disabled/DISABLED/on -> ValidationException (HTTP 400)...**
  After Put {status: enabled, 3/2} (HTTP 204) the raw Get JSON carries status 'enabled'. Put with the same
  settings and status 'Enabled' -> ValidationException (HTTP 400) '1 validation error detected: Value at
  'value.status' failed to satisfy constraint: Member must satisfy enum value set: [enabled, disabled]';
  'ENABLED' -> ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to
  satisfy constraint: Member must satisfy enum value set: [enabled, disabled]'; 'Disabled' ->
  ValidationException (HTTP 400) '1 validation error detected: Value at 'value.status' failed to satisfy
  constraint: Member must satisfy enum value set: [enabled, disabled]'; 'DISABLED' -> ValidationException
  (HTTP 400) '1 validation error detected: Value at 'value.status' failed to satisfy constraint: Member must
  satisfy enum value set: [enabled, disabled]'; 'on' -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'value.status' failed to satisfy constraint: Member must satisfy enum value set:
  [enabled, disabled]'. Get afterwards -> enabled/3/2 (unchanged: True).
  - ACK: terminal_codes, none · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields: value.status, configuration{}.status
  - repro: Put {enabled, 3/2} -> Get -> Put with status Enabled / ENABLED / Disabled / DISABLED / on -> Get
  - handling: handled via `generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](#s3t-tablebucket-077), [S3T-TABLEBUCKET-080](#s3t-tablebucket-080) · hypotheses:
    H-S-106 · evidence: tablebucket/weird-inputs/maintenance-validation
  - notes: H-S-106 qualified: lowercase literal on read and case variants are rejected with HTTP 400 as
    hypothesized, but the wire code is ValidationException ('1 validation error detected: Value at
    'value.status' failed to satisfy constraint: Member must satisfy enum value set: [enabled, disabled]'),
    not the...
  - full notes: [details/S3T-TABLEBUCKET-074.md](details/S3T-TABLEBUCKET-074.md)

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
  - related: [S3T-TABLEBUCKET-074](#s3t-tablebucket-074), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-077](#s3t-tablebucket-077), [S3T-TABLEBUCKET-080](#s3t-tablebucket-080),
    [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001), [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002),
    [S3T-TABLEBUCKET-170](#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055),
    [S3T-TABLEBUCKET-053](#s3t-tablebucket-053) · hypotheses: H-S-029, H-S-111 · evidence:
    tablebucket/weird-inputs/maintenance-validation
  - notes: H-S-029 qualified: minimum 1 enforced with HTTP 400 naming the field and no upper bound or ordering
    constraint (as hypothesized), but the wire code is ValidationException, not BadRequestException. H-S-111
    confirmed: 2147483647 round-trips exactly, 36501 accepted. Surprise: 2147483648 is accepted and...
  - full notes: [details/S3T-TABLEBUCKET-075.md](details/S3T-TABLEBUCKET-075.md)

- <a id="s3t-tablebucket-077"></a>**S3T-TABLEBUCKET-077** `request-validation` · impact medium · handled · verified 2026-10-13
  **Maintenance type variants (IcebergUnreferencedFileRemoval, icebergCompaction, icebergSnapshotManagement, ALL-CAPS, foo) -> Validation 400**
  type 'IcebergUnreferencedFileRemoval' -> ValidationException (HTTP 400) '1 validation error detected: Value
  at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'icebergCompaction' -> ValidationException (HTTP 400) '1 validation
  error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'icebergSnapshotManagement' -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'; type 'ICEBERGUNREFERENCEDFILEREMOVAL' -> ValidationException (HTTP 400)
  '1 validation error detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value
  set: [icebergUnreferencedFileRemoval]'; type 'foo' -> ValidationException (HTTP 400) '1 validation error
  detected: Value at 'type' failed to satisfy constraint: Member must satisfy enum value set:
  [icebergUnreferencedFileRemoval]'. Get before and after -> enabled/8/8 / enabled/8/8 (stored configuration
  unchanged).
  - ACK: terminal_codes, none · ops: PutTableBucketMaintenanceConfiguration · fields: type
  - repro: Put value {status: enabled} with each type string above -> Get
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275`
  - related: [S3T-TABLEBUCKET-074](#s3t-tablebucket-074), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-080](#s3t-tablebucket-080) · hypotheses:
    H-S-032 · evidence: tablebucket/weird-inputs/maintenance-validation
  - notes: H-S-032 qualified: all four non-bucket type strings (plus the all-caps variant) are rejected with
    HTTP 400 and the enum message 'Value at 'type' failed to satisfy constraint: Member must satisfy enum
    value set: [icebergUnreferencedFileRemoval]', not 404; the wire code is ValidationException rather...
  - full notes: [details/S3T-TABLEBUCKET-077.md](details/S3T-TABLEBUCKET-077.md)

- <a id="s3t-tablebucketmaintenanceconfiguration-001"></a>**S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001** `request-validation` · impact medium · partially handled · verified 2026-10-14
  **PutTableBucketMaintenanceConfiguration silently clamps nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204**
  Client validation off; baseline unreferencedDays=7/nonCurrentDays=11 re-Put before every attempt.
  nonCurrentDays (unreferencedDays=1): 2147483648 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3);
  4294967296 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3); 9223372036854775807 -> Put HTTP 204, Get
  shows clamped-to-2147483647 (n=3). unreferencedDays (nonCurrentDays=1): 2147483648 -> Put HTTP 204, Get
  shows clamped-to-2147483647 (n=3); 4294967296 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3);
  9223372036854775807 -> Put HTTP 204, Get shows clamped-to-2147483647 (n=3). Messages: none. Wire body of one
  attempt: {"value": {"status": "enabled", "settings": {"icebergUnreferencedFileRemoval": {"nonCurrentDays":
  9223372036854775807, "unreferencedDays": 1}}}}.
  - ACK: compare.is_ignored+delta_pre_compare, terminal_codes · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields:
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays,
    value.settings.icebergUnreferencedFileRemoval.unreferencedDays
  - repro: PutTableBucketMaintenanceConfiguration(type=icebergUnreferencedFileRemoval,
    value.settings.%s=2147483648|4294967296|9223372036854775807, other=1) with parameter validation disabled
    -> GetTableBucketMaintenanceConfiguration; x3 per cell
  - measurements: n_per_cell=3, cells=18, values=[2147483648, 4294967296, 9223372036854775807], n=3
  - handling: partially handled via `pkg/resource/table_bucket/hooks.go:31-40; pkg/resource/table_bucket/hooks.go:376-388` - see Handling gaps
  - related: [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002) · evidence:
    tablebucket/weird-inputs/fu-04-maintenance-int-overflow
  - notes: FU-04. The modeled range is 1..2147483647 (int32); a Go int64 spec field would let a user send
    these values and the controller must map the wire code to a terminal condition. See extra cells (0, -1,
    2^63, 2^64, 1.5) in the companion finding.
  - full notes: [details/S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001.md](details/S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001.md)

## Update granularity and ordering

- <a id="s3t-tablebucket-113"></a>**S3T-TABLEBUCKET-113** `prerequisite` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Metrics configuration Put/Get/Delete as a role with s3tables:* Allow and cloudwatch:* Deny -> HTTP 204 / HTTP 200 / HTTP 204...**
  A fresh IAM role with an inline policy {Allow s3tables:* on *, Deny cloudwatch:* on *} (its own
  cloudwatch:ListMetrics -> AccessDenied 403 '... with an explicit deny in an identity-based policy'):
  PutTableBucketMetricsConfiguration -> HTTP 204 (attempts while IAM propagated: ['HTTP 204']),
  GetTableBucketMetricsConfiguration -> HTTP 200 {'tableBucketARN':
  'arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-5b1f89-cwd', 'id': 'EntireBucket'},
  DeleteTableBucketMetricsConfiguration -> HTTP 204, Get after Delete -> NotFoundException/404. iam:ListRoles
  PathPrefix=/aws-service-role/ before (14 roles) and 5 s after enabling metrics: no new role.
  - ACK: docs-only · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: CreateRole (trust: account root; inline Allow s3tables:*, Deny cloudwatch:*) -> AssumeRole ->
    Put/Get/Delete metrics -> ListRoles /aws-service-role/ diff
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](#s3t-tablebucket-099),
    [S3T-TABLEBUCKET-100](#s3t-tablebucket-100) · hypotheses: H-S-095 · evidence: tablebucket/dependencies/iam-cloudwatch-deny-metrics
  - notes: H-S-095 confirmed: no CloudWatch permission and no service-linked role is needed to enable bucket
    metrics. Single role, single bucket.
  - full notes: [details/S3T-TABLEBUCKET-113.md](details/S3T-TABLEBUCKET-113.md)

- <a id="s3t-tablebucket-172"></a>**S3T-TABLEBUCKET-172** `update-granularity` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C050 FALSE: Put {status: enabled} without settings after 5/15 keeps the stored days (field merge), not the service defaults**
  After enabled 7/20: Put {status: disabled} (no settings) -> HTTP 204, Get disabled/7/20; Put {status:
  enabled} (no settings) -> enabled/7/20; Put {enabled, unreferencedDays 5 only} -> enabled/5/20; {enabled,
  nonCurrentDays 20 only} -> enabled/5/20 ([S3T-TABLEBUCKET-054](#s3t-tablebucket-054)). value {} -> HTTP 204 and a no-op (Get keeps
  enabled/9/9) ([S3T-TABLEBUCKET-055](#s3t-tablebucket-055)). The defaults (3/10) never reappear once days were set.
  - ACK: custom_update, compare.is_ignored+delta_pre_compare · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields: value.status, value.settings
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053); no
    new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-170](#s3t-tablebucket-170),
    [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) · evidence: service/static/doc-claims-2
  - notes: VERDICT: FALSE - PutTableBucketMaintenanceConfiguration merges field by field; omitted settings are
    kept, so there is no reset-to-default via omission.

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-tablebucket-053"></a>**S3T-TABLEBUCKET-053** `server-default` · impact high · handled · verified 2026-10-15, re-verified
  **GetTableBucketMaintenanceConfiguration on a fresh bucket -> 200 enabled/3/10; Put {status: enabled} w/o settings -> 204, still enabled/3/10**
  A table bucket created with name+tags only, never Put: GetTableBucketMaintenanceConfiguration returned HTTP
  200 at +0 s and +5 s (bodies identical); raw JSON top-level keys ['configuration', 'tableBucketARN'],
  configuration map keys ['icebergUnreferencedFileRemoval'], value enabled/3/10; tableBucketARN echoed
  byte-identical: True. PutTableBucketMaintenanceConfiguration(type=icebergUnreferencedFileRemoval,
  value={status: enabled}) (no settings) -> HTTP 204; Get afterwards -> enabled/3/10.
  - ACK: late_initialize, compare.is_ignored+delta_pre_compare · ops: GetTableBucketMaintenanceConfiguration,
    PutTableBucketMaintenanceConfiguration · fields: configuration, configuration{}.status,
    configuration{}.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    configuration{}.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: CreateTableBucket(name, tags) -> GetTableBucketMaintenanceConfiguration at +0 s and +5 s -> Put
    {status: enabled} -> Get
  - measurements: get_latency_ms=[49, 42], n=2
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275; generator.yaml:243-249; pkg/resource/table_bucket/hooks.go:268-278; pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057), [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-173](#s3t-tablebucket-173),
    [S3T-TABLEBUCKET-059](tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097) · hypotheses: H-S-025, H-S-026 · evidence:
    tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-025 confirmed (fresh read is enabled/3/10); H-S-026 refuted (the contrarian 'empty map until
    first Put' did not hold). Single bucket, two reads.
  - full notes: [details/S3T-TABLEBUCKET-053.md](details/S3T-TABLEBUCKET-053.md)

- <a id="s3t-tablebucket-054"></a>**S3T-TABLEBUCKET-054** `server-default` · impact high · handled · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration merges into the stored value...** (hypothesis refuted; behavior confirmed)
  All 12 Puts returned HTTP 204. Sequence (sent -> Get status/unreferencedDays/nonCurrentDays): enabled 7/20
  -> enabled/7/20; {status: disabled} (no settings) -> disabled/7/20; {status: enabled} (no settings) ->
  enabled/7/20; enabled 5/20 -> enabled/5/20; {enabled, unreferencedDays: 5 only} -> enabled/5/20; {enabled,
  nonCurrentDays: 20 only} -> enabled/5/20; enabled 7/4 -> enabled/7/4; disabled 7/4 -> disabled/7/4; enabled
  7/4 -> enabled/7/4; {status: disabled} -> disabled/7/4; {status: disabled} -> disabled/7/4; {settings 5/15,
  no status} -> disabled/5/15.
  - ACK: compare.nil_equals_zero_value, late_initialize, custom_update · ops:
    PutTableBucketMaintenanceConfiguration, GetTableBucketMaintenanceConfiguration · fields: value.status,
    value.settings.icebergUnreferencedFileRemoval.unreferencedDays,
    value.settings.icebergUnreferencedFileRemoval.nonCurrentDays
  - repro: Fresh bucket -> the 12 Put/Get pairs above in order (each Get immediately after its Put)
  - measurements: put_latency_ms_min_max=[28, 519], n=12
  - handling: handled via `generator.yaml:239-262; generator.yaml:269-275; pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; generator.yaml:250-257; generator.yaml:278-283`
  - related: [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) · hypotheses: H-S-027,
    H-S-028, H-S-109, H-S-030 · evidence: tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-027 refuted: {status: disabled} without settings keeps the stored 7/20 (no reset to 3/10);
    {status: enabled} without settings keeps 7/20 too. H-S-028 refuted: Put is a merge-patch, not a
    whole-value replace: {enabled, unreferencedDays: 5} kept nonCurrentDays 20 and {enabled, nonCurrentDays:
    20}...
  - full notes: [details/S3T-TABLEBUCKET-054.md](details/S3T-TABLEBUCKET-054.md)

- <a id="s3t-tablebucket-055"></a>**S3T-TABLEBUCKET-055** `unsettable-field` · impact medium · handled · verified 2026-10-13
  **PutTableBucketMaintenanceConfiguration value {} -> HTTP 204 but a no-op (Get keeps enabled/9/9)...** (hypothesis refuted; behavior confirmed)
  After Put enabled 9/9: Put value={} -> HTTP 204; Get -> enabled/9/9 (fresh-bucket body was enabled/3/10, so
  {} does NOT reset to the fresh default). Put {status: enabled, settings: {}} (empty union) ->
  ParamValidationError (HTTP None) 'Parameter validation failed:
  Must set one of the following keys for tagged unionstructure value.settings:
  icebergUnreferencedFileRemoval.'; Get -> enabled/9/9. Put {status: enabled, settings:
  {icebergUnreferencedFileRemoval: {}}} -> HTTP 204; Get -> enabled/9/9. Put with no value parameter at all ->
  ParamValidationError (client-side botocore validation).
  - ACK: custom_update, compare.nil_equals_zero_value · ops: PutTableBucketMaintenanceConfiguration · fields:
    value, value.settings, value.settings.icebergUnreferencedFileRemoval
  - repro: Put enabled 9/9 -> Put value {} -> Get -> Put {enabled, settings {}} -> Get -> Put {enabled,
    settings {icebergUnreferencedFileRemoval: {}}} -> Get -> Put without value
  - handling: handled via `pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249; pkg/resource/table_bucket/hooks.go:173-195; pkg/resource/table_bucket/hooks.go:389-391`
  - related: [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056),
    [S3T-TABLEBUCKET-170](#s3t-tablebucket-170), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) · hypotheses: H-S-112,
    H-S-030 · evidence: tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-112 refuted: value {} is accepted (204) but does NOT reset to the fresh default enabled/3/10;
    the stored enabled/9/9 is unchanged, so {} is not a 'reset'. settings {} (empty union) is rejected by
    botocore's client-side tagged-union validation (ParamValidationError, no request sent) - the server...
  - full notes: [details/S3T-TABLEBUCKET-055.md](details/S3T-TABLEBUCKET-055.md)

- <a id="s3t-tablebucket-056"></a>**S3T-TABLEBUCKET-056** `unsettable-field` · impact medium · handled · verified 2026-10-13
  **No DeleteTableBucketMaintenanceConfiguration in the SDK; status enabled<->disabled flips read back synchronously on the first Get (n=4)**
  The s3tables SDK model exposes no Delete operation for the bucket maintenance configuration (checked
  dir(client) for delete*maintenance*: none). Four Puts alternating status disabled/enabled (settings 4/10):
  every Put -> ['HTTP 204']; the Get issued immediately after each Put returned the new status in all 4 cases
  (get0 values ['disabled', 'enabled', 'disabled', 'enabled']), and the +1 s re-read matched (['disabled',
  'enabled', 'disabled', 'enabled']). Only the literals 'enabled'/'disabled' were observed, no transitional
  value.
  - ACK: custom_delete, scope:field-on-parent · ops: PutTableBucketMaintenanceConfiguration,
    GetTableBucketMaintenanceConfiguration · fields: value.status
  - repro: Put {disabled, 4/10} -> Get at +0/+1 s -> Put {enabled, 4/10} -> Get at +0/+1 s (x2)
  - measurements: put_latency_ms=[335, 324, 339, 308], get_latency_ms=[37, 25, 314, 29], n=4
  - handling: handled via `pkg/resource/table_bucket/hooks.go:345-363; generator.yaml:249`
  - related: [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-076](#s3t-tablebucket-076),
    [S3T-TABLEBUCKET-059](tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-058](tablebucket-encryption-storage.md#s3t-tablebucket-058), [S3T-TABLEBUCKET-112](tablebucket-encryption-storage.md#s3t-tablebucket-112), [S3T-TABLEBUCKET-173](#s3t-tablebucket-173), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171),
    [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057) · hypotheses: H-S-031 · evidence:
    tablebucket/sub-resources/maintenance-defaults-fill
  - notes: H-S-031 confirmed. 'Spec field removed' has no Delete counterpart; a controller must Put an
    explicit value (the fresh-bucket default is enabled/3/10).

## Response fidelity and consistency

- <a id="s3t-tablebucket-095"></a>**S3T-TABLEBUCKET-095** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After DeleteTableBucketMetricsConfiguration (204) the first Get was 404 (2 of 3 reps; 3rd a 429 carry-over), no 200 in 57 polls over 3x4 s**
  Put -> Delete -> poll Get every 200 ms for 4 s, repeated 3 times: first poll after the Delete at [0.003,
  0.013, 0.013] s returned ['TooManyRequestsException/429', 'NotFoundException/404', 'NotFoundException/404'];
  codes seen over all polls ['NotFoundException/404', 'TooManyRequestsException/429']; reads returning 200
  after the Delete: [[], [], []].
  - ACK: requeue · ops: DeleteTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration · fields:
    tableBucketARN
  - repro: Put metrics -> Delete -> Get every 200 ms for 4 s (x3)
  - measurements: first_poll_offset_s=[0.003, 0.013, 0.013], n=57
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-099](#s3t-tablebucket-099), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](#s3t-tablebucket-113), [S3T-TABLEBUCKET-177](#s3t-tablebucket-177), [S3T-TABLEBUCKET-096](tablebucket-replication.md#s3t-tablebucket-096), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092) · hypotheses: H-S-104 ·
    evidence: tablebucket/sub-resources/metrics-lifecycle
  - notes: H-S-104 confirmed for the metrics noun: the first Get at +3-13 ms after the Delete's 204 was
    NotFoundException (the one exception was a TooManyRequestsException carried over from the preceding burst,
    not a stale 200) and none of 57 polls over 3 x 4 s returned the deleted configuration. Replication...
  - full notes: [details/S3T-TABLEBUCKET-095.md](details/S3T-TABLEBUCKET-095.md)

## Quotas and rate limits

- <a id="s3t-tablebucket-094"></a>**S3T-TABLEBUCKET-094** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **10x metrics Put,Put,Get,Delete,Get without sleep: no ConflictException, no stale read; Gets throttled 429 from the 20th Get in 1.3 s**
  Ten iterations of PutTableBucketMetricsConfiguration x2, Get, Delete, Get back-to-back on one bucket (50
  calls in 1.31 s): iterations 1-9 returned {put 204, put 204, get 200, delete 204, get
  NotFoundException/404}; in iteration 10 both Gets returned TooManyRequestsException (HTTP 429, 3 ms) while
  the Puts and the Delete in the same iteration still returned 204. The first 429 came on the 20th
  GetTableBucketMetricsConfiguration within 1.29 s (37 calls on the bucket in the preceding second: 15 Put, 7
  Delete, 15 Get); the following 3 Gets within 300 ms were also 429, the Get 0.6 s later succeeded. No
  ConflictException, no stale read.
  - ACK: requeue · ops: PutTableBucketMetricsConfiguration, GetTableBucketMetricsConfiguration,
    DeleteTableBucketMetricsConfiguration · fields: tableBucketARN
  - repro: loop 10x without sleep: Put, Put, Get, Delete, Get
  - measurements: call_latency_ms_min_max=[3, 50], n=50
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-095](#s3t-tablebucket-095), [S3T-TABLEBUCKET-099](#s3t-tablebucket-099), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100),
    [S3T-TABLEBUCKET-113](#s3t-tablebucket-113), [S3T-TABLEBUCKET-165](tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](#s3t-tablebucket-175), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018),
    [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-174](#s3t-tablebucket-174), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092),
    [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097) · hypotheses: H-S-045 · evidence: tablebucket/sub-resources/metrics-lifecycle
  - notes: H-S-045 confirmed for the toggle semantics (double Put ok, Delete then immediate Get 404 on the
    first read, Put right after Delete ok). Surprise: only the Get was throttled (the shard brief's documented
    10/s sub-resource Get limit is consistent with ~15 Gets/s); the throttle message was 'Too Many...
  - full notes: [details/S3T-TABLEBUCKET-094.md](details/S3T-TABLEBUCKET-094.md)

## Handling gaps (bugs to file)

- [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) - Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and
  2147483647 round-trip; 2147483648 -> 204, read back clamped (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven.
  unreferencedDays/nonCurrentDays = 0 passes both the CRD (no Minimum marker) and the hand-written int32 guard
  (lower bound 0), reaches PutTableBucketMaintenanceConfiguration, comes back as ValidationException 400 which
  is not a TableBucket terminal code, and the resource requeues forever - raise the guard's lower bound to 1
  (or add a kubebuilder Minimum=1 marker) and/or add ValidationException to the TableBucket terminal codes.
  code: config/crd/bases/s3tables.services.k8s.aws_tablebuckets.yaml:83-88 nonCurrentDays/unreferencedDays are
  bare 'format: int64, type: integer' (apis/v1alpha1/types.go:106-107 has no Minimum marker);
  pkg/resource/table_bucket/hooks.go:35-40 maintenanceDaysToInt32 rejects only v < 0 || v > MaxInt32, so 0 is
  forwarded by hooks.go:375-388 into the Put at hooks.go:393-404 and the error is returned raw;
  pkg/resource/table_bucket/sdk.go:400-405 terminalAWSError matches only BadRequestException, so the
  ValidationException is Recoverable and requeued; hooks.go:279-311 maintenanceNeedsSync keeps the 0 !=
  observed delta alive each cycle. The 2147483648 clamp is NOT reachable: the guard turns > MaxInt32 into a
  TerminalError before any API call (hooks.go:376-386). finding: [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) (0 / -1 ->
  ValidationException 400 'Member must have value greater than or equal to 1'; 2147483648 clamped),
  [S3T-TABLEBUCKET-080](#s3t-tablebucket-080) (ValidationException is the wire code for every input validation failure on this
  operation and is absent from the model).
  - handling_ref: `pkg/resource/table_bucket/hooks.go:31-40; pkg/resource/table_bucket/hooks.go:376-388`
- [S3T-TABLEBUCKET-076](#s3t-tablebucket-076) - Server-side: settings {} (empty union) -> HTTP 204 no-op; request body without value
  -> ValidationException 400 'Member must not be null' (partially handled)
  Server-side counterpart of the client-side ParamValidationError cells from
  tablebucket/sub-resources/maintenance-defaults-fill. H-S-112 'settings {} rejected with 400' refuted: the
  server accepts the empty union (204) and ignores it. H-S-030 'missing value -> 400' confirmed, with wire
  code ValidationException (not BadRequestException).
  - handling_ref: `pkg/resource/table_bucket/hooks.go:173-195; pkg/resource/table_bucket/hooks.go:389-391`
- [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001) - PutTableBucketMaintenanceConfiguration silently clamps
  nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204 (partially handled)
  FU-04. The modeled range is 1..2147483647 (int32); a Go int64 spec field would let a user send these values
  and the controller must map the wire code to a terminal condition. See extra cells (0, -1, 2^63, 2^64, 1.5)
  in the companion finding.
  - handling_ref: `pkg/resource/table_bucket/hooks.go:31-40; pkg/resource/table_bucket/hooks.go:376-388`
- [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002) - Maintenance days edges: 2147483647 stored; 0 and -1 ->
  ValidationException 400; 2^63 and 2^64 -> 204 clamped; 1.5 -> 204 stored as 1 (partially handled)
  FU-04 extra cells (n=1 each). Default configuration read before any Put: {"settings":
  {"icebergUnreferencedFileRemoval": {"nonCurrentDays": 10, "unreferencedDays": 3}}, "status": "enabled"}.
  - handling_ref: `pkg/resource/table_bucket/hooks.go:31-40; pkg/resource/table_bucket/hooks.go:376-388`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLEBUCKET-053](#s3t-tablebucket-053) | GetTableBucketMaintenanceConfiguration on a fresh bucket -> 200 enabled/3/10; Put {status: enabled} w/o settings -> 204, still enabled/3/10 | get_latency_ms=[49, 42] | 2 |
| [S3T-TABLEBUCKET-054](#s3t-tablebucket-054) | PutTableBucketMaintenanceConfiguration merges into the stored value... | put_latency_ms_min_max=[28, 519] | 12 |
| [S3T-TABLEBUCKET-056](#s3t-tablebucket-056) | No DeleteTableBucketMaintenanceConfiguration in the SDK; status enabled<->disabled flips read back synchronously on the first Get (n=4) | put_latency_ms=[335, 324, 339, 308], get_latency_ms=[37, 25, 314, 29] | 4 |
| [S3T-TABLEBUCKET-075](#s3t-tablebucket-075) | Maintenance days: 0/-1 -> ValidationException 400 naming the field; 36501 and 2147483647 round-trip; 2147483648 -> 204, read back clamped | accepted_put_latency_ms_n=7 | 10 |
| [S3T-TABLEBUCKET-078](#s3t-tablebucket-078) | 20 concurrent maintenance Puts on one bucket (5 threads): 1 InternalServerErrorException 500 after 6.2 s, Put latencies 28 ms-6.5 s... | conc_put_latency_ms_sorted=[28, 30, 36, 86, 314, 315, 318, 319, 322, 322, 323, 324, 335, 336, 338, 2388, 4466, 5086, 6168, 6459] | 20 |
| [S3T-TABLEBUCKET-093](#s3t-tablebucket-093) | GetTableBucketMetricsConfiguration.id is the constant string 'EntireBucket' across repeated Puts, Delete+Put and on a second bucket | put_latency_ms=[45, 44, 48, 70], get_latency_ms=[12, 13, 12, 12, 13, 12] | 4 |
| [S3T-TABLEBUCKET-094](#s3t-tablebucket-094) | 10x metrics Put,Put,Get,Delete,Get without sleep: no ConflictException, no stale read; Gets throttled 429 from the 20th Get in 1.3 s | call_latency_ms_min_max=[3, 50] | 50 |
| [S3T-TABLEBUCKET-095](#s3t-tablebucket-095) | After DeleteTableBucketMetricsConfiguration (204) the first Get was 404 (2 of 3 reps; 3rd a 429 carry-over), no 200 in 57 polls over 3x4 s | first_poll_offset_s=[0.003, 0.013, 0.013] | 57 |
| [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001) | PutTableBucketMaintenanceConfiguration silently clamps nonCurrentDays/unreferencedDays > 2^31-1 (2^31, 2^32, 2^63-1) to 2147483647, HTTP 204 | n_per_cell=3, cells=18, values=[2147483648, 4294967296, 9223372036854775807] | 3 |
| [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002) | Maintenance days edges: 2147483647 stored; 0 and -1 -> ValidationException 400; 2^63 and 2^64 -> 204 clamped; 1.5 -> 204 stored as 1 | n_per_cell=1 | 1 run |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `tablebucket-maintenance-metrics.md` - unverified, single-run or stale items worth a future probe; one line each with the finding id.

- [S3T-TABLEBUCKET-078](#s3t-tablebucket-078): the 1-in-20 InternalServerErrorException 500 and the 2.4-6.5 s serialized Put latencies come from one 5-thread burst (n=20, confidence medium); a second run would show whether the 500 recurs and whether the ~4.1 s blocked first Get is typical.
- [S3T-TABLEBUCKET-094](#s3t-tablebucket-094): the metrics Get throttle (first 429 on the 20th Get within 1.29 s) is a single burst (n=50, confidence medium); the per-second Get limit itself was not isolated from the interleaved Puts/Deletes.
- [S3T-TABLEBUCKET-095](#s3t-tablebucket-095): one of three first polls after Delete was a 429 carried over from the -094 burst, so the 404-on-first-poll claim rests on 2 of 3 repetitions; re-run with a cooled-down bucket.
- [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002](#s3t-tablebucketmaintenanceconfiguration-002): the 2^63, 2^64, 1.5, 0 and -1 cells ran once each (n=1); only the 2^31/2^32/2^63-1 cells of -001 have n=3.
- [S3T-TABLEBUCKET-075](#s3t-tablebucket-075): the 2147483648 clamp was a single cell (n=1) in this probe; -001 later confirmed it at n=3 per cell, so the two findings together are the evidence - keep both when consolidating.
- [S3T-TABLEBUCKET-055](#s3t-tablebucket-055): the 'Put with no value parameter -> 400' cell was untestable through boto3 (client-side ParamValidationError); only -076 (parameter validation disabled) reached the server, one call.
- [S3T-TABLEBUCKET-053](#s3t-tablebucket-053) / -173: the fresh-bucket default enabled/3/10 was read on one bucket (2 reads) plus the Phase 8 re-verification (3 reads); no finding checks whether the default differs in another region.
- [S3T-TABLEBUCKET-056](#s3t-tablebucket-056): "no Delete operation" was established from the SDK model (dir(client)); no probe attempted a raw request to a Delete path, and no finding records what a Put of the fresh default enabled/3/10 after custom days looks like to a later reader (the merge keeps whatever was last sent, -054).
- [S3T-TABLEBUCKET-074](#s3t-tablebucket-074) / -077: all enum-variant cells are single calls; the 'Get afterwards unchanged' check is the only evidence that a rejected Put never partially applies.
- [S3T-TABLEBUCKET-093](#s3t-tablebucket-093): the id format 'EntireBucket' was observed on two buckets in one region; whether any other id value exists (e.g. after a filter-scoped configuration, if the API ever grows one) is untested.
- [S3T-TABLEBUCKET-080](#s3t-tablebucket-080): the ValidationException-vs-BadRequestException split was observed on PutTableBucketMaintenanceConfiguration and the table-level Put only; the controller-side effect (requeue forever on days = 0, -075) has not been reproduced against a running controller.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-tablebucket-057"></a>**S3T-TABLEBUCKET-057** | shape-mismatch | low | confirmed | GetTableBucketMaintenanceConfiguration returns configuration as a map with key(s) ['icebergUnreferencedFileRemoval']... | [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-100](#s3t-tablebucket-100), [S3T-TABLEBUCKET-158](tablebucket.md#s3t-tablebucket-158), [S3T-TABLEBUCKET-126](tablebucket-encryption-storage.md#s3t-tablebucket-126), [S3T-NAMESPACE-030](namespace.md#s3t-namespace-030), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLEBUCKET-173](#s3t-tablebucket-173), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](tablebucket-encryption-storage.md#s3t-tablebucket-059) | - |
| <a id="s3t-tablebucket-076"></a>**S3T-TABLEBUCKET-076** | request-validation | low | confirmed (hypothesis refuted) | Server-side: settings {} (empty union) -> HTTP 204 no-op; request body without value -> ValidationException 400 'Member must not be null' | [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056), [S3T-TABLEBUCKET-074](#s3t-tablebucket-074), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-077](#s3t-tablebucket-077), [S3T-TABLEBUCKET-080](#s3t-tablebucket-080) | - |
| <a id="s3t-tablebucket-078"></a>**S3T-TABLEBUCKET-078** | error-code | low | confirmed | 20 concurrent maintenance Puts on one bucket (5 threads): 1 InternalServerErrorException 500 after 6.2 s, Put latencies 28 ms-6.5 s... | [S3T-TABLEBUCKET-082](tablebucket-encryption-storage.md#s3t-tablebucket-082) | - |
| <a id="s3t-tablebucket-170"></a>**S3T-TABLEBUCKET-170** | other | low | confirmed | Doc claim C048 TRUE: full-value PutTableBucketMaintenanceConfiguration round-trips exactly (enabled 7/20, 5/20, 7/4 -> 204; Get identical) | [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055) | - |
| <a id="s3t-tablebucket-171"></a>**S3T-TABLEBUCKET-171** | other | low | confirmed | Doc claim C049 TRUE: second full-value Put (disabled 7/4 after enabled 7/4; 5/20 after 7/20) -> 204; Get returns exactly the new value | [S3T-TABLEBUCKET-054](#s3t-tablebucket-054), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-170](#s3t-tablebucket-170), [S3T-TABLEBUCKET-172](#s3t-tablebucket-172), [S3T-TABLEBUCKET-055](#s3t-tablebucket-055), [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-173](#s3t-tablebucket-173), [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057) | - |
| <a id="s3t-tablebucket-173"></a>**S3T-TABLEBUCKET-173** | other | low | confirmed | Doc claim C051 TRUE: GetTableBucketMaintenanceConfiguration on a fresh bucket -> 200 {icebergUnreferencedFileRemoval: enabled/3/10}, not 404 | [S3T-TABLEBUCKET-053](#s3t-tablebucket-053), [S3T-TABLEBUCKET-001](tablebucket-encryption-storage.md#s3t-tablebucket-001), [S3T-TABLEBUCKET-057](#s3t-tablebucket-057), [S3T-TABLEBUCKET-171](#s3t-tablebucket-171), [S3T-TABLEBUCKET-056](#s3t-tablebucket-056), [S3T-TABLEBUCKET-059](tablebucket-encryption-storage.md#s3t-tablebucket-059), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097) | - |
| <a id="s3t-tablebucket-174"></a>**S3T-TABLEBUCKET-174** | other | low | confirmed | Doc claim C057 TRUE: PutTableBucketMetricsConfiguration -> 204; Get returns {tableBucketARN: <bucket ARN>, id: 'EntireBucket'} (constant id) | [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-175](#s3t-tablebucket-175), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097) | - |
| <a id="s3t-tablebucket-175"></a>**S3T-TABLEBUCKET-175** | other | low | confirmed | Doc claim C058 TRUE: Second PutTableBucketMetricsConfiguration on the same bucket -> 204 (no ConflictException); Get id stays 'EntireBucket' | [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-165](tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-081](tablebucket-policy.md#s3t-tablebucket-081), [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018), [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-174](#s3t-tablebucket-174), [S3T-TABLEBUCKET-176](#s3t-tablebucket-176), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097) | - |
| <a id="s3t-tablebucket-176"></a>**S3T-TABLEBUCKET-176** | other | low | confirmed | Doc claim C059 TRUE: GetTableBucketMetricsConfiguration before any Put -> NotFoundException 404 'No such metric configuration exists' | [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKET-174](#s3t-tablebucket-174), [S3T-TABLEBUCKET-175](#s3t-tablebucket-175), [S3T-TABLEBUCKET-093](#s3t-tablebucket-093), [S3T-TABLEBUCKET-094](#s3t-tablebucket-094), [S3T-TABLEBUCKET-173](#s3t-tablebucket-173), [S3T-TABLEBUCKET-053](#s3t-tablebucket-053) | - |
| <a id="s3t-tablebucketmaintenanceconfiguration-002"></a>**S3T-TABLEBUCKETMAINTENANCECONFIGURATION-002** | request-validation | low | confirmed | Maintenance days edges: 2147483647 stored; 0 and -1 -> ValidationException 400; 2^63 and 2^64 -> 204 clamped; 1.5 -> 204 stored as 1 | [S3T-TABLEBUCKET-075](#s3t-tablebucket-075), [S3T-TABLEBUCKETMAINTENANCECONFIGURATION-001](#s3t-tablebucketmaintenanceconfiguration-001) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
