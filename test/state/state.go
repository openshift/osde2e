// Package state retrieves the current state of the cluster for debug purposes.
package state

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("[Suite: e2e] Cluster state", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	ginkgo.It("should be gathered", func() {
		state := h.GetClusterState()
		results := make(map[string][]byte, len(state))
		for resource, list := range state {
			data, err := json.MarshalIndent(list, "", "    ")
			Expect(err).NotTo(HaveOccurred())

			var gbuf bytes.Buffer
			zw := gzip.NewWriter(&gbuf)
			_, err = zw.Write(data)
			Expect(err).NotTo(HaveOccurred())

			err = zw.Close()
			Expect(err).NotTo(HaveOccurred())

			// include gzip in filename to mark compressed data
			filename := fmt.Sprintf("%s-%s-%s.json.gzip", resource.Group, resource.Version, resource.Resource)
			results[filename] = gbuf.Bytes()
		}

		// write results to disk
		h.WriteResults(results)
	}, 900)
})
