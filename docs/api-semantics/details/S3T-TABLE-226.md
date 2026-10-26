<!-- generated: aws-api-semantics s3tables 2026-10-23 model=2018-05-10 controller=1eda357 -->
# S3T-TABLE-226: targetFileSizeMB: 64..512 enforced server-side (400 names the bounds); 100.5 accepted and stored as 100; "256"/true -> 400
_Full entry and notes of one finding; its summary entry is in [table-maintenance.md](../table-maintenance.md).
Generated from aws-api-semantics `services/s3tables` (model 2018-05-10, controller commit 1eda357); edit the
finding in the lab, not here._

## Finding

- <a id="s3t-table-226"></a>**S3T-TABLE-226** `request-validation` · impact high · handled · verified 2026-10-13
  **targetFileSizeMB: 64..512 enforced server-side (400 names the bounds); 100.5 accepted and stored as 100; "256"/true -> 400**
  PutTableMaintenanceConfiguration icebergCompaction with parameter validation disabled: accepted (HTTP 204,
  read back unchanged unless noted) for 64, 65, 100, 511, 512, float_100.5; rejected: {"-1":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "0": "BadRequestException/400 'The specified compaction file size is invalid. Minimum
  value: 64 MB, maximum value: 512 MB.'", "1": "BadRequestException/400 'The specified compaction file size is
  invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "32": "BadRequestException/400 'The specified
  compaction file size is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "63":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "513": "BadRequestException/400 'The specified compaction file size is invalid. Minimum
  value: 64 MB, maximum value: 512 MB.'", "1024": "BadRequestException/400 'The specified compaction file size
  is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "10240": "BadRequestException/400 'The specified
  compaction file size is invalid. Minimum value: 64 MB, maximum value: 512 MB.'", "2147483647":
  "BadRequestException/400 'The specified compaction file size is invalid. Minimum value: 64 MB, maximum
  value: 512 MB.'", "2147483648": "BadRequestException/400 'The specified compaction file size is invalid.
  Minimum value: 64 MB, maximum value: 512 MB.'", "string_256": "BadRequestException/400 'Invalid input.
  Consult the S3 Tables and/or Iceberg REST Catalog API reference for constraints on request parameters.'",
  "bool_true": "BadRequestException/400 'Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API
  reference for constraints on request parameters.'", "sdk_validation_0": "ParamValidationError/None
  'Parameter validation failed:\nInvalid value for parameter
  value.settings.icebergCompaction.targetFileSizeMB, value: 0, valid min value: 1'"}. Read-backs of accepted
  values: {"64": {"targetFileSizeMB": 64, "strategy": "auto"}, "65": {"targetFileSizeMB": 65, "strategy":
  "auto"}, "100": {"targetFileSizeMB": 100, "strategy": "auto"}, "511": {"targetFileSizeMB": 511, "strategy":
  "auto"}, "512": {"targetFileSizeMB": 512, "strategy": "auto"}, "float_100.5": {"targetFileSizeMB": 100,
  "strategy": "auto"}}. With boto3 validation enabled, 0 -> ParamValidationError/None before any request
  (model min=1).
  - ACK: terminal_codes · ops: PutTableMaintenanceConfiguration · fields:
    value.settings.icebergCompaction.targetFileSizeMB
  - repro: Put compaction targetFileSizeMB in [-1,0,1,32,63,64,65,100,511,512,513,1024,10240,2^31-1,2^31] with
    parameter_validation=False; Get after each
  - handling: handled via `generator.yaml:99-104; generator.yaml:170-175`
  - related: [S3T-TABLE-227](../table-maintenance.md#s3t-table-227), [S3T-TABLE-228](../table-maintenance.md#s3t-table-228), [S3T-TABLE-229](../table-maintenance.md#s3t-table-229), [S3T-TABLE-230](../table-maintenance.md#s3t-table-230) · hypotheses: H-U-040, H-U-041 ·
    evidence: table/weird-inputs/maintenance-bounds

## Notes

H-U-040 confirmed (0, 63, 513, 10240 and -1 -> BadRequestException 400 'The specified compaction file size is
invalid. Minimum value: 64 MB, maximum value: 512 MB.'; 64 and 512 accepted and read back unchanged); H-U-041
refuted (1024/10240/32 rejected). Extra cells: a float 100.5 is accepted and truncated to 100; a JSON string
"256" or a boolean -> 400 'Invalid input. Consult the S3 Tables and/or Iceberg REST Catalog API reference...'.
boto3's own validation (model min=1) rejects 0/-1 before any request.
