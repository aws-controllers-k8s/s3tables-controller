// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package table

import (
	"context"
	"fmt"
	"math"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcapitypes "github.com/aws-controllers-k8s/s3tables-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3tables"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
)

// tableARNFromStatus returns the Table's ARN from the common
// ACKResourceMetadata, used as the ResourceArn for the Tag/Untag/ListTags APIs.
func tableARNFromStatus(ko *svcapitypes.Table) *string {
	if ko.Status.ACKResourceMetadata != nil && ko.Status.ACKResourceMetadata.ARN != nil {
		return (*string)(ko.Status.ACKResourceMetadata.ARN)
	}
	return nil
}

// setResourceTags augments the Table spec with the resource tags, which are not
// returned by GetTable. They have a dedicated ListTagsForResource API. Reading
// them here keeps observed state consistent with desired state and prevents
// spurious deltas on every reconciliation.
func (rm *resourceManager) setResourceTags(
	ctx context.Context,
	ko *svcapitypes.Table,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceTags")
	defer func() { exit(err) }()

	arn := tableARNFromStatus(ko)
	if arn == nil {
		return nil
	}

	tagsResp, err := rm.sdkapi.ListTagsForResource(
		ctx,
		&svcsdk.ListTagsForResourceInput{ResourceArn: arn},
	)
	rm.metrics.RecordAPICall("READ_ONE", "ListTagsForResource", err)
	if err != nil {
		return err
	}
	if len(tagsResp.Tags) > 0 {
		ko.Spec.Tags = aws.StringMap(tagsResp.Tags)
	} else {
		ko.Spec.Tags = nil
	}

	return nil
}

// maintenanceToInt32 narrows a CRD int64 maintenance value to the SDK's int32,
// returning an error rather than silently overflowing. The CRD models these as
// int64 but the S3 Tables API is int32; without the guard a value above the
// int32 max would wrap to a negative number and be sent to AWS.
func maintenanceToInt32(v int64) (int32, error) {
	if v < 0 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d out of range for int32 (0..%d)", v, math.MaxInt32)
	}
	return int32(v), nil
}

// setMaintenanceConfiguration augments the Table spec with the table-level
// maintenance configuration, which is not returned by GetTable. It has a
// dedicated GetTableMaintenanceConfiguration API.
func (rm *resourceManager) setMaintenanceConfiguration(
	ctx context.Context,
	ko *svcapitypes.Table,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setMaintenanceConfiguration")
	defer func() { exit(err) }()

	if ko.Spec.TableBucketARN == nil || ko.Spec.Namespace == nil || ko.Spec.Name == nil {
		return nil
	}

	resp, err := rm.sdkapi.GetTableMaintenanceConfiguration(
		ctx,
		&svcsdk.GetTableMaintenanceConfigurationInput{
			TableBucketARN: ko.Spec.TableBucketARN,
			Namespace:      ko.Spec.Namespace,
			Name:           ko.Spec.Name,
		},
	)
	rm.metrics.RecordAPICall("READ_ONE", "GetTableMaintenanceConfiguration", err)
	if err != nil {
		return err
	}

	cfg := make(map[string]*svcapitypes.TableMaintenanceConfigurationValue, len(resp.Configuration))
	for jobType, val := range resp.Configuration {
		// Only map maintenance types this controller understands. A new
		// service-side type would otherwise be half-mapped (status captured but
		// its settings union member unrecognized and dropped), producing a spec
		// that cannot round-trip. Skip unknown types so they are neither
		// surfaced nor mutated; they can be added when the CRD supports them.
		switch jobType {
		case string(svcsdktypes.TableMaintenanceTypeIcebergCompaction),
			string(svcsdktypes.TableMaintenanceTypeIcebergSnapshotManagement):
		default:
			rlog.Info("skipping unrecognized table maintenance type", "type", jobType)
			continue
		}
		cfg[jobType] = maintenanceValueFromSDK(val)
	}
	if len(cfg) > 0 {
		ko.Spec.MaintenanceConfiguration = cfg
	} else {
		ko.Spec.MaintenanceConfiguration = nil
	}

	return nil
}

// maintenanceValueFromSDK maps an SDK maintenance configuration value into the
// generated ACK API type, unwrapping the settings union.
func maintenanceValueFromSDK(
	val svcsdktypes.TableMaintenanceConfigurationValue,
) *svcapitypes.TableMaintenanceConfigurationValue {
	out := &svcapitypes.TableMaintenanceConfigurationValue{}
	if val.Status != "" {
		out.Status = aws.String(string(val.Status))
	}
	switch s := val.Settings.(type) {
	case *svcsdktypes.TableMaintenanceSettingsMemberIcebergCompaction:
		settings := &svcapitypes.IcebergCompactionSettings{}
		if s.Value.Strategy != "" {
			settings.Strategy = aws.String(string(s.Value.Strategy))
		}
		if s.Value.TargetFileSizeMB != nil {
			settings.TargetFileSizeMB = aws.Int64(int64(*s.Value.TargetFileSizeMB))
		}
		out.Settings = &svcapitypes.TableMaintenanceSettings{
			IcebergCompaction: settings,
		}
	case *svcsdktypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement:
		settings := &svcapitypes.IcebergSnapshotManagementSettings{}
		if s.Value.MaxSnapshotAgeHours != nil {
			settings.MaxSnapshotAgeHours = aws.Int64(int64(*s.Value.MaxSnapshotAgeHours))
		}
		if s.Value.MinSnapshotsToKeep != nil {
			settings.MinSnapshotsToKeep = aws.Int64(int64(*s.Value.MinSnapshotsToKeep))
		}
		out.Settings = &svcapitypes.TableMaintenanceSettings{
			IcebergSnapshotManagement: settings,
		}
	}
	return out
}

// customUpdateTable reconciles the mutable surface of a Table. There is no
// generic UpdateTable API; the mutable aspects are:
//   - the table name -> RenameTable (newName), guarded by the version token
//   - tags -> TagResource / UntagResource
//   - maintenance configuration -> PutTableMaintenanceConfiguration
//
// All other fields (namespace, format, metadata, encryption, storage class)
// are create-only/immutable.
func (rm *resourceManager) customUpdateTable(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateTable")
	defer func() { exit(err) }()

	// Start from desired spec, carry over observed status.
	ko := desired.ko.DeepCopy()
	ko.Status = *latest.ko.Status.DeepCopy()

	if delta.DifferentAt("Spec.Name") {
		if err := rm.renameTable(ctx, desired, latest); err != nil {
			return nil, err
		}
	}

	if delta.DifferentAt("Spec.MetadataLocation") {
		if err := rm.updateMetadataLocation(ctx, desired, latest); err != nil {
			return nil, err
		}
	}

	if delta.DifferentAt("Spec.Tags") {
		arn := tableARNFromStatus(ko)
		if arn != nil {
			if err := rm.syncTags(ctx, desired, latest, arn); err != nil {
				return nil, err
			}
		}
	}

	if delta.DifferentAt("Spec.MaintenanceConfiguration") {
		if err := rm.syncMaintenance(ctx, desired, latest); err != nil {
			return nil, err
		}
	}

	return &resource{ko}, nil
}

// customDeltaPostCompare adds a Spec.MaintenanceConfiguration difference to the
// delta using a field-aware comparison. The field is compare.is_ignored because
// the generated whole-map DeepEqual churns a partial spec against the AWS-
// defaulted observed config; this hook restores accurate diffing so genuine
// changes still drive an update while partial configs do not churn.
func customDeltaPostCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if maintenanceNeedsSync(a.ko.Spec.MaintenanceConfiguration, b.ko.Spec.MaintenanceConfiguration) {
		delta.Add("Spec.MaintenanceConfiguration", a.ko.Spec.MaintenanceConfiguration, b.ko.Spec.MaintenanceConfiguration)
	}
}

// maintenanceNeedsSync reports whether the desired (a) maintenance configuration
// differs from what is observed on the table (b), comparing only the sub-fields
// the user actually declared. A nil desired sub-field means "adopt whatever AWS
// defaulted" and is not a difference, so a partial config does not churn.
//
// A maintenance type observed on the table but absent from the desired spec is
// left alone: the field is late-initialized, so an unset spec adopts the
// observed config rather than disabling it. There is no delete API and the
// service defaults maintenance to enabled; disabling requires an explicit
// status: disabled in the spec.
func maintenanceNeedsSync(
	desired, latest map[string]*svcapitypes.TableMaintenanceConfigurationValue,
) bool {
	for jobType, d := range desired {
		if d == nil {
			continue
		}
		l := latest[jobType]
		if l == nil {
			return true
		}
		if d.Status != nil && (l.Status == nil || *d.Status != *l.Status) {
			return true
		}
		if d.Settings == nil {
			continue
		}
		var ls *svcapitypes.TableMaintenanceSettings
		if l.Settings != nil {
			ls = l.Settings
		}
		if c := d.Settings.IcebergCompaction; c != nil {
			var lc *svcapitypes.IcebergCompactionSettings
			if ls != nil {
				lc = ls.IcebergCompaction
			}
			if lc == nil {
				return true
			}
			if c.Strategy != nil && (lc.Strategy == nil || *c.Strategy != *lc.Strategy) {
				return true
			}
			if c.TargetFileSizeMB != nil && (lc.TargetFileSizeMB == nil || *c.TargetFileSizeMB != *lc.TargetFileSizeMB) {
				return true
			}
		}
		if m := d.Settings.IcebergSnapshotManagement; m != nil {
			var lm *svcapitypes.IcebergSnapshotManagementSettings
			if ls != nil {
				lm = ls.IcebergSnapshotManagement
			}
			if lm == nil {
				return true
			}
			if m.MaxSnapshotAgeHours != nil && (lm.MaxSnapshotAgeHours == nil || *m.MaxSnapshotAgeHours != *lm.MaxSnapshotAgeHours) {
				return true
			}
			if m.MinSnapshotsToKeep != nil && (lm.MinSnapshotsToKeep == nil || *m.MinSnapshotsToKeep != *lm.MinSnapshotsToKeep) {
				return true
			}
		}
	}
	return false
}

// syncMaintenance applies the desired table-level maintenance configuration via
// PutTableMaintenanceConfiguration, one call per maintenance type.
//
// The Put is keyed by bucket ARN + namespace + name rather than the table ARN,
// so it uses the desired name: a rename earlier in this same reconcile has
// already been applied to AWS, which makes the latest name stale. Bucket and
// namespace are immutable, so the latest values are current.
//
// There is no DeleteTableMaintenanceConfiguration API and the config always
// exists (a new table defaults to enabled), so a maintenance type present on
// the table but dropped from the desired spec is left alone rather than
// removed. Disabling requires an explicit status: disabled.
func (rm *resourceManager) syncMaintenance(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncMaintenance")
	defer func() { exit(err) }()

	name := desired.ko.Spec.Name
	arn := latest.ko.Spec.TableBucketARN
	namespace := latest.ko.Spec.Namespace
	if arn == nil || namespace == nil || name == nil {
		return nil
	}

	for jobType, val := range desired.ko.Spec.MaintenanceConfiguration {
		if val == nil {
			continue
		}
		sdkVal := &svcsdktypes.TableMaintenanceConfigurationValue{}
		if val.Status != nil {
			sdkVal.Status = svcsdktypes.MaintenanceStatus(*val.Status)
		}
		// The settings are a union: send only the member matching the type key,
		// so a spec that declares the wrong member for a type is not silently
		// sent as the other type's settings.
		switch jobType {
		case string(svcsdktypes.TableMaintenanceTypeIcebergCompaction):
			if val.Settings != nil {
				if c := val.Settings.IcebergCompaction; c != nil {
					settings := svcsdktypes.IcebergCompactionSettings{}
					if c.Strategy != nil {
						settings.Strategy = svcsdktypes.IcebergCompactionStrategy(*c.Strategy)
					}
					if c.TargetFileSizeMB != nil {
						v, cerr := maintenanceToInt32(*c.TargetFileSizeMB)
						if cerr != nil {
							return ackerr.NewTerminalError(fmt.Errorf("targetFileSizeMB: %w", cerr))
						}
						settings.TargetFileSizeMB = aws.Int32(v)
					}
					sdkVal.Settings = &svcsdktypes.TableMaintenanceSettingsMemberIcebergCompaction{
						Value: settings,
					}
				}
			}
		case string(svcsdktypes.TableMaintenanceTypeIcebergSnapshotManagement):
			if val.Settings != nil {
				if m := val.Settings.IcebergSnapshotManagement; m != nil {
					settings := svcsdktypes.IcebergSnapshotManagementSettings{}
					if m.MaxSnapshotAgeHours != nil {
						v, cerr := maintenanceToInt32(*m.MaxSnapshotAgeHours)
						if cerr != nil {
							return ackerr.NewTerminalError(fmt.Errorf("maxSnapshotAgeHours: %w", cerr))
						}
						settings.MaxSnapshotAgeHours = aws.Int32(v)
					}
					if m.MinSnapshotsToKeep != nil {
						v, cerr := maintenanceToInt32(*m.MinSnapshotsToKeep)
						if cerr != nil {
							return ackerr.NewTerminalError(fmt.Errorf("minSnapshotsToKeep: %w", cerr))
						}
						settings.MinSnapshotsToKeep = aws.Int32(v)
					}
					sdkVal.Settings = &svcsdktypes.TableMaintenanceSettingsMemberIcebergSnapshotManagement{
						Value: settings,
					}
				}
			}
		default:
			rlog.Info("skipping unrecognized table maintenance type", "type", jobType)
			continue
		}
		_, err = rm.sdkapi.PutTableMaintenanceConfiguration(
			ctx,
			&svcsdk.PutTableMaintenanceConfigurationInput{
				TableBucketARN: arn,
				Namespace:      namespace,
				Name:           name,
				Type:           svcsdktypes.TableMaintenanceType(jobType),
				Value:          sdkVal,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "PutTableMaintenanceConfiguration", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// renameTable changes the table name via RenameTable. It targets the CURRENT
// (latest) bucket/namespace/name and supplies the desired new name, guarded by
// the latest observed version token.
func (rm *resourceManager) renameTable(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.renameTable")
	defer func() { exit(err) }()

	input := &svcsdk.RenameTableInput{
		TableBucketARN: latest.ko.Spec.TableBucketARN,
		Namespace:      latest.ko.Spec.Namespace,
		Name:           latest.ko.Spec.Name,
		NewName:        desired.ko.Spec.Name,
	}
	if latest.ko.Status.VersionToken != nil {
		input.VersionToken = latest.ko.Status.VersionToken
	}

	_, err = rm.sdkapi.RenameTable(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "RenameTable", err)
	return err
}

// updateMetadataLocation applies the desired Iceberg metadata location via the
// dedicated UpdateTableMetadataLocation API. metadataLocation is not part of
// CreateTable; it can only be changed through this operation, guarded by the
// latest observed version token.
func (rm *resourceManager) updateMetadataLocation(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateMetadataLocation")
	defer func() { exit(err) }()

	if desired.ko.Spec.MetadataLocation == nil {
		return nil
	}

	input := &svcsdk.UpdateTableMetadataLocationInput{
		TableBucketARN:   latest.ko.Spec.TableBucketARN,
		Namespace:        latest.ko.Spec.Namespace,
		Name:             latest.ko.Spec.Name,
		MetadataLocation: desired.ko.Spec.MetadataLocation,
	}
	if latest.ko.Status.VersionToken != nil {
		input.VersionToken = latest.ko.Status.VersionToken
	}

	_, err = rm.sdkapi.UpdateTableMetadataLocation(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateTableMetadataLocation", err)
	return err
}

// syncTags reconciles the desired tag set against the latest observed tag set
// using the TagResource and UntagResource APIs.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
	arn *string,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTags")
	defer func() { exit(err) }()

	from, _ := convertToOrderedACKTags(latest.ko.Spec.Tags)
	to, _ := convertToOrderedACKTags(desired.ko.Spec.Tags)

	added, _, removed := ackcompare.GetTagsDifference(from, to)

	// A key present in both added and removed is a value change; keep it in
	// added (TagResource overwrites) and drop it from removed.
	for key := range removed {
		if _, ok := added[key]; ok {
			delete(removed, key)
		}
	}

	if len(removed) > 0 {
		toRemove := make([]string, 0, len(removed))
		for key := range removed {
			toRemove = append(toRemove, key)
		}
		_, err = rm.sdkapi.UntagResource(
			ctx,
			&svcsdk.UntagResourceInput{
				ResourceArn: arn,
				TagKeys:     toRemove,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "UntagResource", err)
		if err != nil {
			return err
		}
	}

	if len(added) > 0 {
		toAdd := make(map[string]string, len(added))
		for key, val := range added {
			toAdd[key] = val
		}
		_, err = rm.sdkapi.TagResource(
			ctx,
			&svcsdk.TagResourceInput{
				ResourceArn: arn,
				Tags:        toAdd,
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "TagResource", err)
		if err != nil {
			return err
		}
	}

	return nil
}
