<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-023: Error catalogue sibling-service and transport codes: 11 message classes - 2 AMBIGUOUS, 8 PERMANENT-SPEC, 1 WAIT-FOR-STATE
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-023"></a>**S3T-SERVICE-023** `error-code` · impact medium · handled · verified 2026-10-14
  **Error catalogue sibling-service and transport codes: 11 message classes - 2 AMBIGUOUS, 8 PERMANENT-SPEC, 1 WAIT-FOR-STATE**
  sibling-service and transport codes: codes without an s3tables shape: STS/KMS/IAM/S3 siblings, SigV4 and
  credential failures, routing 404s, client-side validation. 11 message classes; classification counts
  {'AMBIGUOUS': 2, 'PERMANENT-SPEC': 8, 'WAIT-FOR-STATE': 1}. Codes without an s3tables shape; mostly client
  configuration, two transient windows after IAM changes.
  - ACK: terminal_codes, docs-only
  - repro: Classification of seed-error-messages.md rows plus the live creative probes listed in notes; each
    row cites the finding ids that produced it
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-SERVICE-003](../service.md#s3t-service-003), [S3T-SERVICE-005](../service.md#s3t-service-005), [S3T-SERVICE-013](../service.md#s3t-service-013), [S3T-SERVICE-006](../service.md#s3t-service-006), [S3T-SERVICE-025](../service.md#s3t-service-025),
    [S3T-SERVICE-001](../service.md#s3t-service-001), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-SERVICE-017](../service.md#s3t-service-017), [S3T-SERVICE-018](../service.md#s3t-service-018) · evidence:
    service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
    service/creative/iam-authz-variants, service/creative/throttle-scope, service/creative/error-catalogue

## Notes

Classification key: RETRY = retry the identical request after a short backoff; WAIT-FOR-STATE = the request is
right but a state must change first (own resource, parent/child, IAM/KMS propagation, deletion cooldown, or
the controller's own stale versionToken - re-read then retry); PERMANENT-SPEC = the request is wrong or the
desired state already holds (non-idempotent deletes), terminal until the spec changes; PERMANENT-QUOTA = a
limit; AUTHZ = permissions (IAM/resource/boundary/session policy, KMS key policy, cross-account); AMBIGUOUS =
the same code+message has different causes with different remediations (named in the rule). Sources:
seed-error-messages.md (105 structured probes, 391 rows; finding ids cited per row) plus the creative-round
probes service/creative/raw-wire-codes, service/creative/sig-skew-and-empty-500,
service/creative/iam-authz-variants, service/creative/throttle-scope ([S3T-SERVICE-001](../service.md#s3t-service-001)..013). Messages are
quoted as recorded; <ARN>, <N>, <T> are normalised. Controllers should branch on the wire code and HTTP status
and treat the message as diagnostic text only - several texts echo caller input or differ by a curly
apostrophe.

\## sibling-service and transport codes

| message class | operations | HTTP | classification | controller rule |
|---|---|---|---|---|
| AccessDenied: User: <ARN> is not authorized to perform: sts:AssumeRole on resource: <ARN> (n=29 + live) |
sts:AssumeRole on a role created seconds earlier | 403 | AMBIGUOUS | (a) IAM propagation: a NEW role is
assumable 8.1 / 8.75 / 8.6 / ~9 s after CreateRole (n=4, [S3T-SERVICE-010](../service.md#s3t-service-010)/012 evidence) and the answer FLAPS: a
role that just returned OK was denied again 1-3 s later on a different attempt ([S3T-SERVICE-012](../service.md#s3t-service-012) evidence: OK
then AccessDenied x3 then OK) -> RETRY for ~15 s; (b) a real trust-policy problem -> AUTHZ. Freshly issued
credentials of such a role were also rejected ONCE by s3tables with UnrecognizedClientException (see below). |
| UnrecognizedClientException: The security token included in the request is invalid. (n=2 + live) | any
s3tables op | 403 | AMBIGUOUS | (a) bogus access key id; a session token garbled in its head/middle, truncated
or empty ([S3T-SERVICE-003](../service.md#s3t-service-003)/006) -> credential configuration, PERMANENT until fixed; (b) TRANSIENT: STS
credentials issued ~2 s after a role became assumable were rejected by s3tables with this code once, and
accepted 10 s later ([S3T-SERVICE-012](../service.md#s3t-service-012) evidence, probe paused/resumed). Token TAIL alterations (last 1-4 chars,
one appended char) were accepted - the tail is not covered by the token's integrity check. The harness and
most SDK credential providers treat this code as 'refresh credentials'. |
| InvalidSignatureException: The request signature we calculated does not match the signature you provided...
/ Credential should be scoped to a valid region. / Credential should be scoped to correct service: 's3tables'.
(live) | any op: wrong secret / signed for another region / signed for service 's3' | 403 | PERMANENT-SPEC |
Client configuration (endpoint override with the wrong signing region/name). No SDK retry. [S3T-SERVICE-003](../service.md#s3t-service-003). |
| InvalidSignatureException: Signature expired: <T> is now earlier than <now-5min> (<now> - 5 min.) /
Signature not yet current: <T> is still later than <now+5min> (<now> + 5 min.) (live) | any op from a host
whose clock is off by more than 5 min | 403 | WAIT-FOR-STATE | The tolerance is +/-5 minutes (accepted
-4.5/+4.5, rejected -5.5/+5.5, [S3T-SERVICE-005](../service.md#s3t-service-005)/013), tighter than the 15 min most AWS services allow. A
drifting node fails every s3tables call; botocore does not auto-correct clock skew for this code. |
| MissingAuthenticationTokenException: Missing Authentication Token (live) | any unsigned request | 403 |
PERMANENT-SPEC | No Authorization header ([S3T-SERVICE-003](../service.md#s3t-service-003)). |
| HTTP 404, body <UnknownOperationException/> (XML, no x-amzn-ErrorType) (n=6 + live) | unknown path, wrong
HTTP method on a valid path (POST /buckets, PATCH, DELETE /buckets, HEAD), extra/unknown sub-resource segment,
type segment on a GET, unencoded ARN in the path | 404 | PERMANENT-SPEC | Routing miss, not a resource 404:
botocore surfaces code '404' with the XML as message (seed row '404 <UnknownOperationException/>',
[S3T-TABLE-184](../table-policy.md#s3t-table-184), [S3T-TABLEBUCKET-039](../tablebucket.md#s3t-tablebucket-039), [S3T-SERVICE-001](../service.md#s3t-service-001)). An error-code switch must not treat code '404' as
NotFoundException. HEAD returns 404 with an empty body. |
| MethodNotAllowed: The specified method is not allowed against this resource. (n=12) | s3:ListObjectsV2,
s3:CopyObject against a table bucket's warehouse | 405 | PERMANENT-SPEC | Sibling S3 data plane: table storage
is not a general-purpose bucket. |
| 404 Not Found (n=3) | s3:HeadObject on a warehouse object | 404 | PERMANENT-SPEC | Sibling S3; e.g. probing
a metadata file that UpdateTableMetadataLocation accepted without checking ([S3T-TABLE-100](../table-metadata-location.md#s3t-table-100)). |
| KMSInvalidStateException: <ARN> is not pending deletion. (n=1) | kms:CancelKeyDeletion | 400 |
PERMANENT-SPEC | Sibling KMS (probe cleanup). |
| NoSuchEntity: The role with name <name> cannot be found. (n=1) | iam:DeleteRole | 404 | PERMANENT-SPEC |
Sibling IAM (idempotent delete handling for roles created by probes). |
| ParamValidationError (botocore, client-side; the Go SDK has the equivalent smithy validation error) | any op
with a missing required member, wrong type, or a value outside the modeled enum/length | - | PERMANENT-SPEC |
Never reaches AWS; retrying never helps. Several server behaviors above are reachable only by disabling it
(schema={}, configuration absent, empty strings). |

Classification counts: {'AMBIGUOUS': 2, 'PERMANENT-SPEC': 8, 'WAIT-FOR-STATE': 1}.
