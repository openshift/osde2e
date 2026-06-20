package dashboard

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/dashboard/collectors"
	"github.com/openshift/osde2e/pkg/dashboard/config"
	"github.com/openshift/osde2e/pkg/dashboard/server"
	"github.com/openshift/osde2e/pkg/dashboard/store"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start osde2e dashboard web server",
	Long:  "Start a web dashboard that aggregates cluster reserves, usage metrics, and test results from OCM and S3.",
	Args:  cobra.NoArgs,
	Run:   run,
}

var args struct {
	configString    string
	secretLocations string
	environment     string
	port            int
	maxResults      int
	sqsQueueURL     string
	dbPath          string
	backfill        bool
}

func init() {
	pfs := Cmd.PersistentFlags()

	pfs.StringVar(&args.configString, "configs", "", "A comma separated list of built in configs to use")
	_ = Cmd.RegisterFlagCompletionFunc("configs", helpers.ConfigComplete)

	pfs.StringVar(&args.secretLocations, "secret-locations", "",
		"A comma separated list of possible secret directory locations for loading secret configs.")

	pfs.StringVarP(&args.environment, "environment", "e", "",
		"Filter clusters by environment (stage, prod, integration, all). Defaults to 'all'.")

	pfs.IntVarP(&args.port, "port", "p", config.DefaultPort, "HTTP port for the dashboard server")

	pfs.IntVar(&args.maxResults, "max-results", config.DefaultMaxTestResults,
		"Maximum number of test results to display")

	pfs.StringVar(&args.sqsQueueURL, "sqs-queue-url", "",
		"SQS queue URL receiving S3 ObjectCreated notifications. When set, enables event-driven DB updates.")

	pfs.StringVar(&args.dbPath, "db", "dashboard.db",
		"Path to the SQLite database file. Use ':memory:' for an ephemeral in-memory DB.")

	pfs.BoolVar(&args.backfill, "backfill", false,
		"Scan all historical S3 objects and populate the DB before starting the server.")

	// Bind flags to viper
	_ = viper.BindPFlag(config.Port, pfs.Lookup("port"))
	_ = viper.BindPFlag(config.Environment, pfs.Lookup("environment"))
	_ = viper.BindPFlag(config.MaxTestResults, pfs.Lookup("max-results"))
	_ = viper.BindPFlag(ocmprovider.Env, pfs.Lookup("environment"))
	_ = viper.BindPFlag(config.SQSQueueURL, pfs.Lookup("sqs-queue-url"))
	_ = viper.BindPFlag(config.DBPath, pfs.Lookup("db"))
}

func run(cmd *cobra.Command, argv []string) {
	log.Println("==== Starting osde2e Dashboard ====")

	// Unset personal OCM token so the dashboard authenticates via OCM_CLIENT_ID/SECRET only.
	os.Unsetenv("OCM_TOKEN")

	// Load configurations
	if err := common.LoadConfigs(args.configString, "", args.secretLocations); err != nil {
		log.Printf("Error loading initial configuration: %v", err)
		os.Exit(1)
	}

	// Set dashboard defaults
	config.SetDefaults()

	// Override with CLI flags if explicitly set
	if cmd.PersistentFlags().Changed("port") {
		viper.Set(config.Port, args.port)
	}
	if cmd.PersistentFlags().Changed("environment") {
		viper.Set(config.Environment, args.environment)
		viper.Set(ocmprovider.Env, args.environment)
	}
	if cmd.PersistentFlags().Changed("max-results") {
		viper.Set(config.MaxTestResults, args.maxResults)
	}
	if cmd.PersistentFlags().Changed("sqs-queue-url") {
		viper.Set(config.SQSQueueURL, args.sqsQueueURL)
	}
	if cmd.PersistentFlags().Changed("db") {
		viper.Set(config.DBPath, args.dbPath)
	}

	// Load dashboard configuration
	dashboardConfig := config.LoadConfig()

	// Validate configuration
	if dashboardConfig.OCMConfigPath == "" {
		log.Println("Warning: OCM_CONFIG not set. OCM features may not work.")
	}
	if dashboardConfig.S3Bucket == "" {
		log.Println("Warning: LOG_BUCKET not set. S3 test results will not be available.")
	}

	log.Printf("Dashboard Configuration:")
	log.Printf("  Port:           %d", dashboardConfig.Port)
	log.Printf("  S3 Bucket:      %s", dashboardConfig.S3Bucket)
	log.Printf("  S3 Region:      %s", dashboardConfig.S3Region)
	log.Printf("  Environment:    %s", dashboardConfig.Environment)
	log.Printf("  DB Path:        %s", dashboardConfig.DBPath)
	log.Printf("  SQS Queue URL:  %s", dashboardConfig.SQSQueueURL)

	// Open the SQLite store
	st, err := store.Open(dashboardConfig.DBPath)
	if err != nil {
		log.Printf("Failed to open store at %s: %v", dashboardConfig.DBPath, err)
		os.Exit(1)
	}
	defer st.Close()

	// Top-level context — cancelled on Ctrl+C or SIGTERM, shuts down everything.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Optionally backfill historical S3 data into the DB
	if args.backfill || dashboardConfig.SQSQueueURL != "" {
		if dashboardConfig.S3Bucket == "" {
			log.Println("Warning: --backfill requested but LOG_BUCKET is not set; skipping.")
		} else {
			consumer, err := collectors.NewSQSConsumer(
				dashboardConfig.SQSQueueURL,
				dashboardConfig.S3Bucket,
				dashboardConfig.S3Region,
				st,
			)
			if err != nil {
				log.Printf("Warning: failed to create SQS consumer: %v", err)
			} else {
				if args.backfill {
					log.Println("Truncating DB before backfill...")
					if err := st.Truncate(); err != nil {
						log.Printf("Warning: truncate failed: %v", err)
					}
					log.Println("Running backfill — this may take a few minutes...")
					if err := consumer.Backfill(); err != nil {
						log.Printf("Backfill error: %v", err)
					}
				}

				// Start the SQS consumer goroutine (only when queue URL is configured)
				if dashboardConfig.SQSQueueURL != "" {
					go consumer.Run(ctx)
					log.Printf("SQS consumer started")
				}
			}
		}
	}

	// Create and start the HTTP server
	srv, err := server.NewServer(dashboardConfig)
	if err != nil {
		log.Printf("Failed to create dashboard server: %v", err)
		os.Exit(1)
	}
	srv.WithStore(st)

	addr := fmt.Sprintf(":%d", dashboardConfig.Port)
	log.Printf("Dashboard server starting on http://localhost%s", addr)
	log.Printf("Press Ctrl+C to stop")

	if err := srv.Start(addr, ctx); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
