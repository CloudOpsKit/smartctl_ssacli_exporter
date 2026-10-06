package exporter

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/collector"
	"github.com/CloudOpsKit/smartctl_ssacli_exporter/command"
	"github.com/prometheus/client_golang/prometheus"
)

// An Exporter is a Prometheus exporter for metrics.
// It wraps all metrics collectors and provides a single global
// exporter which can serve metrics.
//
// ssacli and smartctl take seconds per call, so metrics are collected in
// the background (see Start) and scrapes are served from a cache.
//
// It implements the exporter.Collector interface in order to register
// with Prometheus.
type Exporter struct {
	devicePath string

	mu      sync.RWMutex
	metrics []prometheus.Metric

	successDesc   *prometheus.Desc
	durationDesc  *prometheus.Desc
	timestampDesc *prometheus.Desc
}

var _ prometheus.Collector = &Exporter{}

// New creates a new Exporter for the given raid controller device.
func New(devicePath string) *Exporter {
	const namespace = "smartctl_ssacli_exporter"
	return &Exporter{
		devicePath: devicePath,
		successDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_collect_success"),
			"Whether the last background collection succeeded (1) or failed (0)",
			nil, nil,
		),
		durationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_collect_duration_seconds"),
			"Duration of the last background collection",
			nil, nil,
		),
		timestampDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "last_collect_timestamp_seconds"),
			"Unix time when the last background collection finished",
			nil, nil,
		),
	}
}

// Start collects metrics right away and then every interval in a background
// goroutine. Collections never overlap: a tick that fires while a slow
// collection is still running is dropped.
func (e *Exporter) Start(interval time.Duration) {
	go func() {
		e.refresh()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			e.refresh()
		}
	}()
}

// refresh runs a full collection and replaces the cached metrics with its result.
func (e *Exporter) refresh() {
	start := time.Now()

	ch := make(chan prometheus.Metric)
	done := make(chan struct{})
	var metrics []prometheus.Metric
	go func() {
		for m := range ch {
			metrics = append(metrics, m)
		}
		close(done)
	}()

	err := e.collect(ch)
	close(ch)
	<-done

	success := 1.0
	if err != nil {
		log.Printf("[ERROR] collection failed: %v", err)
		success = 0
	}
	finished := time.Now()
	metrics = append(metrics,
		prometheus.MustNewConstMetric(e.successDesc, prometheus.GaugeValue, success),
		prometheus.MustNewConstMetric(e.durationDesc, prometheus.GaugeValue, finished.Sub(start).Seconds()),
		prometheus.MustNewConstMetric(e.timestampDesc, prometheus.GaugeValue, float64(finished.Unix())),
	)

	e.mu.Lock()
	e.metrics = metrics
	e.mu.Unlock()
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	collector.NewSsacliSumCollector().Describe(ch)
	collector.NewSsacliPhysDiskCollector("", "").Describe(ch)
	collector.NewSmartctlDiskCollector(e.devicePath, "", 0).Describe(ch)
	collector.NewSsacliLogDiskCollector("", "").Describe(ch)
	ch <- e.successDesc
	ch <- e.durationDesc
	ch <- e.timestampDesc
}

// Collect sends the metrics cached by the last background collection.
// Nothing is sent until the first collection finishes.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, m := range e.metrics {
		ch <- m
	}
}

// collect runs ssacli and smartctl and sends the resulting metrics to ch.
func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	// One "ctrl all show detail" call feeds both the controller summary
	// and the slot list, since each ssacli call takes seconds
	out, err := command.Run("ssacli", "ctrl", "all", "show", "detail")
	if err != nil {
		return fmt.Errorf("getting controller details: %w", err)
	}

	collector.NewSsacliSumCollectorWithData(string(out)).Collect(ch)
	slotIDs := parseControllerSlots(string(out))

	var wg sync.WaitGroup

	for _, slotID := range slotIDs {
		pds, err := getPhysicalDisksBulk(slotID)
		if err != nil {
			log.Printf("[ERROR] failed getting bulk PD data for slot %s: %v", slotID, err)
			continue
		}

		// The drive's position in ssacli output is its smartctl "-d cciss,N" index
		for smartCtlIndex, pd := range pds {
			wg.Add(1)
			go func(sID, pID, data string, idx int) {
				defer wg.Done()

				// Pass pre-collected raw data to the collector
				// This prevents the collector from running its own 'ssacli' command
				collector.NewSsacliPhysDiskCollectorWithData(pID, sID, data).Collect(ch)

				// SMART metrics still need separate 'smartctl' calls
				// because they talk to the disk firmware directly
				collector.NewSmartctlDiskCollector(e.devicePath, pID, idx).Collect(ch)
			}(slotID, pd.ID, pd.Data, smartCtlIndex)
		}

		ldDataMap, err := getLogicalDrivesBulk(slotID)
		if err == nil {
			for ldID, rawData := range ldDataMap {
				wg.Add(1)
				go func(sID, lID, data string) {
					defer wg.Done()
					collector.NewSsacliLogDiskCollectorWithData(lID, sID, data).Collect(ch)
				}(slotID, ldID, rawData)
			}
		}
	}
	wg.Wait()
	return nil
}
