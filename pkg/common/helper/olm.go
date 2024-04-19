package helper

import (
	"context"
	"fmt"

	"github.com/openshift/osde2e/pkg/common/runner"
)

// InspectOLM inspects the OLM state of the cluster and saves the state to disk for later debugging
func (h *H) InspectOLM(ctx context.Context) error {
	inspectTimeoutInSeconds := 200
	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
	r := h.Runner(fmt.Sprintf("oc adm inspect --dest-dir=%v -A olm", runner.DefaultRunner.OutputDir))
	r.Name = "olm-inspect"
	r.Tarball = true
	stopCh := make(chan struct{})

	err := r.Run(inspectTimeoutInSeconds, stopCh)
	if err != nil {
		return fmt.Errorf("error running OLM inspection: %s", err.Error())
	}

	gatherResults, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("error retrieving OLM inspection results: %s", err.Error())
	}

	h.WriteResults(gatherResults)
	return nil
}
