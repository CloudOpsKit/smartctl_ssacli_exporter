package collector

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/parser"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &SsacliPhysDiskCollector{}

type SsacliPhysDiskCollector struct {
	diskID             string
	slotID             string
	rawData            string
	physDiskStatusDesc *prometheus.Desc
}

func NewSsacliPhysDiskCollector(diskID string, slotID string) *SsacliPhysDiskCollector {
	return NewSsacliPhysDiskCollectorWithData(diskID, slotID, "")
}

func NewSsacliPhysDiskCollectorWithData(diskID string, slotID string, data string) *SsacliPhysDiskCollector {
	var (
		namespace = "ssacli"
		subsystem = "phys_disk"
		labels    = []string{
			"physDiskID",
			"physDiskDriveType",
			"physDiskInterfaceType",
			"physDiskSize",
			"physDiskStatus",
			"physDiskSerialNumber",
			"physDiskModel",
			"physDiskCurrentTemperature",
			"physDiskMaximumTemperature",
			"physDiskBay",
			"physDiskSlotID",
		}
	)

	return &SsacliPhysDiskCollector{
		diskID:  diskID,
		slotID:  slotID,
		rawData: data,
		physDiskStatusDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Hardware raid physical disk status (1 if OK, 0 otherwise)",
			labels,
			nil,
		),
	}
}

func (c *SsacliPhysDiskCollector) Describe(ch chan<- *prometheus.Desc) {
	if c.physDiskStatusDesc != nil {
		ch <- c.physDiskStatusDesc
	}
}

func (c *SsacliPhysDiskCollector) Collect(ch chan<- prometheus.Metric) {
	if c.physDiskStatusDesc == nil {
		log.Printf("[ERROR] physDiskStatusDesc is NIL for disk %s. Check constructor!", c.diskID)
		return
	}

	if _, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting phys disk metrics for %s: %v", c.diskID, err)
	}
}

func (c *SsacliPhysDiskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if c.diskID == "" {
		return nil, nil
	}

	var output string
	if c.rawData != "" {
		output = c.rawData
	} else {
		slotArg := "slot=" + c.slotID
		out, err := exec.Command("ssacli", "ctrl", slotArg, "pd", c.diskID, "show", "detail").CombinedOutput()
		if err != nil {
			return nil, err
		}
		output = string(out)
	}

	data := parser.ParseSsacliPhysDisk(output)
	if data == nil {
		return nil, nil
	}

	for i := range data.SsacliPhysDiskData {
		labels := []string{
			c.diskID,
			data.SsacliPhysDiskData[i].DriveType,
			data.SsacliPhysDiskData[i].IntType,
			data.SsacliPhysDiskData[i].Size,
			data.SsacliPhysDiskData[i].Status,
			data.SsacliPhysDiskData[i].SN,
			data.SsacliPhysDiskData[i].Model,
			fmt.Sprintf("%.0f", data.SsacliPhysDiskData[i].CurTemp),
			fmt.Sprintf("%.0f", data.SsacliPhysDiskData[i].MaxTemp),
			data.SsacliPhysDiskData[i].Bay,
			c.slotID,
		}

		val := 0.0
		if data.SsacliPhysDiskData[i].Status == "OK" {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.physDiskStatusDesc,
			prometheus.GaugeValue,
			val,
			labels...,
		)
	}

	return nil, nil
}
