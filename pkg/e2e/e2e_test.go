package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
)

func TestNoHiveLogs(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if resetEvents() != nil {
		t.Errorf("list of events is not empty on start: %v", err)
	}

	viper.Set(config.ReportDir, tmpDir)

	checkBeforeMetricsGeneration()
	if !reflect.DeepEqual(events.GetListOfEvents(), []string{string(events.NoHiveLogs)}) {
		t.Errorf("the NoHiveLogs event was not detected")
	}

	// Make sure the events are clear again
	if resetEvents() != nil {
		t.Errorf("list of events is not empty during second hive log check: %v", err)
	}

	_, err = os.Create(filepath.Join(tmpDir, hiveLog))
	if err != nil {
		t.Errorf("error creating dummy hive log: %v", err)
	}

	checkBeforeMetricsGeneration()

	if !reflect.DeepEqual(events.GetListOfEvents(), []string{}) {
		t.Errorf("the NoHiveLogs event should not have been detected")
	}
}

func resetEvents() error {
	events.Instance.Events = map[string]bool{}

	if !reflect.DeepEqual(events.GetListOfEvents(), []string{}) {
		return fmt.Errorf("list of events is not empty on reset")
	}

	return nil
}
