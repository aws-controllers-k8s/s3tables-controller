	// On adoption only Spec.Name is set, but GetTableBucket takes only the ARN,
	// so ReadOne would never find the bucket. The table bucket ARN is
	// deterministic, so synthesize it from Spec.Name when unset; GetTableBucket
	// then confirms the bucket exists.
	//
	// This is gated to the adoption path (identified by the adoption-policy
	// annotation) on purpose. On a normal create, ReadOne must still
	// short-circuit to NotFound via the required-fields check below so the
	// create flow populates the ARN itself. If we synthesized the ARN here
	// unconditionally, it would leak onto the desired object on the pre-create
	// read and the resource's status ARN would never be populated after create.
	if _, adopting := r.ko.ObjectMeta.Annotations[ackv1alpha1.AnnotationAdoptionPolicy]; adopting &&
		(r.ko.Status.ACKResourceMetadata == nil || r.ko.Status.ACKResourceMetadata.ARN == nil) &&
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
