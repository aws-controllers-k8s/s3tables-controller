<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-017: Wire error codes match the modeled shape names (BadRequestException, ConflictException, NotFoundException)...
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-017"></a>**S3T-NAMESPACE-017** `error-code` · impact medium · handled · verified 2026-10-13
  **Wire error codes match the modeled shape names (BadRequestException, ConflictException, NotFoundException)...**
  Across 3 distinct error codes seen in this probe, botocore Error.Code vs wire: BadRequestException -> HTTP
  400, header 'BadRequestException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body
  {"message": "Multipart namespaces are not supported."}; ConflictException -> HTTP 409, header
  'ConflictException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body {"message": "A
  namespace with an identical name already exists in the bucket."}; NotFoundException -> HTTP 404, header
  'NotFoundException:http://internal.amazon.com/coral/com.amazon.s3tablebuckets/', body {"message": "The
  specified namespace does not exist."}. No NoSuchBucket/NoSuchKey style codes appeared.
  - ACK: terminal_codes, exceptions.404 · ops: CreateNamespace, GetNamespace, DeleteNamespace, ListNamespaces
  - repro: Force BadRequestException (bad name), ConflictException (duplicate create), NotFoundException
    (missing namespace); read the raw HTTP response
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-NAMESPACE-029](../namespace.md#s3t-namespace-029), [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-NAMESPACE-031](../namespace.md#s3t-namespace-031), [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003) · hypotheses: H-N-030
    · evidence: namespace/weird-inputs/naming-edge-cases

## Notes

H-N-030 confirmed: the HTTP header x-amzn-ErrorType is '<ShapeName>:http://internal.amazon.com/coral/...' and
botocore strips the suffix; the JSON body carries only 'message'.

Full original title: Wire error codes match the modeled shape names (BadRequestException, ConflictException,
NotFoundException); x-amzn-ErrorType carries '<Code>:<uri>', body has only a message
