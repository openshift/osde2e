package ocmprovider

import (
	"fmt"
	"log"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// VersionGates gets the list of available version gates from ocm
func (o *OCMProvider) VersionGates() (*v1.VersionGateList, error) {
	versionGatesClient := o.conn.ClustersMgmt().V1().VersionGates()

	response, err := versionGatesClient.List().Send()
	if err != nil {
		return nil, fmt.Errorf("error retrieving version gates: %v", err)
	}

	return response.Items(), nil
}

// GetVersionGateID checks to see if a version gate exists for the cluster version provided
func (o *OCMProvider) GetVersionGateID(version string, label string) (string, error) {
	versionGates, err := o.VersionGates()
	if err != nil {
		return "", err
	}

	for _, versionGate := range versionGates.Slice() {
		if versionGate.VersionRawIDPrefix() == version && versionGate.Label() == label {
			return versionGate.ID(), nil
		}
	}
	return "", fmt.Errorf("%s version gate does not exist", version)
}

// GetVersionGate gets the version gate resource using the version gate id provided
func (o *OCMProvider) GetVersionGate(id string) (*v1.VersionGate, error) {
	versionGatesClient := o.conn.ClustersMgmt().V1().VersionGates()
	versionGate := versionGatesClient.VersionGate(id)
	response, err := versionGate.Get().Send()
	if err != nil {
		return nil, fmt.Errorf("unable to find version gate using id: %s, error: %v", id, err)
	}
	return response.Body(), nil
}

// GateAgreementExist checks to see if the gate agreement id is already applied to the
// cluster provided
func (o *OCMProvider) GateAgreementExist(clusterID string, gateAgreementID string) (bool, error) {
	gateAgreementClient := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
		GateAgreements()
	response, err := gateAgreementClient.List().Send()
	if err != nil {
		return false, fmt.Errorf("error retrieving gate agreements for cluster: %v", err)
	}

	for _, gateAgreement := range response.Items().Slice() {
		if gateAgreement.VersionGate().ID() == gateAgreementID {
			log.Printf("Cluster gate agreement id: %s already exists", gateAgreementID)
			return true, nil
		}
	}
	return false, nil
}

// AddGateAgreement adds the gate agreement to the cluster for cluster upgrades
func (o *OCMProvider) AddGateAgreement(clusterID string, versionGateID string) error {
	versionGate, err := o.GetVersionGate(versionGateID)
	if err != nil {
		return err
	}

	gateAgreement, err := v1.NewVersionGateAgreement().
		VersionGate(v1.NewVersionGate().Copy(versionGate)).
		Build()
	if err != nil {
		return fmt.Errorf("error building version gate agreement: %v", err)
	}

	gateAgreementExist, err := o.GateAgreementExist(clusterID, versionGateID)
	if err != nil {
		return err
	}

	if !gateAgreementExist {
		_, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			GateAgreements().Add().Body(gateAgreement).Send()
		if err != nil {
			return fmt.Errorf("error adding version gate agreement: %v", err)
		}
	}

	return nil
}
