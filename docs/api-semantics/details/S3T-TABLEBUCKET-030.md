<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLEBUCKET-030: TagResource length violations return the NON-modeled ValidationException (HTTP 400), unlike CreateTableBucket's BadRequestException
_Full entry and notes of one finding; its summary entry is in [tablebucket.md](../tablebucket.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

- <a id="s3t-tablebucket-030"></a>**S3T-TABLEBUCKET-030** `error-code` · impact high · SUSPECTED CONTROLLER BUG · verified 2026-10-13
  **TagResource length violations return the NON-modeled ValidationException (HTTP 400), unlike CreateTableBucket's BadRequestException**
  TagResource({'k'*128: 'v'*256}) -> HTTP 200. TagResource({ok: 1, 'k'*129: v}) -> ValidationException HTTP
  400 "1 validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy
  constraint: [Member must have length less than or equal to 128, Member must have length greater than or
  equal to 1]"; 'ok' present afterwards: False. TagResource({ok2: 1, k: 'v'*257}) -> ValidationException HTTP
  400 "1 validation error detected: Value at 'tags' failed to satisfy constraint: Map value must satisfy
  constraint: [Member must have length less than or equal to 256, Member must have length greater than or
  equal to 0]"; 'ok2' present afterwards: False.
  - ACK: tags.custom-sync, terminal_codes · ops: TagResource, ListTagsForResource · fields: tags
  - repro: TagResource with a 128/256 pair, then {ok, 129-char key}, then {ok2, 257-char value};
    ListTagsForResource after each
  - handling: suspected controller bug - see Handling gaps
  - related: [S3T-TABLEBUCKET-025](../tablebucket.md#s3t-tablebucket-025), [S3T-TABLEBUCKET-029](../tablebucket.md#s3t-tablebucket-029), [S3T-TABLEBUCKET-031](../tablebucket.md#s3t-tablebucket-031) · hypotheses: H-B-045 · evidence:
    tablebucket/tags/upsert-limits-chars-scope

## Notes

Confirms H-B-045 for the limits (128/256 accepted, 129/257 rejected, all-or-nothing per call) but QUALIFIES
the code: TagResource (and an empty key) fail with code 'ValidationException' and the framework message '1
validation error detected: Value at 'tags' failed to satisfy constraint: Map keys must satisfy constraint:
[Member must have length less than or equal to 128, Member must have length greater than or equal to 1]' - a
code that is not in the s3tables model (CreateTableBucket with the same bad tags returns BadRequestException
'The provided tags are invalid. Tag key is empty or exceeds max length'). A terminal-code list built from the
model misses it.

Suspected controller bug confirmed by evidence: not flagged; evidence-driven: TagResource rejects a key >128
chars, an empty key or a value >256 chars with the non-modeled code ValidationException 400 (-030, -029),
which is not in terminal_codes, so a TableBucket whose spec.tags gain such a key/value after creation requeues
on the tag sync forever instead of going Terminal, while the same tags at create are BadRequestException and
Terminal (-025); add ValidationException to terminal_codes or enforce tag key/value lengths in the CRD.
