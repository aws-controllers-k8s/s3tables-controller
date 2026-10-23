<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-041: ListTableBuckets: bad/stale token -> BadRequestException 400; token reused after a delete skips (200); invalid prefixes -> 200 empty
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-041"></a>**S3T-TABLEBUCKET-041** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **ListTableBuckets: bad/stale token -> BadRequestException 400; token reused after a delete skips (200); invalid prefixes -> 200 empty**
  continuationToken='abc' -> BadRequestException HTTP 400 'The continuation token is no longer valid.'; 2049
  chars -> BadRequestException HTTP 400 'The continuation token is no longer valid.'; page-1 token
  (maxBuckets=1) re-sent after DeleteTableBucket of the next bucket -> HTTP 200 ['ackq-<run>-ordz']
  next_token=False; the same token with a different prefix -> n/a. prefix validation: {"Ackq": "HTTP 200 (0
  items)", "ackq_b": "HTTP 200 (0 items)", "ackq.b": "HTTP 200 (0 items)", "aaaaaaaaaa..(64)": "HTTP 200 (0
  items)", "aaaaaaaaaa..(63)": "HTTP 200 (0 items)", "<empty>": "HTTP 200 (7 items)", "ackq-": "HTTP 200 (7
  items)", "ackq-0a78c..(16)": "HTTP 200 (1 items)"}. maxBuckets=0 -> BadRequestException HTTP 400 'The
  specified limit is not valid. A limit should be a positive number'; maxBuckets=1001 -> HTTP 200.
  - ACK: custom_find, terminal_codes · ops: ListTableBuckets · fields: continuationToken, prefix, maxBuckets
  - repro: ListTableBuckets with bad tokens / prefixes / maxBuckets values (parameter_validation=False)
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLEBUCKET-040](../tablebucket.md#s3t-tablebucket-040), [S3T-TABLEBUCKET-069](../tablebucket.md#s3t-tablebucket-069), [S3T-TABLEBUCKET-130](../tablebucket.md#s3t-tablebucket-130), [S3T-TABLEBUCKET-125](../tablebucket.md#s3t-tablebucket-125) · hypotheses:
    H-B-051, H-B-052 · evidence: tablebucket/error-taxonomy/arn-forms-get-delete-tags

## Notes

H-B-051 token half: confirmed for garbage and over-long tokens (BadRequestException 'The continuation token is
no longer valid.'), but a valid page token re-sent after the next bucket was deleted returned HTTP 200 with
the following bucket (no error, no duplicate). H-B-052 REFUTED: prefix='Ackq', 'ackq_b', 'ackq.b' and a
64-char prefix all returned HTTP 200 with an empty list (no character-class or max-length validation), and an
empty prefix behaved like no prefix (full list). Also maxBuckets=1001 is accepted (200) while maxBuckets=0 ->
BadRequestException 'The specified limit is not valid. A limit should be a positive number'.
