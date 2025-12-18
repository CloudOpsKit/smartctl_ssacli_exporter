package exporter

import (
	"log"
	"os/exec"
	"strings"

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

	for _, slotID := range slotIDs {
		pdIDs, err := getPhysicalDisks(slotID)
		if err != nil {
			log.Printf("[ERROR] failed getting PDs for slot %s: %v", slotID, err)
		} else {
			// smartCtlIndex corresponds to the disk number for smartctl (e.g. -d cciss,0)
			// We reset it to 0 for every controller loop, or increment it based on found disks.
			// Assuming standard mapping:
			smartCtlIndex := 0

			for _, pdID := range pdIDs {
				// Collect SSACLI metrics for this physical disk
				collector.NewSsacliPhysDiskCollector(pdID, slotID).Collect(ch)

				// Collect SMART metrics for this physical disk
				// We pass the devicePath (e.g. /dev/sda) and the calculated index
				collector.NewSmartctlDiskCollector(e.devicePath, pdID, smartCtlIndex).Collect(ch)

				smartCtlIndex++
			}
		}

		// --- Collect Logical Drives ---
		ldIDs, err := getLogicalDrives(slotID)
		if err != nil {
			log.Printf("[ERROR] failed getting LDs for slot %s: %v", slotID, err)
		} else {
			for _, ldID := range ldIDs {
				collector.NewSsacliLogDiskCollector(ldID, slotID).Collect(ch)
			}
		}
	}
}

// getControllerSlots parses "ssacli ctrl all show status" output.
// Example line: "Smart Array P440ar in Slot 0 (Embedded)    Status: OK"
func getControllerSlots() ([]string, error) {
	out, err := exec.Command("ssacli", "ctrl", "all", "show", "status").CombinedOutput()
	if err != nil {
		return nil, err
	}

	var slots []string
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		// Split line into fields by whitespace
		fields := strings.Fields(line)

		// Look for the word "Slot" and take the next field
		for i, field := range fields {
			if field == "Slot" && i+1 < len(fields) {
				slots = append(slots, fields[i+1])
				break // Move to next line after finding the slot
			}
		}
	}
	return slots, nil
}

// getPhysicalDisks parses "ssacli ctrl slot=... pd all show status" output.
// Example line: "physicaldrive 1I:1:1 (port 1I:box 1:bay 1, 600 GB): OK"
func getPhysicalDisks(slotID string) ([]string, error) {
	// Construct arguments: ctrl slot=0 pd all show status
	slotArg := "slot=" + slotID
	out, err := exec.Command("ssacli", "ctrl", slotArg, "pd", "all", "show", "status").CombinedOutput()
	if err != nil {
		return nil, err
	}

	var pds []string
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		// Check if line starts with "physicaldrive" and has enough fields
		if len(fields) >= 2 && fields[0] == "physicaldrive" {
			// fields[1] contains the ID, e.g., "1I:1:1"
			pds = append(pds, fields[1])
		}
	}
	return pds, nil
}

// getLogicalDrives parses "ssacli ctrl slot=... ld all show status" output.
// Example line: "logicaldrive 1 (600 GB, RAID 1): OK"
func getLogicalDrives(slotID string) ([]string, error) {
	slotArg := "slot=" + slotID
	out, err := exec.Command("ssacli", "ctrl", slotArg, "ld", "all", "show", "status").CombinedOutput()
	if err != nil {
		return nil, err
	}

	var lds []string
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		// Check if line starts with "logicaldrive"
		if len(fields) >= 2 && fields[0] == "logicaldrive" {
			// fields[1] contains the ID, e.g., "1"
			lds = append(lds, fields[1])
		}
	}
	return lds, nil
}
