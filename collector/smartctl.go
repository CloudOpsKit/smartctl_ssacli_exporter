package collector

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/parser"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &SmartctlDiskCollector{}

type SmartctlDiskCollector struct {
	diskID     string
	diskN      int
	devicePath string

	rawReadErrorRate      *prometheus.Desc
	reallocatedSectorCt   *prometheus.Desc
	powerOnHours          *prometheus.Desc
	powerCycleCount       *prometheus.Desc
	runtimeBadBlock       *prometheus.Desc
	endToEndError         *prometheus.Desc
	reportedUncorrect     *prometheus.Desc
	commandTimeout        *prometheus.Desc
	hardwareECCRecovered  *prometheus.Desc
	reallocatedEventCount *prometheus.Desc
	currentPendingSector  *prometheus.Desc
	offlineUncorrectable  *prometheus.Desc
	uDMACRCErrorCount     *prometheus.Desc
	unusedRsvdBlkCntTot   *prometheus.Desc
	grownDefects          *prometheus.Desc
	spinUpTime            *prometheus.Desc
	startStopCount        *prometheus.Desc
	seekErrorRate         *prometheus.Desc
	spinRetryCount        *prometheus.Desc
	airflowTemperature    *prometheus.Desc
	temperatureCelsius    *prometheus.Desc
	loadCycleCount        *prometheus.Desc
	totalLBAsWritten      *prometheus.Desc
	totalLBAsRead         *prometheus.Desc
}

func NewSmartctlDiskCollector(devicePath string, diskID string, diskN int) *SmartctlDiskCollector {
	var (
		namespace = "smartctl"
		subsystem = "physical_disk"
		labels    = []string{
			"diskID",
			"model",
			"sn",
			"rotRate",
			"fromFact",
		}
	)

	return &SmartctlDiskCollector{
		diskID:                diskID,
		diskN:                 diskN,
		devicePath:            devicePath,
		rawReadErrorRate:      prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rawReadErrorRate"), "Smartctl raw read error rate", labels, nil),
		reallocatedSectorCt:   prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "reallocatedSectorCt"), "Smartctl reallocated sector ct", labels, nil),
		powerOnHours:          prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "powerOnHours"), "Smartctl power on hours", labels, nil),
		powerCycleCount:       prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "powerCycleCount"), "Smartctl power cycle down count", labels, nil),
		runtimeBadBlock:       prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "runtimeBadBlock"), "Smartctl runtime bad block", labels, nil),
		endToEndError:         prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "endToEndError"), "Smartctl end to end error", labels, nil),
		reportedUncorrect:     prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "reportedUncorrect"), "Smartctl reported uncorrect", labels, nil),
		commandTimeout:        prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "commandTimeout"), "Smartctl command timeout", labels, nil),
		hardwareECCRecovered:  prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "hardwareECCRecovered"), "Smartctl hardware ecc recovered", labels, nil),
		reallocatedEventCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "reallocatedEventCount"), "Smartctl reallocated event count", labels, nil),
		currentPendingSector:  prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "currentPendingSector"), "Smartctl current pending sector", labels, nil),
		offlineUncorrectable:  prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "offlineUncorrectable"), "Smartctl offline uncorrectable", labels, nil),
		uDMACRCErrorCount:     prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "uDMACRCErrorCount"), "Smartctl ud macrc error count", labels, nil),
		unusedRsvdBlkCntTot:   prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unusedRsvdBlkCntTot"), "Smartctl unused rsvd block Count Total", labels, nil),
		grownDefects:          prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "grownDefects"), "Smartctl elements in grown defect list", labels, nil),
		spinUpTime:            prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "spinUpTime"), "Smartctl spin up time", labels, nil),
		startStopCount:        prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "startStopCount"), "Smartctl start stop count", labels, nil),
		seekErrorRate:         prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "seekErrorRate"), "Smartctl seek error rate", labels, nil),
		spinRetryCount:        prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "spinRetryCount"), "Smartctl spin retry count", labels, nil),
		airflowTemperature:    prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "airflowTemperature"), "Smartctl airflow temperature", labels, nil),
		temperatureCelsius:    prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "temperatureCelsius"), "Smartctl temperature celsius", labels, nil),
		loadCycleCount:        prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "loadCycleCount"), "Smartctl load cycle count", labels, nil),
		totalLBAsWritten:      prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "totalLBAsWritten"), "Smartctl total LBAs written", labels, nil),
		totalLBAsRead:         prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "totalLBAsRead"), "Smartctl total LBAs read", labels, nil),
	}
}

func (c *SmartctlDiskCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.rawReadErrorRate, c.reallocatedSectorCt, c.powerOnHours, c.powerCycleCount,
		c.runtimeBadBlock, c.endToEndError, c.reportedUncorrect, c.commandTimeout,
		c.hardwareECCRecovered, c.reallocatedEventCount, c.currentPendingSector,
		c.offlineUncorrectable, c.uDMACRCErrorCount, c.unusedRsvdBlkCntTot,
		c.grownDefects, c.spinUpTime, c.startStopCount, c.seekErrorRate,
		c.spinRetryCount, c.airflowTemperature, c.temperatureCelsius,
		c.loadCycleCount, c.totalLBAsWritten, c.totalLBAsRead,
	}
	for _, d := range ds {
		if d != nil {
			ch <- d
		}
	}
}

func (c *SmartctlDiskCollector) Collect(ch chan<- prometheus.Metric) {
	if _, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] smartctl failed for disk %s (idx %d): %v", c.diskID, c.diskN, err)
		return
	}
}

func (c *SmartctlDiskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if c.diskID == "" {
		return nil, nil
	}

	diskArg := fmt.Sprintf("cciss,%d", c.diskN)
	cmd := exec.Command("smartctl", "-iA", "-d", diskArg, c.devicePath)
	out, err := cmd.CombinedOutput()

	data := parser.ParseSmartctlDisk(string(out))
	if data == nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unable to parse smartctl output")
	}

	if len(data.SmartctlDiskDataInfo) == 0 || len(data.SmartctlDiskDataAttr) == 0 {
		return nil, fmt.Errorf("parsed smartctl data is empty")
	}

	info := data.SmartctlDiskDataInfo[0]
	attrs := data.SmartctlDiskDataAttr[0]

	labels := []string{c.diskID, info.Model, info.SN, info.RotRate, info.FromFact}

	sendMetric := func(desc *prometheus.Desc, val *float64) {
		if desc != nil && val != nil {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, *val, labels...)
		}
	}

	sendMetric(c.rawReadErrorRate, attrs.RawReadErrorRate)
	sendMetric(c.reallocatedSectorCt, attrs.ReallocatedSectorCt)
	sendMetric(c.powerOnHours, attrs.PowerOnHours)
	sendMetric(c.powerCycleCount, attrs.PowerCycleCount)
	sendMetric(c.runtimeBadBlock, attrs.RuntimeBadBlock)
	sendMetric(c.endToEndError, attrs.EndToEndError)
	sendMetric(c.reportedUncorrect, attrs.ReportedUncorrect)
	sendMetric(c.commandTimeout, attrs.CommandTimeout)
	sendMetric(c.hardwareECCRecovered, attrs.HardwareECCRecovered)
	sendMetric(c.reallocatedEventCount, attrs.ReallocatedEventCount)
	sendMetric(c.currentPendingSector, attrs.CurrentPendingSector)
	sendMetric(c.offlineUncorrectable, attrs.OfflineUncorrectable)
	sendMetric(c.uDMACRCErrorCount, attrs.UDMACRCErrorCount)
	sendMetric(c.unusedRsvdBlkCntTot, attrs.UnusedRsvdBlkCntTot)
	sendMetric(c.grownDefects, attrs.GrownDefects)
	sendMetric(c.spinUpTime, attrs.SpinUpTime)
	sendMetric(c.startStopCount, attrs.StartStopCount)
	sendMetric(c.seekErrorRate, attrs.SeekErrorRate)
	sendMetric(c.spinRetryCount, attrs.SpinRetryCount)
	sendMetric(c.airflowTemperature, attrs.AirflowTemperature)
	sendMetric(c.temperatureCelsius, attrs.TemperatureCelsius)
	sendMetric(c.loadCycleCount, attrs.LoadCycleCount)
	sendMetric(c.totalLBAsWritten, attrs.TotalLBAsWritten)
	sendMetric(c.totalLBAsRead, attrs.TotalLBAsRead)

	return nil, nil
}
