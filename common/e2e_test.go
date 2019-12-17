package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/events"
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

	cfg := config.Cfg
	cfg.ReportDir = tmpDir

	checkBeforeMetricsGeneration(cfg)
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

	checkBeforeMetricsGeneration(cfg)

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
