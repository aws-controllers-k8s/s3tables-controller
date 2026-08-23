	// metadataLocation and maintenanceConfiguration are not part of CreateTable;
	// each is applied via its dedicated API (UpdateTableMetadataLocation,
	// PutTableMaintenanceConfiguration). When the user declares either, requeue
	// so the next reconciliation observes the delta and customUpdateTable
	// applies it.
	if desired.ko.Spec.MetadataLocation != nil || desired.ko.Spec.MaintenanceConfiguration != nil {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, aws.String("table created, requeue for updates"), nil)
		err = ackrequeue.NeededAfter(fmt.Errorf("Reconciling to sync additional fields"), time.Second)
		return &resource{ko}, err
	}
