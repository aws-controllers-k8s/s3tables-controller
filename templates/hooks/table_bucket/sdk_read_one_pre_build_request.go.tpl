	// On adoption only Spec.Name is set, but GetTableBucket takes only the ARN,
	// so ReadOne would never find the bucket. The table bucket ARN is
	// deterministic, so synthesize it here when unset; GetTableBucket then
	// confirms the bucket exists.
	if (r.ko.Status.ACKResourceMetadata == nil || r.ko.Status.ACKResourceMetadata.ARN == nil) &&
		r.ko.Spec.Name != nil && *r.ko.Spec.Name != "" {
		arn := ackv1alpha1.AWSResourceName(fmt.Sprintf(
			"arn:%s:s3tables:%s:%s:bucket/%s",
			rm.awsPartition, rm.awsRegion, rm.awsAccountID, *r.ko.Spec.Name,
		))
		if r.ko.Status.ACKResourceMetadata == nil {
			r.ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
		}
		r.ko.Status.ACKResourceMetadata.ARN = &arn
	}
