<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-SERVICE-025: Fresh IAM role: assumable 8.1-9.1 s after CreateRole, AssumeRole flaps OK->AccessDenied->OK ~3 s; first STS creds -> UnrecognizedClient 403
_Full entry and notes of one finding; its summary entry is in [service.md](../service.md). Generated from
aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the
lab, not here._

## Finding

- <a id="s3t-service-025"></a>**S3T-SERVICE-025** `error-code` · impact medium · unhandled (not handled in controller) · verified 2026-10-14
  **Fresh IAM role: assumable 8.1-9.1 s after CreateRole, AssumeRole flaps OK->AccessDenied->OK ~3 s; first STS creds -> UnrecognizedClient 403**
  Roles created by the probe became assumable 8.1 / 8.75 / 8.6 / 9.1 s after CreateRole (AccessDenied 'not
  authorized to perform: sts:AssumeRole' before). For a role re-created under a name deleted seconds earlier
  AssumeRole answered OK on 6 consecutive 1-s polls and then AccessDenied on the next 2-3 calls before OK
  again; a GetTableBucket with credentials obtained ~2 s after that OK was rejected by s3tables with
  UnrecognizedClientException 403 'The security token included in the request is invalid.' (the probe paused;
  the identical call 10 s later with fresh credentials -> the expected AccessDeniedException).
  - ACK: requeue · ops: GetTableBucket
  - repro: CreateRole; poll AssumeRole at 1 s; immediately use the credentials for GetTableBucket
  - measurements: assume_role_first_ok_s=[8.1, 8.75, 8.6, 9.08], n=4
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-SERVICE-010](../service.md#s3t-service-010), [S3T-SERVICE-012](../service.md#s3t-service-012), [S3T-SERVICE-003](../service.md#s3t-service-003), [S3T-SERVICE-005](../service.md#s3t-service-005), [S3T-SERVICE-013](../service.md#s3t-service-013),
    [S3T-SERVICE-006](../service.md#s3t-service-006), [S3T-SERVICE-023](../service.md#s3t-service-023), [S3T-TABLEBUCKETPOLICY-001](../tablebucket-policy.md#s3t-tablebucketpolicy-001), [S3T-TABLE-164](../table-policy.md#s3t-table-164) · evidence:
    service/creative/iam-authz-variants, service/creative/error-catalogue

## Notes

Seed rows 'AccessDenied sts:AssumeRole (n=29)' and 'UnrecognizedClientException (n=2)' are explained: both are
IAM propagation of a role created seconds earlier, i.e. RETRY, not configuration errors. A controller that
creates an IAM Role and immediately assumes it (or hands it to another controller) should retry both codes for
~15 s.

Full original title: Fresh IAM role: sts:AssumeRole flaps OK -> AccessDenied -> OK for ~3 s and its first STS
credentials were rejected by s3tables with UnrecognizedClientException
