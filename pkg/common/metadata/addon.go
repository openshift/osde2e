package metadata

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

// addonMetadata houses metadata to be written out to the additional-metadata.json
type AddonMetadata struct {
	// Whether the CRD was found. Typically Spyglass seems to have issues displaying non-strings, so
	// this will be written out as a string despite the native JSON boolean type.
	Version string `json:"version,string"`
	ID      string `json:"id,string"`
}

func (m *AddonMetadata) SetVersion(version string) {
	m.Version = version
}

func (m *AddonMetadata) SetID(id string) {
	m.ID = id
}

// WriteToJSON will marshall the addon metadata struct and write it into the given file.
func (m *AddonMetadata) WriteToJSONFile(outputFilename string) (err error) {
	var data []byte
	if data, err = json.Marshal(m); err != nil {
		return err
	}
	outputFilePath := filepath.Join(viper.GetString(config.ReportDir), outputFilename)
	f, err := os.OpenFile(outputFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	log.Println("writing addon metadata to ", outputFilePath)
	if _, err := f.WriteString(string(data)); err != nil {
		log.Println(err)
	}

	return nil
}
