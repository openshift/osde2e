package routemonitors

import (
	"fmt"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	vegeta "github.com/tsenart/vegeta/lib"
	"github.com/tsenart/vegeta/lib/plot"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	consoleNamespace = "openshift-console"
	consoleLabel     = "console"
	oauthNamespace   = "openshift-authentication"
	oauthName        = "oauth-openshift"
)

type RouteMonitors struct {
	Monitors          map[string]<-chan *vegeta.Result
	Metrics           map[string]*vegeta.Metrics
	Plots             map[string]*plot.Plot
	consoleTargeter   vegeta.Targeter
	oauthTargeter     vegeta.Targeter
	apiTargeters      []vegeta.Targeter
	workloadTargeters []vegeta.Targeter
	attackers         []*vegeta.Attacker
}

// Frequency of requests per second (per route)
const pollFrequency = 3
const timeoutSeconds = 3 * time.Second

// Detects the available routes in the cluster and initializes monitors for their availability
func Create() (*RouteMonitors, error) {
	h := helper.NewOutsideGinkgo()

	// Create a monitor for the web console
	consoleRoute, err := h.Route().RouteV1().Routes(consoleNamespace).Get(consoleLabel, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve console route %s", consoleLabel)
	}
	consoleUrl := fmt.Sprintf("https://%s", consoleRoute.Spec.Host)
	consoleTargeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    consoleUrl,
	})

	// Create a monitor for the oauth URL
	oauthRoute, err := h.Route().RouteV1().Routes(oauthNamespace).Get(oauthName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve oauth route %s", oauthName)
	}
	oauthUrl := fmt.Sprintf("https://%s/healthz", oauthRoute.Spec.Host)
	oauthTargeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    oauthUrl,
	})

	// Create monitors for API Server URLs
	apiservers, err := h.Cfg().ConfigV1().APIServers().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of API servers")
	}
	apiTargeters := make([]vegeta.Targeter, 0)
	for _, apiServer := range apiservers.Items {
		for _, servingCert := range apiServer.Spec.ServingCerts.NamedCertificates {
			for _, name := range servingCert.Names {
				apiUrl := fmt.Sprintf("https://%s:6443/healthz", name)
				apiTargeters = append(apiTargeters,
					vegeta.NewStaticTargeter(vegeta.Target{
						Method: "GET",
						URL:    apiUrl,
					}))
			}
		}
	}

	// If we created any routes during workload testing, let's add them too
	workloadRoutes, err := h.Route().RouteV1().Routes(h.CurrentProject()).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of workload routes")
	}
	workloadTargeters := make([]vegeta.Targeter, 0)
	for _, workloadRoute := range workloadRoutes.Items {
		workloadUrl := fmt.Sprintf("https://%s", workloadRoute.Spec.Host)
		workloadTargeters = append(workloadTargeters,
			vegeta.NewStaticTargeter(vegeta.Target{
				Method: "GET",
				URL:    workloadUrl,
			}))
	}

	return &RouteMonitors{
		Monitors:          make(map[string]<-chan *vegeta.Result, 0),
		Metrics:           make(map[string]*vegeta.Metrics, 0),
		Plots:             make(map[string]*plot.Plot, 0),
		consoleTargeter:   consoleTargeter,
		oauthTargeter:     oauthTargeter,
		apiTargeters:      apiTargeters,
		workloadTargeters: workloadTargeters,
	}, nil
}

// Sets the RouteMonitors to begin polling the configured routes with traffic
func (rm *RouteMonitors) Start() {
	pollRate := vegeta.Rate{Freq: pollFrequency, Per: time.Second}
	timeout := vegeta.Timeout(timeoutSeconds)
	consoleAttacker := vegeta.NewAttacker(timeout)
	oauthAttacker := vegeta.NewAttacker(timeout)

	rm.Monitors["console"] = consoleAttacker.Attack(rm.consoleTargeter, pollRate, 0, "console")
	rm.Plots["console"] = createPlot("console")
	rm.Monitors["oauth"] = oauthAttacker.Attack(rm.oauthTargeter, pollRate, 0, "oauth")
	rm.Plots["oauth"] = createPlot("oauth")
	for i, apiTargeter := range rm.apiTargeters {
		apititle := fmt.Sprintf("api-%d", i)
		apiAttacker := vegeta.NewAttacker(timeout)
		rm.Monitors[apititle] = apiAttacker.Attack(apiTargeter, pollRate, 0, apititle)
		rm.Plots[apititle] = createPlot(apititle)
	}
	for i, workloadTargeter := range rm.workloadTargeters {
		workloadTitle := fmt.Sprintf("workload-%d", i)
		workloadAttacker := vegeta.NewAttacker(timeout)
		rm.Monitors[workloadTitle] = workloadAttacker.Attack(workloadTargeter, pollRate, 0, workloadTitle)
		rm.Plots[workloadTitle] = createPlot(workloadTitle)
	}

	for title, _ := range rm.Monitors {
		rm.Metrics[title] = &vegeta.Metrics{}
	}
}

// Sets the RouteMonitors to cease polling the configured routes with traffic
func (rm *RouteMonitors) End() {
	for _, attacker := range rm.attackers {
		attacker.Stop()
	}
	for _, metric := range rm.Metrics {
		metric.Close()
	}
}

// Stores the measured RouteMonitor metrics inside osde2e metadata for DataHub
func (rm *RouteMonitors) StoreMetadata() {
	for title, metric := range rm.Metrics {
		latency := float64(metric.Latencies.Mean / time.Millisecond)
		if latency < 0 {
			latency = 0
		}
		metadata.Instance.SetRouteLatency(title, latency)
		metadata.Instance.SetRouteThroughput(title, metric.Throughput)
		metadata.Instance.SetRouteAvailability(title, metric.Success)
	}
}

// Saves the measured RouteMonitor metrics in HDR Histogram reports in the specified base directory
func (rm *RouteMonitors) SaveReports(baseDir string) error {
	outputDirectory := filepath.Join(baseDir, "route-monitors")
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(outputDirectory, os.FileMode(0755)); err != nil {
			return fmt.Errorf("error while creating route monitor report directory %s: %v", outputDirectory, err)
		}
	}
	for title, metric := range rm.Metrics {
		histoPath := filepath.Join(outputDirectory, fmt.Sprintf("%s.histo", title))
		reporter := vegeta.NewHDRHistogramPlotReporter(metric)
		out, err := os.Create(histoPath)
		if err != nil {
			return fmt.Errorf("error while creating route monitor report: %v", err)
		}
		reporter.Report(out)
		log.Printf("Wrote route monitor histogram: %s\n", histoPath)
	}
	return nil
}

// Saves HTML-formatted latency plots in the specified base directory
func (rm *RouteMonitors) SavePlots(baseDir string) error {
	outputDirectory := filepath.Join(baseDir, "route-monitors")
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(outputDirectory, os.FileMode(0755)); err != nil {
			return fmt.Errorf("error while creating route monitor report directory %s: %v", outputDirectory, err)
		}
	}
	for title, pl := range rm.Plots {
		plotPath := filepath.Join(outputDirectory, fmt.Sprintf("%s.html", title))
		out, err := os.Create(plotPath)
		if err != nil {
			return fmt.Errorf("error while creating route monitor report: %v", err)
		}
		pl.WriteTo(out)
		log.Printf("Wrote route monitor plot: %s\n", plotPath)

	}
	return nil
}

// Creates a new plot with the specified title
func createPlot(title string) *plot.Plot {
	return plot.New(
		plot.Title(title),
		plot.Label(plot.ErrorLabeler),
	)
}
