<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-062: DeleteTableBucketPolicy is idempotent: HTTP 204 when no policy exists and on a second Delete (not NotFoundException)
_Full entry and notes of one finding; its summary entry is in
[tablebucket-policy.md](../tablebucket-policy.md). Generated from aws-api-semantics `services/s3tables` (model
2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

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
  - related: [S3T-TABLEBUCKETPOLICY-015](../tablebucket-policy.md#s3t-tablebucketpolicy-015), [S3T-TABLEBUCKET-065](../tablebucket-policy.md#s3t-tablebucket-065), [S3T-TABLE-150](../table-policy.md#s3t-table-150), [S3T-TABLE-156](../table-policy.md#s3t-table-156) · hypotheses:
    H-S-014 · evidence: tablebucket/sub-resources/policy-get-put-delete

## Notes

H-S-014 REFUTED: Delete-when-absent and the repeated Delete both returned HTTP 204, no NotFoundException; the
Put->Delete->Get part held (policy gone on the first read, 404 'The specified bucket policy does not exist.').
Contrast: DeleteTableBucketEncryption when never set returns NotFoundException 404 (encryption-get-put-delete
probe).

Suspected controller bug confirmed by evidence: Reconciling an adopted or out-of-band-policied table bucket
with spec.policy unset issues DeleteTableBucketPolicy, which returns HTTP 204 with no precondition and removes
the live policy on the very first read ([S3T-TABLEBUCKET-062](../tablebucket-policy.md#s3t-tablebucket-062); 51/51 reads 404 afterwards per
[S3T-TABLEBUCKETPOLICY-003](../tablebucket-policy.md#s3t-tablebucketpolicy-003)), and if that policy denies the controller principal
s3tables:DeleteTableBucketPolicy the same delete instead fails every reconcile with AccessDeniedException 403,
a code the model does not list ([S3T-TABLEBUCKET-091](../tablebucket-policy.md#s3t-tablebucket-091)); treat nil as unmanaged and reserve the delete for an
explicit empty document, which the API itself cannot accept as a Put (ValidationException 400 'Member must
have length greater than or equal to 1', [S3T-TABLEBUCKET-090](../tablebucket-policy.md#s3t-tablebucket-090)), so the empty-vs-nil distinction is the only
non-destructive way to express removal. The empty-string-vs-nil perpetual-diff half of the suspicion is
controller-internal and would be silent on the wire (Delete is idempotent 204), so it is not settled by API
evidence.
