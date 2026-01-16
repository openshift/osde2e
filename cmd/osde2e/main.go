package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2/textlogger"

	"github.com/openshift/osde2e/cmd/osde2e/arguments"
	"github.com/openshift/osde2e/cmd/osde2e/cleanup"
	"github.com/openshift/osde2e/cmd/osde2e/completion"
	"github.com/openshift/osde2e/cmd/osde2e/healthcheck"
	"github.com/openshift/osde2e/cmd/osde2e/krknai"
	"github.com/openshift/osde2e/cmd/osde2e/provision"
	"github.com/openshift/osde2e/cmd/osde2e/test"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
)

var root = &cobra.Command{
	Use:           "osde2e",
	Long:          "Command line tool for osde2e.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	// Add the command line flags:
	pfs := root.PersistentFlags()
	arguments.AddDebugFlag(pfs)

	root.AddCommand(provision.Cmd)
	root.AddCommand(test.Cmd)
	root.AddCommand(healthcheck.Cmd)
	root.AddCommand(completion.Cmd)
	root.AddCommand(cleanup.Cmd)
	root.AddCommand(krknai.Cmd)
}

func main() {
	const buildLog = "test_output.log"

	reportDir := viper.GetString(config.ReportDir)
	sharedDir := viper.GetString(config.SharedDir)
	runtimeDir := fmt.Sprintf("%s/osde2e-%s", os.TempDir(), util.RandomStr(10))

	if reportDir == "" {
		reportDir = runtimeDir
		viper.Set(config.ReportDir, reportDir)
	}

	if err := os.MkdirAll(reportDir, os.ModePerm); err != nil {
		log.Printf("Could not create report directory: %s, %v", reportDir, err)
		os.Exit(1)
	}

	if sharedDir != "" {
		if err := os.MkdirAll(sharedDir, os.ModePerm); err != nil {
			log.Printf("Could not create shared directory: %s, %v", sharedDir, err)
		}
	}
	buildLogPath := filepath.Join(reportDir, buildLog)
	logFile, err := os.Create(buildLogPath)
	if err != nil {
		log.Println("unable to create output file")
		os.Exit(1)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	config := textlogger.NewConfig(textlogger.Output(mw))
	logger := textlogger.NewLogger(config)
	ctx := logr.NewContext(context.Background(), logger)
	root.SetContext(ctx)

	log.SetOutput(mw)

	logger.Info("configured logging", "outputFile", buildLogPath, "reportDir", reportDir, "sharedDir", sharedDir)

	// Register providers
	spi.RegisterProvider("rosa", func() (spi.Provider, error) { return rosaprovider.New(ctx) })
	spi.RegisterProvider("ocm", func() (spi.Provider, error) { return ocmprovider.New() })

	if err := root.Execute(); err != nil {
		logger.Error(err, "command execution failed")
		os.Exit(1)
	}

	os.Exit(0)
}
