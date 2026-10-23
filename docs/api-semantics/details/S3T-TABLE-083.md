<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-083: Wire error codes equal the modeled shape names (BadRequest=400, Forbidden=403, NotFound=404)...
_Full entry and notes of one finding; its summary entry is in [table-policy.md](../table-policy.md). Generated
from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in
the lab, not here._

## Finding

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
  - related: [S3T-TABLEBUCKET-109](../tablebucket-encryption-storage.md#s3t-tablebucket-109), [S3T-TABLEBUCKET-067](../tablebucket-encryption-storage.md#s3t-tablebucket-067), [S3T-TABLE-250](../table-policy.md#s3t-table-250), [S3T-TABLE-109](../table-policy.md#s3t-table-109), [S3T-TABLE-120](../table.md#s3t-table-120),
    [S3T-TABLE-026](../table-policy.md#s3t-table-026), [S3T-TABLE-081](../table.md#s3t-table-081) · hypotheses: H-T-047 · evidence: table/error-taxonomy/create-validation

## Notes

H-T-047 confirmed for the codes reached here (BadRequestException, NotFoundException, ForbiddenException);
ConflictException/TooManyRequestsException/ForbiddenException are covered by the idempotency and identity
probes.

Full original title: Wire error codes equal the modeled shape names (BadRequest=400, Forbidden=403,
NotFound=404); x-amzn-ErrorType header matches, body has only 'message'
