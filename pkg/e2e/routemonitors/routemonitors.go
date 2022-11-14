package routemonitors

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	vegeta "github.com/tsenart/vegeta/lib"
	"github.com/tsenart/vegeta/lib/plot"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	consoleNamespace    = "openshift-console"
	consoleLabel        = "console"
	monitoringNamespace = "openshift-monitoring"
	oauthNamespace      = "openshift-authentication"
	oauthName           = "oauth-openshift"
)

type RouteMonitors struct {
	Monitors   map[string]<-chan *vegeta.Result
	Metrics    map[string]*vegeta.Metrics
	Plots      map[string]*plot.Plot
	targeters  map[string]vegeta.Targeter
	attackers  []*vegeta.Attacker
	ReportData map[string][]RouteMonData
}

// Frequency of requests per second (per route)
const (
	pollFrequency  = 3
	timeoutSeconds = 3 * time.Second
)

// Data Structure used to store time-value values of Route Monitor Data
type RouteMonData struct {
	Time, Value float64
}

// Detects the available routes in the cluster and initializes monitors for their availability
func Create(ctx context.Context) (*RouteMonitors, error) {
	h := helper.NewOutsideGinkgo()

	if h == nil {
		return nil, fmt.Errorf("unable to generate helper outside ginkgo")
	}

	// record all targeters created in a map, accessible via a key which is their URL
	targeters := make(map[string]vegeta.Targeter, 0)

	// Create a monitor for the web console
	consoleRoute, err := h.Route().RouteV1().Routes(consoleNamespace).Get(ctx, consoleLabel, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve console route %s", consoleLabel)
	}
	consoleUrl := fmt.Sprintf("https://%s", consoleRoute.Spec.Host)
	u, err := url.Parse(consoleUrl)
	if err == nil {
		consoleTargeter := vegeta.NewStaticTargeter(vegeta.Target{
			Method: "GET",
			URL:    consoleUrl,
		})
		targeters[u.Host] = consoleTargeter
	}

	// Create a monitor for the oauth URL
	oauthRoute, err := h.Route().RouteV1().Routes(oauthNamespace).Get(ctx, oauthName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve oauth route %s", oauthName)
	}
	oauthUrl := fmt.Sprintf("https://%s/healthz", oauthRoute.Spec.Host)
	u, err = url.Parse(oauthUrl)
	if err == nil {
		oauthTargeter := vegeta.NewStaticTargeter(vegeta.Target{
			Method: "GET",
			URL:    oauthUrl,
		})
		targeters[u.Host] = oauthTargeter
	}

	// Create monitors for API Server URLs
	apiservers, err := h.Cfg().ConfigV1().APIServers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of API servers")
	}
	for _, apiServer := range apiservers.Items {
		for _, servingCert := range apiServer.Spec.ServingCerts.NamedCertificates {
			for _, name := range servingCert.Names {
				apiUrl := fmt.Sprintf("https://%s:6443/healthz", name)
				apiTargeter := vegeta.NewStaticTargeter(vegeta.Target{
					Method: "GET",
					URL:    apiUrl,
				})
				u, err := url.Parse(apiUrl)
				if err == nil {
					targeters[u.Host] = apiTargeter
				}
			}
		}
	}

	// If we created any routes during workload testing, let's add them too
	workloadRoutes, err := h.Route().RouteV1().Routes(h.CurrentProject()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of workload routes")
	}
	for _, workloadRoute := range workloadRoutes.Items {
		workloadUrl := fmt.Sprintf("https://%s", workloadRoute.Spec.Host)
		u, err := url.Parse(workloadUrl)
		if err == nil {
			workloadTargeter := vegeta.NewStaticTargeter(vegeta.Target{
				Method: "GET",
				URL:    workloadUrl,
			})
			targeters[u.Host] = workloadTargeter
		}
	}

	// Create monitors for each route existing in openshift-monitoring
	monitoringRoutes, err := h.Route().RouteV1().Routes(monitoringNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not retrieve list of workload routes")
	}
	for _, monitoringRoute := range monitoringRoutes.Items {
		monitoringURL := fmt.Sprintf("https://%s", monitoringRoute.Spec.Host)
		u, err := url.Parse(monitoringURL)
		if err == nil {
			monitoringTargeter := vegeta.NewStaticTargeter(vegeta.Target{
				Method: "GET",
				URL:    monitoringURL,
			})
			targeters[u.Host] = monitoringTargeter
		}
	}

	return &RouteMonitors{
		Monitors:   make(map[string]<-chan *vegeta.Result),
		Metrics:    make(map[string]*vegeta.Metrics),
		Plots:      make(map[string]*plot.Plot),
		targeters:  targeters,
		ReportData: make(map[string][]RouteMonData),
	}, nil
}

// Sets the RouteMonitors to begin polling the configured routes with traffic
func (rm *RouteMonitors) Start() {
	pollRate := vegeta.Rate{Freq: pollFrequency, Per: time.Second}
	timeout := vegeta.Timeout(timeoutSeconds)

	for url, targeter := range rm.targeters {
		log.Printf("Setting up monitor for %s\n", url)
		attacker := vegeta.NewAttacker(timeout)
		rm.Monitors[url] = attacker.Attack(targeter, pollRate, 0, url)
		rm.Plots[url] = createPlot(url)
		rm.attackers = append(rm.attackers, attacker)
	}

	for title := range rm.Monitors {
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
		if math.IsNaN(metric.Throughput) {
			metric.Throughput = 0
		}
		if math.IsNaN(metric.Success) {
			metric.Success = 0
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
		if err := os.Mkdir(outputDirectory, os.FileMode(0o755)); err != nil {
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
		if err := os.Mkdir(outputDirectory, os.FileMode(0o755)); err != nil {
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

// Extracts the numbers and title from the route monitor plot files
func (rm *RouteMonitors) ExtractData(baseDir string) error {
	outputDirectory := filepath.Join(baseDir, "route-monitors")
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirectory, os.FileMode(0o755)); err != nil {
			return fmt.Errorf("error while creating route monitor report directory %s: %v", outputDirectory, err)
		}
	}

	for title := range rm.Plots {
		// Open the plot file
		htmlFilePath := filepath.Join(outputDirectory, fmt.Sprintf("%s.html", title))
		file, err := os.Open(htmlFilePath)
		if err != nil {
			log.Printf("Unable to read route monitor plot file; %v\n", err)
		}

		// Regex to match numbers for the data variable inside the plot file
		startRegex := regexp.MustCompile(`var data`)
		dataNumRegex := regexp.MustCompile(`[0-9]*\.?[0-9]+,[0-9]*\.?[0-9]+`)

		// This list stores the numbers

		dataList := make([]RouteMonData, 0)

		// Scan each line in the html file to extract data
		scanner := bufio.NewReader(file)
	scanLoop:
		for {
			fileBytes, _, err := scanner.ReadLine()
			if err != nil {
				break scanLoop
			}
			line := strings.TrimSpace(string(fileBytes))

			if readData := len(line) > 0 && startRegex.MatchString(line); !readData {
				continue scanLoop
			}
			if dataNumRegex.FindString(line) == "" {
				continue scanLoop
			}
			numStrings := strings.Split(startRegex.FindString(line), ",")
			if len(numStrings) < 2 {
				continue scanLoop
			}
			numData := RouteMonData{}
			numData.Time, err = strconv.ParseFloat(numStrings[0], 64)
			if err != nil {
				log.Printf("Error while parsing route monitor data values - %v", err)
			}
			numData.Value, err = strconv.ParseFloat(numStrings[1], 64)
			if err != nil {
				log.Printf("Error while parsing route monitor data values - %v", err)
			}
			dataList = append(dataList, numData)
		}

		// Store the extracted data corresponding to the plotfile title
		rm.ReportData[title] = dataList

		if err != nil {
			log.Printf("Error - %v", err.Error())
		}
	}

	return nil
}
