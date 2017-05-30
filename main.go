package main

import (
	"flag"
	"log"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"os"
	"io/ioutil"
	"time"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var errorCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "mesos",
	Subsystem: "collector",
	Name:      "errors_total",
	Help:      "Total number of internal mesos-collector errors.",
})

func init() {
	prometheus.MustRegister(errorCounter)
}

func getX509CertPool(pemFiles []string) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, f := range pemFiles {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		ok := pool.AppendCertsFromPEM(content)
		if !ok {
			log.Fatal("Error parsing .pem file %s", f)
		}
	}
	return pool
}

func getTrustedRedirectsMap(trustedRedirects []string) map[string]bool {
	trustedRedirectsMap := make(map[string]bool)
	for _, r := range trustedRedirects {
		trustedRedirectsMap[r] = true
	}
	return trustedRedirectsMap
}

func mkHttpClient(url string, timeout time.Duration, auth authInfo, certPool *x509.CertPool, trustedRedirects map[string]bool) *httpClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: certPool},
	}
	return &httpClient{
		http.Client{
			Timeout: timeout,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				redirectHost := req.URL.Hostname()
				if _, ok := trustedRedirects[redirectHost]; ok {
					if auth.username != "" && auth.password != "" {
						req.SetBasicAuth(auth.username, auth.password)
					}
					return nil
				} else {
					log.Printf("Redirect to '%s' not trusted", redirectHost)
					return http.ErrUseLastResponse
				}
			},
		},
		url,
		auth,
		trustedRedirects,
	}
}

func main() {
	fs := flag.NewFlagSet("mesos-exporter", flag.ExitOnError)
	addr := fs.String("addr", ":9110", "Address to listen on")
	masterURL := fs.String("master", "", "Expose metrics from master running on this URL")
	slaveURL := fs.String("slave", "", "Expose metrics from slave running on this URL")
	timeout := fs.Duration("timeout", 5*time.Second, "Master polling timeout")
	exportedTaskLabels := fs.String("exportedTaskLabels", "", "Comma-separated list of task labels to include in the task_labels metric")
	ignoreCompletedFrameworkTasks := fs.Bool("ignoreCompletedFrameworkTasks", false, "Don't export task_state_time metric");
	trustedCerts := fs.String("trustedCerts", "", "Comma-separated list of certificates (.pem files) trusted for requests to Mesos endpoints")
	trustedRedirects := fs.String("trustedRedirects", "", "Comma-separated list of trusted hosts (ip addresses, host names) where metrics requests can be redirected")

	fs.Parse(os.Args[1:])
	if *masterURL != "" && *slaveURL != "" {
		log.Fatal("Only -master or -slave can be given at a time")
	}

	auth := authInfo{
		os.Getenv("MESOS_EXPORTER_USERNAME"),
		os.Getenv("MESOS_EXPORTER_PASSWORD"),
	}

	var certPool *x509.CertPool = nil
	if *trustedCerts != "" {
		certPool = getX509CertPool(strings.Split(*trustedCerts, ","))
	}

	var trustedRedirectsMap map[string]bool = nil
	if *trustedRedirects != "" {
		trustedRedirectsMap = getTrustedRedirectsMap(strings.Split(*trustedRedirects, ","))
	}

	switch {
	case *masterURL != "":
		for _, f := range []func(*httpClient) prometheus.Collector{
			newMasterCollector,
			func(c *httpClient) prometheus.Collector {
				return newMasterStateCollector(c, *ignoreCompletedFrameworkTasks)
			},
		} {
			c := f(mkHttpClient(*masterURL, *timeout, auth, certPool, trustedRedirectsMap));
			if err := prometheus.Register(c); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("Exposing master metrics on %s", *addr)

	case *slaveURL != "":
		slaveCollectors := []func(*httpClient) prometheus.Collector{
			func(c *httpClient) prometheus.Collector {
				return newSlaveCollector(c)
			},
			func(c *httpClient) prometheus.Collector {
				return newSlaveMonitorCollector(c)
			},
		}
		if *exportedTaskLabels != "" {
			slaveLabels := strings.Split(*exportedTaskLabels, ",");
			slaveCollectors = append(slaveCollectors, func (c *httpClient) prometheus.Collector{
				return newSlaveStateCollector(c, slaveLabels)
			})
		}

		for _, f := range slaveCollectors {
			c := f(mkHttpClient(*slaveURL, *timeout, auth, certPool, trustedRedirectsMap));
			if err := prometheus.Register(c); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("Exposing slave metrics on %s", *addr)

	default:
		log.Fatal("Either -master or -slave is required")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Mesos Exporter</title></head>
            <body>
            <h1>Mesos Exporter</h1>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})
	http.Handle("/metrics", prometheus.Handler())
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
