package exporter

import (
	"log"
	"sync"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// An Exporter is a Prometheus exporter for metrics.
// It wraps all metrics collectors and provides a single global
// exporter which can serve metrics.
//
// It implements the exporter.Collector interface in order to register
// with Prometheus.
type Exporter struct {
	devicePath string
}

var _ prometheus.Collector = &Exporter{}

// New creates a new Exporter which collects metrics by creating a apcupsd
// client using the input ClientFunc.
func New(devicePath string) *Exporter {
	return &Exporter{
		devicePath: devicePath,
	}
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	collector.NewSsacliSumCollector().Describe(ch)
	collector.NewSsacliPhysDiskCollector("", "").Describe(ch)
	collector.NewSmartctlDiskCollector(e.devicePath, "", 0).Describe(ch)
	collector.NewSsacliLogDiskCollector("", "").Describe(ch)
}

// Collect sends the collected metrics from each of the collectors to
// exporter.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	collector.NewSsacliSumCollector().Collect(ch)

	slotIDs, err := getControllerSlots()
	if err != nil {
		log.Printf("[ERROR] failed getting controller slots: %v", err)
		return
	}

	var wg sync.WaitGroup

	for _, slotID := range slotIDs {
		pdDataMap, err := getPhysicalDisksBulk(slotID)
		if err != nil {
			log.Printf("[ERROR] failed getting bulk PD data for slot %s: %v", slotID, err)
			continue
		}

		smartCtlIndex := 0
		for pdID, rawData := range pdDataMap {
			wg.Add(1)
			go func(sID, pID, data string, idx int) {
				defer wg.Done()

				// NEW: Pass pre-collected raw data to the collector
				// This prevents the collector from running its own 'ssacli' command
				collector.NewSsacliPhysDiskCollectorWithData(pID, sID, data).Collect(ch)

				// SMART metrics still need separate 'smartctl' calls
				// because they talk to the disk firmware directly
				collector.NewSmartctlDiskCollector(e.devicePath, pID, idx).Collect(ch)
			}(slotID, pdID, rawData, smartCtlIndex)

			smartCtlIndex++
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
}
