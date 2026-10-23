<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-172: PutTableRecordExpirationConfiguration pre-table validation passes enabled-w/o-settings, {} value, no status, days=2^31, 1.5 to the 405 gate
_Full entry and notes of one finding; its summary entry is in
[table-record-expiration.md](../table-record-expiration.md). Generated from aws-api-semantics
`services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the finding in the lab, not here._

## Finding

- <a id="s3t-table-172"></a>**S3T-TABLE-172** `request-validation` · impact medium · unhandled (not handled in controller) · verified 2026-10-13
  **PutTableRecordExpirationConfiguration pre-table validation passes enabled-w/o-settings, {} value, no status, days=2^31, 1.5 to the 405 gate**
  On a customer ICEBERG table (where any well-formed Put -> MethodNotAllowedException 405 'Table record
  expiration configuration API calls to this table are not allowed') these 12 of 18 value variants were
  answered by the 405 table gate, i.e. passed the service's request validation: {status: enabled, settings:
  {days: 1|36500|2147483647|2147483648|1.5}}, {status: enabled} (no settings), {status: enabled, settings:
  {}}, {status: disabled, settings: {days: 7}}, {status: disabled}, {status: disabled, settings: {}},
  {settings: {days: 3}} (no status) and {} (empty value). The only variants rejected ahead of the gate were
  days<1, a non-enum status and a string-typed days (HTTP 400, see the ValidationException finding). The Get
  after all variants -> MethodNotAllowedException 405.
  - ACK: custom_update, none · ops: PutTableRecordExpirationConfiguration · fields: value, value.status,
    value.settings, value.settings.days
  - repro: CreateTable (customer) -> Put each value variant through a client with parameter_validation=False
    -> record code/HTTP/message
  - measurements: variants_passed_to_gate=12, variants_total=18, n=18
  - handling: not handled in the controller (as of commit 1eda357)
  - related: [S3T-TABLE-169](../table-record-expiration.md#s3t-table-169), [S3T-TABLE-147](../table-record-expiration.md#s3t-table-147), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-006](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-006),
    [S3T-TABLERECORDEXPIRATIONCONFIGURATION-005](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-005), [S3T-TABLE-170](../table-record-expiration.md#s3t-table-170), [S3T-TABLERECORDEXPIRATIONCONFIGURATION-002](../table-record-expiration.md#s3t-tablerecordexpirationconfiguration-002) ·
    hypotheses: H-U-114, H-U-113, H-U-115, H-U-116, H-U-117 · evidence:
    table/mutation-matrix/expiration-status-days-variants

## Notes

H-U-114 qualified: 'enabled without settings' is NOT rejected by the request validation that runs before the
table check (it reaches the 405 gate); whether an eligible (AWS-managed) table would reject it is unknown.
H-U-113 qualified: days=2147483647 and 2147483648 both pass request validation (no upper bound enforced there;
botocore caps at 2147483647 client-side). H-U-115/H-U-116/H-U-117 untestable: what is STORED after a Put
cannot be observed on customer tables (Put and Get both 405). Confidence medium because a 405 answer only
proves 'not rejected by pre-table validation'.
