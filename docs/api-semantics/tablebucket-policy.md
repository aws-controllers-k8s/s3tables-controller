<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# TableBucket resource policy and authorization errors
_Put/Get/DeleteTableBucketPolicy semantics (whole-document replace, 404 when absent, idempotent delete, read-back normalization, principal propagation windows), the two 403 codes, and the first-sync deletion hazard; scope verdict for TableBucketPolicy._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `tablebucket-policy.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the table bucket's resource policy - PutTableBucketPolicy, GetTableBucketPolicy and
DeleteTableBucketPolicy on a bucket ARN - with document validation and server-side normalization, the authorization codes
a caller meets on the policy and on the bucket-core operations, and the metrics-configuration findings that reached this
document through their authorization behavior. There is no state machine and no asynchrony: Put is a synchronous
whole-document replace (only the last document's statements remain; re-sending the same document is 200, never
ConflictException), Delete is an unconditional 204 whether or not a policy exists, with no prior read and no version token,
and GetTableBucketPolicy reflected the last write on its first poll in every repetition ([S3T-TABLEBUCKET-064](#s3t-tablebucket-064), -062,
[S3T-TABLEBUCKETPOLICY-015](#s3t-tablebucketpolicy-015), -003). The policy is read back only through GetTableBucketPolicy - GetTableBucket and
ListTableBuckets carry no policy member - and the read-back is the server's re-serialization, never the sent bytes
([S3T-TABLEBUCKET-178](#s3t-tablebucket-178), -063, [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012)). Scope verdict: the findings show the API shape of a standalone
resource - a Put/Get/Delete triple, NotFoundException 404 'The specified bucket policy does not exist.' when absent
(never-set and after-delete share the message) and 204 on delete-when-absent - so a TableBucketPolicy CRD is implementable
on this API ([S3T-TABLEBUCKET-061](#s3t-tablebucket-061), -065, [S3T-TABLEBUCKETPOLICY-010](#s3t-tablebucketpolicy-010), -011); the controller does not implement that CRD today
(render.yaml partition note) and reconciles the document as a field of TableBucket, see "Handling gaps", where the
first-sync deletion, the one transient 400, the normalization re-Put loop and the lockout Denies are the hazards
([S3T-TABLEBUCKET-062](#s3t-tablebucket-062), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), -012, -006, -007).

### Rules

1. Never issue DeleteTableBucketPolicy for an unset spec against a policy you did not write: the delete is unconditional -
   204 with no policy present, 204 on a second Delete, 204 from a principal that never read the policy, no version token -
   and the policy is gone on the first read afterwards (51/51 reads 404 at 200 ms), so an adopted or out-of-band-policied
   bucket loses its policy on the first reconcile ([S3T-TABLEBUCKET-062](#s3t-tablebucket-062), [S3T-TABLEBUCKETPOLICY-015](#s3t-tablebucketpolicy-015), -003).
2. PutTableBucketPolicy accepts a Deny + NotAction s3tables:GetTableBucket for the account root principal (200) with no
   safety check; afterwards every IAM principal of the owning account gets AccessDeniedException 403 'with an explicit deny
   in a resource-based policy' on every operation except GetTableBucket - including DeleteTableBucketPolicy, DeleteTable,
   DeleteNamespace and DeleteTableBucket - while the bucket stays listed; no owner override exists, only the root user or
   AWS Support recovers; surface that 403 as terminal instead of requeueing silently ([S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007)).
3. A policy Deny naming the caller on Get/Put/DeleteTableBucketPolicy is enforced on the very first call (0 Puts accepted,
   first 403 0.03 s after the Deny Put returned) as AccessDeniedException 403 - a code the model does not list for
   Put/DeleteTableBucketPolicy - and locks the controller out of reading, updating and deleting the policy; the bucket's
   own reads stay 200, DeleteTableBucket on the empty bucket (204) is the only escape (the same name is then 409
   'transitional state'), and a principal the Deny does not name still deletes with 204 ([S3T-TABLEBUCKET-091](#s3t-tablebucket-091),
   [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), -016).
4. Compare the policy semantically or re-Put every reconcile: no read-back was byte-identical to the sent text -
   GetTableBucketPolicy minifies, moves Version first, inserts "Version":"2008-10-17" when none was sent, folds Effect
   'allow'/'ALLOW' to 'Allow', collapses one-element Action/Resource/Principal.AWS/Condition lists to scalars, turns a
   Statement object into a 1-element array, de-duplicates [root, root], stringifies Condition Bool false to "false" and
   renders a deleted role's ARN as its AROA id; action-name case, wildcard actions, duplicate Sids, Allow+NotAction/
   NotResource and a top-level Id come back as sent ([S3T-TABLEBUCKET-063](#s3t-tablebucket-063), -111, [S3T-TABLEBUCKETPOLICY-008](#s3t-tablebucketpolicy-008), -012, -013).
5. Put is a whole-document replace: after Put A, Put A again (200, no ConflictException) and Put B, Get shows only B's
   statements; there is no statement merge ([S3T-TABLEBUCKET-064](#s3t-tablebucket-064), [S3T-TABLEBUCKETPOLICY-009](#s3t-tablebucketpolicy-009)).
6. NotFoundException 404 has two meanings on GetTableBucketPolicy and only the message tells them apart: 'The specified
   bucket policy does not exist.' (never set, and the same text after a Delete) vs 'The specified bucket does not exist.'
   (bucket gone, the text GetTableBucket returns); a never-set policy is 404, not 200 with an empty resourcePolicy, and the
   parent's Get/List expose no policy key, so existence is learnt only here ([S3T-TABLEBUCKET-061](#s3t-tablebucket-061), -065, -178,
   [S3T-TABLEBUCKETPOLICY-010](#s3t-tablebucketpolicy-010), -011).
7. Classify PutTableBucketPolicy BadRequestException 400 by message - one code covers permanent and transient causes:
   'The specified policy is malformed' (not JSON, whitespace-only, '{}', UTF-8 BOM, bare-string Principal, Condition
   operator 'bool'), 'Invalid Action' ('foo:Bar', 's3:GetObject', Action '*', 'S3Tables:GetTableBucket'), 'Policy contains
   a statement with no principals' and 'Public policies cannot be applied...' (Principal '*', {'AWS':'*'},
   Allow+NotPrincipal) are permanent; 'Invalid principal in policy' is permanent for a role that does not exist but
   transient for a role created 0-7.5 s earlier ([S3T-TABLEBUCKET-088](#s3t-tablebucket-088), -089, -090, -110, [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), -012,
   -013).
8. Do not rely on Put to catch mistakes: it stores as sent the unknown actions 's3tables:FooBarBaz', 's3tables:*',
   's3tables:Get*' and 's3tables:gettable', Resource '*', the bucket ARN itself, a bucket ARN in another region or account,
   a nonexistent same-account bucket, 'arn:aws:s3:::some-bucket/*', a Service principal, and Principal '*' once a
   StringEquals aws:PrincipalAccount condition is attached ([S3T-TABLEBUCKET-088](#s3t-tablebucket-088), -110, -089).
9. The size limit is 20480 bytes of RAW resourcePolicy as sent (20480 -> 200, 20481 -> ValidationException 400 'Member
   must have length less than or equal to 20480', an unmodeled code, not BadRequestException); whitespace counts, so the
   same 60-statement document passes minified at 12098 B and fails pretty-printed to 20481 B, and '' -> ValidationException
   400 'length greater than or equal to 1'; every rejected Put - size, malformed, public, empty - left the stored baseline
   untouched, none acted as a delete ([S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002), -014, [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), -089).
10. Two 403 codes with disjoint causes: AccessDeniedException 403 for an IAM identity-policy explicit Deny on each of the
    seven bucket-core operations (Get/List/Create/DeleteTableBucket, Tag/Untag/ListTagsForResource; the non-denied
    GetTableBucketEncryption stays 200) and for every resource-policy Deny observed (caller session, another role, account
    root) although the model lists ForbiddenException for Put/DeleteTableBucketPolicy and
    GetTableBucketMaintenanceConfiguration; ForbiddenException 403 'Unauthorized' only for a foreign account id in the
    bucket ARN (metrics Put/Get/Delete); a well-formed own-account ARN of a never-existing bucket is NotFoundException 404
    on Get/Delete/ListTags and PutTableBucketMetricsConfiguration ([S3T-TABLEBUCKET-046](#s3t-tablebucket-046), -091, -099, -163,
    [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005), -016).
11. Metrics configuration needs no CloudWatch permission and no service-linked role: a role with Allow s3tables:* and Deny
    cloudwatch:* ran Put 204 / Get 200 {id: EntireBucket} / Delete 204 / Get-after-Delete 404 on the metrics configuration,
    and iam:ListRoles under /aws-service-role/ showed no new role 5 s after enabling ([S3T-TABLEBUCKET-113](tablebucket-maintenance-metrics.md#s3t-tablebucket-113)).
12. Do not poll the policy for eventual consistency: after Put v1 -> v2 every read at 200 ms carried v2 and after Delete
    every read was 404 (3 repetitions), a Put 103 ms after a Delete was 200 with five 200 reads, and a resource-policy Deny
    on a role was enforced on that role's first poll and lifted on its first poll after DeleteTableBucketPolicy with 0 denied
    polls in 3/3 ([S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), -005, [S3T-TABLEBUCKET-065](#s3t-tablebucket-065)).
13. Requeue, do not go terminal, on 'Invalid principal in policy' for a role created moments ago (400 then 200 within
    7.2-8.6 s of CreateRole, n=2); after DeleteRole the read-back flips from the ARN to the AROA id, and a same-name
    re-created role does NOT re-bind - the read-back stayed the old AROA for the 20 s watched until the ARN document was
    re-Put (200) - so a spec(ARN) != read-back(AROA) diff is a correct signal and the re-Put is the fix once the new role has
    propagated ([S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), -013).
14. Reads did not throttle at the lab's burst size: 30 concurrent GetTableBucketPolicy (and GetTableBucket,
    ListTableBuckets, ListTagsForResource), two runs each, all 200; 0 TooManyRequestsException in 270 read calls; the 100-
    and 200-concurrent cells are untested ([S3T-TABLEBUCKET-081](#s3t-tablebucket-081)).

### Timing ranges

- GetTableBucketPolicy read-after-write: first poll after the v2 Put at 0.064-0.068 s, n=3 repetitions (50 polls per Put
  and 51 per Delete at 200 ms, 101 polls counted), 0 stale reads in either direction ([S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003)).
- Resource-policy Deny on another role: denied on the first poll at +0.01-0.07 s after the Deny Put, n=3; allowed again on
  the first poll at +0.12-0.13 s after DeleteTableBucketPolicy, 0 denied polls in 3/3 ([S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005)).
  Caller-scoped Deny: first 403 at 0.03 s after the Deny Put returned, 0 Puts accepted before it - n=1, no range
  ([S3T-TABLEBUCKET-091](#s3t-tablebucket-091)).
- Delete -> re-Put gap: one measurement, 103 ms, then 200 and five 200 reads at 1 s (n=1, [S3T-TABLEBUCKET-065](#s3t-tablebucket-065)) - no range
  exists, do not derive a minimum wait from it.
- Principal propagation after CreateRole: 'Invalid principal in policy' at +4.9 s (n=1) and 8 rejections at 1 s over
  0-7.5 s (n=1); first accepted Put at 7.2-8.6 s, n=2 roles; the same-name re-created role's first retry at +20.7 s was
  accepted (an upper bound, not a window) ([S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), -013).
- Read-back ARN -> AROA after DeleteRole: 8.4 s, n=1, and the AROA document re-Put at +9 s was 200 - single run; the
  time-since-delete rule for the AROA window is canonical in `table-policy.md` ([S3T-TABLEPOLICY-010](table-policy.md#s3t-tablepolicy-010) per -013's
  contradiction note) ([S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013)).
- GetTableBucketPolicy latency under 30-way concurrency: 35-130 ms and 29-131 ms in two bursts, burst wall 0.15 s both,
  n=240 measured read calls across the four read ops, 0 throttled ([S3T-TABLEBUCKET-081](#s3t-tablebucket-081)).

### Handling gaps

- [S3T-TABLEBUCKET-062](#s3t-tablebucket-062) (suspect-bug, not yet handled): today `syncPolicy` (pkg/resource/table_bucket/hooks.go:316-343)
  issues DeleteTableBucketPolicy whenever `spec.policy` is nil or "", `Policy` has no `late_initialize`
  (generator.yaml:229-238) and the read hook sets `Spec.Policy = nil` on NotFoundException (hooks.go:123-138), so an unset
  spec against a live document is a `Spec.Policy` nil-difference (delta.go:69-70) that runs the delete. Evidence: the delete
  is 204 with no precondition and the policy is gone on the first read (rule 1; -003, -015), so an adopted or
  out-of-band-policied bucket loses its policy on the first reconcile; if that policy denies the controller
  s3tables:DeleteTableBucketPolicy the delete instead fails every reconcile with AccessDeniedException 403, a code the model
  does not list (-091). The empty-vs-nil perpetual-diff half is controller-internal and silent on the wire (Delete is
  idempotent 204), so API evidence does not settle it.
- [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001) (suspect-bug, not yet handled): today `terminal_codes: [BadRequestException]`
  (generator.yaml:267-268) covers every TableBucket call, `syncPolicy` returns the PutTableBucketPolicy error unchanged with
  no message inspection (hooks.go:334-342), and the create-time Synced=False nudge (generator.yaml:270-275) lands the first
  Put seconds after CreateTableBucket. Evidence: 'Invalid principal in policy' 400 is transient for a role created moments
  earlier (400 at +4.9 s then 200 at +7.2 s, n=1; 8 x 400 over 0-7.5 s then 200 at +8.6 s, -013), so a TableBucket whose
  policy names a role from the same apply goes ACK.Terminal on a transient and is not reconciled again until the spec
  changes; 'malformed', 'Invalid Action', 'no principals' and 'Public policies' are correctly permanent (-088, -089).
- [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012) (suspect-bug, not yet handled): today `Policy` is `is_iam_policy: true`
  (generator.yaml:234-238), so delta.go:69-75 compares with `ackcompare.IAMPolicyDocumentEqual`, and
  `customUpdateTableBucket` returns the desired spec untouched (hooks.go:213-215, 250). Evidence (the finding's notes record
  a unit test of the sent/read-back pairs against runtime v0.64.0): the comparator tolerates whitespace, key order, statement
  order, Statement object -> array and one-element list collapse, but reports a difference for an inserted Version
  2008-10-17 (-111), Effect case folding, [root, root] de-duplication, Bool false -> "false", account id -> :root and ARN ->
  AROA (-013), so a policy in any of those forms carries a permanent `Spec.Policy` delta and re-Puts on every reconcile.
- [S3T-TABLEBUCKET-090](#s3t-tablebucket-090) (partial): today nil and "" both route to DeleteTableBucketPolicy (hooks.go:325-332) and `Policy` is
  deliberately not late-initialized (generator.yaml:229-234), so the '' -> ValidationException 400 cell (an unmodeled wire
  code, 'length greater than or equal to 1') is never reached from the controller. Evidence: whitespace-only, '{}' and a
  BOM-prefixed document do reach Put and return BadRequestException 400 'The specified policy is malformed' (Terminal via
  `terminal_codes`, matching the permanent behavior), none of the rejections acted as a delete, a non-ASCII Sid is stored
  unchanged; the partial gap: removal can only be expressed as nil/"" -> Delete - the path rule 1 and -062 flag as
  destructive on first sync - and the API accepts no non-destructive empty document as a Put instead.

### Where next

- `tablebucket.md` carries the seven bucket-core operations behind the IAM Deny matrix of -046, the identifier-form
  404/400/403 split that -163 confirms (foreign-account ForbiddenException via [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038) per -046's note), and
  the ~10 min post-delete name reservation behind the 409 that follows rule 3's DeleteTableBucket escape (-091).
- `tablebucket-maintenance-metrics.md` owns the metrics configuration pair; -099 (foreign account id -> ForbiddenException
  'Unauthorized') and -113 (no CloudWatch permission or service-linked role needed) landed here through their authorization
  match, and GetTableBucketMaintenanceConfiguration is the probe target of -005's Deny.
- `table-policy.md` is the table-level twin: the same normalization engine at both levels (-012 notes; statement key
  re-ordering is canonical in [S3T-TABLE-152](table-policy.md#s3t-table-152) per -063's contradiction note), the analogous PutTablePolicy bricking (-007),
  caller-scoped lockouts on tables ([S3T-TABLE-214](table-policy.md#s3t-table-214) per -007's note), and the AROA re-Put window settled as time since
  DeleteRole, not level ([S3T-TABLEPOLICY-010](table-policy.md#s3t-tablepolicy-010), [S3T-TABLE-165](table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](table-policy.md#s3t-tablepolicy-004) per -013's contradiction notes).
  `tablebucket-encryption-storage.md` and `tablebucket-replication.md` cover the other sub-resource pairs; the undeletable
  lab bucket is accounted for here (-007), not there.
- Long catalogues belong in `details/<id>.md` rather than inline: the 25-variant re-shape/identical/rejected matrix
  ([S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012)), the 17 accepted-as-sent documents ([S3T-TABLEBUCKET-110](#s3t-tablebucket-110)), the six validation cells (-088), the
  five weird-input cells (-090), the seven IAM Deny messages (-046), the 17 operations denied after the root Deny+NotAction
  ([S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007)), the per-op burst/latency matrix (-081) and the CreateRole / DeleteRole / re-create timeline
  ([S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013)).
- Open experiments (STATUS.md): the leftover bucket `ackq-9ba36e-polnorm` (namespace `ackq_9ba36e_ns`, table
  `ackq_9ba36e_t`) cannot be deleted by any IAM principal, stays `active` in the registry and fails every `semantics sweep`
  until the root user or AWS Support removes it (-007); -007 was swapped out of the Phase 8 reverify sample by policy and
  stands on a single run; the 100- and 200-concurrent read cells of -081 are out of reach under the lab's 30-call burst cap;
  the recreate step after -091's 409 was not re-tried; -013's same-name re-binding was watched for 20 s only and -001 ran a
  single role interleaved with a replication Put loop; the empty-vs-nil half of -062 cannot be settled on the wire.
<!-- preserved:end -->

## At a glance

- canonical findings: 33 (high 9 / medium 18 / low 6); duplicates folded into the appendix: 2
- handling: handled 8 · partial 0 · tracked 0 · unhandled 22 · suspect-bug 3 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 0 · last_verified: 2026-10-13..2026-10-14 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: normalization 5, delete-semantics 4, error-code 4, other 4, request-validation 4, prerequisite
  3, quota-limit 3, eventual-consistency 2, idempotency 2, read-gap 1, sub-resource-api 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTableBucket | create | name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteNamespace | delete | tableBucketARN, namespace | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucket | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTableBucketPolicy | delete | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTablePolicy | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketEncryption | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableBucketMaintenanceConfiguration | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketPolicy | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| ListTableBuckets | list | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| ListTagsForResource | list | resourceArn | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTableBucketPolicy | create | tableBucketARN, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTablePolicy | create | tableBucketARN, namespace, name, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Idempotency

- <a id="s3t-tablebucket-062"></a>**S3T-TABLEBUCKET-062** `idempotency` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **DeleteTableBucketPolicy is idempotent: HTTP 204 when no policy exists and on a second Delete (not NotFoundException)** (hypothesis refuted; behavior confirmed)
  Fresh bucket: DeleteTableBucketPolicy -> HTTP 204. PutTableBucketPolicy(valid) -> HTTP 200; Delete -> HTTP
  204; the first Get right after the Delete -> NotFoundException (HTTP 404) 'The specified bucket policy does
  not exist.'; Gets at +1 s and +2 s -> ["NotFoundException (HTTP 404) 'The specified bucket policy does not
  exist.'", "NotFoundException (HTTP 404) 'The specified bucket policy does not exist.'"]. Second Delete ->
  HTTP 204.
  - ACK: exceptions.404, custom_delete · ops: DeleteTableBucketPolicy, PutTableBucketPolicy,
    GetTableBucketPolicy · fields: resourcePolicy
  - repro: Fresh bucket: Delete; Put; Delete; Get immediately (+1 s, +2 s); Delete again
  - measurements: policy_gone_on_first_read=true, n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKETPOLICY-015](#s3t-tablebucketpolicy-015), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065), [S3T-TABLE-150](table-policy.md#s3t-table-150), [S3T-TABLE-156](table-policy.md#s3t-table-156) · hypotheses:
    H-S-014 · evidence: tablebucket/sub-resources/policy-get-put-delete
  - notes: H-S-014 REFUTED: Delete-when-absent and the repeated Delete both returned HTTP 204, no
    NotFoundException; the Put->Delete->Get part held (policy gone on the first read, 404 'The specified
    bucket policy does not exist.'). Contrast: DeleteTableBucketEncryption when never set returns
    NotFoundException...
  - full notes: [details/S3T-TABLEBUCKET-062.md](details/S3T-TABLEBUCKET-062.md)

## Errors

- <a id="s3t-tablebucket-045"></a>**S3T-TABLEBUCKET-045** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM Deny (identity policy or table-bucket policy) surfaces as AccessDeniedException 403, not ForbiddenException...** (hypothesis refuted; behavior confirmed)
  As a role whose identity policy explicitly denies s3tables:GetTableBucketPolicy: GetTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-bd98cf-perm-src with an explicit deny in an identity-based
  policy'. After PutTableBucketPolicy (admin) with an explicit Deny of s3tables:GetTableBucketEncryption for
  that role: GetTableBucketEncryption as the role -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableBucketEncryption on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-bd98cf-perm-src with an explicit deny in a resource-based
  policy' (before the policy: HTTP 200; control GetTableBucketStorageClass as the role: HTTP 200; admin
  GetTableBucketEncryption with the Deny policy in place: HTTP 200).
  - ACK: terminal_codes, exceptions.404 · ops: GetTableBucketPolicy, GetTableBucketEncryption,
    PutTableBucketPolicy
  - repro: AssumeRole(role with identity Deny) -> GetTableBucketPolicy; PutTableBucketPolicy(Deny role on
    GetTableBucketEncryption) -> GetTableBucketEncryption as the role
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016), [S3T-TABLE-109](table-policy.md#s3t-table-109),
    [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](table-policy.md#s3t-table-131) · hypotheses:
    H-S-018 · evidence: service/error-taxonomy/repl-permissions
  - notes: H-S-018 REFUTED: an explicit Deny in an identity-based policy AND an explicit Deny in the
    table-bucket (resource-based) policy both surface as AccessDeniedException 403 with the IAM-style message
    ('... with an explicit deny in an identity-based policy' / '... in a resource-based policy')....
  - full notes: [details/S3T-TABLEBUCKET-045.md](details/S3T-TABLEBUCKET-045.md)

- <a id="s3t-tablebucket-046"></a>**S3T-TABLEBUCKET-046** `error-code` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **IAM explicit Deny on the seven bucket-core operations yields AccessDeniedException/403 on every one...**
  Role = AdministratorAccess + inline Deny on ['s3tables:GetTableBucket', 's3tables:ListTableBuckets',
  's3tables:CreateTableBucket', 's3tables:DeleteTableBucket', 's3tables:TagResource',
  's3tables:UntagResource', 's3tables:ListTagsForResource']. As that role: GetTableBucket ->
  AccessDeniedException/403; ListTableBuckets -> AccessDeniedException/403; CreateTableBucket ->
  AccessDeniedException/403; DeleteTableBucket -> AccessDeniedException/403; TagResource ->
  AccessDeniedException/403; UntagResource -> AccessDeniedException/403; ListTagsForResource ->
  AccessDeniedException/403; GetTableBucketEncryption (not denied) -> 200. Messages: ['User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:CreateTableBucket on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-denied with an explicit deny in an
  identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucket on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with
  an explicit deny in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:GetTableBucket on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:ListTableBuckets on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/* with an explicit deny
  in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTagsForResource on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:TagResource on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an
  explicit deny in an identity-based policy', 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not
  authorized to perform: s3tables:UntagResource on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-d5c897-target with an explicit deny in an identity-based
  policy'].
  - ACK: terminal_codes · ops: GetTableBucket, ListTableBuckets, CreateTableBucket, DeleteTableBucket,
    TagResource, UntagResource, ListTagsForResource
  - repro: IAM role with AdministratorAccess + Deny on the seven ops; AssumeRole; call each op against an
    existing bucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016), [S3T-TABLE-109](table-policy.md#s3t-table-109),
    [S3T-TABLE-250](table-policy.md#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](table-policy.md#s3t-table-026), [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](table-policy.md#s3t-table-131) · hypotheses:
    H-B-036, H-B-037 · evidence: tablebucket/error-taxonomy/iam-deny-codes
  - notes: H-B-036 (AccessDeniedException for Get/List, ForbiddenException for the rest): refuted. H-B-037
    (ForbiddenException everywhere for IAM denies): refuted/qualified: ['AccessDeniedException/403'].
  - full notes: [details/S3T-TABLEBUCKET-046.md](details/S3T-TABLEBUCKET-046.md)

- <a id="s3t-tablebucket-061"></a>**S3T-TABLEBUCKET-061** `error-code` · impact high · handled · verified 2026-10-13
  **GetTableBucketPolicy never set -> NotFoundException 404; same op on a nonexistent bucket -> NotFoundException 404**
  Fresh bucket (no policy ever): GetTableBucketPolicy -> NotFoundException (HTTP 404) 'The specified bucket
  policy does not exist.' (x-amzn-ErrorType
  'NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'). Well-formed ARN of a
  bucket that does not exist: GetTableBucketPolicy -> NotFoundException (HTTP 404) 'The specified bucket does
  not exist.' (x-amzn-ErrorType
  'NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'); GetTableBucket on that ARN
  -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'. Messages differ, so the code alone
  cannot distinguish 'policy absent' from 'bucket gone'.
  - ACK: exceptions.404, custom_find · ops: GetTableBucketPolicy, GetTableBucket · fields: resourcePolicy
  - repro: CreateTableBucket -> GetTableBucketPolicy; GetTableBucketPolicy(ARN of a nonexistent bucket);
    GetTableBucket(same ARN)
  - measurements: n=1
  - handling: handled via `generator.yaml:229-238; generator.yaml:269-275; pkg/resource/table_bucket/hooks.go:123-138; test/e2e/tests/test_table_bucket.py:330-340; generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-178](#s3t-tablebucket-178), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065),
    [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKETPOLICY-010](#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](#s3t-tablebucketpolicy-011), [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007),
    [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLE-156](table-policy.md#s3t-table-156), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLE-149](table-policy.md#s3t-table-149), [S3T-TABLE-181](table-policy.md#s3t-table-181), [S3T-TABLE-184](table-policy.md#s3t-table-184) ·
    hypotheses: H-S-013 · evidence: tablebucket/sub-resources/policy-get-put-delete
  - notes: H-S-013 confirmed. Never-set message: 'The specified bucket policy does not exist.'.
    Nonexistent-bucket message: 'The specified bucket does not exist.'.

- <a id="s3t-tablebucketpolicy-016"></a>**S3T-TABLEBUCKETPOLICY-016** `error-code` · impact medium · handled · verified 2026-10-14
  **Bucket policy Deny of s3tables:DeleteTableBucketPolicy for principal B -> B's Delete AccessDeniedException 403; A (not denied) deletes 204**
  A PutTableBucketPolicy(P2 = P1 + {Effect Deny, Principal AWS <B's role ARN>, Action
  s3tables:DeleteTableBucketPolicy, Resource <bucket ARN>}) -> HTTP 200. +3 s: B DeleteTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:DeleteTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4c97df-fu02 with an explicit deny in a resource-based
  policy'; second attempt -> AccessDeniedException (HTTP 403); B GetTableBucketPolicy -> HTTP 200; B
  GetTableBucket -> HTTP 200; A GetTableBucketPolicy -> HTTP 200; A DeleteTableBucketPolicy -> HTTP 204; then
  A Get -> NotFoundException (HTTP 404) and B Get -> NotFoundException (HTTP 404).
  - ACK: terminal_codes, custom_delete · ops: PutTableBucketPolicy, DeleteTableBucketPolicy,
    GetTableBucketPolicy · fields: resourcePolicy
  - repro: A: PutTableBucketPolicy with a Deny s3tables:DeleteTableBucketPolicy for B's role ARN -> B:
    DeleteTableBucketPolicy -> A: DeleteTableBucketPolicy
  - handling: handled via `generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009) ·
    evidence: tablebucketpolicy/sub-resources/fu-02-delete-policy-principal-b
  - notes: FU-02 round 2. The denied principal gets the same wire code as an IAM denial (see behavior); a
    controller whose role is denied by the policy it manages cannot remove that policy - and a Deny for '*'
    bricks the bucket for everyone (the lab's pre-existing polnorm bucket is such a case; not touched...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-016.md](details/S3T-TABLEBUCKETPOLICY-016.md)

## Request validation

- <a id="s3t-tablebucket-088"></a>**S3T-TABLEBUCKET-088** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy: bad JSON, non-s3tables action, missing Principal -> 400; unknown s3tables:* action, other bucket's Resource -> 200**
  With a valid baseline policy stored: not_json -> BadRequestException 400; unknown_namespace ->
  BadRequestException 400; unknown_s3tables_action -> HTTP 200 (STORED); s3_action -> BadRequestException 400;
  foreign_resource -> HTTP 200 (STORED); no_principal -> BadRequestException 400. Messages: not_json: 'The
  specified policy is malformed'; unknown_namespace: 'Invalid Action'; s3_action: 'Invalid Action';
  no_principal: 'Policy contains a statement with no principals'. 'STORED' marks documents that replaced the
  baseline (Get showed them).
  - ACK: terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put baseline; for each document: Put -> record -> Get -> compare with baseline
  - measurements: n=6
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-089](#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLE-167](table-policy.md#s3t-table-167), [S3T-TABLE-162](table-policy.md#s3t-table-162), [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009) ·
    hypotheses: H-S-016 · evidence: tablebucket/weird-inputs/policy-validation-lockout
  - notes: H-S-016 qualified: (a) bad JSON -> BadRequestException 400 'The specified policy is malformed'; (b)
    'foo:Bar' and 's3:GetObject' -> BadRequestException 400 'Invalid Action' (namespace must be s3tables), yet
    's3tables:NotAnAction' -> HTTP 200 and is stored; (c) REFUTED: Resource = another existing...
  - full notes: [details/S3T-TABLEBUCKET-088.md](details/S3T-TABLEBUCKET-088.md)

- <a id="s3t-tablebucket-089"></a>**S3T-TABLEBUCKET-089** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy with Principal '*' and no Condition -> BadRequestException 400; {'AWS':'*'} -> BadRequestException 400...**
  Allow s3tables:GetTable on <bucket>/table/* to Principal '*' -> BadRequestException (HTTP 400) ('Public
  policies cannot be applied to S3 table buckets and tables.'); Principal {'AWS':'*'} -> BadRequestException
  (HTTP 400) ('Public policies cannot be applied to S3 table buckets and tables.'); Principal '*' with
  Condition StringEquals aws:PrincipalAccount=<own account> -> HTTP 200. Baseline policy unchanged after the
  rejected Puts: True.
  - ACK: terminal_codes, docs-only · ops: PutTableBucketPolicy · fields: resourcePolicy
  - repro: Put policy with Principal '*' (no Condition); with {'AWS':'*'}; with '*' + aws:PrincipalAccount
    condition; Get after each
  - measurements: n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](#s3t-tablebucket-088), [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLE-157](table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-008](table-policy.md#s3t-tablepolicy-008) · hypotheses:
    H-S-017 · evidence: tablebucket/weird-inputs/policy-validation-lockout
  - notes: H-S-017 confirmed: public-principal policy rejected with BadRequestException; with the account
    condition it returned HTTP 200.
  - full notes: [details/S3T-TABLEBUCKET-089.md](details/S3T-TABLEBUCKET-089.md)

- <a id="s3t-tablebucket-090"></a>**S3T-TABLEBUCKET-090** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy with resourcePolicy '' -> ValidationException 400, whitespace -> BadRequestException 400...**
  SDK validation off. '' -> ValidationException 400 ('1 validation error detected: Value at 'resourcePolicy'
  failed to satisfy constraint: Member must have length greater than or equal to 1'); ' ' ->
  BadRequestException 400 ('The specified policy is malformed'); '{}' -> BadRequestException 400 ('The
  specified policy is malformed'); BOM-prefixed valid document -> BadRequestException 400; Sid 'ÄckÜnicode-日本'
  -> HTTP 200 (Sid returned unchanged). After each rejected Put the baseline policy was still stored (True);
  none acted as a delete.
  - ACK: custom_delete, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put baseline; Put '' / ' ' / '{}' / BOM+valid / unicode Sid with client validation off; Get after
    each
  - measurements: n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](#s3t-tablebucket-089), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLE-120](table.md#s3t-table-120), [S3T-TABLEPOLICY-001](table-policy.md#s3t-tablepolicy-001) · hypotheses:
    H-S-024 · evidence: tablebucket/weird-inputs/policy-validation-lockout
  - notes: H-S-024 qualified: '' -> ValidationException 400 (unmodeled wire code, min length 1);
    whitespace-only and '{}' -> BadRequestException 400 'The specified policy is malformed' (none acted as a
    delete: confirmed); a UTF-8 BOM before a valid document is REJECTED (BadRequestException 400,
    hypothesis...
  - full notes: [details/S3T-TABLEBUCKET-090.md](details/S3T-TABLEBUCKET-090.md)

- <a id="s3t-tablebucket-110"></a>**S3T-TABLEBUCKET-110** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy accepts unknown s3tables:* actions, Resource '*', s3/foreign ARNs; rejects Action '*' and a nonexistent principal**
  Accepted with HTTP 200 and stored as sent (17 single-statement documents on one bucket): Action
  's3tables:*', 's3tables:FooBarBaz', 's3tables:gettable' (case kept), 's3tables:Get*',
  ['s3tables:GetTable','s3tables:NotReal']; Resource '*', the bucket ARN itself, a bucket ARN in another
  region, in another account, 'arn:aws:s3:::some-bucket/*', a nonexistent same-account bucket; Principal
  {'Service':'maintenance.s3tables.amazonaws.com'}; Version '2008-10-17'. Rejected: Action '*' ->
  BadRequestException 400 'Invalid Action'; Principal = ARN of a role that does not exist ->
  BadRequestException 400 'Invalid principal in policy'.
  - ACK: is_iam_policy, docs-only · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields: resourcePolicy
  - repro: Put each document on a fresh bucket, Get, Delete; record code/HTTP/message
  - measurements: n=17
  - handling: not handled in the controller (as of commit 1eda357)
  - hypotheses: H-S-016 · evidence: tablebucket/creative/policy-validation-boundary
  - notes: Creative follow-up to H-S-016: validation is syntactic plus a principal-existence check and an
    action-namespace check ('s3tables:' prefix required; '*' and other services' actions are 'Invalid
    Action'); action names after the prefix and the Resource ARN are not validated against anything. A...
  - full notes: [details/S3T-TABLEBUCKET-110.md](details/S3T-TABLEBUCKET-110.md)

## Update granularity and ordering

- <a id="s3t-tablebucket-129"></a>**S3T-TABLEBUCKET-129** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Role with only Create/Get/DeleteTableBucket - plain CreateTableBucket 200; with tags, encryption or storageClass -> AccessDeniedException**
  Role policy = Allow s3tables:CreateTableBucket, GetTableBucket, DeleteTableBucket on * (nothing else).
  CreateTableBucket(name) -> HTTP 200 (bucket exists; DeleteTableBucket as the role -> 204).
  CreateTableBucket(name, tags={a: b}) -> AccessDeniedException 403 '... is not authorized to perform:
  s3tables:TagResource on resource: <bucket ARN of the requested name> because no identity-based policy allows
  the s3tables:TagResource action'. CreateTableBucket(name, encryptionConfiguration={sseAlgorithm: AES256}) ->
  AccessDeniedException 403 naming s3tables:PutTableBucketEncryption (even though AES256 is the default).
  CreateTableBucket(name, storageClassConfiguration={storageClass: STANDARD}) -> AccessDeniedException 403
  naming s3tables:PutTableBucketStorageClass. CreateTableBucket(name, tags + encryptionConfiguration) ->
  AccessDeniedException 403 naming s3tables:TagResource (tags checked first). After every 403 the admin's
  GetTableBucket on the constructed ARN -> NotFoundException 404 (no bucket created). The same role's implicit
  deny on ListTagsForResource / TagResource -> AccessDeniedException 403 'because no identity-based policy
  allows the ... action'.
  - ACK: docs-only, terminal_codes · ops: CreateTableBucket, TagResource · fields: tags,
    encryptionConfiguration, storageClassConfiguration
  - repro: IAM role with Allow s3tables:CreateTableBucket/GetTableBucket/DeleteTableBucket only;
    CreateTableBucket with and without tags / encryptionConfiguration(AES256) /
    storageClassConfiguration(STANDARD)
  - measurements: create_table_bucket_latency_ms_allowed=1628, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016), [S3T-TABLE-140](table-policy.md#s3t-table-140),
    [S3T-TABLE-131](table-policy.md#s3t-table-131) · hypotheses: H-B-039 · evidence: table/error-taxonomy/iam-deny-rename-tag
  - notes: H-B-039 qualified: the permission requirement is confirmed (tags need s3tables:TagResource;
    encryptionConfiguration needs PutTableBucketEncryption; storageClassConfiguration needs
    PutTableBucketStorageClass; no bucket is created on the 403), but the code is AccessDeniedException 403
    (unmodeled for...
  - full notes: [details/S3T-TABLEBUCKET-129.md](details/S3T-TABLEBUCKET-129.md)

- <a id="s3t-tablebucketpolicy-001"></a>**S3T-TABLEBUCKETPOLICY-001** `prerequisite` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400...**
  PutTableBucketPolicy whose Principal.AWS is a role ARN that does not exist -> BadRequestException (HTTP 400)
  'Invalid principal in policy'. With the shard role created moments earlier: attempts [+4.9s
  BadRequestException/400, +7.2s 200]; first 200 at +7.2s after CreateRole.
  - ACK: requeue, is_iam_policy · ops: PutTableBucketPolicy · fields: resourcePolicy
  - repro: PutTableBucketPolicy(Principal.AWS=arn of a missing role) -> record; CreateRole -> immediately
    PutTableBucketPolicy(Principal=new role), retry every 2 s
  - measurements: first_policy_ok_since_role_create_s=[7.2], rejected_attempts_before_ok=[1],
    rejected_codes=[BadRequestException/400], n=1
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-088](#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-SERVICE-010](service.md#s3t-service-010), [S3T-SERVICE-025](service.md#s3t-service-025), [S3T-SERVICE-012](service.md#s3t-service-012),
    [S3T-TABLE-164](table-policy.md#s3t-table-164), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-165](table-policy.md#s3t-table-165), [S3T-TABLEPOLICY-004](table-policy.md#s3t-tablepolicy-004),
    [S3T-TABLE-157](table-policy.md#s3t-table-157) · hypotheses: H-S-023 · evidence: tablebucket/dependencies/repl-setup
  - notes: H-S-023 confirmed (400 for a missing principal, transient 4xx right after CreateRole) (n=1 role;
    the policy loop ran interleaved with the replication Put loop).
  - full notes: [details/S3T-TABLEBUCKETPOLICY-001.md](details/S3T-TABLEBUCKETPOLICY-001.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-tablebucket-063"></a>**S3T-TABLEBUCKET-063** `normalization` · impact high · handled · verified 2026-10-13
  **GetTableBucketPolicy re-serializes the policy: whitespace stripped, Version moved first, one-element Action array collapsed to a string**
  Put of a 2-space-indented document with trailing newline, Statement before Version, Sid and Action as a
  one-element list -> HTTP 200; Get returned a minified document with Version first and
  "Action":"s3tables:GetTable" (array collapsed to a scalar). Put of the same statement already minified
  (Action still a list) -> Get again collapsed the list to a scalar, so even a minified document is not echoed
  byte-for-byte. Put with Action as a scalar string -> Get returned the sent string byte-identical. Member
  order inside a statement was preserved (Sid, Effect, Principal, Action, Resource).
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Put pretty-printed policy -> Get -> compare strings; Put minified -> Get; Put with scalar Action ->
    Get
  - measurements: n=3
  - handling: handled via `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`
  - related: [S3T-TABLEBUCKETPOLICY-017](#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111),
    [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013), [S3T-TABLE-152](table-policy.md#s3t-table-152), [S3T-TABLEBUCKETPOLICY-008](#s3t-tablebucketpolicy-008),
    [S3T-TABLEPOLICY-005](table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-154](table-policy.md#s3t-table-154), [S3T-TABLE-166](table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](table-policy.md#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](table-policy.md#s3t-tablepolicy-008) · hypotheses:
    H-S-015 · evidence: tablebucket/sub-resources/policy-get-put-delete
  - notes: H-S-015 confirmed (3 documents): compare policies semantically (parse, treat one-element arrays as
    equal to scalars); a string comparison of spec vs Get differs unless the user writes the canonical form
    (minified, Version first, scalar single values).
  - full notes: [details/S3T-TABLEBUCKET-063.md](details/S3T-TABLEBUCKET-063.md)

- <a id="s3t-tablebucket-111"></a>**S3T-TABLEBUCKET-111** `normalization` · impact medium · handled · verified 2026-10-13
  **GetTableBucketPolicy returns Effect 'allow' normalized to 'Allow' and adds Version '2008-10-17' when the document had none**
  Put of a statement with Effect 'allow' (lower-case) -> HTTP 200; Get returned 'Effect':'Allow'. Put of a
  document WITHOUT a Version member -> HTTP 200; Get returned the document with '"Version":"2008-10-17"'
  inserted first. A document sent with Version '2008-10-17' was echoed unchanged; 's3tables:gettable'
  (lower-case action) was stored with its case preserved.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Put policy with Effect 'allow' -> Get; Put policy without Version -> Get; compare
  - measurements: n=2
  - handling: handled via `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`
  - related: [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013),
    [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKETPOLICY-017](#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKETPOLICY-008](#s3t-tablebucketpolicy-008), [S3T-TABLEPOLICY-005](table-policy.md#s3t-tablepolicy-005),
    [S3T-TABLE-152](table-policy.md#s3t-table-152), [S3T-TABLE-154](table-policy.md#s3t-table-154), [S3T-TABLE-166](table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](table-policy.md#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](table-policy.md#s3t-tablepolicy-008) · hypotheses:
    H-S-015, H-S-016 · evidence: tablebucket/creative/policy-validation-boundary
  - notes: Extends H-S-015 (normalization): besides whitespace/order/array collapsing, Effect case is
    canonicalized and a missing Version is materialized as 2008-10-17 (not 2012-10-17); a semantic comparison
    must tolerate both or the controller re-Puts every reconcile.

- <a id="s3t-tablebucket-178"></a>**S3T-TABLEBUCKET-178** `read-gap` · impact medium · handled · verified 2026-10-14
  **GetTableBucket and ListTableBuckets carry no policy member; a bucket policy is visible only through GetTableBucketPolicy (404 when absent)**
  Raw GetTableBucket JSON keys as A: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'],
  as B: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type'] (B -> HTTP 200);
  ListTableBuckets entry keys: ['arn', 'createdAt', 'name', 'ownerAccountId', 'tableBucketId', 'type']. Keys
  containing 'polic': none. Whether a policy exists is learnt only from GetTableBucketPolicy
  (NotFoundException (HTTP 404) when absent).
  - ACK: custom_find, exceptions.404 · ops: GetTableBucket, ListTableBuckets, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: GetTableBucket(arn), ListTableBuckets(prefix) with and without a bucket policy;
    GetTableBucketPolicy when none is set
  - handling: handled via `generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-061](#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065),
    [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLE-149](table-policy.md#s3t-table-149), [S3T-TABLE-181](table-policy.md#s3t-table-181), [S3T-TABLE-184](table-policy.md#s3t-table-184) · evidence:
    tablebucketpolicy/sub-resources/fu-02-delete-policy-principal-b
  - notes: FU-02 (GetTableBucket body check). Single run; the key set is deterministic from the model (6
    output members).
  - full notes: [details/S3T-TABLEBUCKET-178.md](details/S3T-TABLEBUCKET-178.md)

- <a id="s3t-tablebucketpolicy-008"></a>**S3T-TABLEBUCKETPOLICY-008** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C041 PARTLY: GetTableBucketPolicy parsed JSON equals the Put document only if it had Version, no 1-elem arrays, canonical Effect**
  PutTableBucketPolicy on a bucket with no policy -> HTTP 200 and GetTableBucketPolicy returns the stored
  document minified (whitespace stripped, Version moved first - parsed-equal so far) BUT a one-element Action
  array is rewritten as a plain string even when the sent document was already minified ([S3T-TABLEBUCKET-063](#s3t-tablebucket-063),
  [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004)); a document without Version comes back with "Version":"2008-10-17" inserted and
  Effect 'allow' is normalized to 'Allow' ([S3T-TABLEBUCKET-111](#s3t-tablebucket-111)). Action sent as a scalar string is echoed
  byte-identical.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Settled from existing finding(s) [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111),
    [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKET-064](#s3t-tablebucket-064); no new API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKET-064](#s3t-tablebucket-064),
    [S3T-TABLEPOLICY-005](table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-152](table-policy.md#s3t-table-152), [S3T-TABLE-154](table-policy.md#s3t-table-154), [S3T-TABLE-166](table-policy.md#s3t-table-166) · evidence: service/static/doc-claims-2
  - notes: VERDICT: PARTLY - the Put succeeds and the statements survive, but parsed-JSON equality holds only
    for documents that are already in the server's canonical form.

- <a id="s3t-tablebucketpolicy-012"></a>**S3T-TABLEBUCKETPOLICY-012** `normalization` · impact medium · SUSPECTED CONTROLLER BUG · verified 2026-10-14
  **GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar...**
  Put of 25 policy variants on one bucket, each followed by GetTableBucketPolicy and a structural compare.
  RE-SHAPED: Statement given as a single object -> 1-element array; one-element Resource list -> scalar
  string; one-element Principal.AWS list -> scalar; Principal.AWS list [root, root] -> de-duplicated scalar;
  Condition Bool value JSON false -> the string "false"; Condition StringEquals one-element list -> scalar;
  Effect "ALLOW" -> "Allow". RETURNED STRUCTURALLY IDENTICAL (only minified): action name case as sent
  (s3tables:gettablebucket, s3tables:LISTNAMESPACES), wildcard s3tables:Get*, two-element Resource list,
  Allow+NotAction, Allow+NotResource, two statements with the same Sid (accepted), two byte-identical
  statements (both kept), statements without Sid, Sid "", Sid caf\u00e9 (returned as the raw UTF-8 char),
  explicit Version 2008-10-17 (kept), top-level Id (kept). REJECTED at Put: S3Tables:GetTableBucket
  (capitalised service prefix) -> BadRequestException (HTTP 400) 'Invalid Action'; Principal as a bare ARN
  string -> BadRequestException (HTTP 400) 'The specified policy is malformed'; Condition operator 'bool' /
  key 'AWS:securetransport' -> BadRequestException (HTTP 400) 'The specified policy is malformed';
  Allow+NotPrincipal -> BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets
  and tables.'. No read-back was byte-identical to the sent text.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTableBucketPolicy, GetTableBucketPolicy
    · fields: resourcePolicy
  - repro: Put each variant (Statement object, action case, one-element lists, bare principal, JSON bool
    condition, Not* members, duplicate Sids, unicode, Version 2008, Id) -> Get -> compare
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111), [S3T-TABLE-152](table-policy.md#s3t-table-152), [S3T-TABLE-154](table-policy.md#s3t-table-154),
    [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKETPOLICY-017](#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKET-088](#s3t-tablebucket-088),
    [S3T-TABLEBUCKET-089](#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLE-157](table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-008](table-policy.md#s3t-tablepolicy-008), [S3T-TABLE-166](table-policy.md#s3t-table-166), [S3T-TABLEPOLICY-003](table-policy.md#s3t-tablepolicy-003),
    [S3T-TABLE-167](table-policy.md#s3t-table-167), [S3T-TABLE-162](table-policy.md#s3t-table-162), [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009) · evidence:
    tablebucketpolicy/creative/normalization-extras
  - notes: Extends the known re-serialization findings (minify, Version injection, Effect case, 1-element
    Action collapse, account-id -> :root) with the remaining shapes a semantic comparator must treat as equal
    or a controller will re-Put on every reconcile. Single run per cell; same engine at both levels:...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-012.md](details/S3T-TABLEBUCKETPOLICY-012.md)

- <a id="s3t-tablebucketpolicy-013"></a>**S3T-TABLEBUCKETPOLICY-013** `normalization` · impact medium · handled · verified 2026-10-14
  **Principal role deleted then re-created (same name): read-back stays the old AROA id...**
  Role v1 (path /ackq/) created; PutTableBucketPolicy naming its ARN every 1 s: 8 x BadRequestException 400
  'Invalid principal in policy' (0-7.5 s), HTTP 200 at +8.6 s; GetTableBucketPolicy echoed the ARN with path
  verbatim. DeleteRole -> 200; the read-back principal stayed the ARN for 8 s and flipped to the AROA unique
  id (== v1 RoleId) at +8.4 s. While the role was gone (+9 s): re-Put of the read-back AROA document -> HTTP
  200 (contrast [S3T-TABLE-165](table-policy.md#s3t-table-165): 400 for a table policy at +30 s); re-Put of the ARN document -> 400 'Invalid
  principal in policy'; read-back still the AROA. CreateRole with the SAME name (same ARN, new RoleId): the
  read-back stayed the OLD AROA for the 20 s watched (no re-binding by name). Re-Put of the ARN document at
  +20.7 s -> HTTP 200 and the read-back shows the ARN (now bound to the new role).
  - ACK: is_iam_policy, requeue, references · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: CreateRole -> PutTableBucketPolicy(principal ARN) -> DeleteRole -> Get (AROA) -> CreateRole same
    name -> Get x10 -> PutTableBucketPolicy(ARN) -> Get
  - measurements: accept_after_createrole_s=[8.6, 20.7], aroa_flip_after_delete_s=8.4, n=2
  - handling: handled via `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`
  - related: [S3T-TABLE-165](table-policy.md#s3t-table-165), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111),
    [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004), [S3T-TABLEBUCKETPOLICY-017](#s3t-tablebucketpolicy-017), [S3T-TABLEPOLICY-004](table-policy.md#s3t-tablepolicy-004),
    [S3T-TABLEPOLICY-002](table-policy.md#s3t-tablepolicy-002), [S3T-TABLE-164](table-policy.md#s3t-table-164), [S3T-TABLE-157](table-policy.md#s3t-table-157), [S3T-TABLEPOLICY-010](table-policy.md#s3t-tablepolicy-010) · evidence:
    tablebucketpolicy/creative/recreated-role-principal
  - notes: Extends [S3T-TABLE-165](table-policy.md#s3t-table-165) (deleted role -> AROA in the read-back) with the re-creation case a
    controller meets when an IAM Role CR is recreated: the stored policy is bound to the OLD unique id and
    does not follow the name; the diff spec(ARN) != read-back(AROA) is therefore a correct signal and the...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-013.md](details/S3T-TABLEBUCKETPOLICY-013.md)

## Response fidelity and consistency

- <a id="s3t-tablebucketpolicy-003"></a>**S3T-TABLEBUCKETPOLICY-003** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTableBucketPolicy after Put v1->v2 and after Delete, polled at 200 ms for 4 s x3: 0 stale reads in 101 polls (first read 64-68 ms)** (hypothesis refuted; behavior confirmed)
  Three repetitions on one bucket: PutTableBucketPolicy v1, 2 s, PutTableBucketPolicy v2 (HTTP 200) then
  GetTableBucketPolicy every 200 ms for 4 s: all 50 reads (first at +0.064-0.068 s) carried the v2 statement
  Sid, none returned v1. DeleteTableBucketPolicy (HTTP 204) then the same poll: all 51 reads returned
  NotFoundException 404 'The specified bucket policy does not exist.', none returned the deleted policy. (Byte
  comparison of the document is not possible because the server normalizes it, see the normalization finding.)
  - ACK: none · ops: PutTableBucketPolicy, GetTableBucketPolicy, DeleteTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put v1 -> 2 s -> Put v2 -> Get every 200 ms x 20 -> Delete -> Get every 200 ms x 20; x3
  - measurements: first_poll_offset_s=[0.068, 0.064, 0.064], n=101
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-009](#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKET-064](#s3t-tablebucket-064), [S3T-TABLEPOLICY-006](table-policy.md#s3t-tablepolicy-006), [S3T-TABLE-153](table-policy.md#s3t-table-153), [S3T-TABLE-151](table-policy.md#s3t-table-151) ·
    hypotheses: H-S-021 · evidence: tablebucket/consistency-windows/create-then-subresource-put
  - notes: H-S-021 (contrarian: policy Get returns the old value for ~2-5 s) refuted for the policy noun in
    101 polls at 200 ms; maintenance/metrics/storage-class read-after-Put showed no stale reads either
    (tablebucket/sub-resources/maintenance-defaults-fill, metrics-lifecycle).
  - full notes: [details/S3T-TABLEBUCKETPOLICY-003.md](details/S3T-TABLEBUCKETPOLICY-003.md)

- <a id="s3t-tablebucketpolicy-005"></a>**S3T-TABLEBUCKETPOLICY-005** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Bucket-policy Deny on a role takes effect on the first poll (<= 0.07 s) and stops on the first poll after DeleteTableBucketPolicy...** (hypothesis refuted; behavior confirmed)
  Owner PutTableBucketPolicy {Deny <role ARN> s3tables:GetTableBucketMaintenanceConfiguration on the bucket}
  -> HTTP 200; the role's GetTableBucketMaintenanceConfiguration (HTTP 200 just before) failed on the very
  first poll after the Put (+0.07/+0.01/+0.01 s) with AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<account>:assumed-role/... is not authorized to perform:
  s3tables:GetTableBucketMaintenanceConfiguration on resource: <bucket ARN> with an explicit deny in a
  resource-based policy'. Owner DeleteTableBucketPolicy -> HTTP 204 and owner GetTableBucketPolicy ->
  NotFoundException 404 immediately; the role's Get returned HTTP 200 on its first poll after the delete
  (+0.12/+0.13/+0.12 s), 0 denied polls, in all 3 repetitions. Note the wire code for a resource-policy deny
  is AccessDeniedException, not the ForbiddenException seen for IAM-level/foreign-account denials.
  - ACK: none · ops: PutTableBucketPolicy, DeleteTableBucketPolicy, GetTableBucketPolicy,
    GetTableBucketMaintenanceConfiguration · fields: resourcePolicy
  - repro: role R denied by bucket policy; R polls GetTableBucketMaintenanceConfiguration at 1 s; owner
    deletes the policy; R polls until 200 (x3)
  - measurements: deny_effective_after_s=[0.07, 0.01, 0.01], allowed_again_after_s=[0.12, 0.13, 0.12],
    denied_polls_after_delete=[0, 0, 0], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016), [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009),
    [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065), [S3T-TABLE-158](table-policy.md#s3t-table-158) · hypotheses: H-S-140 ·
    evidence: tablebucket/dependencies/iam-cloudwatch-deny-metrics
  - notes: H-S-140 (contrarian: the Deny keeps being enforced for up to ~30 s after the Delete's 204) refuted
    at 1 s polling resolution in 3 repetitions: policy changes were enforced read-your-writes in both
    directions. GetTableBucketMaintenanceConfiguration is listed with ForbiddenException but returned...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-005.md](details/S3T-TABLEBUCKETPOLICY-005.md)

## Sub-resources

- <a id="s3t-tablebucketpolicy-015"></a>**S3T-TABLEBUCKETPOLICY-015** `sub-resource-api` · impact medium · handled · verified 2026-10-14
  **DeleteTableBucketPolicy by a principal that never read the policy -> 204; Get afterwards -> NotFoundException 404; second Delete -> 204**
  Fresh bucket; principal A PutTableBucketPolicy(P1: Allow account root s3tables:GetTableBucket) -> HTTP 200
  (raw status 200, body None). Principal B (assumed role, Allow s3tables:* only, had never called
  GetTableBucketPolicy) DeleteTableBucketPolicy -> HTTP 204 (raw body None); B GetTableBucketPolicy ->
  NotFoundException (HTTP 404) 'The specified bucket policy does not exist.'; A GetTableBucketPolicy ->
  NotFoundException (HTTP 404) 'The specified bucket policy does not exist.'; B DeleteTableBucketPolicy again
  with no policy present -> HTTP 204. A's GetTableBucketPolicy before the delete returned the policy NOT
  byte-identical to the compact JSON sent (semantically equal: False).
  - ACK: custom_delete, exceptions.404, is_iam_policy · ops: PutTableBucketPolicy, DeleteTableBucketPolicy,
    GetTableBucketPolicy · fields: resourcePolicy
  - repro: A: PutTableBucketPolicy -> B (other principal, no prior Get): DeleteTableBucketPolicy -> B:
    GetTableBucketPolicy -> B: DeleteTableBucketPolicy again
  - handling: handled via `generator.yaml:229-234; pkg/resource/table_bucket/hooks.go:325-332`
  - related: [S3T-TABLEBUCKET-062](#s3t-tablebucket-062), [S3T-TABLEBUCKET-065](#s3t-tablebucket-065), [S3T-TABLE-150](table-policy.md#s3t-table-150), [S3T-TABLE-156](table-policy.md#s3t-table-156) · evidence:
    tablebucketpolicy/sub-resources/fu-02-delete-policy-principal-b
  - notes: FU-02 round 1. No version token / read-before-delete requirement exists for the bucket policy: the
    delete is unconditional and the second delete shows the idempotency code. Single run, deterministic codes.

## Delete semantics

- <a id="s3t-tablebucket-065"></a>**S3T-TABLEBUCKET-065** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTableBucketPolicy then PutTableBucketPolicy within 103 ms -> HTTP 200; Gets over 5 s: 200/200/200/200/200...**
  Put P; Delete -> HTTP 204; Put P issued 103 ms after the Delete returned -> HTTP 200; 5 Gets at 1 s
  intervals -> ['200', '200', '200', '200', '200']. Bare Delete -> HTTP 204; Get -> NotFoundException (HTTP 404)
  'The specified bucket policy does not exist.' (message 'The specified bucket policy does not exist.';
  never-set message 'The specified bucket policy does not exist.').
  - ACK: custom_update, requeue · ops: DeleteTableBucketPolicy, PutTableBucketPolicy, GetTableBucketPolicy ·
    fields: resourcePolicy
  - repro: Put P; Delete; Put P immediately; Get x5 at 1 s; Delete; Get
  - measurements: put_after_delete_gap_ms=103, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-061](#s3t-tablebucket-061), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-178](#s3t-tablebucket-178),
    [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKET-062](#s3t-tablebucket-062), [S3T-TABLEBUCKETPOLICY-015](#s3t-tablebucketpolicy-015), [S3T-TABLE-150](table-policy.md#s3t-table-150), [S3T-TABLE-156](table-policy.md#s3t-table-156),
    [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005) · hypotheses:
    H-S-139 · evidence: tablebucket/sub-resources/policy-get-put-delete
  - notes: H-S-139 confirmed. Full original title: DeleteTableBucketPolicy then PutTableBucketPolicy within
    103 ms -> HTTP 200; Gets over 5 s: 200/200/200/200/200; 404 after Delete equals the never-set message

- <a id="s3t-tablebucket-091"></a>**S3T-TABLEBUCKET-091** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **A policy Deny on Put/DeleteTableBucketPolicy for the caller is enforced on the first call (AccessDeniedException 403...**
  Put(Deny PutTableBucketPolicy+DeleteTableBucketPolicy for the caller role) -> HTTP 200. Re-Put attempts at 1
  s: AccessDeniedException 403 (0 accepted before the first 403; first 403 0.03 s after the Deny Put
  returned). DeleteTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucketPolicy on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-ffbb82-polv
  with an explicit deny in a resource-based policy'. GetTableBucketPolicy -> HTTP 200 (version v1 in place).
  DeleteTableBucket (empty bucket) -> HTTP 204. Recreating the same name 3 s later -> ConflictException 409
  - ACK: terminal_codes, pre-delete-cleanup · ops: PutTableBucketPolicy, DeleteTableBucketPolicy,
    GetTableBucketPolicy, DeleteTableBucket · fields: resourcePolicy
  - repro: Put a Deny on Put/DeleteTableBucketPolicy for the caller; retry Put (same Deny + marker) at 1 s
    until 403; Delete; Get; DeleteTableBucket; recreate; Get
  - measurements: policy_enforcement_lag_s=0.03, puts_accepted_before_403=0, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016), [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009) ·
    hypotheses: H-S-141 · evidence: tablebucket/weird-inputs/policy-validation-lockout
  - notes: H-S-141 qualified: lockout confirmed and DeleteTableBucket (204) is the escape, but the wire code
    is AccessDeniedException 403 ('... with an explicit deny in a resource-based policy'), which the model
    does NOT list for PutTableBucketPolicy/DeleteTableBucketPolicy (they list ForbiddenException). No...
  - full notes: [details/S3T-TABLEBUCKET-091.md](details/S3T-TABLEBUCKET-091.md)

- <a id="s3t-tablebucketpolicy-006"></a>**S3T-TABLEBUCKETPOLICY-006** `delete-semantics` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Caller-scoped Deny on Get/Put/DeleteTableBucketPolicy locks the policy...**
  PutTableBucketPolicy(Deny Principal=<caller session ARN>, Action Get/Put/DeleteTableBucketPolicy,
  Resource=<bucket ARN>) -> HTTP 200. 1 s later, as the same caller: GetTableBucketPolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a resource-based
  policy'; PutTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableBucketPolicy
  on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a
  resource-based policy'; DeleteTableBucketPolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:DeleteTableBucketPolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-a0e652-fc-51-tags with an explicit deny in a resource-based
  policy'; GetTableBucket -> HTTP 200; ListTagsForResource -> HTTP 200; GetTableBucketEncryption -> HTTP 200;
  ListNamespaces -> HTTP 200; DeleteTableBucket (empty bucket) -> HTTP 204; GetTableBucket afterwards ->
  NotFoundException/404.
  - ACK: pre-delete-cleanup, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy,
    DeleteTableBucketPolicy, DeleteTableBucket
  - repro: PutTableBucketPolicy with a Deny for the caller on the three policy ops -> try each policy op, the
    bucket reads and DeleteTableBucket
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-091](#s3t-tablebucket-091), [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046), [S3T-TABLEBUCKET-129](#s3t-tablebucket-129),
    [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016) · evidence:
    tablebucket/creative/failed-create-name-reservation
  - notes: Extends [S3T-TABLEBUCKET-091](#s3t-tablebucket-091) (policy Deny enforced on Put/Delete) with the ReadOne side
    (GetTableBucketPolicy) and the recovery path: a TableBucketPolicy CR under such a policy cannot be read,
    updated or deleted by the controller (AccessDeniedException on all three), while the parent bucket's
    reads...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-006.md](details/S3T-TABLEBUCKETPOLICY-006.md)

- <a id="s3t-tablebucketpolicy-007"></a>**S3T-TABLEBUCKETPOLICY-007** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Deny+NotAction bucket/table policy for the account root principal is accepted (200) and bricks the bucket...**
  PutTableBucketPolicy({Effect: Deny, Principal: {AWS: <account root ARN>}, NotAction:
  s3tables:GetTableBucket, Resource: <bucket ARN>}) -> HTTP 200 (the analogous PutTablePolicy on the table ->
  200). Afterwards, for the owning account's IAM principal: still allowed ['GetTableBucket'];
  AccessDeniedException 403 'with an explicit deny in a resource-based policy' for ['DeleteNamespace',
  'DeleteTable', 'DeleteTableBucket', 'DeleteTableBucketPolicy', 'DeleteTablePolicy', 'GetNamespace',
  'GetTable(tableArn)', 'GetTable(triple)', 'GetTableBucketEncryption',
  'GetTableBucketMaintenanceConfiguration', 'GetTableBucketPolicy', 'GetTablePolicy', 'ListNamespaces',
  'ListTables', 'ListTagsForResource(bucket)', 'PutTableBucketPolicy(open)', 'TagResource(bucket)']. The
  bucket is still listed by ListTableBuckets (['ackq-9ba36e-polnorm']). The account root principal in a
  resource-policy Deny matches every IAM principal of the account, so no IAM principal can delete the policy,
  the table, the namespace or the bucket; only the root user (or AWS Support) can recover.
  - ACK: pre-delete-cleanup, terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, PutTablePolicy,
    DeleteTableBucketPolicy, DeleteTablePolicy, DeleteTable, DeleteNamespace, DeleteTableBucket · fields:
    resourcePolicy
  - repro: PutTableBucketPolicy(Deny, Principal root, NotAction s3tables:GetTableBucket, Resource bucket ARN)
    -> any other call on the bucket as any IAM principal of the account
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-214](table-policy.md#s3t-table-214), [S3T-TABLEBUCKET-091](#s3t-tablebucket-091), [S3T-TABLE-157](table-policy.md#s3t-table-157), [S3T-TABLEBUCKET-045](#s3t-tablebucket-045), [S3T-TABLEBUCKET-046](#s3t-tablebucket-046),
    [S3T-TABLEBUCKET-129](#s3t-tablebucket-129), [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006), [S3T-TABLEBUCKETPOLICY-016](#s3t-tablebucketpolicy-016),
    [S3T-TABLEPOLICY-009](table-policy.md#s3t-tablepolicy-009) · evidence: tablebucketpolicy/creative/lockout-recovery
  - notes: Extends [S3T-TABLE-214](table-policy.md#s3t-table-214) / [S3T-TABLEBUCKET-091](#s3t-tablebucket-091) (caller-scoped lockouts, recoverable by deleting the
    parent) with the unrecoverable case: a Deny that also covers Delete* on the parent. PutTableBucketPolicy
    rejects public Allow '*' policies but accepts an account-wide Deny-NotAction without any safety...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-007.md](details/S3T-TABLEBUCKETPOLICY-007.md)

## Quotas and rate limits

- <a id="s3t-tablebucket-081"></a>**S3T-TABLEBUCKET-081** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **30 concurrent GetTableBucket/ListTableBuckets/ListTagsForResource/GetTableBucketPolicy (2 runs each): 0 TooManyRequests, 0 other errors**
  Bursts of 30 concurrent calls (20 threads) on one table bucket, two runs each 20 s apart: GetTableBucket ->
  run1 {"200": 30} (wall 0.16s), run2 {"200": 30} (wall 0.08s); ListTableBuckets -> run1 {"200": 30} (wall
  0.06s), run2 {"200": 30} (wall 0.05s); ListTagsForResource -> run1 {"200": 30} (wall 0.10s), run2 {"200":
  30} (wall 0.12s); GetTableBucketPolicy -> run1 {"200": 30} (wall 0.15s), run2 {"200": 30} (wall 0.15s). 30
  sequential GetTableBucket at 10 TPS -> {"200": 30}. No TooManyRequestsException was returned by any of the
  270 read calls.
  - ACK: requeue, e2e-timing · ops: GetTableBucket, ListTableBuckets, ListTagsForResource,
    GetTableBucketPolicy
  - repro: CreateTableBucket; 20 threads x 30 GetTableBucket (then ListTableBuckets, ListTagsForResource,
    GetTableBucketPolicy); repeat after 20 s
  - measurements: burst_wall_s.GetTableBucket=[0.16, 0.08], burst_wall_s.ListTableBuckets=[0.06, 0.05],
    burst_wall_s.ListTagsForResource=[0.1, 0.12], burst_wall_s.GetTableBucketPolicy=[0.15, 0.15],
    latency_ms_min_max.GetTableBucket=[[5,146], [7,65]], latency_ms_min_max.ListTableBuckets=[[7,44], [7,39]],
    latency_ms_min_max.ListTagsForResource=[[8,80], [10,100]],
    latency_ms_min_max.GetTableBucketPolicy=[[35,130], [29,131]], n=240
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-064](#s3t-tablebucket-064), [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-165](tablebucket.md#s3t-tablebucket-165), [S3T-TABLEBUCKET-175](tablebucket-maintenance-metrics.md#s3t-tablebucket-175),
    [S3T-TABLEBUCKET-018](tablebucket.md#s3t-tablebucket-018), [S3T-NAMESPACE-033](namespace.md#s3t-namespace-033), [S3T-TABLE-218](table.md#s3t-table-218), [S3T-TABLEBUCKET-094](tablebucket-maintenance-metrics.md#s3t-tablebucket-094) · hypotheses: H-B-060, H-S-073
    · evidence: tablebucket/limits/read-write-bursts-policy-size
  - notes: H-B-060 and H-S-073 refuted at 30 concurrent reads (no 429): the hypotheses posit 429s at 200
    (resp. 100) concurrent reads; the shard caps bursts at 30 calls / 5 s, so only the 30-concurrent cell was
    tried (twice, with [1, 2] other probes running in this checkout). The 100/200-call cells are...
  - full notes: [details/S3T-TABLEBUCKET-081.md](details/S3T-TABLEBUCKET-081.md)

- <a id="s3t-tablebucketpolicy-002"></a>**S3T-TABLEBUCKETPOLICY-002** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableBucketPolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)**
  Valid Allow policies padded with unique Sids: 19000 bytes (96 statements) -> HTTP 200; 20000 bytes (101
  statements) -> HTTP 200; 20480 bytes (103 statements) -> HTTP 200; 20481 bytes (103 statements) ->
  ValidationException (HTTP 400); 21000 bytes (106 statements) -> ValidationException (HTTP 400). Bisection to
  the byte: largest accepted 20480, smallest rejected 20481. Rejection message: '1 validation error detected:
  Value at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to
  20480'. GetTableBucketPolicy after the accepted puts returns 20480 bytes (the stored document is
  byte-identical in length to the sent one).
  - ACK: terminal_codes, is_iam_policy · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: PutTableBucketPolicy with ASCII policies of 19000/20000/20480/20481/21000 bytes (client validation
    off), then bisect
  - measurements: largest_accepted_bytes=20480, smallest_rejected_bytes=20481, n=5
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-088](#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](#s3t-tablebucket-089), [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-014](#s3t-tablebucketpolicy-014), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-001](table-policy.md#s3t-tablepolicy-001), [S3T-TABLE-120](table.md#s3t-table-120) · hypotheses:
    H-S-019 · evidence: tablebucket/limits/read-write-bursts-policy-size
  - notes: H-S-019 qualified: a size limit exists but at exactly 20480 bytes (the SDK model max), not 20 KB =
    20000; 19000/20000/20480 accepted, 20481 and 21000 rejected with wire code ValidationException (HTTP 400) -
    not BadRequestException and not in the modeled error list - '1 validation error detected:...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-002.md](details/S3T-TABLEBUCKETPOLICY-002.md)

- <a id="s3t-tablebucketpolicy-014"></a>**S3T-TABLEBUCKETPOLICY-014** `quota-limit` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **PutTableBucketPolicy 20480-byte limit is applied to the RAW bytes as sent (whitespace counts)...**
  Same 60-statement document: minified (12098 B) -> HTTP 200 (read-back 12098 B); pretty-printed and
  space-padded to 20479 B -> HTTP 200; to 20480 B -> HTTP 200; to 20481 B -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
  less than or equal to 20480'; to 30000 B -> ValidationException (HTTP 400) '1 validation error detected:
  Value at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to
  20480'. Control: a minified 102-statement document of 20540 B -> ValidationException (HTTP 400) '1
  validation error detected: Value at 'resourcePolicy' failed to satisfy constraint: Member must have length
  less than or equal to 20480'.
  - ACK: is_iam_policy, terminal_codes · ops: PutTableBucketPolicy, GetTableBucketPolicy · fields:
    resourcePolicy
  - repro: Put the same policy minified and pretty-printed/padded to 20479, 20480, 20481 and 30000 bytes
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-088](#s3t-tablebucket-088), [S3T-TABLEBUCKET-089](#s3t-tablebucket-089),
    [S3T-TABLEBUCKET-090](#s3t-tablebucket-090), [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-001](table-policy.md#s3t-tablepolicy-001) · evidence:
    tablebucketpolicy/creative/size-limit-whitespace
  - notes: Extends [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002) (20480/20481 boundary) with WHAT is measured. Whitespace counts:
    a spec document that the user pretty-prints can exceed the limit while the service's own minified
    read-back of the same policy is far below it, so a controller that round-trips the read-back would...
  - full notes: [details/S3T-TABLEBUCKETPOLICY-014.md](details/S3T-TABLEBUCKETPOLICY-014.md)

## Handling gaps (bugs to file)

- [S3T-TABLEBUCKET-062](#s3t-tablebucket-062) - DeleteTableBucketPolicy is idempotent: HTTP 204 when no policy exists and on a second
  Delete (not NotFoundException) (suspected bug)
  Suspected controller bug confirmed by evidence: Reconciling an adopted or out-of-band-policied table bucket
  with spec.policy unset issues DeleteTableBucketPolicy, which returns HTTP 204 with no precondition and
  removes the live policy on the very first read ([S3T-TABLEBUCKET-062](#s3t-tablebucket-062); 51/51 reads 404 afterwards per
  [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003)), and if that policy denies the controller principal
  s3tables:DeleteTableBucketPolicy the same delete instead fails every reconcile with AccessDeniedException
  403, a code the model does not list ([S3T-TABLEBUCKET-091](#s3t-tablebucket-091)); treat nil as unmanaged and reserve the delete for
  an explicit empty document, which the API itself cannot accept as a Put (ValidationException 400 'Member
  must have length greater than or equal to 1', [S3T-TABLEBUCKET-090](#s3t-tablebucket-090)), so the empty-vs-nil distinction is the
  only non-destructive way to express removal. The empty-string-vs-nil perpetual-diff half of the suspicion is
  controller-internal and would be silent on the wire (Delete is idempotent 204), so it is not settled by API
  evidence.
  - handling_ref: `generator.yaml:229-238; generator.yaml:269-275; pkg/resource/table_bucket/hooks.go:123-138;
    test/e2e/tests/test_table_bucket.py:330-340; generator.yaml:229-234;
    pkg/resource/table_bucket/hooks.go:325-332`
- [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001) - PutTableBucketPolicy with a nonexistent role principal -> BadRequestException
  400... (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven. BadRequestException is
  terminal for TableBucket, so the transient 'Invalid principal in policy' that PutTableBucketPolicy returns
  for ~7-9 s after a referenced IAM role is created (or re-created) sets ACK.Terminal=True and the bucket is
  never reconciled again until the spec changes - the controller has the 'strands on transient' sign, not the
  'loops on malformed document' sign; classify PutTableBucketPolicy 400s by message (retry 'Invalid principal
  in policy' with a requeue, keep 'malformed' / 'Invalid Action' / 'no principals' / 'Public policies'
  terminal). code: generator.yaml:267-268 TableBucket terminal_codes = [BadRequestException] ->
  pkg/resource/table_bucket/ sdk.go:400-402 terminalAWSError returns true for every BadRequestException;
  pkg/resource/table_bucket/ hooks.go:334-342 syncPolicy returns the PutTableBucketPolicy error unchanged with
  no message inspection; manager.go:397-417 onError turns it into ackerr.Terminal; the create-time
  Synced=False nudge (templates/hooks/table_bucket/sdk_create_post_set_output.go.tpl:8-11) makes the first Put
  land seconds after CreateTableBucket, squarely inside the propagation window; runtime v0.64.0
  reconciler.go:118 uses GenerationChangedPredicate and reconciler.go:1701 returns Result{} on
  ackerr.Terminal, so no retry ever happens. finding: [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001) (BadRequestException 400
  'Invalid principal in policy' at +4.9 s, 200 at +7.2 s after CreateRole), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013) (8 x 400
  over 0-7.5 s, 200 at +8.6 s; same-name re-create), [S3T-TABLEBUCKET-088](#s3t-tablebucket-088) / [S3T-TABLEBUCKET-089](#s3t-tablebucket-089) (the permanent
  cases share the same code).
  - handling_ref: `generator.yaml:229-238; generator.yaml:269-275`
- [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012) - GetTableBucketPolicy re-shapes 7 more forms (Statement object->array, 1-elem
  Resource/Principal/Condition lists->scalar... (suspected bug)
  Suspected controller bug confirmed by evidence: not flagged; evidence-driven.
  ackcompare.IAMPolicyDocumentEqual tolerates only whitespace, key order, statement order, Statement
  object->array and one-element list collapsing; it reports a difference for a service-inserted Version
  (desired omits Version, read-back has '2008-10-17'), for Effect case folding ('allow'/'ALLOW' -> 'Allow'),
  for de-duplicated principals ([root, root] -> root), for Condition Bool false -> "false" and for account-id
  -> :root, so a TableBucket whose policy uses any of these forms carries a permanent Spec.Policy delta and
  re-issues PutTableBucketPolicy on every reconcile (resync and every spec edit) without ever converging -
  normalize desired (default Version, canonical Effect case, de-dup, stringify Bool values, expand bare
  account ids) before comparing, or compare against the read-back of the last Put. code:
  pkg/resource/table_bucket/delta.go:69-75 compares Spec.Policy with ackcompare.IAMPolicyDocumentEqual;
  runtime v0.64.0 pkg/compare/iam_policy.go:31-46 unmarshals into micahhausler/aws-iam-policy Policy and calls
  Policy.Equal, which is strict on Version (policy_statement.go:38-39) and Effect (policy_statement.go:86),
  compares string lists by sorted slice with length check (string_slice.go:95-120, so de-dup fails) and keeps
  bool and string condition values in separate buckets (condition_value.go:56-60, so false != "false");
  verified by a unit test outside the repo feeding the sent/read-back pairs through
  ackcompare.IAMPolicyDocumentEqual at runtime v0.64.0: version-inserted=false, effect-case-fold=false,
  duplicate-principals-removed=false, bool-false-stringified=false, account-id-to-root=false,
  role-arn-to-aroa=false; one-element-lists-collapsed=true, statement-object-to-array=true,
  whitespace-key-order-only=true. pkg/resource/table_bucket/hooks.go:197-251 customUpdateTableBucket returns
  the desired spec untouched, so the CR is never rewritten to the read-back and the delta recurs. The ARN ->
  AROA rewrite after a role deletion is a legitimate delta, but the re-Put then hits the terminal
  BadRequestException of [GT-S3T-010](service.md#gt-s3t-010) (controller hooks catalog entry). finding: [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012) (Statement object->array, 1-elem
  lists->scalar, [root,root]->root, Bool false->"false", Effect ALLOW->Allow, account id -> :root),
  [S3T-TABLEBUCKET-111](#s3t-tablebucket-111) (Version 2008-10-17 inserted, Effect allow->Allow), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013) (deleted
  role ARN -> AROA id).
  - handling_ref: `generator.yaml:234-238; pkg/resource/table_bucket/delta.go:69-75`

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLEBUCKET-062](#s3t-tablebucket-062) | DeleteTableBucketPolicy is idempotent: HTTP 204 when no policy exists and on a second Delete (not NotFoundException) | policy_gone_on_first_read=true | 1 |
| [S3T-TABLEBUCKET-065](#s3t-tablebucket-065) | DeleteTableBucketPolicy then PutTableBucketPolicy within 103 ms -> HTTP 200; Gets over 5 s: 200/200/200/200/200... | put_after_delete_gap_ms=103 | 1 |
| [S3T-TABLEBUCKET-081](#s3t-tablebucket-081) | 30 concurrent GetTableBucket/ListTableBuckets/ListTagsForResource/GetTableBucketPolicy (2 runs each): 0 TooManyRequests, 0 other errors | burst_wall_s.GetTableBucket=[0.16, 0.08], burst_wall_s.ListTableBuckets=[0.06, 0.05], burst_wall_s.ListTagsForResource=[0.1, 0.12], burst_wall_s.GetTableBucketPolicy=[0.15, 0.15], latency_ms_min_max.GetTableBucket=[[5,146], [7,65]], latency_ms_min_max.ListTableBuckets=[[7,44], [7,39]], latency_ms_min_max.ListTagsForResource=[[8,80], [10,100]], latency_ms_min_max.GetTableBucketPolicy=[[35,130], [29,131]] | 240 |
| [S3T-TABLEBUCKET-091](#s3t-tablebucket-091) | A policy Deny on Put/DeleteTableBucketPolicy for the caller is enforced on the first call (AccessDeniedException 403... | policy_enforcement_lag_s=0.03, puts_accepted_before_403=0 | 1 |
| [S3T-TABLEBUCKET-106](#s3t-tablebucket-106) | GetTableBucketEncryption (aws:kms set) from a role with Deny kms:* -> HTTP 200; the same role's kms:DescribeKey -> AccessDeniedException 400 | assume_role_available_after_s=10.1 | 1 |
| [S3T-TABLEBUCKET-129](#s3t-tablebucket-129) | Role with only Create/Get/DeleteTableBucket - plain CreateTableBucket 200; with tags, encryption or storageClass -> AccessDeniedException | create_table_bucket_latency_ms_allowed=1628 | 1 |
| [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001) | PutTableBucketPolicy with a nonexistent role principal -> BadRequestException 400... | first_policy_ok_since_role_create_s=[7.2], rejected_attempts_before_ok=[1], rejected_codes=[BadRequestException/400] | 1 |
| [S3T-TABLEBUCKETPOLICY-002](#s3t-tablebucketpolicy-002) | PutTableBucketPolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400) | largest_accepted_bytes=20480, smallest_rejected_bytes=20481 | 5 |
| [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003) | GetTableBucketPolicy after Put v1->v2 and after Delete, polled at 200 ms for 4 s x3: 0 stale reads in 101 polls (first read 64-68 ms) | first_poll_offset_s=[0.068, 0.064, 0.064] | 101 |
| [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005) | Bucket-policy Deny on a role takes effect on the first poll (<= 0.07 s) and stops on the first poll after DeleteTableBucketPolicy... | deny_effective_after_s=[0.07, 0.01, 0.01], allowed_again_after_s=[0.12, 0.13, 0.12], denied_polls_after_delete=[0, 0, 0] | 3 |
| [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013) | Principal role deleted then re-created (same name): read-back stays the old AROA id... | accept_after_createrole_s=[8.6, 20.7], aroa_flip_after_delete_s=8.4 | 2 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `tablebucket-policy.md`: unverified or stale items worth a future probe, one line each.

- [S3T-TABLEBUCKET-062](#s3t-tablebucket-062): the empty-string-vs-nil perpetual-diff half of the suspicion is controller-internal and silent on the wire (Delete is idempotent 204); needs a controller-level test, not an API probe.
- [S3T-TABLEBUCKET-065](#s3t-tablebucket-065): the Delete -> re-Put gap has one measurement (103 ms, n=1); a sweep at 0/10/50/100 ms would show whether any minimum exists.
- [S3T-TABLEBUCKET-081](#s3t-tablebucket-081): the 100- and 200-concurrent GetTableBucketPolicy cells were never tried (lab burst cap 30 calls / 5 s); no read throttle code has been observed on this surface.
- [S3T-TABLEBUCKET-091](#s3t-tablebucket-091): the caller-scoped Deny enforcement lag is a single run (0.03 s, n=1) and the same-name CreateTableBucket after the escape was not re-tried past the 409.
- [S3T-TABLEBUCKET-099](tablebucket-maintenance-metrics.md#s3t-tablebucket-099): whether a bucket of the probed name exists in the two foreign accounts is unknown, so "ForbiddenException regardless of existence" is inferred, not shown.
- [S3T-TABLEBUCKETPOLICY-001](#s3t-tablebucketpolicy-001): n=1 role, medium confidence, and the policy loop ran interleaved with a replication Put loop; a clean n>=3 CreateRole -> Put sweep at 1 s would fix the 'Invalid principal' window.
- [S3T-TABLEBUCKETPOLICY-005](#s3t-tablebucketpolicy-005): its behavior text calls ForbiddenException the IAM-level code; -046 is canonical (AccessDeniedException for IAM-level denials) and the remark stands as a recorded contradiction.
- [S3T-TABLEBUCKETPOLICY-006](#s3t-tablebucketpolicy-006): single run, Deny scoped to the caller's session ARN; whether a role-ARN Deny on GetTableBucketPolicy alone (no Put/Delete Deny) still blocks a Put-then-Get reconcile is untested.
- [S3T-TABLEBUCKETPOLICY-007](#s3t-tablebucketpolicy-007): single run, swapped out of the Phase 8 reverify sample by policy; the leftover bucket `ackq-9ba36e-polnorm` stays undeletable until root or AWS Support acts, so a repeat would create a second one.
- [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012): one run per cell; the "account id -> :root" rewrite appears in the notes' comparator list (unit test cell account-id-to-root=false) but not in this finding's behavior text, so its bucket-level source is not identified here.
- [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013): same-name re-creation watched for 20 s only (no re-binding seen); the AROA re-Put acceptance at +9 s contradicts the table-level +30 s rejection and is settled as a time window in [S3T-TABLEPOLICY-010](table-policy.md#s3t-tablepolicy-010) (`table-policy.md`), not re-measured at bucket level beyond FU2.
- [S3T-TABLEBUCKET-063](#s3t-tablebucket-063): "member order inside a statement preserved" was observed on an input already in canonical order; [S3T-TABLE-152](table-policy.md#s3t-table-152) shows the normalizer re-keys to Sid/Effect/Principal/Action/Resource - a bucket-level non-canonical-order Put would confirm the same engine.
- [S3T-TABLEBUCKET-090](#s3t-tablebucket-090): the UTF-8 BOM rejection contradicted the hypothesis (expected accepted) on one run; a BOM-only vs BOM+whitespace variant pair would show whether the BOM or the leading non-JSON byte triggers 'malformed'.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-tablebucket-064"></a>**S3T-TABLEBUCKET-064** | idempotency | low | confirmed | PutTableBucketPolicy is a full replace (only the last document's statements remain) and re-sending the same document returns HTTP 200 | [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-081](#s3t-tablebucket-081), [S3T-TABLEBUCKETPOLICY-009](#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), [S3T-TABLEPOLICY-006](table-policy.md#s3t-tablepolicy-006), [S3T-TABLE-153](table-policy.md#s3t-table-153), [S3T-TABLE-151](table-policy.md#s3t-table-151) | - |
| <a id="s3t-tablebucket-106"></a>**S3T-TABLEBUCKET-106** | prerequisite | low | confirmed | GetTableBucketEncryption (aws:kms set) from a role with Deny kms:* -> HTTP 200; the same role's kms:DescribeKey -> AccessDeniedException 400 | [S3T-TABLEBUCKET-034](tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLEBUCKET-035](tablebucket-encryption-storage.md#s3t-tablebucket-035), [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181) | - |
| <a id="s3t-tablebucket-163"></a>**S3T-TABLEBUCKET-163** | other | low | confirmed | Doc claim C111 TRUE: GetTableBucket on a well-formed ARN of a never-existing bucket -> NotFoundException 404, not Forbidden/AccessDenied | [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLEBUCKET-127](tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLEBUCKET-162](tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-164](tablebucket.md#s3t-tablebucket-164), [S3T-NAMESPACE-029](namespace.md#s3t-namespace-029) | - |
| <a id="s3t-tablebucketpolicy-004"></a>**S3T-TABLEBUCKETPOLICY-004** | normalization | medium | confirmed | GetTableBucketPolicy returns a normalized document: whitespace removed and a one-element Action array collapsed to a string | [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-017](#s3t-tablebucketpolicy-017), [S3T-TABLEBUCKETPOLICY-008](#s3t-tablebucketpolicy-008), [S3T-TABLEPOLICY-005](table-policy.md#s3t-tablepolicy-005), [S3T-TABLE-152](table-policy.md#s3t-table-152), [S3T-TABLE-154](table-policy.md#s3t-table-154), [S3T-TABLE-166](table-policy.md#s3t-table-166) | [S3T-TABLEBUCKET-063](#s3t-tablebucket-063) |
| <a id="s3t-tablebucketpolicy-009"></a>**S3T-TABLEBUCKETPOLICY-009** | other | low | confirmed | Doc claim C042 TRUE: Second PutTableBucketPolicy replaces the whole document: GetTableBucketPolicy shows only the last document's statements | [S3T-TABLEBUCKET-064](#s3t-tablebucket-064), [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), [S3T-TABLEPOLICY-006](table-policy.md#s3t-tablepolicy-006), [S3T-TABLE-153](table-policy.md#s3t-table-153), [S3T-TABLE-151](table-policy.md#s3t-table-151) | - |
| <a id="s3t-tablebucketpolicy-010"></a>**S3T-TABLEBUCKETPOLICY-010** | other | low | confirmed | Doc claim C043 TRUE: After DeleteTableBucketPolicy, GetTableBucketPolicy -> NotFoundException 404 'bucket policy does not exist' at once | [S3T-TABLEBUCKETPOLICY-003](#s3t-tablebucketpolicy-003), [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-061](#s3t-tablebucket-061), [S3T-TABLEBUCKETPOLICY-011](#s3t-tablebucketpolicy-011), [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLE-156](table-policy.md#s3t-table-156), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLE-149](table-policy.md#s3t-table-149) | - |
| <a id="s3t-tablebucketpolicy-011"></a>**S3T-TABLEBUCKETPOLICY-011** | other | low | confirmed | Doc claim C044 TRUE: GetTableBucketPolicy on a bucket that never had a policy -> NotFoundException 404, not 200 with an empty resourcePolicy | [S3T-TABLEBUCKET-061](#s3t-tablebucket-061), [S3T-TABLEBUCKET-092](tablebucket-replication.md#s3t-tablebucket-092), [S3T-TABLEBUCKET-097](tablebucket-replication.md#s3t-tablebucket-097), [S3T-TABLEBUCKETPOLICY-010](#s3t-tablebucketpolicy-010), [S3T-TABLEPOLICY-007](table-policy.md#s3t-tablepolicy-007), [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLE-156](table-policy.md#s3t-table-156), [S3T-TABLE-158](table-policy.md#s3t-table-158), [S3T-TABLE-149](table-policy.md#s3t-table-149), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023) | - |
| <a id="s3t-tablebucketpolicy-017"></a>**S3T-TABLEBUCKETPOLICY-017** | normalization | medium | confirmed | GetTableBucketPolicy returns a normalized document: a one-element Action array comes back as a scalar string (not byte-identical) | [S3T-TABLEBUCKET-063](#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-012](#s3t-tablebucketpolicy-012), [S3T-TABLEBUCKETPOLICY-013](#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-004](#s3t-tablebucketpolicy-004) | [S3T-TABLEBUCKET-063](#s3t-tablebucket-063) |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
