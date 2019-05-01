package osde2e

import (
	"log"
	"os"

	"github.com/openshift/osde2e/pkg/cluster"
)

const (
	UHCTokenEnv  = "UHC_TOKEN"
	AWSIDEnv     = "AWS_ACCESS_KEY_ID"
	AWSKeyEnv    = "AWS_SECRET_ACCESS_KEY"
	ReportDirEnv = "REPORT_DIR"
	ProdEnv      = "USE_PROD"
)

// Cfg is the configuration being used for end to end testing.
var Cfg Config

// Config dictates the behavior of cluster tests.
type Config struct {
	// ReportDir is the location JUnit XML results are written.
	ReportDir string

	// Prefix is used at the beginning of tests to identify them.
	Prefix string

	// UHCToken is used to authenticate with UHC.
	UHCToken string

	// ClusterName is the name of the cluster being created.
	ClusterName string

	// AWSKeyId is used by UHC.
	AWSKeyId string

	// AWSAccessKey is used by UHC.
	AWSAccessKey string

	// UseProd sends requests to production UHC.
	UseProd bool

	// runtime vars
	clusterId  string
	uhc        *cluster.UHC
	kubeconfig []byte
}

func setupCfgFromEnv() {
	Cfg.UHCToken = getVar(UHCTokenEnv)
	Cfg.AWSKeyId = getVar(AWSIDEnv)
	Cfg.AWSAccessKey = getVar(AWSKeyEnv)
	Cfg.ReportDir = os.Getenv(ReportDirEnv)

	// use staging unless told to use prod
	prod := os.Getenv(ProdEnv)
	Cfg.UseProd = len(prod) != 0
}

func getVar(name string) string {
	contents, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("'%s' must be provided", name)
	}
	return contents
}
