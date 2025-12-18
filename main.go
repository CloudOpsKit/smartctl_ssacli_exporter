package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddr  = flag.String("listen", ":9633", "address for exporter")
	metricsPath = flag.String("path", "/metrics", "URL path for surfacing collected metrics")
	devicePath  = flag.String("device", "/dev/sda", "Path to the raid controller device (e.g. /dev/sda or /dev/sg0)")
)

func main() {
	flag.Parse()

	prometheus.MustRegister(exporter.New(*devicePath))

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

	log.Printf("Beginning to serve on %s (Device: %s)", *listenAddr, *devicePath)

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("Cannot start exporter: %s", err)
	}
}
