package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers/mock"
	"github.com/prometheus/client_golang/prometheus"
)

func TestProcessJUnitXMLFile(t *testing.T) {
	viper.Reset()
	viper.Set(mock.Env, "prod")
	viper.Set(config.Provider, "mock")
	viper.Set(config.JobID, 123)
	viper.Set(config.CloudProvider.CloudProviderID, "aws")
	viper.Set(config.CloudProvider.Region, "us-east-1")
	viper.Set(config.Cluster.ID, "1a2b3c")
	viper.Set(config.Cluster.Version, "install-version")
	viper.Set(config.Upgrade.ReleaseName, "upgrade-version")

	tests := []struct {
		testName       string
		phase          string
		fileContents   string
		expectedOutput string
	}{
		{
			testName: "regular parsing",
			phase:    "install",
			fileContents: `<testsuite name="test suite" time="6">
	<testcase name="test 1" time="1" />
	<testcase name="test 2" time="2" />
	<testcase name="test 3" time="3">
		<failure type="blah">
			failure text
		</failure>
	</testcase>
	<testcase name="test 4" time="4">
		<skipped>
			blah
		</skipped>
	</testcase>
</testsuite>`,
			expectedOutput: `cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test suite",testname="test 1",upgrade_version="upgrade-version"} 1
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test suite",testname="test 2",upgrade_version="upgrade-version"} 2
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="failed",suite="test suite",testname="test 3",upgrade_version="upgrade-version"} 3
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="skipped",suite="test suite",testname="test 4",upgrade_version="upgrade-version"} 4
`,
		},
		{
			testName: "parsing with complicated attribute values",
			phase:    "install",
			fileContents: `<testsuite name="test &quot;suite&quot;" time="6">
	<testcase name="test \1" time="1" />
	<testcase name="test 2" time="2" />
	<testcase name="test 3
newline" time="3">
		<failure type="blah">
			failure text
		</failure>
	</testcase>
</testsuite>`,
			expectedOutput: `cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test \"suite\"",testname="test \\1",upgrade_version="upgrade-version"} 1
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test \"suite\"",testname="test 2",upgrade_version="upgrade-version"} 2
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="failed",suite="test \"suite\"",testname="test 3\nnewline",upgrade_version="upgrade-version"} 3
`,
		},
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("error creating temporary directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	for _, test := range tests {
		m := NewMetrics()

		if m == nil {
			t.Error("error creating new metrics provider")
		}
		tmpFile, err := ioutil.TempFile(tmpDir, "*")
		if err != nil {
			t.Errorf("error writing temporary file: %v", err)
		}

		tmpFile.WriteString(test.fileContents)
		tmpFile.Close()

		err = m.processJUnitXMLFile(test.phase, tmpFile.Name())

		if err != nil {
			t.Errorf("error while processing junit file: %v", err)
		}

		output, err := m.registryToExpositionFormat()
		if err != nil {
			t.Errorf("error convering registry to exposition format: %v", err)
		}

		if err = arraysHaveSameElements(strings.Split(string(output), "\n"), strings.Split(test.expectedOutput, "\n")); err != nil {
			t.Errorf("%s\nOutput:\n---\n%s\n---\ndoes not match expected output (disregarding order):\n---\n%s\n---\n%v\n", test.testName, output, test.expectedOutput, err)
		}
	}
}

func TestProcessJSONFile(t *testing.T) {
	viper.Reset()
	viper.Set(mock.Env, "prod")
	viper.Set(config.Provider, "mock")
	viper.Set(config.JobID, 123)
	viper.Set(config.CloudProvider.CloudProviderID, "aws")
	viper.Set(config.CloudProvider.Region, "us-east-1")
	viper.Set(config.Cluster.ID, "1a2b3c")
	viper.Set(config.Cluster.Version, "install-version")
	viper.Set(config.Upgrade.ReleaseName, "upgrade-version")

	tests := []struct {
		testName         string
		useAddonGatherer bool
		fileContents     string
		expectedOutput   string
		phase            string
	}{
		{
			testName:         "regular parsing with custom metadata",
			useAddonGatherer: false,
			fileContents: `{
	"test1": "value1",
	"test2": 6,
	"nested field": {
		"another-level": {
			"test3": "value2"
		}
	},
	"another-nested field": {
		"another-level": {
			"test4": "7"
		}
	}
}`,
			expectedOutput: `cicd_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="test2",region="us-east-1",upgrade_version="upgrade-version"} 6
cicd_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="another-nested field.another-level.test4",region="us-east-1",upgrade_version="upgrade-version"} 7
`,
			phase: "",
		},
		{
			testName:         "regular parsing with addon metadata",
			useAddonGatherer: true,
			fileContents: `{
	"test1***?": "value1",
	"test2": 6,
	"nested field": {
		"another-level": {
			"test3": "value2"
		}
	},
	"another-nested field": {
		"another-level": {
			"test4": "7"
		}
	}
}`,
			expectedOutput: `cicd_addon_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="test2",phase="install",region="us-east-1",upgrade_version="upgrade-version"} 6
cicd_addon_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="another-nested field.another-level.test4",phase="install",region="us-east-1",upgrade_version="upgrade-version"} 7
`,
			phase: "install",
		},
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("error creating temporary directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)

	for _, test := range tests {
		m := NewMetrics()
		if m == nil {
			t.Error("error creating new metrics provider")
		}
		tmpFile, err := ioutil.TempFile(tmpDir, "*")
		if err != nil {
			t.Errorf("error writing temporary file: %v", err)
		}

		tmpFile.WriteString(test.fileContents)
		tmpFile.Close()

		var gatherer *prometheus.GaugeVec
		if test.useAddonGatherer {
			gatherer = m.addonGatherer
		} else {
			gatherer = m.metadataGatherer
		}
		err = m.processJSONFile(gatherer, tmpFile.Name(), test.phase)

		if err != nil {
			t.Errorf("error while processing JSON file: %v", err)
		}

		output, err := m.registryToExpositionFormat()
		if err != nil {
			t.Errorf("error convering registry to exposition format: %v", err)
		}

		err = arraysHaveSameElements(strings.Split(string(output), "\n"), strings.Split(test.expectedOutput, "\n"))
		if err != nil {
			t.Errorf("%s\nOutput:\n---\n%s\n---\ndoes not match expected output (disregarding order):\n---\n%s\n---\n%v", test.testName, output, test.expectedOutput, err)
		}
	}
}

func TestWritePrometheusFile(t *testing.T) {
	viper.Reset()
	viper.Set(mock.Env, "prod")
	viper.Set(config.Provider, "mock")
	viper.Set(config.JobID, 123)
	viper.Set(config.JobName, "test-job")
	viper.Set(config.CloudProvider.CloudProviderID, "aws")
	viper.Set(config.CloudProvider.Region, "us-east-1")
	viper.Set(config.Cluster.ID, "1a2b3c")
	viper.Set(config.Cluster.Version, "install-version")
	viper.Set(config.Upgrade.ReleaseName, "upgrade-version")

	type jUnitFile struct {
		fileContents string
		directory    string
	}

	jUnitFile1Contents := `<testsuite name="test suite 1" time="6">
	<testcase name="test 1" time="1" />
	<testcase name="test 2" time="2" />
	<testcase name="test 3" time="3">
		<failure type="blah">
			failure text
		</failure>
	</testcase>
</testsuite>`
	jUnitFile2Contents := `<testsuite name="test suite 2" time="6">
	<testcase name="test 1" time="1" />
	<testcase name="test 2" time="2" />
	<testcase name="test 3" time="3">
		<failure type="blah">
			failure text
		</failure>
	</testcase>
</testsuite>`
	metadataFileContents := `{
	"test1": "value1",
	"test2": 6,
	"nested": {
		"another-level": {
			"test3": "value2"
		}
	}
}`
	addonMetadataFileContents := metadataFileContents

	jUnitExpectedOutput := `cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test suite 1",testname="test 1",upgrade_version="upgrade-version"} 1
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="passed",suite="test suite 1",testname="test 2",upgrade_version="upgrade-version"} 2
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="install",region="us-east-1",result="failed",suite="test suite 1",testname="test 3",upgrade_version="upgrade-version"} 3
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="upgrade",region="us-east-1",result="passed",suite="test suite 2",testname="test 1",upgrade_version="upgrade-version"} 1
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="upgrade",region="us-east-1",result="passed",suite="test suite 2",testname="test 2",upgrade_version="upgrade-version"} 2
cicd_jUnitResult{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",phase="upgrade",region="us-east-1",result="failed",suite="test suite 2",testname="test 3",upgrade_version="upgrade-version"} 3
`

	tests := []struct {
		testName                  string
		jUnitFiles                []jUnitFile
		metadataFileContents      string
		addonMetadataFileContents string
		expectedOutput            string
	}{
		{
			testName: "regular parsing",
			jUnitFiles: []jUnitFile{
				{
					fileContents: jUnitFile1Contents,
					directory:    "install",
				},
				{
					fileContents: jUnitFile2Contents,
					directory:    "upgrade",
				},
			},
			metadataFileContents:      metadataFileContents,
			addonMetadataFileContents: addonMetadataFileContents,
			expectedOutput: jUnitExpectedOutput + `cicd_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="test2",region="us-east-1",upgrade_version="upgrade-version"} 6
cicd_addon_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="test2",phase="install",region="us-east-1",upgrade_version="upgrade-version"} 6
`,
		},
		{
			testName: "no addon metadata",
			jUnitFiles: []jUnitFile{
				{
					fileContents: jUnitFile1Contents,
					directory:    "install",
				},
				{
					fileContents: jUnitFile2Contents,
					directory:    "upgrade",
				},
			},
			metadataFileContents:      metadataFileContents,
			addonMetadataFileContents: "",
			expectedOutput: jUnitExpectedOutput + `cicd_metadata{cloud_provider="aws",cluster_id="1a2b3c",environment="prod",install_version="install-version",job_id="123",metadata_name="test2",region="us-east-1",upgrade_version="upgrade-version"} 6
`,
		},
		{
			testName: "no addon metadata or regular metadata",
			jUnitFiles: []jUnitFile{
				{
					fileContents: jUnitFile1Contents,
					directory:    "install",
				},
				{
					fileContents: jUnitFile2Contents,
					directory:    "upgrade",
				},
			},
			metadataFileContents:      "",
			addonMetadataFileContents: "",
			expectedOutput:            jUnitExpectedOutput,
		},
	}

	for _, test := range tests {
		m := NewMetrics()
		if m == nil {
			t.Error("error creating new metrics provider")
		}
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Errorf("error creating temporary directory: %v", err)
		}

		defer os.RemoveAll(tmpDir)

		for _, jUnitFile := range test.jUnitFiles {
			jUnitDir := filepath.Join(tmpDir, jUnitFile.directory)
			if _, err := os.Stat(jUnitDir); os.IsNotExist(err) {
				err := os.Mkdir(jUnitDir, os.FileMode(0o755))
				if err != nil {
					t.Errorf("error creating jUnit file directory: %v", err)
				}
			}
			tmpFile, err := ioutil.TempFile(jUnitDir, "junit*.xml")
			if err != nil {
				t.Errorf("error writing junit file: %v", err)
			}
			fmt.Printf("Writing file %s\n", tmpFile.Name())

			tmpFile.WriteString(jUnitFile.fileContents)
			tmpFile.Close()
		}

		if test.metadataFileContents != "" {
			err = ioutil.WriteFile(filepath.Join(tmpDir, metadata.CustomMetadataFile), []byte(test.metadataFileContents), os.FileMode(0o644))
			if err != nil {
				t.Errorf("error writing metadata file: %v", err)
			}
		}

		if test.addonMetadataFileContents != "" {
			err = ioutil.WriteFile(filepath.Join(tmpDir, "install", metadata.AddonMetadataFile), []byte(test.addonMetadataFileContents), os.FileMode(0o644))
			if err != nil {
				t.Errorf("error writing metadata file: %v", err)
			}
		}

		prometheusFile, err := m.WritePrometheusFile(tmpDir)
		if err != nil {
			t.Errorf("error while processing report directory: %v", err)
		}

		if prometheusFile != fmt.Sprintf("%s.%s.metrics.prom", viper.GetString(config.Cluster.ID), viper.GetString(config.JobName)) {
			t.Errorf("unexpected prometheus filename: %s", prometheusFile)
		}

		data, err := ioutil.ReadFile(filepath.Join(tmpDir, prometheusFile))
		if err != nil {
			t.Errorf("error while reading prometheus file: %v", err)
		}

		output := string(data)
		outputLines := strings.Split(output, "\n")
		expectedLines := strings.Split(test.expectedOutput, "\n")

		err = arraysHaveSameElements(outputLines, expectedLines)
		if err != nil {
			t.Errorf("%s\nOutput:\n---\n%s\n---\ndoes not match expected output (disregarding order):\n---\n%s\n---\n%v", test.testName, output, test.expectedOutput, err)
		}
	}
}

func lengthWithoutComments(array []string) int {
	length := 0
	for _, line := range array {
		if strings.HasPrefix(line, "#") {
			continue
		}
		length++
	}
	return length
}

func arraysHaveSameElements(array1 []string, array2 []string) error {
	array1Length := lengthWithoutComments(array1)
	array2Length := lengthWithoutComments(array2)

	if array1Length != array2Length {
		return fmt.Errorf("arrays don't have the same number of elements (%d and %d)", array1Length, array2Length)
	}

	for _, array1Item := range array1 {
		if strings.HasPrefix(array1Item, "#") {
			continue
		}

		found := false
		for _, array2Item := range array2 {
			if array1Item == array2Item {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("couldn't find line %s in both arrays", array1Item)
		}
	}

	return nil
}
