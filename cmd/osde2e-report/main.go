package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/report"
)

var (
	// Cfg is the global configuration for the command.
	Cfg = config.Cfg

	// Out has the reports contents written to it.
	Out io.Writer = os.Stdout

	// envs shown in report.
	envs = []report.EnvConfig{
		{
			Name: "int",
		},
		{
			Name: "stage",
		},
		{
			Name: "prod",
		},
	}

	// jobs are included for each env in the report
	jobs = []report.JobConfig{
		{
			Name:    "osd",
			Version: "4.1",
		},
		{
			Name:    "osd-upgrade",
			Version: "4.1-4.1",
		},
	}
)

func init() {
	flag.Parse()
}

func main() {
	durStr := flag.Arg(0)
	if len(durStr) == 0 {
		log.Fatal("A duration to report on must be specified")
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		log.Fatalf("Could not parse duration specified: %v", err)
	}

	reportFile := flag.Arg(1)

	// configure report
	reportCfg := *report.DefaultConfig
	reportCfg.Envs = envs
	reportCfg.Jobs = jobs

	// load or initialize new report
	r := loadReportOrCreateNew(reportCfg, reportFile)
	r.Title = "osde2e Failure Report"

	// set time range
	end := time.Now().UTC()
	start := end.Add(-dur)
	rng := report.TimeRange{
		Start: start,
		End:   end,
	}

	// perform update
	if err = r.Update(Cfg, rng); err != nil {
		log.Fatalf("Error updating: %v", err)
	}

	// write markdown
	if err := r.Markdown(Out); err != nil {
		log.Fatalf("couldn't render report: %v", err)
	}

	// write report to disk if filename specified
	if len(reportFile) != 0 {
		if err = writeReport(r, reportFile); err != nil {
			log.Printf("Failed writing report to '%s': %v", reportFile, err)
		}
	}
}

func loadReportOrCreateNew(cfg report.Config, filename string) (r report.Report) {
	// read report from disk
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(data, &r)
	}

	if err != nil {
		log.Printf("Failed to read report from disk, creating new one: %v", err)
	}

	// override with config
	r.Config = cfg
	return
}

func writeReport(r report.Report, filename string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("couldn't encode report: %v", err)
	}
	return ioutil.WriteFile(filename, data, os.ModePerm)
}
