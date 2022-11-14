package ocmprovider

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterWideProxy struct {
	Enabled               bool
	HTTPProxy             string
	HTTPSProxy            string
	AdditionalTrustBundle string
}

// AddClusterProxy sets the cluster proxy configuration for the supplied cluster
func (o *OCMProvider) AddClusterProxy(clusterId string, httpsProxy string, httpProxy string, userCABundle string) error {
	clusterBuilder := cmv1.NewCluster()

	clusterProxyBuilder := cmv1.NewProxy()
	clusterProxyBuilder = clusterProxyBuilder.HTTPProxy(httpProxy)
	clusterProxyBuilder = clusterProxyBuilder.HTTPSProxy(httpsProxy)
	clusterBuilder = clusterBuilder.Proxy(clusterProxyBuilder)
	userCABundleData, err := o.LoadUserCaBundleData(userCABundle)
	if err != nil {
		return fmt.Errorf("error loading ca bundle contents: %v", err)
	}
	clusterBuilder = clusterBuilder.AdditionalTrustBundle(userCABundleData)

	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		return err
	}
	return o.updateCluster(clusterId, clusterSpec)
}

// RemoveClusterProxy removes the cluster proxy configuration for the supplied cluster
func (o *OCMProvider) RemoveClusterProxy(clusterId string) error {
	clusterBuilder := cmv1.NewCluster()

	clusterProxyBuilder := cmv1.NewProxy()
	clusterProxyBuilder = clusterProxyBuilder.HTTPProxy("")
	clusterProxyBuilder = clusterProxyBuilder.HTTPSProxy("")
	clusterBuilder = clusterBuilder.Proxy(clusterProxyBuilder)
	clusterBuilder = clusterBuilder.AdditionalTrustBundle("")

	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		return err
	}
	return o.updateCluster(clusterId, clusterSpec)
}

// RemoveUserCABundle removes only the Additional Trusted CA Bundle from the cluster
func (o *OCMProvider) RemoveUserCABundle(clusterId string) error {
	clusterBuilder := cmv1.NewCluster()
	clusterBuilder = clusterBuilder.AdditionalTrustBundle("")
	clusterSpec, err := clusterBuilder.Build()
	if err != nil {
		return err
	}
	return o.updateCluster(clusterId, clusterSpec)
}

func (o *OCMProvider) updateCluster(clusterId string, clusterSpec *cmv1.Cluster) error {
	resp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterId).Update().Body(clusterSpec).Send()
	if err != nil {
		err = fmt.Errorf("couldn't update proxy for cluster '%s': %v", clusterId, err)
		log.Printf("%v", err)
		return err
	}

	if resp != nil && resp.Error() != nil {
		err = fmt.Errorf("error while trying to update proxy from cluster: %v", resp.Error())
		log.Printf("%v", err)
		return resp.Error()
	}

	return nil
}

func (o *OCMProvider) LoadUserCaBundleData(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf(
			"can't read userCABundle file '%s': %w",
			file, err,
		)
	}

	userCaBundleData := strings.TrimSpace(string(data))
	return userCaBundleData, nil
}
