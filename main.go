package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/command"
	"github.com/CloudOpsKit/smartctl_ssacli_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddr  = flag.String("listen", ":9633", "address for exporter")
	metricsPath = flag.String("path", "/metrics", "URL path for surfacing collected metrics")
	devicePath  = flag.String("device", "/dev/sda", "Path to the raid controller device (e.g. /dev/sda or /dev/sg0)")
	timeout     = flag.Duration("timeout", command.Timeout, "Timeout for each ssacli/smartctl call")
	interval    = flag.Duration("interval", time.Minute, "How often to collect metrics in the background")
)

func main() {
	flag.Parse()
	command.Timeout = *timeout

	e := exporter.New(*devicePath)
	e.Start(*interval)
	prometheus.MustRegister(e)

	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Smartctl & SSACLI Exporter</title></head>
             <body>
             <h1>Smartctl & SSACLI Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
			 <p>Controller Device: ` + *devicePath + `</p>
             </body>
             </html>`))
	})

	log.Printf("Beginning to serve on %s (Device: %s, collect interval: %s)", *listenAddr, *devicePath, *interval)

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("Cannot start exporter: %s", err)
	}
}
