package aws

func (CcsAwsSession *ccsAwsSession) CheckIfEC2ExistBasedOnNodeName(nodeName string) (bool, error) {
	var err error
	CcsAwsSession.session, CcsAwsSession.iam = CcsAwsSession.getClient()

	// Get the list of instances
	instances, err := CcsAwsSession.ec2.DescribeInstances(nil)
	if err != nil {
		return false, err
	}

	// Loop through the instances
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.PrivateDnsName == nodeName {
				return true, nil
			}
		}
	}

	return false, nil
}
