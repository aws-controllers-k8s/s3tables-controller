<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-003: GET /buckets credential failures (403): UnrecognizedClientException (bogus key), InvalidSignatureException (secret/region/service), unsigned
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-003"></a>**S3T-SERVICE-003** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **GET /buckets credential failures (403): UnrecognizedClientException (bogus key), InvalidSignatureException (secret/region/service), unsigned**
  bogus access key -> HTTP 403 UnrecognizedClientException "{\"message\":\"The security token included in the
  request is invalid.\"}"; valid key, wrong secret -> HTTP 403 InvalidSignatureException "{\"message\":\"The
  request signature we calculated does not match the signature you provided. Check your AWS Secret Acc; valid
  key+secret, session token garbled -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  signed for us-east-1 -> HTTP 403 InvalidSignatureException "{\"message\":\"Credential should be scoped to a
  valid region. \"}"; signed for service s3 -> HTTP 403 InvalidSignatureException "{\"message\":\"Credential
  should be scoped to correct service: 's3tables'. \"}"; clock skew -20 min -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  clock skew -4 min (within tolerance?) -> HTTP 200 -
  "{\"tableBuckets\":[{\"arn\":\"arn:aws:s3tables:us-west-2:<ACCOUNT>:bucket/ackq-432cd7-rawwire\",\"createdAt\":\"2026;
  no Authorization header -> HTTP 403 MissingAuthenticationTokenException "{\"message\":\"Missing
  Authentication Token\"}".
  - ACK: terminal_codes, requeue · ops: ListTableBuckets
  - repro: SigV4-sign GET /buckets with a bogus key / wrong secret / wrong region / wrong service / skewed
    X-Amz-Date / no Authorization
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-005](../service.md#s3t-service-005), [S3T-SERVICE-013](../service.md#s3t-service-013), [S3T-SERVICE-006](../service.md#s3t-service-006), [S3T-SERVICE-025](../service.md#s3t-service-025), [S3T-SERVICE-023](../service.md#s3t-service-023) · evidence:
    service/creative/raw-wire-codes

## Notes

Creative round, seed error-messages, gap 'UnrecognizedClientException and sibling codes'. All are HTTP 403 (or 400)
with no s3tables shape; none is a permission problem: they mean the controller's credentials/endpoint
configuration is wrong (or the pod clock is off) and recur until that is fixed - requeue with backoff, do not
mark the resource terminal.

Full original title: Credential/signature failures on GET /buckets: InvalidSignatureException/403,
MissingAuthenticationTokenException/403, UnrecognizedClientException/

Contradiction with [S3T-SERVICE-013](../service.md#s3t-service-013), [S3T-SERVICE-005](../service.md#s3t-service-005): 003's behavior records 'clock skew -20 min -> HTTP 200'
(and '-4 min (within tolerance?) -> 200'), i.e. a 20-minute skew accepted; 005 records the same -20 min shift
(X-Amz-Date 20261014T005710Z vs now 20261014T011710Z) -> InvalidSignatureException 403 'Signature expired' and
013 pins the tolerance at +/-5 min with the server message quoting '(now - 5 min.)'. The -20 min cell of 003
disagrees on the outcome; 003's cell is self-flagged uncertain and shows no X-Amz-Date, the two dedicated skew
probes agree with each other and with the server's own message Resolution: keep both; [S3T-SERVICE-013](../service.md#s3t-service-013) is
canonical for clock skew (005 is its duplicate and independently reproduces -20 min -> 403); 003's skew cells
are unreliable and its title does not claim them
