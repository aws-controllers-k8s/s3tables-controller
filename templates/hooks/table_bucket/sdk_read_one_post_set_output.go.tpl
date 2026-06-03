	// The generated SetResource output mapping does not emit an assignment for
	// the AWS-assigned TableBucketId, so populate the status field here.
	if resp.TableBucketId != nil {
		ko.Status.TableBucketID = resp.TableBucketId
	}
	// The TableBucket-level encryption and storage class configuration are not
	// returned by GetTableBucket. They have dedicated Get* APIs. Populate the
	// corresponding Spec fields so the observed state reflects what AWS holds
	// and we avoid spurious deltas on every reconciliation.
	if err := rm.setBucketConfigurations(ctx, ko); err != nil {
		return nil, err
	}
