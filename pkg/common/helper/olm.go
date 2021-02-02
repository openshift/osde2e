package helper

import (
	"fmt"
	"github.com/openshift/osde2e/pkg/common/runner"
)

// InspectOLM inspects the OLM state of the cluster and saves the state to disk for later debugging
func (h *H) InspectOLM() error {
	inspectTimeoutInSeconds := 200
	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
	r := h.Runner(fmt.Sprintf("oc adm inspect --dest-dir=%v -A olm", runner.DefaultRunner.OutputDir))
	r.Name = "olm-inspect"
	r.Tarball = true
	stopCh := make(chan struct{})

	err := r.Run(inspectTimeoutInSeconds, stopCh)
	if err != nil {
		return fmt.Errorf("Error running OLM inspection: %s", err.Error())
	}

	gatherResults, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("Error retrieving OLM inspection results: %s", err.Error())
	}

	h.WriteResults(gatherResults)
	return nil
}
