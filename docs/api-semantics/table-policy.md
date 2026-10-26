<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# Table resource policy and authorization errors
_Put/Get/DeleteTablePolicy semantics and Resource-form binding, IAM action requirements per table operation (PutTableData for schemas, GetTableData for sort compaction, kms:DescribeKey for aws:kms), AccessDeniedException vs ForbiddenException; scope verdict for TablePolicy._
Generated from aws-api-semantics `services/s3tables` (render date in the marker above); model 2018-05-10 (service/s3tables v1.16.4); controller commit 1eda357; evidence: `services/s3tables/probes/<probe id>/` in the lab repo.

## Overview

<!-- preserved:start id=overview -->
Hand-authored overview seed for `table-policy.md`; every claim cites the finding ids that carry the full evidence below.

### Orientation

This document covers the table's resource policy - PutTablePolicy, GetTablePolicy and DeleteTablePolicy addressed by
(tableBucketARN, namespace, name) - with document validation and server-side normalization, which Resource forms actually
bind, the lifecycle of an IAM principal named in a stored policy, and the authorization behavior of the whole table family:
which IAM action each CreateTable feature needs and when a 403 is AccessDeniedException versus ForbiddenException. There is
no state machine and no asynchrony: Put is a synchronous whole-document replace (only the last document's statements
remain; a byte-identical re-Put is 200, never ConflictException), Delete is 204 whether or not a policy exists, a Deny is
enforced on the first call 0.03 s after Put and lifted 0.01-0.04 s after Delete, and GetTablePolicy returned the written
document on every immediate read, 0 stale in 60 ([S3T-TABLE-153](#s3t-table-153), -151, -150, -156, -216, -158). The policy is read back only
through GetTablePolicy's single resourcePolicy member - GetTable and ListTables expose no policy key - and what comes back
is the server's re-serialization, never the sent bytes ([S3T-TABLE-159](#s3t-table-159), -152, [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003)). Scope verdict: the
findings show the API shape of a standalone resource - a Put/Get/Delete triple, NotFoundException 404 'The specified table
policy does not exist.' when absent, 204 on delete-when-absent, one independent document per (namespace, name), unchanged
by RenameTable - so a TablePolicy CRD, or a Table.spec.policy field shaped like TableBucket.spec.policy, is implementable on
this API ([S3T-TABLE-149](#s3t-table-149), -156, -150, -181, -004, [S3T-TABLEPOLICY-007](#s3t-tablepolicy-007)); the controller implements neither today (see
"Handling gaps"), and the hazards are not consistency but the normalization re-Put loop, the transient 'Invalid principal'
400, the AROA drift after DeleteRole, inert Resource forms and the lockout Deny ([S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), -002, -004, -010, -009,
[S3T-TABLE-214](#s3t-table-214)). On the IAM side the findings show: CreateTable with metadata.iceberg.schema needs s3tables:PutTableData
(without it ForbiddenException 403 'Access Denied' and nothing is created), an encryptionConfiguration needs
s3tables:PutTableEncryption (AccessDeniedException 403, even for AES256), {aws:kms K} needs the caller's kms:DescribeKey
(ForbiddenException 403 naming it) while {AES256} and no-config succeed with kms:* denied; every identity- or
resource-policy denial on the table family is AccessDeniedException 403 with the IAM 'User: ... is not authorized to perform:
s3tables:<Op>' message - a code that, of these operations, only GetTable models - whereas ForbiddenException is the
service's own 403: 'Unauthorized' for a foreign-account ARN, 'Access Denied' for the PutTableData case, the kms:DescribeKey
text for a KMS key ([S3T-TABLE-325](#s3t-table-325), -131, -140, -328, -086, -109, -132, -250, -251, -216, -026); the s3tables:GetTableData
requirement for sort/z-order compaction is evidence of the table-maintenance partition ([S3T-TABLE-234](table-maintenance.md#s3t-table-234) in
`table-maintenance.md`).

### Rules

1. Compare the policy semantically or store the Get-after-first-Put fixed point - a string compare of spec vs read-back never
   matches and re-Puts every reconcile: GetTablePolicy minifies and re-keys statements to Sid, Effect, Principal, Action,
   Resource, rewrites an account-id principal to its :root ARN, collapses one-element Principal.AWS/Action/Resource/Condition
   lists to scalars, turns a Statement object into a 1-element array, de-duplicates [root, root], stringifies Condition Bool
   false to "false", folds Effect 'allow'/'ALLOW' to 'Allow' and inserts "Version":"2008-10-17" when none was sent;
   action-name case, wildcard actions, duplicate Sids, Allow+NotAction/NotResource and a top-level Id come back as sent;
   Put(observed) reads back byte-identical at table and bucket level through the same normalizer ([S3T-TABLE-152](#s3t-table-152), -154, -166,
   [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), -005, -008).
2. Write Resource as the table's real uuid ARN (status) and warn on every other form: Put accepts and stores as sent another
   table's ARN, the bucket ARN, <bucket>/table/<name>, <bucket>/table/<ns>/<name>, "*", <arn>/*, <bucket>/table/*, [t1, t2]
   and a statement with no Resource key, but a Deny bound only for the exact uuid table ARN, <bucket>/table/*, <bucket>/*,
   "*" and a list containing the exact ARN; the name-based pseudo ARN, the bare bucket ARN and the sibling table's ARN were
   inert (Put 200, denied call 200), and a table-policy Deny never affected the sibling table ([S3T-TABLEPOLICY-009](#s3t-tablepolicy-009),
   [S3T-TABLE-167](#s3t-table-167)).
3. A Deny to Principal "*" is not rejected as public: Deny [PutTablePolicy, DeleteTablePolicy, ...] to "*" (DateLessThan
   guard) -> 200, then the caller's own PutTablePolicy and DeleteTablePolicy -> AccessDeniedException 403 from 0.03 s while
   GetTablePolicy and GetTable stay 200; the only remediation recorded is DeleteTable(versionToken) -> 204, after which
   GetTablePolicy is 404 'The specified table does not exist.'; that Deny did not name DeleteTable - the bucket-level
   Deny+NotAction twin denies every Delete and bricked a lab bucket ([S3T-TABLE-214](#s3t-table-214), -157; [S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007) in
   `tablebucket-policy.md`).
4. Observed principal AROA != desired ARN is externally caused drift the controller cannot fix by writing: after DeleteRole
   the read flips between the role ARN and its AROA unique id (first AROA at +0/+33 s, 4-9 flips, stable from +64-65 s,
   every read 200); re-Put of the observed or of the desired document -> BadRequestException 400 'Invalid principal in
   policy'; DeleteTablePolicy -> 204 only removes protection; a role re-created with the SAME name heals it (Put 200 after
   7.53 s / 7 rejections) and the stored document is bound to the role's unique id, not its name - requeue until the
   referenced role exists again ([S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), [S3T-TABLE-165](#s3t-table-165)).
5. Do not trust a 200 re-Put in the first ~20-40 s after DeleteRole: at +5/+10/+20 s both the AROA and the ARN documents
   were accepted (12/12 Puts per level, 2 roles) while reads still rendered the ARN; at +40 s reads rendered the AROA and
   7/8 Puts were 400, at +90 s 8/8 - time since DeleteRole decides, not bucket vs table level, so the next reconcile cannot
   reproduce that 200 ([S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)).
6. Put is a whole-document replace with no statement merge and no discovery path: Put(Sids A,B) then Put(Sid C) -> Get has
   Sids ['C']; a byte-identical second Put is 200/200 with identical read-backs; DeleteTablePolicy is 204 on a table that
   never had a policy and 204 twice back-to-back with one present, then Get -> 404; wire shapes are Put 200 and Delete 204
   with empty bodies and Get {resourcePolicy} only; same-named tables in two namespaces carry independent policies, and after
   RenameTable the policy read by the new triple is unchanged and echoes the unchanged table ARN ([S3T-TABLE-153](#s3t-table-153), -151, -150,
   -156, -159, -181, -004, [S3T-TABLEPOLICY-006](#s3t-tablepolicy-006), -007).
7. BadRequestException 400 'Invalid principal in policy' is permanent for a nonexistent role or user ARN, a Federated cognito
   principal or another account's :root, but transient for a role created moments ago: PutTablePolicy right after
   CreateRole was rejected 8 times at 1/s with that exact code and message before the first 200 at 8.57-8.66 s (n=3),
   indistinguishable by code or text; the finding's decision rule is to retry it for ~30 s before surfacing it as terminal
   and to resolve a Role reference instead of marking the policy failed ([S3T-TABLEPOLICY-002](#s3t-tablepolicy-002), [S3T-TABLE-164](#s3t-table-164), -157, -165).
8. Two 403 codes with different causes: IAM identity-policy denials (explicit Deny or no Allow) and resource-policy Denies
   surface as AccessDeniedException 403 'User: ... is not authorized to perform: s3tables:<Op> on resource: <ARN> ...' on
   GetTable, GetTableMetadataLocation, UpdateTableMetadataLocation, ListTables, RenameTable, Tag/UntagResource,
   GetTableEncryption/StorageClass, Put/DeleteTablePolicy and CreateTable - of these only GetTable models that code (body
   {"Message"} only, code from the x-amzn-ErrorType header); ForbiddenException 403 is the service's own: 'Unauthorized' for a
   foreign-account table ARN, 'Access Denied' for CreateTable-with-schema lacking s3tables:PutTableData, the kms:DescribeKey
   text for {aws:kms} ([S3T-TABLE-109](#s3t-table-109), -132, -250, -251, -216, -214, -142, -026, -131, -325, -328, -086, -083).
9. Authorization precedes the existence check: without the action, GetTable, GetTableEncryption and DeleteTable on a MISSING
   name -> AccessDeniedException naming '<bucket ARN>/table/*', not 404; the same missing name in scope -> NotFoundException;
   CreateTable of an EXISTING name by a role lacking s3tables:CreateTable -> ConflictException 409, a new name ->
   AccessDeniedException 403 ([S3T-TABLE-132](#s3t-table-132), -109, -142 n=1).
10. IAM actions per CreateTable feature: metadata.iceberg.schema needs s3tables:PutTableData - without it ForbiddenException
    403 'Access Denied' (402-449 ms), an immediate retry the same, admin GetTable 404 and ListTables without the name (nothing
    created); with it 200 with a metadataLocation (3/3); properties-only metadata -> 400 'The specified metadata is not
    valid' before any authorization ([S3T-TABLE-325](#s3t-table-325), -131). An encryptionConfiguration ({AES256} or {aws:kms}) needs
    s3tables:PutTableEncryption - Deny -> AccessDeniedException 403 naming the BUCKET ARN; no config -> 200;
    storageClassConfiguration with PutTableStorageClass denied -> 200 ([S3T-TABLE-140](#s3t-table-140), -328). {aws:kms K} needs the caller's
    kms:DescribeKey - ForbiddenException 403 naming it, also with only kms:GenerateDataKey+Decrypt allowed on K - while
    {AES256} and no-config -> 200 with kms:* denied, in an aws:kms-default bucket too (the table inherits K); GetTableEncryption
    needs no KMS permission ([S3T-TABLE-328](#s3t-table-328), 3 roles x 6 cells; [S3T-TABLE-141](#s3t-table-141) is the superseded reading, confidence low). A
    missing (own account or other region) or malformed key ARN -> BadRequestException 400, a foreign-account key ->
    ForbiddenException 403, never NotFoundException, no table created ([S3T-TABLE-086](#s3t-table-086)).
11. GetTablePolicy NotFoundException 404 has two meanings and only the message tells them apart: 'The specified table policy
    does not exist.' (never set, and after Delete) vs 'The specified table does not exist.' (table missing, also for an
    upper-cased name, a hyphenated namespace, a dotted name or an empty path segment) and 'The specified namespace does not
    exist.' (upper-cased namespace); a never-set policy is 404, not 200 with an empty document ([S3T-TABLE-149](#s3t-table-149), -156, -181,
    -184, [S3T-TABLEPOLICY-007](#s3t-tablepolicy-007)).
12. Validation at Put is partial, so classify by message: BadRequestException 400 'The specified policy is malformed'
    (truncated JSON, {}, no Statement, Statement [], Version 2030-01-01, bare-string Principal, Condition operator 'bool'),
    'Invalid Action' (s3:GetObject, 'S3Tables:' prefix), 'Policy contains a statement with no principals', 'Invalid service
    principal in policy' and 'Public policies cannot be applied to S3 table buckets and tables.' (Allow to "*" or {"AWS":"*"}
    without Condition, "*" with only aws:SourceIp, [role, "*"], Allow+NotPrincipal) are permanent; accepted with 200 (read
    back per rule 1): no Version, 's3tables:NoSuchAction', 's3tables:*', 's3tables:Get*', duplicate Sids, Sid "", Effect 'allow',
    NotPrincipal Deny, Allow+NotAction/NotResource, Deny "*", Allow "*" with an aws:PrincipalArn or aws:PrincipalAccount
    Condition, Service s3.amazonaws.com, an sts assumed-role session ARN, the own :root; the size limit is 20480 bytes -
    20481 -> ValidationException 400 (not BadRequestException, unmodeled, not enforced client-side) and the stored document
    survived the rejected Puts ([S3T-TABLE-162](#s3t-table-162), -157, -164, -167, -154, [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008), -001).
13. Do not poll for eventual consistency - 0 stale reads in 60 immediate reads 86-233 ms after Put/Put/Delete, a Deny
    enforced on the first call 0.03 s after Put (n=3) and lifted 0.01-0.04 s after Delete (n=2); the only waits are IAM-side:
    a new role becomes a valid principal 0.14-10.45 s after CreateRole and AssumeRole works 0-10.1 s after CreateRole
    ([S3T-TABLE-158](#s3t-table-158), -216, -251, [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002), -010, [S3T-TABLE-165](#s3t-table-165), -131, -325, -250).

### Timing ranges

- GetTablePolicy read-after-write: Get issued 86-233 ms after Put(A) / Put(B) / Delete returned, 0 stale reads in 20 cycles
  x 3 phases, n=60 ([S3T-TABLE-158](#s3t-table-158)).
- Table-policy Deny: enforced on the first call 0.03/0.03/0.03 s after PutTablePolicy returned (n=3) and lifted 0.01-0.04 s
  after DeleteTablePolicy (n=2), no propagation window at 1 s resolution ([S3T-TABLE-216](#s3t-table-216)); the same 0.03 s first-call figure
  recurs for the lockout Deny ([S3T-TABLE-214](#s3t-table-214), n=1) and for a bucket-policy Deny on GetTableEncryption ([S3T-TABLE-251](#s3t-table-251), n=1).
- Call latencies: PutTablePolicy 105-112 ms, n=2 ([S3T-TABLE-151](#s3t-table-151)); CreateTable with metadata.iceberg.schema ForbiddenException
  at 402-449 ms without s3tables:PutTableData and 200 at 401-486 ms with it, n=3 each, properties-only metadata rejected at
  7-8 ms ([S3T-TABLE-325](#s3t-table-325)).
- New role as a policy principal (first accepted PutTablePolicy after CreateRole): 8.57-8.66 s with 8 rejections per trial
  at 1/s, n=3 ([S3T-TABLEPOLICY-002](#s3t-tablepolicy-002)); 10.45 s / 6 attempts, n=1 ([S3T-TABLE-165](#s3t-table-165)); 7.53 s / 7 rejections after re-creating the
  same-name role, n=1 ([S3T-TABLEPOLICY-004](#s3t-tablepolicy-004)); 0.14-1.23 s at table level and 7.59-7.69 s at bucket level, n=2 roles
  ([S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)) - 0.14-10.45 s across the four findings.
- After DeleteRole: AROA first seen at +0 to +33 s, read stable from +64-65 s after 4-9 flips, 1 poll/s for 75 s, n=2
  ([S3T-TABLEPOLICY-004](#s3t-tablepolicy-004)); ARN at +0/+5 s and AROA at +15/+30 s, n=1 ([S3T-TABLE-165](#s3t-table-165)); re-Puts of the AROA and the ARN
  documents 200 at +5/+10/+20 s, 400 for 7/8 at +40 s and 8/8 at +90 s, n=2 roles x 2 levels ([S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)).
- IAM propagation on the caller side: AssumeRole possible at the first attempt (0.0 s, n=3, [S3T-TABLE-325](#s3t-table-325)), after 9.2 s
  (n=1, [S3T-TABLE-250](#s3t-table-250)) and 10.1 s (n=1, [S3T-TABLE-131](#s3t-table-131)) - 0-10.1 s; an inline Deny effective 0.06 s after the first assume
  (n=1, [S3T-TABLE-250](#s3t-table-250)); the kms:DescribeKey matrix waited 60.8-67.6 s after PutRolePolicy before its first call (n=3 roles,
  [S3T-TABLE-328](#s3t-table-328)) - that wait is the probe's design, not a measured minimum.

### Handling gaps

- None: no finding in this partition carries `handling: suspect-bug` (0 HANDLING-GAP lines in the input; 18 of the 43 are
  `handled`, 25 `unhandled`). What the controller does today: Table models no policy field and issues no
  Put/Get/DeleteTablePolicy (generator.yaml Table resource; pkg/resource/table/hooks.go contains no policy, KMS or
  authorization code); its exceptions map 404 -> NotFoundException and terminal_codes [BadRequestException]
  (generator.yaml:99-104), the [GT-S3T-039](service.md#gt-s3t-039) (controller hooks catalog entry) basis every `handled` finding here cites. TableBucket carries the pattern a table
  policy would mirror: spec.policy with is_iam_policy from GetTableBucketPolicy and no late_initialize because nil is treated
  as delete (generator.yaml:229-238).
- The 25 `unhandled` findings are unimplemented scope, not defects, and are not yet handled: normalization ([S3T-TABLE-152](#s3t-table-152),
  -154, -166, [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), -005, -008), principal lifecycle ([S3T-TABLE-165](#s3t-table-165), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), -010), Resource
  binding ([S3T-TABLEPOLICY-009](#s3t-tablepolicy-009), [S3T-TABLE-167](#s3t-table-167)), the lockout Deny ([S3T-TABLE-214](#s3t-table-214)), IAM prerequisites ([S3T-TABLE-325](#s3t-table-325), -328,
  -141), enforcement and consistency timing ([S3T-TABLE-158](#s3t-table-158), -216, -251), idempotency and replace ([S3T-TABLE-150](#s3t-table-150), -151, -153,
  [S3T-TABLEPOLICY-006](#s3t-tablepolicy-006), -007), identity and wire shape ([S3T-TABLE-004](#s3t-table-004), -159).
- Two records are superseded readings kept as evidence (confidence low) - read their titles with the contradiction notes:
  [S3T-TABLE-141](#s3t-table-141) attributes its 403 to missing KMS permissions, settled by [S3T-TABLE-328](#s3t-table-328) (the 403 came from a lingering Deny
  s3tables:PutTableEncryption; a fresh Deny-kms:* role gets 200 for AES256 and ForbiddenException naming kms:DescribeKey for
  aws:kms); [S3T-TABLE-142](#s3t-table-142)'s 200 on GetTableEncryption/StorageClass with only s3tables:GetTable did not reproduce (30/30
  AccessDeniedException in [S3T-TABLE-327](table-maintenance.md#s3t-table-327), `table-maintenance.md`).
- If a Table.spec.policy field follows the TableBucket pattern, the bucket-level suspect-bug findings describe behavior this
  partition records for tables too - the transient 'Invalid principal' 400 under terminal_codes ([S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001) /
  here [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002)), the comparator's blind spots ([S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) / here [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008)), the
  unconditional 204 delete of an unset spec ([S3T-TABLEBUCKET-062](tablebucket-policy.md#s3t-tablebucket-062) / here [S3T-TABLE-150](#s3t-table-150), -156); see `tablebucket-policy.md`
  "Handling gaps".

### Where next

- `tablebucket-policy.md` is the bucket-level twin: the same normalizer (Put(observed) byte-identical at both levels,
  [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003)), the Deny+NotAction bricking policy and the undeletable lab bucket ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007)), the
  bucket-side DeleteRole / re-create record ([S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), whose contradiction with [S3T-TABLE-165](#s3t-table-165) and
  [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004) is settled here by [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)), the first-sync deletion hazard and the controller's syncPolicy
  handling gaps; which bucket-policy Resource forms bind tables is in [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009) here, and a bucket-policy Deny on a
  table sub-read in [S3T-TABLE-251](#s3t-table-251).
- `table-maintenance.md` carries the rest of the IAM-action evidence: every table sub-resource Get needs its own action
  ([S3T-TABLE-327](table-maintenance.md#s3t-table-327), 30/30), Get/Put/DeleteTablePolicy without the action -> AccessDeniedException 403 ([S3T-TABLE-233](table-maintenance.md#s3t-table-233)) and the
  s3tables:GetTableData requirement for sort/z-order compaction ([S3T-TABLE-234](table-maintenance.md#s3t-table-234)) - the GetTableData clause of this document's
  description lives there; GetTableMaintenanceConfiguration is the probe target of the Denies in [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009),
  [S3T-TABLE-214](#s3t-table-214) and -216.
- `table-encryption-storage.md` covers KMS key state and the inheritance of bucket defaults on CreateTable ([S3T-TABLE-086](#s3t-table-086),
  -140, -328 landed here through their authorization match); `table-metadata-schema.md` covers the metadata.iceberg write
  path and UpdateTableMetadataLocation that s3tables:PutTableData and the AccessDeniedException of [S3T-TABLE-132](#s3t-table-132) guard;
  `table.md` has the 404 message shapes, RenameTable (context of [S3T-TABLE-004](#s3t-table-004)) and the identifier forms of [S3T-TABLE-026](#s3t-table-026).
- Long catalogues belong in `details/<id>.md` rather than inline: the 17-document validation matrix ([S3T-TABLE-162](#s3t-table-162)), the
  principal forms ([S3T-TABLE-164](#s3t-table-164)), the 25-variant re-shaped/identical/rejected matrix ([S3T-TABLEPOLICY-008](#s3t-tablepolicy-008)), the 8 table + 5
  bucket Resource forms ([S3T-TABLEPOLICY-009](#s3t-tablepolicy-009)), the 3 roles x 6 cells kms/PutTableEncryption matrix with messages
  ([S3T-TABLE-328](#s3t-table-328)), the +5/+10/+20/+40/+90 s re-Put matrix ([S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)), the IAM message catalogue ([S3T-TABLE-109](#s3t-table-109),
  -132, -250), the KMS key-ARN variants ([S3T-TABLE-086](#s3t-table-086)) and the tableArn variants ([S3T-TABLE-026](#s3t-table-026)).
- Open experiments (STATUS.md): nothing running or blocked. The leftover bucket `ackq-9ba36e-polnorm` (namespace
  `ackq_9ba36e_ns`, table `ackq_9ba36e_t`) carries a root-principal Deny+NotAction table policy as well as the bucket policy
  and is undeletable by any IAM principal until the root user or AWS Support acts ([S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007)); the table-level
  lockout of [S3T-TABLE-214](#s3t-table-214) did not name DeleteTable and left that exit open, the NotAction form does not. Phase 8 follow-ups
  settled three contradictions that render here: the AROA re-Put window is time since DeleteRole (FU2-02 ->
  [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010)), CreateTable aws:kms needs the caller's kms:DescribeKey (FU2-07 -> [S3T-TABLE-328](#s3t-table-328)), sub-resource Gets
  each need their own action (FU2-08 -> [S3T-TABLE-327](table-maintenance.md#s3t-table-327)). Single-run and unverified items are in `table-policy.open-questions.md`.
<!-- preserved:end -->

## At a glance

- canonical findings: 43 (high 21 / medium 17 / low 5); duplicates folded into the appendix: 2
- handling: handled 18 · partial 0 · tracked 0 · unhandled 25 · suspect-bug 0 · n-a 0 (tracked = handled/partial whose reference is an open GitHub issue; counted as not handled)
- re-verified: 2 · last_verified: 2026-10-13..2026-10-15 · model: 2018-05-10 (service/s3tables v1.16.4)
- categories: error-code 11, normalization 6, prerequisite 6, request-validation 5, idempotency 3,
  eventual-consistency 2, identity 2, other 2, sub-resource-api 2, delete-semantics 1, quota-limit 1,
  response-fidelity 1, server-default 1

## Operations

| operation | kind | required inputs | declared error shapes | paginated |
| --- | --- | --- | --- | --- |
| CreateTable | create | tableBucketARN, namespace, name, format | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTable | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| DeleteTablePolicy | delete | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTable | read | - | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucket | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableBucketPolicy | read | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableEncryption | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| GetTableMaintenanceConfiguration | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableMetadataLocation | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTablePolicy | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| GetTableStorageClass | read | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, AccessDeniedException, TooManyRequestsException, BadRequestException | no |
| ListTables | list | tableBucketARN | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | yes |
| PutTableBucketPolicy | create | tableBucketARN, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| PutTablePolicy | create | tableBucketARN, namespace, name, resourcePolicy | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| RenameTable | update | tableBucketARN, namespace, name | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| TagResource | tag | resourceArn, tags | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UntagResource | tag | resourceArn, tagKeys | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |
| UpdateTableMetadataLocation | update | tableBucketARN, namespace, name, versionToken, metadataLocation | InternalServerErrorException, ForbiddenException, NotFoundException, TooManyRequestsException, ConflictException, BadRequestException | no |

## State machine

- **StorageClass**: STANDARD, INTELLIGENT_TIERING (transitional: INTELLIGENT_TIERING)

No high/medium state-machine findings.

## Identity and lookup

- <a id="s3t-table-004"></a>**S3T-TABLE-004** `identity` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **After RenameTable(newName) the 4 table sub-resource reads by the new triple are unchanged**
  GetTablePolicy, GetTableEncryption, GetTableStorageClass, GetTableMetadataLocation by (bucket, ns, newName)
  -> get_table_policy HTTP 200, get_table_encryption HTTP 200, get_table_storage_class HTTP 200,
  get_table_metadata_location HTTP 200. Field diffs vs the pre-rename reads: none.
  warehouseLocation/metadataLocation do not contain the name (paths are UUID based).
  - ACK: none · ops: RenameTable, GetTablePolicy, GetTableEncryption, GetTableStorageClass,
    GetTableMetadataLocation · fields: resourcePolicy, encryptionConfiguration, storageClassConfiguration,
    metadataLocation, warehouseLocation
  - repro: CreateTable(schema) + PutTablePolicy -> snapshot 4 Get* -> RenameTable(newName) -> re-read by new
    triple -> diff
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-002](table-metadata-location.md#s3t-table-002), [S3T-TABLE-032](table.md#s3t-table-032), [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLE-163](#s3t-table-163) · hypotheses:
    H-T-053 · evidence: table/identity/rename-identity
  - notes: H-T-053 confirmed. The policy document echoes the (unchanged) table ARN.

## Idempotency

- <a id="s3t-table-150"></a>**S3T-TABLE-150** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **DeleteTablePolicy on a table that never had a policy -> HTTP 204**
  Fresh table, no policy ever set: DeleteTablePolicy -> HTTP 204. (Second-delete-after-Put behavior is
  recorded separately below.)
  - ACK: custom_delete, exceptions.404 · ops: DeleteTablePolicy
  - repro: CreateTable -> DeleteTablePolicy
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-062](tablebucket-policy.md#s3t-tablebucket-062), [S3T-TABLEBUCKETPOLICY-015](tablebucket-policy.md#s3t-tablebucketpolicy-015), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-156](#s3t-table-156),
    [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLE-153](#s3t-table-153), [S3T-TABLE-158](#s3t-table-158) · hypotheses: H-U-012, H-U-013 · evidence:
    table/sub-resources/policy-lifecycle
  - notes: H-U-013 confirmed (idempotent 204), H-U-012 refuted for the never-set case.

- <a id="s3t-table-151"></a>**S3T-TABLE-151** `idempotency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy twice with a byte-identical document -> HTTP 200 then HTTP 200; read-backs identical**
  PutTablePolicy(P1) -> HTTP 200; PutTablePolicy(P1) again -> HTTP 200 (no ConflictException). GetTablePolicy
  after each returned identical resourcePolicy strings.
  - ACK: none · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy(P1); GetTablePolicy; PutTablePolicy(P1); GetTablePolicy
  - measurements: put_latency_ms=[105, 112], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-009](tablebucket-policy.md#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKET-064](tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEBUCKETPOLICY-003](tablebucket-policy.md#s3t-tablebucketpolicy-003), [S3T-TABLEPOLICY-006](#s3t-tablepolicy-006),
    [S3T-TABLE-153](#s3t-table-153), [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-150](#s3t-table-150), [S3T-TABLE-156](#s3t-table-156), [S3T-TABLE-158](#s3t-table-158) · hypotheses: H-U-003 ·
    evidence: table/sub-resources/policy-lifecycle
  - notes: H-U-003 confirmed.

- <a id="s3t-table-156"></a>**S3T-TABLE-156** `idempotency` · impact medium · handled · verified 2026-10-13
  **DeleteTablePolicy with a policy present -> HTTP 204; immediately again -> HTTP 204; GetTablePolicy after -> NotFoundException (HTTP 404)**
  With a policy stored: DeleteTablePolicy -> HTTP 204; second DeleteTablePolicy back-to-back -> HTTP 204;
  GetTablePolicy -> NotFoundException (HTTP 404) 'The specified table policy does not exist.'.
  - ACK: custom_delete · ops: DeleteTablePolicy, GetTablePolicy
  - repro: PutTablePolicy; DeleteTablePolicy; DeleteTablePolicy; GetTablePolicy
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-062](tablebucket-policy.md#s3t-tablebucket-062), [S3T-TABLEBUCKETPOLICY-015](tablebucket-policy.md#s3t-tablebucketpolicy-015), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-150](#s3t-table-150),
    [S3T-TABLEBUCKETPOLICY-010](tablebucket-policy.md#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEPOLICY-007](#s3t-tablepolicy-007), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061),
    [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLE-158](#s3t-table-158), [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLE-153](#s3t-table-153) · hypotheses: H-U-012,
    H-U-013 · evidence: table/sub-resources/policy-lifecycle
  - notes: Second delete succeeded. H-U-013 (idempotent delete) confirmed; H-U-012 refuted.

## Errors

- <a id="s3t-table-026"></a>**S3T-TABLE-026** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable(tableArn): foreign-account ARN -> ForbiddenException, own-account unknown uuid -> NotFoundException...**
  GetTable(tableArn=...) on variants of a real table ARN: foreign_account_random_uuid -> ForbiddenException
  (HTTP 403) 'Unauthorized'; foreign_account_real_uuid -> ForbiddenException (HTTP 403) 'Unauthorized';
  own_account_random_uuid -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  real_uuid_other_existing_bucket -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  real_uuid_nonexistent_bucket -> NotFoundException (HTTP 404) 'The specified table does not exist.';
  bucket_arn_as_table_arn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.';
  not_an_arn -> BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; trailing_slash ->
  BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; uppercase_uuid ->
  BadRequestException (HTTP 400) 'The specified table ARN is not valid.'; uuid_only -> BadRequestException
  (HTTP 400) 'The specified table ARN is not valid.'; s3_bucket_arn -> BadRequestException (HTTP 400) 'The
  specified table ARN is not valid.'; empty_string -> ParamValidationError (HTTP None) 'Parameter validation
  failed:
  Invalid length for parameter tableArn, value: 0, valid min length: 1'.
  - ACK: terminal_codes, exceptions.404, is_arn_primary_key · ops: GetTable · fields: tableArn
  - repro: Take arn(t1); rewrite account / uuid / bucket segment / strip to bucket ARN / garbage string;
    GetTable(tableArn=variant) for each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLEBUCKET-066](tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-TABLE-081](table.md#s3t-table-081), [S3T-TABLEREPLICATION-023](table-replication.md#s3t-tablereplication-023),
    [S3T-TABLEREPLICATION-024](table-replication.md#s3t-tablereplication-024), [S3T-TABLEREPLICATION-013](table-replication.md#s3t-tablereplication-013), [S3T-TABLEREPLICATION-028](table-replication-status.md#s3t-tablereplication-028),
    [S3T-TABLEBUCKETREPLICATION-012](tablebucket-replication.md#s3t-tablebucketreplication-012), [S3T-TABLE-083](#s3t-table-083) · hypotheses: H-T-008, H-T-009, H-T-010 · evidence:
    table/identity/get-table-identifier-forms
  - notes: H-T-008 confirmed (foreign account -> ForbiddenException, own account random uuid ->
    NotFoundException). H-T-009 confirmed (real uuid under another existing bucket -> NotFoundException; under
    a nonexistent bucket -> NotFoundException). H-T-010 confirmed (bucket ARN as tableArn ->...
  - full notes: [details/S3T-TABLE-026.md](details/S3T-TABLE-026.md)

- <a id="s3t-table-083"></a>**S3T-TABLE-083** `error-code` · impact medium · handled · verified 2026-10-13
  **Wire error codes equal the modeled shape names (BadRequest=400, Forbidden=403, NotFound=404)...**
  Across 32 rejected calls in this probe: BadRequestException: HTTP 400, header x-amzn-ErrorType
  ['BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'], body __type ['None'],
  body keys ['["message"]'] (n=17); ForbiddenException: HTTP 403, header x-amzn-ErrorType
  ['ForbiddenException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'], body __type ['None'],
  body keys ['["message"]'] (n=1); NotFoundException: HTTP 404, header x-amzn-ErrorType
  ['NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'], body __type ['None'],
  body keys ['["message"]'] (n=14). No ValidationException or ResourceNotFoundException was emitted.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTable, GetTable, DeleteTable
  - repro: Capture x-amzn-ErrorType and the raw body on every rejected CreateTable/GetTable/DeleteTable in
    this probe
  - measurements: rejected_calls=32, n=32
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKET-067](tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-120](table.md#s3t-table-120),
    [S3T-TABLE-026](#s3t-table-026), [S3T-TABLE-081](table.md#s3t-table-081) · hypotheses: H-T-047 · evidence: table/error-taxonomy/create-validation
  - notes: H-T-047 confirmed for the codes reached here (BadRequestException, NotFoundException,
    ForbiddenException); ConflictException/TooManyRequestsException/ForbiddenException are covered by the
    idempotency and identity probes.
  - full notes: [details/S3T-TABLE-083.md](details/S3T-TABLE-083.md)

- <a id="s3t-table-086"></a>**S3T-TABLE-086** `error-code` · impact high · handled · verified 2026-10-13
  **CreateTable(aws:kms): missing key (own acct or other region), malformed ARN -> BadRequestException 400; foreign-account key -> Forbidden 403**
  zero-UUID key in this account/region -> BadRequestException (HTTP 400) 'The specified KMS key ARN does not
  exist.'; random-UUID key in this account/region -> BadRequestException (HTTP 400) 'The specified KMS key ARN
  does not exist.'; key in another (non-owned) account, same region -> ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<acct>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on this resource
  because the resource does not exist in this Region, no resource-based policies allow access, or a
  resource-based policy explicitly denies access'; non-existent key ARN in us-east-1 (other region) ->
  BadRequestException (HTTP 400) 'Cross region KMS keys are not supported.'; malformed ARN (no key/ prefix) ->
  BadRequestException (HTTP 400) 'The specified KMS key ARN is invalid. Only key ARN format
  'arn:<partition>:kms:<region>:<account-id>:key/<key-id>' is supported.'. None of these is s3tables
  NotFoundException; no table is created.
  - ACK: terminal_codes, exceptions.404 · ops: CreateTable, GetTable · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: CreateTable(aws:kms, kmsKeyArn=<non-existent ARN variant>) -> GetTable on the triple
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-035](tablebucket-encryption-storage.md#s3t-tablebucket-035), [S3T-TABLEBUCKET-034](tablebucket-encryption-storage.md#s3t-tablebucket-034), [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-247](#s3t-table-247) · hypotheses:
    H-T-161, H-T-160 · evidence: table/weird-inputs/encryption-storage-enums
  - notes: H-T-161 confirmed (not NotFoundException). H-T-160 only partially touched: a well-formed but
    non-existent key ARN in us-east-1 gives the same class of error as a non-existent same-region key (no real
    cross-region key was created - cross-region is outside this shard).
  - full notes: [details/S3T-TABLE-086.md](details/S3T-TABLE-086.md)

- <a id="s3t-table-109"></a>**S3T-TABLE-109** `error-code` · impact medium · handled · verified 2026-10-13
  **GetTable under an IAM explicit Deny or with no s3tables permissions -> AccessDeniedException 403 (IAM message), NOT ForbiddenException**
  Role with Allow s3tables:* plus explicit Deny s3tables:GetTable: get_composite -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34 with
  an explicit deny in an identity-based policy'; get_arn -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34 with
  an explicit deny in an identity-based policy'; get_nonexistent_composite -> AccessDeniedException (HTTP 403)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/* with an explicit deny in an
  identity-based policy'; get_table_bucket -> HTTP 200; list_tables -> HTTP 200; get_table_metadata_location
  -> HTTP 200; delete_nonexistent -> NotFoundException (HTTP 404) 'The specified table does not exist.'. Role
  with no s3tables permissions: get_composite -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTable action'; get_arn -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTable action'; get_nonexistent_composite ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/*
  because no identity-based policy allows the s3tables:GetTable action'; get_table_bucket ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableBucket on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam
  because no identity-based policy allows the s3tables:GetTableBucket action'; list_tables ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTables on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam
  because no identity-based policy allows the s3tables:ListTables action'; get_table_metadata_location ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/7eb65282-cb76-43df-99f0-fb953a993c34
  because no identity-based policy allows the s3tables:GetTableMetadataLocation action'; delete_nonexistent ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:DeleteTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-762338-iam/table/* because no identity-based policy allows
  the s3tables:DeleteTable action'. Control (probe principal) GetTable -> HTTP 200. x-amzn-ErrorType header on
  the deny GetTable: AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/. The
  IAM-denied error differs from the service's own 403: header
  x-amzn-ErrorType=AccessDeniedException:...com.amazon.coral.service/ with body key 'Message' ('User:
  <principal> is not authorized to perform: s3tables:GetTable on resource: <table ARN> with an explicit deny
  in an identity-based policy' / '... because no identity-based policy allows the s3tables:GetTable action'),
  whereas a foreign-account ARN yields ForbiddenException 'Unauthorized' (header namespace
  com.amazon.s3tablebuckets, body key 'message'). With no permissions, DeleteTable of a nonexistent table ->
  AccessDeniedException (auth precedes the existence check); with the Deny-on-GetTable role it ->
  NotFoundException.
  - ACK: terminal_codes · ops: GetTable, GetTableBucket, ListTables, GetTableMetadataLocation, DeleteTable ·
    fields: tableArn, tableBucketARN, namespace, name
  - repro: CreateRole (trust: own account) with inline Allow s3tables:* + Deny s3tables:GetTable; AssumeRole;
    GetTable composite and by ARN; same with a role lacking s3tables permissions
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](#s3t-table-026),
    [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLE-132](#s3t-table-132), [S3T-TABLE-140](#s3t-table-140), [S3T-TABLE-142](#s3t-table-142), [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109),
    [S3T-TABLEBUCKET-067](tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](#s3t-table-083), [S3T-TABLE-120](table.md#s3t-table-120) · hypotheses: H-T-044 · evidence:
    table/error-taxonomy/iam-deny-forbidden
  - notes: H-T-044 REFUTED: IAM denials (explicit Deny and no-Allow) surface as AccessDeniedException, not
    ForbiddenException; ForbiddenException 'Unauthorized' is the code for a cross-account/foreign ARN (see the
    identifier-forms and create-validation probes). A controller's 'insufficient permissions'...
  - full notes: [details/S3T-TABLE-109.md](details/S3T-TABLE-109.md)

- <a id="s3t-table-131"></a>**S3T-TABLE-131** `error-code` · impact high · handled · verified 2026-10-15, re-verified
  **CreateTable WITH metadata by a role lacking s3tables:PutTableData -> ForbiddenException 403...**
  Role policy: Allow s3tables:CreateTable, GetTable, DeleteTable on the table bucket and bucket/*; nothing
  else. CreateTable(format=ICEBERG, no metadata) by the role -> HTTP 200 (attempts while IAM propagated:
  ['HTTP 200']). CreateTable with metadata.iceberg.schema -> ForbiddenException (HTTP 403) 'Access Denied'.
  Admin GetTable on that name -> NotFoundException 404; admin ListTables contains it: False. Immediate retry
  by the role -> ForbiddenException 403. Admin CreateTable of the same name with the same metadata -> HTTP 200.
  Role CreateTable with metadata={iceberg:{properties:{k:v}}} (no schema) -> BadRequestException (HTTP 400)
  'The specified metadata is not valid'.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTable, ListTables · fields: metadata,
    metadata.iceberg.schema
  - repro: IAM role with s3tables:CreateTable+GetTable+DeleteTable only -> AssumeRole ->
    CreateTable(metadata.iceberg.schema) -> admin GetTable/ListTables -> admin CreateTable same name
  - measurements: assume_role_wait_s=10.1, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLE-026](#s3t-table-026), [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-132](#s3t-table-132), [S3T-TABLE-140](#s3t-table-140), [S3T-TABLE-142](#s3t-table-142), [S3T-TABLEBUCKET-129](tablebucket-policy.md#s3t-tablebucket-129) ·
    hypotheses: H-T-105, H-T-106 · evidence: table/error-taxonomy/iam-putdata-denied
  - notes: H-T-105 confirmed: ForbiddenException (HTTP 403) with the terse message 'Access Denied' (not the
    IAM-style 'User ... is not authorized' text that identity-policy denials produce for other operations);
    admin GetTable -> NotFoundException 404 and ListTables does not list the name. H-T-106 REFUTED:...
  - full notes: [details/S3T-TABLE-131.md](details/S3T-TABLE-131.md)

- <a id="s3t-table-132"></a>**S3T-TABLE-132** `error-code` · impact high · handled · verified 2026-10-13
  **IAM identity-policy denials are AccessDeniedException 403 (IAM-style message) for GetTableMetadataLocation, UpdateTableMetadataLocation...**
  Same restricted role (CreateTable/GetTable/DeleteTable on bucket X only). GetTable on X's table -> HTTP 200.
  GetTableMetadataLocation (not granted) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx/table/718a9891-d591-494b-9e1c-aca67d534eb4
  because no identity-based policy allows the s3tables:GetTableMetadataLocation action'.
  UpdateTableMetadataLocation (not granted) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:UpdateTableMetadataLocation on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx/table/718a9891-d591-494b-9e1c-aca67d534eb4
  because no identity-based policy allows the s3tables:UpdateTableMetadataLocation action'. GetTable on bucket
  Y's table (out of scope) -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/8fc73291-2cf5-496c-9f0c-5f596914bbc9
  because no identity-based policy allows the s3tables:GetTable action'; GetTable(tableArn) of that table ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/8fc73291-2cf5-496c-9f0c-5f596914bbc9
  because no identity-based policy allows the s3tables:GetTable action'. GetTable on a MISSING table in scope
  -> NotFoundException (HTTP 404) 'The specified table does not exist.' vs out of scope ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:GetTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamy/table/*
  because no identity-based policy allows the s3tables:GetTable action'. ListTables on X (not granted) ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:ListTables on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-4ecaac-iamx
  because no identity-based policy allows the s3tables:ListTables action'.
  - ACK: terminal_codes, exceptions.404 · ops: GetTable, GetTableMetadataLocation,
    UpdateTableMetadataLocation, ListTables
  - repro: AssumeRole(restricted) -> GetTable / GetTableMetadataLocation / UpdateTableMetadataLocation /
    ListTables on in-scope and out-of-scope tables and missing names
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLE-140](#s3t-table-140), [S3T-TABLE-142](#s3t-table-142), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLE-327](table-maintenance.md#s3t-table-327) ·
    hypotheses: H-T-146 · evidence: table/error-taxonomy/iam-putdata-denied
  - notes: H-T-146 IAM part QUALIFIED/REFUTED: GetTableMetadataLocation (which models only ForbiddenException)
    returns AccessDeniedException (HTTP 403) 'User: <principal> is not authorized to perform:
    s3tables:GetTableMetadataLocation on resource: <tableARN> because no identity-based policy allows ...';...
  - full notes: [details/S3T-TABLE-132.md](details/S3T-TABLE-132.md)

- <a id="s3t-table-142"></a>**S3T-TABLE-142** `error-code` · impact high · handled · verified 2026-10-13
  **Role allowed only s3tables:GetTable: GetTableEncryption/StorageClass of an existing table -> 200; missing name -> AccessDeniedException 403**
  Role allowed only s3tables:GetTable. GetTable(existing KMS table) -> HTTP 200; GetTableEncryption -> HTTP
  200; GetTableStorageClass -> HTTP 200; GetTableEncryption on a non-existent name -> AccessDeniedException
  (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:GetTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam/table/*
  because no identity-based policy allows the s3tables:GetTableEncryption action' (authorization is evaluated
  before existence); CreateTable with the existing name -> ConflictException (HTTP 409) 'A table with an
  identical name already exists in the namespace.'; CreateTable with a new name -> AccessDeniedException (HTTP
  403, NOT in the modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized
  to perform: s3tables:CreateTable on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam
  because no identity-based policy allows the s3tables:CreateTable action'.
  - ACK: terminal_codes, hook:sdk_read_one_post_set_output · ops: GetTable, GetTableEncryption,
    GetTableStorageClass, CreateTable
  - repro: Assume a role with only s3tables:GetTable; call GetTable, GetTableEncryption, GetTableStorageClass,
    CreateTable (existing and new name)
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-132](#s3t-table-132), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLE-140](#s3t-table-140), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLE-327](table-maintenance.md#s3t-table-327) ·
    hypotheses: H-T-170 · evidence: table/dependencies/iam-permissions
  - notes: H-T-170 qualified/refuted - see behavior: sub-read deny code None (modeled on those operations:
    None); CreateTable deny code AccessDeniedException (modeled: False); x-amzn-ErrorType headers: enc 'None',
    create 'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/'.
  - full notes: [details/S3T-TABLE-142.md](details/S3T-TABLE-142.md)

- <a id="s3t-table-162"></a>**S3T-TABLE-162** `error-code` · impact high · handled · verified 2026-10-13
  **PutTablePolicy invalid documents: 9/17 variants rejected, all BadRequestException+ParamValidationError...**
  Outcome per document (code/HTTP or HTTP on success): {"truncated": "BadRequestException/400",
  "s3:GetObject": "BadRequestException/400", "no Version": "HTTP 200", "NoSuchAction": "HTTP 200",
  "s3tables:*": "HTTP 200", "\"\"": "ParamValidationError/None", "{}": "BadRequestException/400", "duplicate
  Sid": "HTTP 200", "unicode Sid": "HTTP 200", "Version 2008-10-17": "HTTP 200", "Version 2030-01-01":
  "BadRequestException/400", "Effect allow": "HTTP 200", "no Statement": "BadRequestException/400", "Statement
  []": "BadRequestException/400", "unknown keys Id/Foo": "BadRequestException/400", "no Principal":
  "BadRequestException/400", "NotPrincipal Deny": "HTTP 200"}. Messages: {"truncated": "The specified policy
  is malformed", "s3:GetObject": "Invalid Action", "\"\"": "Parameter validation failed:\nInvalid length for
  parameter resourcePolicy, value: 0, valid min length: 1", "{}": "The specified policy is malformed",
  "Version 2030-01-01": "The specified policy is malformed", "no Statement": "The specified policy is
  malformed", "Statement []": "The specified policy is malformed", "no Principal": "Policy contains a
  statement with no principals"}. Variants that were accepted are read back exactly as recorded in result.yaml
  (case_* entries).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on a fresh table with each document variant; Get+Delete after each accepted one
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-157](#s3t-table-157), [S3T-TABLE-164](#s3t-table-164), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001),
    [S3T-TABLEBUCKET-088](tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-008, H-U-027 ·
    evidence: table/error-taxonomy/policy-validation
  - notes: H-U-008 qualified (truncated=BadRequestException/400, s3:GetObject=BadRequestException/400, no
    Version=HTTP 200, NoSuchAction=HTTP 200; distinct messages=2). H-U-027: empty
    string=ParamValidationError/None, {}=BadRequestException/400, duplicate Sid=HTTP 200, unicode Sid=HTTP
    200, Version 2008=HTTP...
  - full notes: [details/S3T-TABLE-162.md](details/S3T-TABLE-162.md)

- <a id="s3t-table-250"></a>**S3T-TABLE-250** `error-code` · impact high · handled · verified 2026-10-14
  **IAM explicit Deny on RenameTable/TagResource/UntagResource/ListTables/GetTable -> AccessDeniedException 403 on all 8 calls, never Forbidden**
  Role = AdministratorAccess + inline Deny on s3tables:RenameTable, TagResource, UntagResource, ListTables,
  GetTable (assumable 9.2 s after CreateRole; deny effective on the first call). As that role, all 8 calls
  returned AccessDeniedException HTTP 403: RenameTable, TagResource(table ARN), TagResource(bucket ARN),
  UntagResource(table), ListTables(bucket), ListTables(bucket+namespace), GetTable(bucketARN+namespace+name),
  GetTable(tableArn). Message: 'User: <assumed-role ARN> is not authorized to perform: s3tables:<Action> on
  resource: <table or bucket ARN> with an explicit deny in an identity-based policy'. The wire body is only
  {"Message": ...} with no __type; the code comes from the x-amzn-ErrorType header
  'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/' (framework namespace),
  whereas service-side errors such as BadRequestException carry
  'http://internal.amazon.com/coral/com.amazon.s3tablebuckets/'. Controls as the same role: GetTableBucket,
  ListTagsForResource(table), ListNamespaces -> HTTP 200. The denied RenameTable/TagResource left the table
  name and tags unchanged (admin read).
  - ACK: terminal_codes · ops: RenameTable, TagResource, UntagResource, ListTables, GetTable
  - repro: IAM role: AdministratorAccess + Deny on the five actions; AssumeRole; RenameTable / TagResource /
    UntagResource / ListTables / GetTable on an existing table
  - measurements: create_role_to_assumable_s=9.2, deny_effective_after_first_assume_s=0.06, n=1
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-046](tablebucket-policy.md#s3t-tablebucket-046), [S3T-TABLEBUCKET-045](tablebucket-policy.md#s3t-tablebucket-045), [S3T-TABLE-109](#s3t-table-109), [S3T-TABLEBUCKET-038](tablebucket.md#s3t-tablebucket-038), [S3T-TABLE-026](#s3t-table-026),
    [S3T-TABLEBUCKET-181](tablebucket-encryption-storage.md#s3t-tablebucket-181), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLE-132](#s3t-table-132), [S3T-TABLE-140](#s3t-table-140), [S3T-TABLE-142](#s3t-table-142), [S3T-TABLEBUCKET-109](tablebucket-encryption-storage.md#s3t-tablebucket-109),
    [S3T-TABLEBUCKET-067](tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-083](#s3t-table-083), [S3T-TABLE-120](table.md#s3t-table-120) · hypotheses: H-T-066 · evidence:
    table/error-taxonomy/iam-deny-rename-tag
  - notes: H-T-066 (IAM half) confirmed: RenameTable/TagResource/UntagResource/ListTables - which model only
    ForbiddenException - return the unmodeled AccessDeniedException; GetTable (models both) also returns
    AccessDeniedException. ForbiddenException was never observed in this probe. Single role, single run;...
  - full notes: [details/S3T-TABLE-250.md](details/S3T-TABLE-250.md)

- <a id="s3t-table-251"></a>**S3T-TABLE-251** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Table-bucket-policy Deny (Principal=role, Resource=<bucket>/table/*) on GetTableEncryption -> AccessDeniedException 403 like IAM deny**
  PutTableBucketPolicy with Deny {Principal: {AWS: role A ARN}, Action: [s3tables:GetTableEncryption,
  s3tables:GetTableStorageClass], Resource: <bucketARN>/table/*} -> HTTP 200 (the /table/* resource form is
  accepted). As role A (IAM allows both actions): GetTableEncryption -> AccessDeniedException 403 and
  GetTableStorageClass -> AccessDeniedException 403, message '... is not authorized to perform:
  s3tables:GetTableEncryption on resource: <table ARN> with an explicit deny in a resource-based policy';
  x-amzn-ErrorType 'AccessDeniedException:http://internal.amazon.com/coral/com.amazon.coral.service/', body
  {"Message": ...} without __type. GetTableMaintenanceConfiguration (not in the policy) -> HTTP 200 as the
  role; GetTableEncryption as the admin caller -> HTTP 200. The deny was effective on the first call 0.03 s
  after PutTableBucketPolicy returned.
  - ACK: terminal_codes · ops: PutTableBucketPolicy, GetTableEncryption, GetTableStorageClass
  - repro: PutTableBucketPolicy(Deny role on GetTableEncryption for <bucket>/table/*) -> GetTableEncryption as
    the role
  - measurements: bucket_policy_deny_effective_s=0.03, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-233](table-maintenance.md#s3t-table-233), [S3T-TABLE-234](table-maintenance.md#s3t-table-234) · hypotheses: H-T-066 · evidence:
    table/error-taxonomy/iam-deny-rename-tag
  - notes: Characterizes H-T-066 beyond the IAM case: identity-policy deny, resource-policy deny and implicit
    deny (see the role-B finding) all surface as AccessDeniedException 403; only the message tail differs
    ('with an explicit deny in an identity-based policy' / '... in a resource-based policy' / 'because...
  - full notes: [details/S3T-TABLE-251.md](details/S3T-TABLE-251.md)

- <a id="s3t-tablepolicy-002"></a>**S3T-TABLEPOLICY-002** `error-code` · impact high · handled · verified 2026-10-14
  **PutTablePolicy right after CreateRole: BadRequestException 400 for 8.57-8.66 s (n=3) then 200 - a transient 400 a controller must retry**
  CreateRole -> PutTablePolicy(Principal=<new role ARN>) retried every 1 s: rejected [8, 8, 8] times (per
  trial) with ['BadRequestException/400'] ["Invalid principal in policy"] before the first HTTP 200 at [8.66,
  8.63, 8.57] s. The same code and message are returned for a role that never existed, so the 400 is not
  distinguishable from a permanent spec error by code or text.
  - ACK: requeue, terminal_codes, references · ops: PutTablePolicy, PutTableBucketPolicy · fields:
    resourcePolicy, Principal
  - repro: iam CreateRole -> s3tables PutTablePolicy with Principal.AWS=<role ARN> at +0 ms, retry 1/s until
    200 (x3 roles)
  - measurements: ok_at_s=[8.66, 8.63, 8.57], n_rejected=[8, 8, 8], n=3
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-164](#s3t-table-164), [S3T-TABLEREPLICATION-009](table-replication.md#s3t-tablereplication-009), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013),
    [S3T-TABLE-165](#s3t-table-165), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), [S3T-TABLE-157](#s3t-table-157) · evidence: table/creative/policy-principal-lifecycle
  - notes: extends [S3T-TABLE-164](#s3t-table-164) (nonexistent role -> 400) with the propagation window after CreateRole
    measured from the controller's call sequence. Decision rule: treat BadRequestException 'invalid principal'
    on Put*Policy as retryable for at least ~30 s before surfacing it as terminal; a Role reference...
  - full notes: [details/S3T-TABLEPOLICY-002.md](details/S3T-TABLEPOLICY-002.md)

## Request validation

- <a id="s3t-table-157"></a>**S3T-TABLE-157** `request-validation` · impact high · handled · verified 2026-10-13
  **PutTablePolicy rejects public policies: Allow Principal "*" no Condition -> BadRequestException (HTTP 400); Deny "*" -> HTTP 200...**
  Allow s3tables:GetTable Principal "*" no Condition -> BadRequestException (HTTP 400) 'Public policies cannot
  be applied to S3 table buckets and tables.'. Principal {"AWS":"*"} -> BadRequestException (HTTP 400) 'Public
  policies cannot be applied to S3 table buckets and tables.'. Deny GetTableData Principal "*" -> HTTP 200.
  Allow "*" + Condition ArnLike aws:PrincipalArn=<own role> -> HTTP 200. Allow "*" + IpAddress aws:SourceIp ->
  BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets and tables.'. Allow to
  another account's :root -> BadRequestException (HTTP 400) 'Invalid principal in policy'. Service principal
  s3.amazonaws.com -> HTTP 200.
  - ACK: terminal_codes · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy with each Principal "*" variant on the fixture table
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-089](tablebucket-policy.md#s3t-tablebucket-089), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002), [S3T-TABLE-165](#s3t-table-165), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), [S3T-TABLE-164](#s3t-table-164),
    [S3T-TABLE-162](#s3t-table-162), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001) · hypotheses: H-U-026 · evidence:
    table/sub-resources/policy-lifecycle
  - notes: H-U-026 refuted: there IS a public-policy guard (BadRequestException 400 'Public policies cannot be
    applied to S3 table buckets and tables.').
  - full notes: [details/S3T-TABLE-157.md](details/S3T-TABLE-157.md)

- <a id="s3t-table-164"></a>**S3T-TABLE-164** `request-validation` · impact medium · handled · verified 2026-10-13
  **PutTablePolicy Principal: nonexistent role ARN BadRequestException/400; sts assumed-role session ARN HTTP 200...**
  Outcome per Principal value: {"sts assumed-role session ARN": "HTTP 200", "Service example.com":
  "BadRequestException/400", "Service s3.amazonaws.com": "HTTP 200", "nonexistent role ARN (own acct)":
  "BadRequestException/400", "own :root ARN": "HTTP 200", "nonexistent user ARN": "BadRequestException/400",
  "Federated cognito": "BadRequestException/400", "AWS [role, *]": "BadRequestException/400"}. Messages:
  {"Service example.com": "Invalid service principal in policy", "nonexistent role ARN (own acct)": "Invalid
  principal in policy", "nonexistent user ARN": "Invalid principal in policy", "AWS [role, *]": "Public
  policies cannot be applied to S3 table buckets and tables."}. Read-backs of accepted forms are in
  result.yaml (case_principal_*).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on the fixture table with each Principal form
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-010](service.md#s3t-service-010), [S3T-SERVICE-025](service.md#s3t-service-025), [S3T-SERVICE-012](service.md#s3t-service-012), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002), [S3T-TABLE-165](#s3t-table-165), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), [S3T-TABLE-157](#s3t-table-157),
    [S3T-TABLE-162](#s3t-table-162), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001) · hypotheses: H-U-010, H-U-068 ·
    evidence: table/error-taxonomy/policy-validation
  - notes: H-U-010 first half (nonexistent role rejected): confirmed. H-U-068 principal part: sts session ARN
    ACCEPTED (readback
    {"Version":"2012-10-17","Statement":[{"Sid":"X","Effect":"Allow","Principal":{"AWS":"arn:aws:sts::<ACCOUNT>:assumed-r),
    non-amazonaws Service rejected.
  - full notes: [details/S3T-TABLE-164.md](details/S3T-TABLE-164.md)

- <a id="s3t-table-167"></a>**S3T-TABLE-167** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy does not validate Resource or s3tables action names...**
  On table t1: Resource = t2's ARN, the bucket ARN, <bucketARN>/table/<name>, <bucketARN>/table/<ns>/<name>,
  "*", <t1 ARN>/*, <bucketARN>/table/*, [t1, t2] and a statement with NO Resource key were all accepted (HTTP 200)
  and read back as sent. Action "s3tables:NoSuchAction" and "s3tables:*" are accepted; Action "s3:GetObject"
  -> BadRequestException 400 'Invalid Action'. Only the service prefix is checked.
  - ACK: none · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy on t1 with each Resource/Action variant; GetTablePolicy after each
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-163](#s3t-table-163), [S3T-TABLE-157](#s3t-table-157), [S3T-TABLE-162](#s3t-table-162), [S3T-TABLE-164](#s3t-table-164), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001),
    [S3T-TABLEBUCKET-088](tablebucket-policy.md#s3t-tablebucket-088), [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-004](#s3t-table-004) · hypotheses: H-U-009,
    H-U-068, H-U-008 · evidence: table/error-taxonomy/policy-validation
  - notes: Refutes H-U-009's rejection half (no Resource-vs-table check; '*' accepted as hypothesized) and
    H-U-068's Resource-suffix half (<arn>/* accepted). A controller templating Resource from the spec name
    instead of status.tableARN will NOT get an error; the policy is simply ineffective.
  - full notes: [details/S3T-TABLE-167.md](details/S3T-TABLE-167.md)

- <a id="s3t-tablepolicy-009"></a>**S3T-TABLEPOLICY-009** `request-validation` · impact high · unhandled (not handled in controller) · verified 2026-10-15, re-verified
  **Policy Deny binds for Resource = uuid table ARN, <bucket>/table/*, <bucket>/*, '*'...**
  Deny(Principal '*', Action s3tables:GetTableMaintenanceConfiguration, Resource=<form>) then the denied call
  on the table by the caller itself (1 s after Put). Table policy: exact_table_arn -> Put HTTP 200, call
  AccessDeniedException/403; name_based_pseudo_arn -> Put HTTP 200, call HTTP 200; bucket_arn -> Put HTTP 200,
  call HTTP 200; bucket_table_wildcard -> Put HTTP 200, call AccessDeniedException/403; bucket_slash_wildcard
  -> Put HTTP 200, call AccessDeniedException/403; star -> Put HTTP 200, call AccessDeniedException/403;
  other_table_arn -> Put HTTP 200, call HTTP 200; exact_table_arn_list_with_other -> Put HTTP 200, call
  AccessDeniedException/403. Bucket policy with the same statements: exact_table_arn -> Put HTTP 200, call
  AccessDeniedException/403; bucket_table_wildcard -> Put HTTP 200, call AccessDeniedException/403;
  name_based_pseudo_arn -> Put HTTP 200, call HTTP 200; bucket_arn -> Put HTTP 200, call HTTP 200;
  bucket_slash_wildcard -> Put HTTP 200, call AccessDeniedException/403. A table-policy Deny never affected
  the sibling table (all HTTP 200). Every Put returned ['HTTP 200']; DeleteTablePolicy/DeleteTableBucketPolicy
  restored HTTP 200 on the next call in all cases.
  - ACK: is_iam_policy, docs-only, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy,
    PutTableBucketPolicy, GetTableMaintenanceConfiguration · fields: resourcePolicy, Resource
  - repro: PutTablePolicy(Deny * GetTableMaintenanceConfiguration Resource=<form>) ->
    GetTableMaintenanceConfiguration -> DeleteTablePolicy, per form; same via PutTableBucketPolicy
  - measurements: table_policy_binding_forms=[exact_table_arn, bucket_table_wildcard, bucket_slash_wildcard,
    star, exact_table_arn_list_with_other], table_policy_inert_forms=[name_based_pseudo_arn, bucket_arn,
    other_table_arn], table_policy_rejected_forms=[], bucket_policy_binding_forms=[exact_table_arn,
    bucket_table_wildcard, bucket_slash_wildcard], bucket_policy_inert_forms=[name_based_pseudo_arn,
    bucket_arn], bucket_policy_rejected_forms=[], n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-163](#s3t-table-163), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLE-213](table-maintenance.md#s3t-table-213), [S3T-TABLE-251](#s3t-table-251), [S3T-TABLEBUCKET-091](tablebucket-policy.md#s3t-tablebucket-091),
    [S3T-TABLEBUCKETPOLICY-005](tablebucket-policy.md#s3t-tablebucketpolicy-005), [S3T-TABLEBUCKETPOLICY-016](tablebucket-policy.md#s3t-tablebucketpolicy-016), [S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007), [S3T-TABLEBUCKET-088](tablebucket-policy.md#s3t-tablebucket-088),
    [S3T-TABLE-162](#s3t-table-162), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-004](#s3t-table-004) · evidence: table/creative/policy-resource-binding
  - notes: extends [S3T-TABLE-163](#s3t-table-163)/-167 (any Resource accepted) with which forms are actually enforced. Decision
    rule: a controller that lets users write Resource with the name-based pseudo ARN or the bucket ARN in a
    TablePolicy produces a policy that reads back fine but protects nothing; the controller should...
  - full notes: [details/S3T-TABLEPOLICY-009.md](details/S3T-TABLEPOLICY-009.md)

## Update granularity and ordering

- <a id="s3t-table-140"></a>**S3T-TABLE-140** `prerequisite` · impact high · handled · verified 2026-10-13
  **CreateTable with encryptionConfiguration needs s3tables:PutTableEncryption (deny -> AccessDeniedException 403, even for AES256...**
  Role policy Allow s3tables:* + kms use on K, explicit Deny s3tables:PutTableEncryption: CreateTable {AES256}
  -> AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'; {aws:kms K} -> AccessDeniedException (HTTP 403, NOT in the modeled error list)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; no encryptionConfiguration -> HTTP 200 (GetTableEncryption
  {"sseAlgorithm": "AES256"}); storageClassConfiguration={INTELLIGENT_TIERING} only -> HTTP 200
  (GetTableEncryption {"sseAlgorithm": "AES256"}). With the Deny switched to s3tables:PutTableStorageClass
  instead: {INTELLIGENT_TIERING} -> HTTP 200 (GetTableEncryption {"sseAlgorithm": "AES256"}); no config ->
  HTTP 200 (GetTableEncryption {"sseAlgorithm": "AES256"}).
  - ACK: terminal_codes, docs-only · ops: CreateTable · fields: encryptionConfiguration,
    storageClassConfiguration
  - repro: Assume a role with s3tables:* and an explicit Deny on s3tables:PutTableEncryption (then on
    s3tables:PutTableStorageClass); CreateTable with/without each struct
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-109](#s3t-table-109), [S3T-TABLE-132](#s3t-table-132), [S3T-TABLE-250](#s3t-table-250), [S3T-TABLE-142](#s3t-table-142), [S3T-TABLE-131](#s3t-table-131), [S3T-TABLEBUCKET-129](tablebucket-policy.md#s3t-tablebucket-129) ·
    hypotheses: H-T-169 · evidence: table/dependencies/iam-permissions
  - notes: H-T-169 confirmed. The controller's IAM policy must include s3tables:PutTableEncryption (storage
    class needs no extra action) whenever spec carries the struct.
  - full notes: [details/S3T-TABLE-140.md](details/S3T-TABLE-140.md)

- <a id="s3t-table-141"></a>**S3T-TABLE-141** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **CreateTable({aws:kms K} or {AES256}) by the Deny-kms:* probe role -> AccessDeniedException 403 (unmodeled) on s3tables:PutTableEncryption**
  Key policy grants the maintenance principal and delegates IAM-user permissions to the account. Role with
  Allow s3tables:* and Deny kms:*: CreateTable {aws:kms K} -> AccessDeniedException (HTTP 403, NOT in the
  modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; CreateTable {AES256} -> AccessDeniedException (HTTP 403, NOT in
  the modeled error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'; GetTableEncryption on an existing KMS table -> HTTP 200. Allow
  only kms:GenerateDataKey on K -> AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'. Allow only kms:Decrypt on K -> AccessDeniedException (HTTP 403, NOT in the modeled
  error list) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an
  explicit deny in an identity-based policy'. Allow kms:GenerateDataKey+kms:Decrypt on K ->
  AccessDeniedException (HTTP 403, NOT in the modeled error list) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fc4776-iam with an explicit deny in an
  identity-based policy'.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.kmsKeyArn
  - repro: Assume a role with s3tables:* and (Deny kms:* | Allow kms:GenerateDataKey | Allow kms:Decrypt |
    Allow both) on K; CreateTable(aws:kms K)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-054](table-maintenance.md#s3t-table-054), [S3T-TABLE-055](table-maintenance.md#s3t-table-055), [S3T-TABLE-056](table-maintenance.md#s3t-table-056), [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLE-216](#s3t-table-216),
    [S3T-TABLE-328](#s3t-table-328) · hypotheses: H-T-168 · evidence: table/dependencies/iam-permissions
  - notes: H-T-168 qualified/refuted - see behavior. Reading the encryption configuration needs no KMS
    permission (GetTableEncryption succeeded with kms:* denied).
  - full notes: [details/S3T-TABLE-141.md](details/S3T-TABLE-141.md)

- <a id="s3t-table-325"></a>**S3T-TABLE-325** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable with metadata.iceberg.schema needs s3tables:PutTableData: same policy + it -> 200, without -> ForbiddenException 403**
  Three runs, identical. Two assumed roles, both Allow s3tables:CreateTable/GetTable/DeleteTable on
  <bucketARN> and <bucketARN>/*. Without PutTableData: CreateTable(no metadata) -> HTTP 200 at the first
  attempt; CreateTable(metadata.iceberg.schema) -> ForbiddenException (HTTP 403) 'Access Denied' (402-449 ms),
  retry -> the same; admin GetTable on that name -> NotFoundException 404 and admin ListTables does not list
  it; admin CreateTable of the same name/metadata -> HTTP 200. Properties-only metadata
  ({iceberg:{properties:{k:v}}}) by that role -> BadRequestException (HTTP 400) 'The specified metadata is not
  valid' (7-8 ms, before any authorization). With s3tables:PutTableData added: CreateTable(no metadata) ->
  HTTP 200; CreateTable(metadata.iceberg.schema) -> HTTP 200 (378-486 ms) and admin GetTableMetadataLocation
  shows a metadataLocation.
  - ACK: docs-only, terminal_codes · ops: CreateTable · fields: metadata.iceberg.schema
  - repro: Two IAM roles (with/without s3tables:PutTableData) -> AssumeRole -> CreateTable with and without
    metadata.iceberg.schema
  - measurements: assume_role_first_attempt_ok_s=0.0, create_with_metadata_403_latency_ms=[449, 422, 402],
    create_with_metadata_200_latency_ms=[404, 401, 486], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-131](#s3t-table-131) · evidence: table/creative/reverify-set-c
  - notes: Confirms the causal clause in the [S3T-TABLE-131](#s3t-table-131) title (which only tested the role lacking the
    action) with a positive control: the controller's IAM policy must include s3tables:PutTableData for Table
    specs that carry metadata.iceberg.schema; the error it gets otherwise is ForbiddenException 'Access...
  - full notes: [details/S3T-TABLE-325.md](details/S3T-TABLE-325.md)

- <a id="s3t-table-328"></a>**S3T-TABLE-328** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **CreateTable {aws:kms K} needs the caller's kms:DescribeKey (ForbiddenException 403 naming it); {AES256}/no config -> 200 with kms:* denied**
  Key K: policy = account root kms:* + maintenance.s3tables.amazonaws.com GenerateDataKey/Decrypt. Buckets:
  AES256-default and aws:kms(K)-default. Fresh roles (trust this account, one inline policy, never carried any
  other statement), first CreateTable {'denykms': 60.8, 'allowkms': 64.3, 'denyput': 67.6} s after
  PutRolePolicy: {"allowkms": {"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403)
  [kms:DescribeKey]", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K":
  "ForbiddenException (HTTP 403) [kms:DescribeKey]", "kmsbucket/none": "HTTP 200"}, "denykms":
  {"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403) [kms:DescribeKey]",
  "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K": "ForbiddenException (HTTP 403)
  [kms:DescribeKey]", "kmsbucket/none": "HTTP 200"}, "denyput": {"aesbucket/AES256": "AccessDeniedException
  (HTTP 403) [s3tables:PutTableEncryption]", "aesbucket/kms_K": "AccessDeniedException (HTTP 403)
  [s3tables:PutTableEncryption]", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "AccessDeniedException
  (HTTP 403) [s3tables:PutTableEncryption]", "kmsbucket/kms_K": "AccessDeniedException (HTTP 403)
  [s3tables:PutTableEncryption]", "kmsbucket/none": "HTTP 200"}, "lab": {"aesbucket/AES256": "HTTP 200",
  "aesbucket/kms_K": "HTTP 200", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200",
  "kmsbucket/kms_K": "HTTP 200", "kmsbucket/none": "HTTP 200"}}. Full cells incl. messages:
  denykms={"aesbucket/AES256": "HTTP 200", "aesbucket/kms_K": "ForbiddenException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 with an explicit deny in an
  identity-based policy'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200", "kmsbucket/kms_K":
  "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 with an explicit deny in an
  identity-based policy'", "kmsbucket/none": "HTTP 200"}; allowkms={"aesbucket/AES256": "HTTP 200",
  "aesbucket/kms_K": "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is
  not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 because no identity-based policy
  allows the kms:DescribeKey action'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "HTTP 200",
  "kmsbucket/kms_K": "ForbiddenException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is
  not authorized to perform: kms:DescribeKey on resource:
  arn:aws:kms:us-west-2:<ACCOUNT>:key/02a3c4ea-41c2-4ece-8830-303258e701f4 because no identity-based policy
  allows the kms:DescribeKey action'", "kmsbucket/none": "HTTP 200"}; denyput={"aesbucket/AES256":
  "AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized
  to perform: s3tables:PutTableEncryption on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2aes with an explicit deny in an identity-based
  policy'", "aesbucket/kms_K": "AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2aes with an explicit deny in an
  identity-based policy'", "aesbucket/none": "HTTP 200", "kmsbucket/AES256": "AccessDeniedException (HTTP 403)
  'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform:
  s3tables:PutTableEncryption on resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2kms with
  an explicit deny in an identity-based policy'", "kmsbucket/kms_K": "AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTableEncryption on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-fda88c-fu2kms with an explicit deny in an
  identity-based policy'", "kmsbucket/none": "HTTP 200"}.
  - ACK: terminal_codes, docs-only · ops: CreateTable, GetTableEncryption · fields:
    encryptionConfiguration.sseAlgorithm, encryptionConfiguration.kmsKeyArn
  - repro: CreateKey(K) -> CreateTableBucket x2 (AES256 default, aws:kms K default) -> CreateRole x3 (Allow
    s3tables:* + Deny kms:* | + Allow kms:GenerateDataKey/Decrypt on K | + Deny s3tables:PutTableEncryption)
    -> wait 60 s -> AssumeRole -> CreateTable {aws:kms K}, {AES256}, no config in each bucket
  - measurements: roles=3, cells_per_role=6, wait_after_put_role_policy_s=[60.8, 64.3, 67.6], n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-141](#s3t-table-141), [S3T-TABLE-057](table-encryption-storage.md#s3t-table-057) · evidence:
    tablebucket/weird-inputs/fu2-01-07-kms-key-state-and-kms-deny
  - notes: FU2-07: both readings are right under different conditions, and 141's attribution was wrong. With
    kms:* explicitly denied (fresh role, 60 s after PutRolePolicy, never carried another statement):
    CreateTable {AES256} -> HTTP 200 and CreateTable without encryptionConfiguration -> HTTP 200 in both an...
  - full notes: [details/S3T-TABLE-328.md](details/S3T-TABLE-328.md)

- <a id="s3t-tablepolicy-004"></a>**S3T-TABLEPOLICY-004** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-14
  **Policy principal role deleted: Get flaps ARN/AROA for ~65 s; re-Put observed or desired -> 400...**
  With Allow Principal=<role ARN> stored, iam DeleteRole: GetTablePolicy (1/s for 75 s) first shows the AROA
  unique id at +33 s, but the rendering flips between ARN and AROA 9 times before it is stable from +64 s
  (second deletion: first AROA +0 s, 4 flips, stable from +65 s); every read is HTTP 200. Controller moves:
  re-Put of the observed document -> BadRequestException/400 'Invalid principal in policy'; re-Put of the
  desired document (ARN) -> BadRequestException/400 'Invalid principal in policy'; DeleteTablePolicy -> HTTP
  204; Put(desired) with the role still absent -> BadRequestException/400 'Invalid principal in policy'. After
  CreateRole with the SAME name (same ARN, new RoleId): Put(desired) -> HTTP 200 after 7.53 s (7 rejections,
  ['BadRequestException/400']); the first GetTablePolicy after that 200 still rendered the NEW role's unique
  id instead of the ARN (the read renderer lags the Put validator). Deleting the new role makes the read show
  the NEW unique id and never the old one: the stored document is bound to the role's unique id, not its name.
  - ACK: requeue, references, terminal_codes, is_iam_policy · ops: PutTablePolicy, GetTablePolicy,
    DeleteTablePolicy · fields: resourcePolicy, Principal
  - repro: PutTablePolicy(Allow role) -> iam DeleteRole -> GetTablePolicy (poll) -> Put(observed) ->
    Put(desired) -> DeleteTablePolicy -> Put(desired) -> CreateRole(same name) -> Put(desired) poll ->
    DeleteRole -> Get
  - measurements: aroa_first_seen_s=[33, 0], aroa_stable_from_s=[64, 65], flips=[9, 4],
    form_sequences_1_per_s=[AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAIAAIAAAAAAAAAAAAAAIIAIIIIIIIIIII,
    IAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAIIIIIIIIII],
    put_ok_after_recreate_s=7.53, n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-165](#s3t-table-165), [S3T-TABLE-164](#s3t-table-164), [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001),
    [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002), [S3T-TABLE-157](#s3t-table-157), [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010) · evidence:
    table/creative/policy-principal-lifecycle
  - notes: extends [S3T-TABLE-165](#s3t-table-165) (AROA rewrite) with the propagation flapping and the full recovery sequence a
    controller would run. A controller diffing (n=2 deletions, 1 poll/s for 75 s) GetTablePolicy against spec
    sees the principal change back and forth for ~20 s with no write of its own. Decision rule:...
  - full notes: [details/S3T-TABLEPOLICY-004.md](details/S3T-TABLEPOLICY-004.md)

- <a id="s3t-tablepolicy-010"></a>**S3T-TABLEPOLICY-010** `prerequisite` · impact high · unhandled (not handled in controller) · verified 2026-10-15
  **Deleted principal role: AROA and ARN documents both re-Put 200 for ~20-40 s after DeleteRole (IAM lag), then 400 at bucket and table level**
  Allow policies naming a fresh role's ARN stored on the bucket (accepted after [7.69, 7.59] s) and on the
  table (accepted after [1.23, 0.14] s); read back as [{'bucket': 'ARN', 'table': 'ARN'}, {'bucket': 'ARN',
  'table': 'ARN'}]. iam DeleteRole, then at each offset: PutTableBucketPolicy(AROA doc) /
  PutTableBucketPolicy(ARN doc) / PutTablePolicy(AROA doc) / PutTablePolicy(ARN doc). Bucket level: +5s AROA
  HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +10s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +20s AROA
  HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +40s AROA HTTP 200|BadRequestException (HTTP 400) / ARN
  BadRequestException (HTTP 400)|BadRequestException (HTTP 400); +90s AROA BadRequestException (HTTP
  400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP 400)|BadRequestException (HTTP 400).
  Table level: +5s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +10s AROA HTTP 200|HTTP 200 / ARN HTTP
  200|HTTP 200; +20s AROA HTTP 200|HTTP 200 / ARN HTTP 200|HTTP 200; +40s AROA BadRequestException (HTTP
  400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP 400)|BadRequestException (HTTP 400);
  +90s AROA BadRequestException (HTTP 400)|BadRequestException (HTTP 400) / ARN BadRequestException (HTTP
  400)|BadRequestException (HTTP 400). (per offset: results of role 1 | role 2.)
  GetTableBucketPolicy/GetTablePolicy rendering per offset, role 1: [(5, 'ARN', 'ARN'), (10, 'ARN', 'ARN'),
  (20, 'ARN', 'ARN'), (40, 'AROA', 'AROA'), (90, 'AROA', 'AROA')]; role 2: [(5, 'ARN', 'ARN'), (10, 'ARN',
  'ARN'), (20, 'ARN', 'ARN'), (40, 'AROA', 'AROA'), (90, 'AROA', 'AROA')].
  - ACK: is_iam_policy, requeue, references · ops: PutTableBucketPolicy, PutTablePolicy, GetTableBucketPolicy,
    GetTablePolicy · fields: resourcePolicy, Principal
  - repro: CreateRole -> PutTableBucketPolicy+PutTablePolicy(Principal=role ARN) until 200 -> DeleteRole -> at
    +5/10/20/40/90 s Put the AROA (RoleId) document and the ARN document at both levels
  - measurements: offsets_s=[5, 10, 20, 40, 90], roles=2, aroa_accepted_offsets_bucket=[5, 10, 20, 40],
    aroa_accepted_offsets_table=[5, 10, 20], arn_accepted_offsets_both_levels=[5, 10, 20],
    all_rejected_from_s=90, arn_doc_accepted_after_create_s=[7.69, 7.59, 1.23, 0.14], n=2
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLE-165](#s3t-table-165), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004) · evidence:
    table/identity/fu2-06-missing-ns-messages-subread-iam-aroa
  - notes: FU2-02: both readings are right under different conditions - time since DeleteRole (IAM propagation
    of the deletion), not the policy level, decides. At +5/+10/+20 s both the AROA document and the ARN
    document were accepted (HTTP 200) at bucket AND table level while...
  - full notes: [details/S3T-TABLEPOLICY-010.md](details/S3T-TABLEPOLICY-010.md)

## Field behavior (defaults, normalization, shapes, immutability)

- <a id="s3t-table-152"></a>**S3T-TABLE-152** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTablePolicy returns the policy minified with statement keys reordered to Sid, Effect, Principal, Action, Resource**
  Submitted a 2-statement policy pretty-printed (4-space indent, a tab, double spaces, trailing newline, 753
  bytes) with statement keys in the order Sid, Resource, Action, Effect, Principal. GetTablePolicy returned
  552 bytes: no whitespace at all, top-level order Version, Statement kept, every statement re-keyed in the
  order Sid, Effect, Principal, Action, Resource; statement order and all values unchanged (structurally
  equal). A string compare against the submitted document never matches.
  - ACK: is_document, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: PutTablePolicy(pretty-printed 2-statement doc) -> GetTablePolicy -> compare bytes
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKETPOLICY-008](tablebucket-policy.md#s3t-tablebucketpolicy-008), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](tablebucket-policy.md#s3t-tablebucketpolicy-004),
    [S3T-TABLEPOLICY-005](#s3t-tablepolicy-005), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008),
    [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-005, H-U-006 · evidence: table/sub-resources/policy-lifecycle
  - notes: H-U-005 qualified: minified and structurally equal, but key order is NOT preserved (canonical IAM
    order); H-U-006 refuted (not byte-for-byte). Single document, n=1 but deterministic across the 2 identical
    Puts.
  - full notes: [details/S3T-TABLE-152.md](details/S3T-TABLE-152.md)

- <a id="s3t-table-154"></a>**S3T-TABLE-154** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **GetTablePolicy: account-id principal rewritten to :root ARN, 1-elem principal list collapsed, 1-elem Action list collapsed**
  PutTablePolicy with Principal {"AWS":"<account-id>"} read back as {"AWS": "arn:aws:iam::<ACCOUNT>:root"};
  {"AWS":["<role ARN>"]} read back as {"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}; "*" (with
  aws:PrincipalAccount Condition) read back as '*'; {"AWS":"*"} read back as {'AWS': '*'}; Action
  ["s3tables:GetTable"] read back as 's3tables:GetTable'. Statement given as a single object (not array) ->
  Put HTTP 200, read back Statement type list.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: PutTablePolicy with account-id, [role ARN], "*"+Condition, {AWS:"*"}+Condition principals and a
    1-element Action list -> GetTablePolicy
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-008](tablebucket-policy.md#s3t-tablebucketpolicy-008), [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](tablebucket-policy.md#s3t-tablebucketpolicy-004),
    [S3T-TABLEPOLICY-005](#s3t-tablepolicy-005), [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008),
    [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) · hypotheses: H-U-007 · evidence: table/sub-resources/policy-lifecycle
  - notes: H-U-007 confirmed: acct->root True, list collapse True, '*' kept True. Unconditioned '*' could not
    be tested (rejected as public, see H-U-026 finding). The unconditioned '*' is rejected as public, so the
    '*' round trip was observed with an aws:PrincipalAccount Condition: returned as the bare string...
  - full notes: [details/S3T-TABLE-154.md](details/S3T-TABLE-154.md)

- <a id="s3t-table-165"></a>**S3T-TABLE-165** `normalization` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Deleting the IAM role named in a stored table policy: GetTablePolicy principal rewritten to the AROA unique id...**
  CreateRole then PutTablePolicy(Principal=<new role ARN>): accepted after 10.45 s (6 attempts; first
  rejection ('BadRequestException', 'Invalid principal in policy')). Principal read back before role deletion:
  {"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}. After DeleteRole, GetTablePolicy at 0/5/15/30 s returned
  principals [{"AWS": "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}, {"AWS":
  "arn:aws:iam::<ACCOUNT>:role/<PRINCIPAL>"}, {"AWS": "<AROA>"}, {"AWS": "<AROA>"}]. Re-Put of the read-back
  document -> BadRequestException/400 'Invalid principal in policy'; re-Put of the original ARN document (role
  gone) -> BadRequestException/400 'Invalid principal in policy'.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: iam CreateRole R -> PutTablePolicy(Principal R ARN) -> iam DeleteRole R -> GetTablePolicy (x4 over
    30 s) -> PutTablePolicy(read-back)
  - measurements: new_role_accepted_after_s=10.45, put_attempts=6, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-013](tablebucket-policy.md#s3t-tablebucketpolicy-013), [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004), [S3T-TABLEBUCKETPOLICY-001](tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002),
    [S3T-TABLE-164](#s3t-table-164), [S3T-TABLE-157](#s3t-table-157), [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010) · hypotheses: H-U-010 · evidence:
    table/error-taxonomy/policy-validation
  - notes: H-U-010 second half (AROA rewrite) confirmed within the 30 s observed. Single run.
  - full notes: [details/S3T-TABLE-165.md](details/S3T-TABLE-165.md)

- <a id="s3t-table-166"></a>**S3T-TABLE-166** `server-default` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy without "Version" is accepted and GetTablePolicy returns the document with "Version":"2008-10-17" injected**
  A policy document with Statement but no Version key -> PutTablePolicy HTTP 200; GetTablePolicy returns
  {"Version":"2008-10-17","Statement":[...]} (the legacy IAM default), so the read-back has a key the caller
  never sent. "Version":"2008-10-17" sent explicitly is accepted and kept; "Version":"2030-01-01" ->
  BadRequestException 400 'The specified policy is malformed'. "Effect":"allow" (lowercase) is accepted and
  read back as "Allow".
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: PutTablePolicy({"Statement":[...]}) -> GetTablePolicy; PutTablePolicy with Effect "allow" ->
    GetTablePolicy
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-008](tablebucket-policy.md#s3t-tablebucketpolicy-008), [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](tablebucket-policy.md#s3t-tablebucketpolicy-004),
    [S3T-TABLEPOLICY-005](#s3t-tablepolicy-005), [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003), [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008),
    [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-157](#s3t-table-157), [S3T-TABLE-162](#s3t-table-162), [S3T-TABLE-164](#s3t-table-164), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001)
    · hypotheses: H-U-027, H-U-005 · evidence: table/error-taxonomy/policy-validation
  - notes: Qualifies H-U-005/H-U-027: besides minification and key reordering the service injects a default
    Version and canonicalizes Effect case; a structural compare of spec vs read-back must tolerate an added
    Version=2008-10-17 and Effect case.

- <a id="s3t-tablepolicy-003"></a>**S3T-TABLEPOLICY-003** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTablePolicy is a fixed point: Put(observed) reads back byte-identical (table and bucket); desired minified != observed**
  Two-statement document (Effect 'allow' lower-case, 1-elem Principal list, 2-elem Action list, 1-elem
  Resource list, two Condition operators, Deny Principal '*'): Put -> HTTP 200; Get -> normalized form;
  Put(that form) -> HTTP 200; Get again byte-identical: True. Same through
  PutTableBucketPolicy/GetTableBucketPolicy: byte-identical True, same normalizer as the table path: True.
  Rules seen: top-level keys ['Version', 'Statement'], statement keys ['Sid', 'Effect', 'Principal', 'Action',
  'Resource', 'Condition'], Effect ['Allow', 'Deny'], principal list collapsed True, 2-elem Action kept as
  list True, 1-elem Resource collapsed True, Condition keys ['StringEquals', 'Bool'], statement order kept
  True. json.loads(observed) == desired: False.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy,
    PutTableBucketPolicy, GetTableBucketPolicy · fields: resourcePolicy
  - repro: PutTablePolicy(desired) -> GetTablePolicy -> PutTablePolicy(observed) -> GetTablePolicy; same via
    the bucket policy ops
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLE-166](#s3t-table-166),
    [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012) · evidence: table/creative/policy-principal-lifecycle
  - notes: extends [S3T-TABLE-152](#s3t-table-152)/-154 and [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063)/-111 with the controller-relevant property: the
    normalization is idempotent, so a controller can store normalize(desired)=Get-after-first-Put and compare
    strings, or compare semantically (lists vs scalars and Effect case differ); a raw string...
  - full notes: [details/S3T-TABLEPOLICY-003.md](details/S3T-TABLEPOLICY-003.md)

- <a id="s3t-tablepolicy-005"></a>**S3T-TABLEPOLICY-005** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Doc claim C045 PARTLY: GetTablePolicy parsed JSON equals the Put document only w/o 1-elem lists, account-id principals, missing Version**
  PutTablePolicy -> HTTP 200; GetTablePolicy returns the document minified with statement keys reordered to
  Sid, Effect, Principal, Action, Resource - structurally equal ([S3T-TABLE-152](#s3t-table-152)). But Principal
  {"AWS":"<account-id>"} is read back as {"AWS":"arn:aws:iam::<account>:root"}, a 1-element Principal list and
  a 1-element Action list are collapsed to strings, a single-object Statement becomes a list ([S3T-TABLE-154](#s3t-table-154)),
  and a document without Version gets "Version":"2008-10-17" injected; Effect 'allow' -> 'Allow'
  ([S3T-TABLE-166](#s3t-table-166)).
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: Settled from existing finding(s) [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-151](#s3t-table-151); no new
    API call for this claim.
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLEBUCKETPOLICY-008](tablebucket-policy.md#s3t-tablebucketpolicy-008),
    [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLEBUCKETPOLICY-004](tablebucket-policy.md#s3t-tablebucketpolicy-004) · evidence:
    service/static/doc-claims-2
  - notes: VERDICT: PARTLY - parsed equality holds only for documents already in canonical form; otherwise the
    read-back differs structurally from what was sent.

- <a id="s3t-tablepolicy-008"></a>**S3T-TABLEPOLICY-008** `normalization` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GetTablePolicy re-shapes 7 more forms (Statement object->array, 1-elem Resource/Principal/Condition lists->scalar, JSON bool->string...**
  Put of 25 policy variants on one table, each followed by GetTablePolicy and a structural compare. RE-SHAPED:
  Statement given as a single object -> 1-element array; one-element Resource list -> scalar string;
  one-element Principal.AWS list -> scalar; Principal.AWS list [root, root] -> de-duplicated scalar; Condition
  Bool value JSON false -> the string "false"; Condition StringEquals one-element list -> scalar; Effect
  "ALLOW" -> "Allow". RETURNED STRUCTURALLY IDENTICAL (only minified): action name case as sent
  (s3tables:gettablebucket, s3tables:LISTNAMESPACES), wildcard s3tables:Get*, two-element Resource list,
  Allow+NotAction, Allow+NotResource, two statements with the same Sid (accepted), two byte-identical
  statements (both kept), statements without Sid, Sid "", Sid caf\u00e9 (returned as the raw UTF-8 char),
  explicit Version 2008-10-17 (kept), top-level Id (kept). REJECTED at Put: S3Tables:GetTableBucket
  (capitalised service prefix) -> BadRequestException (HTTP 400) 'Invalid Action'; Principal as a bare ARN
  string -> BadRequestException (HTTP 400) 'The specified policy is malformed'; Condition operator 'bool' /
  key 'AWS:securetransport' -> BadRequestException (HTTP 400) 'The specified policy is malformed';
  Allow+NotPrincipal -> BadRequestException (HTTP 400) 'Public policies cannot be applied to S3 table buckets
  and tables.'. No read-back was byte-identical to the sent text.
  - ACK: is_iam_policy, compare.is_ignored+delta_pre_compare · ops: PutTablePolicy, GetTablePolicy · fields:
    resourcePolicy
  - repro: Put each variant (Statement object, action case, one-element lists, bare principal, JSON bool
    condition, Not* members, duplicate Sids, unicode, Version 2008, Id) -> Get -> compare
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-063](tablebucket-policy.md#s3t-tablebucket-063), [S3T-TABLEBUCKET-111](tablebucket-policy.md#s3t-tablebucket-111), [S3T-TABLE-152](#s3t-table-152), [S3T-TABLE-154](#s3t-table-154), [S3T-TABLEBUCKET-089](tablebucket-policy.md#s3t-tablebucket-089),
    [S3T-TABLE-157](#s3t-table-157), [S3T-TABLEBUCKETPOLICY-012](tablebucket-policy.md#s3t-tablebucketpolicy-012), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLEPOLICY-003](#s3t-tablepolicy-003) · evidence:
    tablebucketpolicy/creative/normalization-extras
  - notes: Extends the known re-serialization findings (minify, Version injection, Effect case, 1-element
    Action collapse, account-id -> :root) with the remaining shapes a semantic comparator must treat as equal
    or a controller will re-Put on every reconcile. Single run per cell; same engine at both levels:...
  - full notes: [details/S3T-TABLEPOLICY-008.md](details/S3T-TABLEPOLICY-008.md)

## Response fidelity and consistency

- <a id="s3t-table-158"></a>**S3T-TABLE-158** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **GetTablePolicy right after Put/Delete: 0 stale reads in 20 trials x3 phases (write->read gap 86-233 ms)**
  20 cycles of Put(A) -> Get, Put(B) -> Get, Delete -> Get with the Get issued immediately after the write
  returned (write+read gap 86-233 ms): stale reads after Put(A)=0, after Put(B)=0, after Delete=0. Codes seen:
  {'put': ['HTTP 200'], 'get': ['HTTP 200', 'NotFoundException (HTTP 404)'], 'delete': ['HTTP 204']}.
  - ACK: none · ops: PutTablePolicy, DeleteTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: loop: PutTablePolicy(A); GetTablePolicy; PutTablePolicy(B); GetTablePolicy; DeleteTablePolicy;
    GetTablePolicy
  - measurements: trials=20, stale_reads=0, write_to_read_gap_ms_min=86, write_to_read_gap_ms_max=233, n=60
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKETPOLICY-003](tablebucket-policy.md#s3t-tablebucketpolicy-003), [S3T-TABLEBUCKET-065](tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLEBUCKETPOLICY-005](tablebucket-policy.md#s3t-tablebucketpolicy-005),
    [S3T-TABLEBUCKETPOLICY-010](tablebucket-policy.md#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEPOLICY-007](#s3t-tablepolicy-007), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061),
    [S3T-TABLE-156](#s3t-table-156), [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-150](#s3t-table-150), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLE-153](#s3t-table-153) · hypotheses: H-U-014 ·
    evidence: table/sub-resources/policy-lifecycle
  - notes: H-U-014 confirmed (20 trials, not 50, to respect the 1 Get/s budget; reads at ~0 ms only).

- <a id="s3t-table-216"></a>**S3T-TABLE-216** `eventual-consistency` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **Table policy Deny is enforced at the first call 30 ms after PutTablePolicy returns (n=3) and lifted 10-40 ms after DeleteTablePolicy (n=2)**
  Across three tables and three Deny policies (DeleteTable+GetTableMaintenanceConfiguration for the caller
  role; Put/DeleteTablePolicy+GetTableMaintenanceConfiguration to "*"; PutTableMaintenanceConfiguration for
  the caller role) the denied operation issued 30 ms after PutTablePolicy returned was already rejected with
  AccessDeniedException (HTTP 403) - no observable propagation window at 1 s polling resolution - and 10-40 ms
  after DeleteTablePolicy returned the same operation succeeded again. Message form: 'User:
  arn:aws:sts::<account>:assumed-role/<role>/<session> is not authorized to perform: s3tables:<Action> on
  resource: <table ARN> with an explicit deny in a resource-based policy'.
  - ACK: requeue · ops: PutTablePolicy, DeleteTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy Deny -> poll denied op at 1/s; DeleteTablePolicy -> poll at 1/s
  - measurements: enforce_s=[0.03, 0.03, 0.03], restore_s=[0.01, 0.04], n=3
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-213](table-maintenance.md#s3t-table-213), [S3T-TABLE-214](#s3t-table-214), [S3T-TABLE-215](table-maintenance.md#s3t-table-215), [S3T-TABLE-183](table-maintenance.md#s3t-table-183), [S3T-TABLE-179](table-maintenance.md#s3t-table-179), [S3T-TABLE-141](#s3t-table-141) ·
    hypotheses: H-U-016, H-U-029, H-U-080 · evidence: table/dependencies/policy-deny-enforcement
  - notes: Qualifies H-U-016/H-U-029/H-U-080's '<5 s': enforcement and lift are effectively synchronous with
    the policy write. The wire code is AccessDeniedException (declared on none of the policy/maintenance
    operations, which only declare ForbiddenException).

## Sub-resources

- <a id="s3t-table-149"></a>**S3T-TABLE-149** `sub-resource-api` · impact high · handled · verified 2026-10-13
  **GetTablePolicy on a table that never had a policy -> NotFoundException (HTTP 404); message DIFFERS from the one for a missing table**
  Fresh ICEBERG table, no PutTablePolicy ever: GetTablePolicy -> NotFoundException (HTTP 404) 'The specified
  table policy does not exist.'. GetTablePolicy with a nonexistent table name in the same namespace ->
  NotFoundException (HTTP 404) 'The specified table does not exist.'. Both are NotFoundException/404; only the
  message text tells 'policy absent' from 'table absent'.
  - ACK: exceptions.404, custom_find · ops: GetTablePolicy · fields: resourcePolicy
  - repro: CreateTable (no policy) -> GetTablePolicy; GetTablePolicy with name=<missing>
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKETPOLICY-010](tablebucket-policy.md#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEPOLICY-007](#s3t-tablepolicy-007), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061),
    [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLE-156](#s3t-table-156), [S3T-TABLE-158](#s3t-table-158), [S3T-TABLE-150](#s3t-table-150), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLE-153](#s3t-table-153),
    [S3T-TABLE-181](#s3t-table-181), [S3T-TABLE-184](#s3t-table-184), [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178) · hypotheses: H-U-001, H-U-002 · evidence:
    table/sub-resources/policy-lifecycle
  - notes: H-U-001 confirmed (404 NotFoundException, message names the policy), H-U-002 refuted.

- <a id="s3t-table-153"></a>**S3T-TABLE-153** `sub-resource-api` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTablePolicy fully replaces the stored document: after Put(Sids A,B) then Put(Sid C) GetTablePolicy has Sids ['C']**
  Put(P1: Sids A,B) -> Put(P2: Sid C) -> GetTablePolicy returned Statement Sids ['C'] (Put -> HTTP 200).
  - ACK: none · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy(A,B); PutTablePolicy(C); GetTablePolicy
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKETPOLICY-009](tablebucket-policy.md#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKET-064](tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEBUCKETPOLICY-003](tablebucket-policy.md#s3t-tablebucketpolicy-003), [S3T-TABLEPOLICY-006](#s3t-tablepolicy-006),
    [S3T-TABLE-151](#s3t-table-151), [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-150](#s3t-table-150), [S3T-TABLE-156](#s3t-table-156), [S3T-TABLE-158](#s3t-table-158) · hypotheses: H-U-004 ·
    evidence: table/sub-resources/policy-lifecycle
  - notes: H-U-004 confirmed.

## Delete semantics

- <a id="s3t-table-214"></a>**S3T-TABLE-214** `delete-semantics` · impact high · unhandled (not handled in controller) · verified 2026-10-13
  **Lockout policy (Deny Put/DeleteTablePolicy to "*") accepted; Put/DeleteTablePolicy -> AccessDeniedException/403 at 0.03 s; DeleteTable 204**
  PutTablePolicy Deny [PutTablePolicy, DeleteTablePolicy, GetTableMaintenanceConfiguration] Principal "*" with
  a DateLessThan guard -> HTTP 200 (a Deny to "*" is not rejected as public). Canary denied after 0.03 s
  (codes ['AccessDeniedException']). While locked: PutTablePolicy -> AccessDeniedException (HTTP 403) 'User:
  arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to perform: s3tables:PutTablePolicy on
  resource: arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3e8f-deny/'; DeleteTablePolicy ->
  AccessDeniedException (HTTP 403) 'User: arn:aws:sts::<ACCOUNT>:assumed-role/<PRINCIPAL> is not authorized to
  perform: s3tables:DeleteTablePolicy on resource:
  arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-7b3e8f-de'; GetTablePolicy -> HTTP 200; GetTable -> HTTP 200.
  DeleteTable(versionToken) -> HTTP 204 (waited 0 s). GetTablePolicy after the delete -> NotFoundException
  (HTTP 404) 'The specified table does not exist.' [truncated raw body]
  - ACK: terminal_codes, custom_delete · ops: PutTablePolicy, DeleteTablePolicy, GetTablePolicy, DeleteTable ·
    fields: resourcePolicy
  - repro: PutTablePolicy Deny Put/DeleteTablePolicy to * -> poll -> Put/Delete policy -> DeleteTable
  - measurements: deny_enforced_after_s=0.03, n=1
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-213](table-maintenance.md#s3t-table-213), [S3T-TABLE-215](table-maintenance.md#s3t-table-215), [S3T-TABLE-216](#s3t-table-216), [S3T-TABLE-183](table-maintenance.md#s3t-table-183) · hypotheses: H-U-029 · evidence:
    table/dependencies/policy-deny-enforcement
  - notes: H-U-029 confirmed (Put AccessDeniedException, Delete AccessDeniedException, DeleteTable succeeded).
    Full original title: Lockout policy (Deny Put/DeleteTablePolicy to "*"): Put -> AccessDeniedException/403,
    Delete -> AccessDeniedException/403 after 0.03s; DeleteTable remediation -> HTTP 204

## Quotas and rate limits

- <a id="s3t-tablepolicy-001"></a>**S3T-TABLEPOLICY-001** `quota-limit` · impact medium · handled · verified 2026-10-13
  **PutTablePolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400)**
  Valid Allow policies padded with unique Sids on one table: 16000 bytes (68 statements) -> HTTP 200; 19500
  bytes (83 statements) -> HTTP 200; 20480 bytes (87 statements) -> HTTP 200; 20481 bytes (87 statements) ->
  ValidationException (HTTP 400); 32000 bytes (136 statements) -> ValidationException (HTTP 400); 65536 bytes
  (279 statements) -> ValidationException (HTTP 400). Rejection message: '1 validation error detected: Value
  at 'resourcePolicy' failed to satisfy constraint: Member must have length less than or equal to 20480'.
  GetTablePolicy afterwards returned 20480 bytes (the last accepted document was 20480 bytes).
  - ACK: terminal_codes, is_iam_policy · ops: PutTablePolicy, GetTablePolicy · fields: resourcePolicy
  - repro: PutTablePolicy with ASCII policies of 16000/19500/20480/20481/32000/65536 bytes (client validation
    off)
  - measurements: largest_accepted_bytes=20480, smallest_rejected_bytes=20481, n=6
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKETPOLICY-002](tablebucket-policy.md#s3t-tablebucketpolicy-002), [S3T-TABLEBUCKETPOLICY-014](tablebucket-policy.md#s3t-tablebucketpolicy-014), [S3T-TABLE-120](table.md#s3t-table-120), [S3T-TABLEBUCKET-090](tablebucket-policy.md#s3t-tablebucket-090),
    [S3T-TABLE-157](#s3t-table-157), [S3T-TABLE-162](#s3t-table-162), [S3T-TABLE-164](#s3t-table-164), [S3T-TABLE-166](#s3t-table-166), [S3T-TABLE-167](#s3t-table-167) · hypotheses: H-U-011 ·
    evidence: table/limits/policy-size-write-races
  - notes: H-U-011 qualified: a hard limit exists but at exactly 20480 bytes (the SDK model max for
    resourcePolicy), not ~20 KB = 20000: 16000/19500/20480 accepted, 20481/32000/65536 rejected with wire code
    ValidationException (HTTP 400) - not BadRequestException, not in the modeled error list - '1 validation...
  - full notes: [details/S3T-TABLEPOLICY-001.md](details/S3T-TABLEPOLICY-001.md)

## Handling gaps (bugs to file)

None recorded for this document's findings; see [service.md 'Handling gaps summary'](service.md#handling-gaps-summary) for the service-wide list.

## E2E timing

Values are seconds unless the key says otherwise; n = trials behind the numbers ('1 run' when the finding records none).

| finding | what | measurements | n |
| --- | --- | --- | --- |
| [S3T-TABLE-083](#s3t-table-083) | Wire error codes equal the modeled shape names (BadRequest=400, Forbidden=403, NotFound=404)... | rejected_calls=32 | 32 |
| [S3T-TABLE-131](#s3t-table-131) | CreateTable WITH metadata by a role lacking s3tables:PutTableData -> ForbiddenException 403... | assume_role_wait_s=10.1 | 1 |
| [S3T-TABLE-151](#s3t-table-151) | PutTablePolicy twice with a byte-identical document -> HTTP 200 then HTTP 200; read-backs identical | put_latency_ms=[105, 112] | 2 |
| [S3T-TABLE-158](#s3t-table-158) | GetTablePolicy right after Put/Delete: 0 stale reads in 20 trials x3 phases (write->read gap 86-233 ms) | trials=20, stale_reads=0, write_to_read_gap_ms_min=86, write_to_read_gap_ms_max=233 | 60 |
| [S3T-TABLE-165](#s3t-table-165) | Deleting the IAM role named in a stored table policy: GetTablePolicy principal rewritten to the AROA unique id... | new_role_accepted_after_s=10.45, put_attempts=6 | 1 |
| [S3T-TABLE-214](#s3t-table-214) | Lockout policy (Deny Put/DeleteTablePolicy to "*") accepted; Put/DeleteTablePolicy -> AccessDeniedException/403 at 0.03 s; DeleteTable 204 | deny_enforced_after_s=0.03 | 1 |
| [S3T-TABLE-216](#s3t-table-216) | Table policy Deny is enforced at the first call 30 ms after PutTablePolicy returns (n=3) and lifted 10-40 ms after DeleteTablePolicy (n=2) | enforce_s=[0.03, 0.03, 0.03], restore_s=[0.01, 0.04] | 3 |
| [S3T-TABLE-250](#s3t-table-250) | IAM explicit Deny on RenameTable/TagResource/UntagResource/ListTables/GetTable -> AccessDeniedException 403 on all 8 calls, never Forbidden | create_role_to_assumable_s=9.2, deny_effective_after_first_assume_s=0.06 | 1 |
| [S3T-TABLE-251](#s3t-table-251) | Table-bucket-policy Deny (Principal=role, Resource=<bucket>/table/*) on GetTableEncryption -> AccessDeniedException 403 like IAM deny | bucket_policy_deny_effective_s=0.03 | 1 |
| [S3T-TABLE-325](#s3t-table-325) | CreateTable with metadata.iceberg.schema needs s3tables:PutTableData: same policy + it -> 200, without -> ForbiddenException 403 | assume_role_first_attempt_ok_s=0.0, create_with_metadata_403_latency_ms=[449, 422, 402], create_with_metadata_200_latency_ms=[404, 401, 486] | 3 |
| [S3T-TABLE-328](#s3t-table-328) | CreateTable {aws:kms K} needs the caller's kms:DescribeKey (ForbiddenException 403 naming it); {AES256}/no config -> 200 with kms:* denied | roles=3, cells_per_role=6, wait_after_put_role_policy_s=[60.8, 64.3, 67.6] | 18 |
| [S3T-TABLEPOLICY-001](#s3t-tablepolicy-001) | PutTablePolicy size limit: 20480 bytes accepted, 20481 bytes rejected with ValidationException (HTTP 400) | largest_accepted_bytes=20480, smallest_rejected_bytes=20481 | 6 |
| [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002) | PutTablePolicy right after CreateRole: BadRequestException 400 for 8.57-8.66 s (n=3) then 200 - a transient 400 a controller must retry | ok_at_s=[8.66, 8.63, 8.57], n_rejected=[8, 8, 8] | 3 |
| [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004) | Policy principal role deleted: Get flaps ARN/AROA for ~65 s; re-Put observed or desired -> 400... | aroa_first_seen_s=[33, 0], aroa_stable_from_s=[64, 65], flips=[9, 4], form_sequences_1_per_s=[AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAIAAIAAAAAAAAAAAAAAIIAIIIIIIIIIII, IAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAIIIIIIIIII], put_ok_after_recreate_s=7.53 | 2 |
| [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009) | Policy Deny binds for Resource = uuid table ARN, <bucket>/table/*, <bucket>/*, '*'... | table_policy_binding_forms=[exact_table_arn, bucket_table_wildcard, bucket_slash_wildcard, star, exact_table_arn_list_with_other], table_policy_inert_forms=[name_based_pseudo_arn, bucket_arn, other_table_arn], table_policy_rejected_forms=[], bucket_policy_binding_forms=[exact_table_arn, bucket_table_wildcard, bucket_slash_wildcard], bucket_policy_inert_forms=[name_based_pseudo_arn, bucket_arn], bucket_policy_rejected_forms=[] | 1 |
| [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010) | Deleted principal role: AROA and ARN documents both re-Put 200 for ~20-40 s after DeleteRole (IAM lag), then 400 at bucket and table level | offsets_s=[5, 10, 20, 40, 90], roles=2, aroa_accepted_offsets_bucket=[5, 10, 20, 40], aroa_accepted_offsets_table=[5, 10, 20], arn_accepted_offsets_both_levels=[5, 10, 20], all_rejected_from_s=90, arn_doc_accepted_after_create_s=[7.69, 7.59, 1.23, 0.14] | 2 |

## Open questions

<!-- preserved:start id=open-questions -->
Open questions for `table-policy.md` - unverified, single-run or stale items worth a future probe; one line each.

- [S3T-TABLE-141](#s3t-table-141) (confidence low): superseded reading - its title still attributes the 403 to missing KMS permissions; the
  settled rule is [S3T-TABLE-328](#s3t-table-328). Candidate for a title note via `semantics finding update`.
- [S3T-TABLE-142](#s3t-table-142) (confidence low, n=1): the 200 on GetTableEncryption/StorageClass with only s3tables:GetTable did not reproduce
  ([S3T-TABLE-327](table-maintenance.md#s3t-table-327)); its ConflictException-409-before-403 cell for CreateTable of an existing name is unreplicated.
- [S3T-TABLE-214](#s3t-table-214) (n=1): DeleteTable as the remediation of a lockout Deny was observed once; a table-only Deny that also names
  DeleteTable (as the NotAction form of [S3T-TABLEBUCKETPOLICY-007](tablebucket-policy.md#s3t-tablebucketpolicy-007) does) and the expiry of the DateLessThan guard were not
  probed.
- [S3T-TABLEPOLICY-002](#s3t-tablepolicy-002) / [S3T-TABLE-165](#s3t-table-165) / [S3T-TABLEPOLICY-004](#s3t-tablepolicy-004): the upper bound of the 'Invalid principal in policy' window is
  10.45 s (n=1); the ~30 s retry budget in -002's notes is a decision rule, not a measurement.
- [S3T-TABLEPOLICY-010](#s3t-tablepolicy-010) (n=2 roles): the post-DeleteRole re-Put window closes between +20 and +40 s with one bucket-level AROA
  Put still 200 at +40 s; finer offsets were not sampled.
- [S3T-TABLE-157](#s3t-table-157): 'another account's :root -> 400 Invalid principal in policy' - the finding does not record whether that
  account exists, so a real cross-account grant through a table policy is unverified.
- [S3T-TABLE-162](#s3t-table-162) vs [S3T-TABLEPOLICY-008](#s3t-tablepolicy-008): 'unknown keys Id/Foo' -> 400 in -162 while a top-level Id was kept in -008; the
  rejecting key (presumably Foo) was not isolated.
- [S3T-TABLE-154](#s3t-table-154): a bare '*' and {"AWS":"*"} (each with a Condition) are stored as sent and not canonicalized to each other;
  whether a comparator must treat them as equal is undecided.
- [S3T-TABLE-250](#s3t-table-250), [S3T-TABLE-251](#s3t-table-251), [S3T-TABLE-216](#s3t-table-216) (restore_s n=2), [S3T-TABLE-131](#s3t-table-131): single roles / single runs; [S3T-TABLE-216](#s3t-table-216)'s
  enforcement and [S3T-TABLE-251](#s3t-table-251)'s bucket-policy Deny were observed at 1 s polling only.
- [S3T-TABLE-325](#s3t-table-325): the behavior text says the 200 cell took 378-486 ms while `measurements` list 401-486 ms; the seed uses the
  measurements.
- [S3T-TABLE-158](#s3t-table-158): strong consistency was measured with one writer on one table in us-west-2; no concurrent-writer finding is
  in this partition.
<!-- preserved:end -->

## Appendix: low-impact and duplicate findings

| id | category | impact | status | title | related | duplicate_of |
| --- | --- | --- | --- | --- | --- | --- |
| <a id="s3t-table-159"></a>**S3T-TABLE-159** | response-fidelity | low | confirmed | Wire shapes: PutTablePolicy HTTP 200 body ''; DeleteTablePolicy HTTP 204 body ''; GetTablePolicy keys ['resourcePolicy'] (no tableARN) | [S3T-TABLE-210](table-maintenance.md#s3t-table-210) | - |
| <a id="s3t-table-163"></a>**S3T-TABLE-163** | request-validation | medium | confirmed | PutTablePolicy Resource: exact table ARN HTTP 200; other table's ARN HTTP 200; bucket ARN HTTP 200; name-based ARN HTTP 200; "*" HTTP 200 | [S3T-TABLEPOLICY-009](#s3t-tablepolicy-009), [S3T-TABLE-167](#s3t-table-167), [S3T-TABLE-004](#s3t-table-004) | [S3T-TABLE-167](#s3t-table-167) |
| <a id="s3t-table-181"></a>**S3T-TABLE-181** | identity | low | confirmed | Same table name in two namespaces has independent policies (NS1/NS2)... | [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-184](#s3t-table-184), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178) | - |
| <a id="s3t-table-184"></a>**S3T-TABLE-184** | request-validation | low | confirmed | Policy path with an empty namespace segment -> HTTP 404; empty name segment -> HTTP 404 (raw HTTP); SDK empty string -> NotFoundException | [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-181](#s3t-table-181), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-178](tablebucket-policy.md#s3t-tablebucket-178) | - |
| <a id="s3t-table-247"></a>**S3T-TABLE-247** | error-code | medium | confirmed | CreateTable(aws:kms) missing key -> 400 'does not exist'; key ARN of another region -> 400 'Cross region KMS keys are not supported' | - | [S3T-TABLE-086](#s3t-table-086) |
| <a id="s3t-tablepolicy-006"></a>**S3T-TABLEPOLICY-006** | other | low | confirmed | Doc claim C046 TRUE: Second PutTablePolicy replaces the document entirely: GetTablePolicy returns only the second (Sids A,B -> C) | [S3T-TABLE-153](#s3t-table-153), [S3T-TABLE-151](#s3t-table-151), [S3T-TABLEBUCKETPOLICY-009](tablebucket-policy.md#s3t-tablebucketpolicy-009), [S3T-TABLEBUCKET-064](tablebucket-policy.md#s3t-tablebucket-064), [S3T-TABLEBUCKETPOLICY-003](tablebucket-policy.md#s3t-tablebucketpolicy-003) | - |
| <a id="s3t-tablepolicy-007"></a>**S3T-TABLEPOLICY-007** | other | low | confirmed | Doc claim C047 TRUE: After DeleteTablePolicy, GetTablePolicy -> NotFoundException 404 'The specified table policy does not exist.' | [S3T-TABLE-156](#s3t-table-156), [S3T-TABLE-149](#s3t-table-149), [S3T-TABLE-158](#s3t-table-158), [S3T-TABLEBUCKETPOLICY-010](tablebucket-policy.md#s3t-tablebucketpolicy-010), [S3T-TABLEBUCKETPOLICY-011](tablebucket-policy.md#s3t-tablebucketpolicy-011), [S3T-TABLEBUCKET-061](tablebucket-policy.md#s3t-tablebucket-061), [S3T-TABLEBUCKET-108](tablebucket-encryption-storage.md#s3t-tablebucket-108), [S3T-TABLEBUCKET-177](tablebucket-maintenance-metrics.md#s3t-tablebucket-177), [S3T-TABLEBUCKETREPLICATION-023](tablebucket-replication.md#s3t-tablebucketreplication-023), [S3T-TABLEBUCKETREPLICATION-026](tablebucket-replication.md#s3t-tablebucketreplication-026), [S3T-TABLEREPLICATION-034](table-replication.md#s3t-tablereplication-034) | - |

## Supplementary notes

<!-- preserved:start -->
<!-- preserved:end -->
