<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-NAMESPACE-029: Parent tableBucketARN: malformed/wrong region -> BadRequestException; missing bucket -> NotFoundException; other account -> Forbidden 403
_Full entry and notes of one finding; its summary entry is in [namespace.md](../namespace.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-namespace-029"></a>**S3T-NAMESPACE-029** `error-code` · impact high · handled · verified 2026-10-13
  **Parent tableBucketARN: malformed/wrong region -> BadRequestException; missing bucket -> NotFoundException; other account -> Forbidden 403**
  Per ARN form, outcomes of GetNamespace / ListNamespaces / CreateNamespace / GetTableBucket: [bucket name,
  not an ARN] GetNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  GetTableBucket=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'; [well-formed
  ARN, bucket does not exist] GetNamespace=NotFoundException (HTTP 404) 'The specified bucket does not exist.'
  | ListNamespaces=NotFoundException (HTTP 404) 'The specified bucket does not exist.' |
  CreateNamespace=NotFoundException (HTTP 404) 'The specified bucket does not exist.' |
  GetTableBucket=NotFoundException (HTTP 404) 'The specified bucket does not exist.'; [real ARN with region
  swapped] GetNamespace=BadRequestException (HTTP 400) 'The provided Region does not match the Region of the
  resource.' | ListNamespaces=BadRequestException (HTTP 400) 'The provided Region does not match the Region of
  the resource.' | CreateNamespace=BadRequestException (HTTP 400) 'The provided Region does not match the
  Region of the resource.' | GetTableBucket=BadRequestException (HTTP 400) 'The provided Region does not match
  the Region of the resource.'; [real ARN with bucket name uppercased] GetNamespace=BadRequestException (HTTP 400)
  'The specified bucket name isn't valid. Specify a different bucket name, and then try again.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different
  bucket name, and then try again.' | CreateNamespace=BadRequestException (HTTP 400) 'The specified bucket
  name isn't valid. Specify a different bucket name, and then try again.' | GetTableBucket=BadRequestException
  (HTTP 400) 'The specified bucket name isn't valid. Specify a different bucket name, and then try again.';
  [well-formed ARN, other account (doc sample id)] GetNamespace=ForbiddenException (HTTP 403) 'Unauthorized' |
  ListNamespaces=ForbiddenException (HTTP 403) 'Unauthorized' | CreateNamespace=ForbiddenException (HTTP 403)
  'Unauthorized' | GetTableBucket=ForbiddenException (HTTP 403) 'Unauthorized'; [general-purpose S3 bucket
  ARN] GetNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  ListNamespaces=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.' |
  GetTableBucket=BadRequestException (HTTP 400) 'The specified table bucket ARN is not valid.'; [table ARN
  form under the bucket] GetNamespace=BadRequestException (HTTP 400) 'The specified bucket name isn't valid.
  Specify a different bucket name, and then try again.' | ListNamespaces=BadRequestException (HTTP 400) 'The
  specified bucket name isn't valid. Specify a different bucket name, and then try again.' |
  CreateNamespace=BadRequestException (HTTP 400) 'The specified bucket name isn't valid. Specify a different
  bucket name, and then try again.' | GetTableBucket=BadRequestException (HTTP 400) 'The specified bucket name
  isn't valid. Specify a different bucket name, and then try again.'. GetNamespace(real bucket, missing
  namespace) -> NotFoundException (HTTP 404) 'The specified namespace does not exist.'; GetNamespace(missing
  bucket, missing namespace) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  CreateNamespace(missing bucket) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.';
  DeleteNamespace(missing bucket) -> NotFoundException (HTTP 404) 'The specified bucket does not exist.'.
  - ACK: terminal_codes, exceptions.404, references · ops: GetNamespace, ListNamespaces, CreateNamespace,
    DeleteNamespace, GetTableBucket · fields: tableBucketARN
  - repro: GetNamespace/ListNamespaces/CreateNamespace/GetTableBucket with each tableBucketARN form in
    behavior
  - handling: handled via `pkg/resource/namespace/sdk.go:122-147; pkg/resource/namespace/resource.go:88-116; generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLEBUCKET-162](../tablebucket.md#s3t-tablebucket-162), [S3T-TABLEBUCKET-163](../tablebucket-policy.md#s3t-tablebucket-163), [S3T-TABLEBUCKET-164](../tablebucket.md#s3t-tablebucket-164), [S3T-TABLEBUCKET-038](../tablebucket.md#s3t-tablebucket-038),
    [S3T-TABLEBUCKET-127](../tablebucket.md#s3t-tablebucket-127), [S3T-TABLEBUCKET-066](../tablebucket-encryption-storage.md#s3t-tablebucket-066), [S3T-NAMESPACE-017](../namespace.md#s3t-namespace-017), [S3T-NAMESPACE-030](../namespace.md#s3t-namespace-030), [S3T-NAMESPACE-031](../namespace.md#s3t-namespace-031),
    [S3T-NAMESPACE-003](../namespace.md#s3t-namespace-003), [S3T-NAMESPACE-001](../namespace.md#s3t-namespace-001), [S3T-NAMESPACE-025](../namespace.md#s3t-namespace-025), [S3T-NAMESPACE-026](../namespace.md#s3t-namespace-026), [S3T-NAMESPACE-028](../namespace.md#s3t-namespace-028) ·
    hypotheses: H-N-028, H-N-029 · evidence: namespace/error-taxonomy/parent-arn-iam-deny-cross-region

## Notes

H-N-028 qualified (see per-form codes). H-N-029 confirmed: missing-namespace and missing-bucket are
distinguishable by message ('The specified namespace does not exist.' vs 'The specified bucket does not
exist.').

Full original title: Parent tableBucketARN problems: not-an-ARN -> BadRequestException, missing bucket ->
NotFoundException, other account -> ForbiddenException, region-swapped -> BadRequestException (Get/List/Create
alike)
