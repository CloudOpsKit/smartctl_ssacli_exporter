package collector

import (
	"log"
	"os/exec"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/parser"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &SsacliLogDiskCollector{}

type SsacliLogDiskCollector struct {
	diskID            string
	slotID            string
	rawData           string
	logDiskStatusDesc *prometheus.Desc
}

func NewSsacliLogDiskCollector(diskID string, slotID string) *SsacliLogDiskCollector {
	return NewSsacliLogDiskCollectorWithData(diskID, slotID, "")
}

func NewSsacliLogDiskCollectorWithData(diskID string, slotID string, data string) *SsacliLogDiskCollector {
	var (
		namespace = "ssacli"
		subsystem = "log_disk"
		labels    = []string{
			"logDiskID",
			"logDiskSize",
			"logDiskFaultTolerance",
			"logDiskStatus",
			"logDiskCaching",
			"logDiskUME",
			"logDiskSlotID",
		}
	)

	return &SsacliLogDiskCollector{
		diskID:  diskID,
		slotID:  slotID,
		rawData: data,
		logDiskStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Hardware raid logical drive status (1 if OK, 0 otherwise)",
			labels,
			nil,
		),
	}
}

func (c *SsacliLogDiskCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.logDiskStatusDesc
}

func (c *SsacliLogDiskCollector) Collect(ch chan<- prometheus.Metric) {
	if c.logDiskStatusDesc == nil {
		log.Printf("[ERROR] logDiskStatusDesc is not initialized for %s", c.diskID)
		return
	}

	if _, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting logical disk metrics for %s: %v", c.diskID, err)
	}
}

func (c *SsacliLogDiskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if c.diskID == "" {
		return nil, nil
	}

	var output string
	if c.rawData != "" {
		output = c.rawData
	} else {
		slotArg := "slot=" + c.slotID
		out, err := exec.Command("ssacli", "ctrl", slotArg, "ld", c.diskID, "show").CombinedOutput()
		if err != nil {
			return nil, err
		}
		output = string(out)
	}

	data := parser.ParseSsacliLogDisk(output)
	if data == nil || len(data.SsacliLogDiskData) == 0 {
		return nil, nil
	}

	for i := range data.SsacliLogDiskData {
		labels := []string{
			c.diskID,
			data.SsacliLogDiskData[i].Size,
			data.SsacliLogDiskData[i].FaultTolerance,
			data.SsacliLogDiskData[i].Status,
			data.SsacliLogDiskData[i].Caching,
			data.SsacliLogDiskData[i].UME,
			c.slotID,
		}

		val := 0.0
		if data.SsacliLogDiskData[i].Status == "OK" {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.logDiskStatusDesc,
			prometheus.GaugeValue,
			val,
			labels...,
		)
	}

	return nil, nil
}
