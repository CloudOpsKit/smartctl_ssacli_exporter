package collector

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/CloudOpsKit/smartctl_ssacli_exporter/parser"
	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &SmartctlDiskCollector{}

// SmartctlDiskCollector Contain raid controller detail information
type SmartctlDiskCollector struct {
	diskID string
	diskN  int

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

// NewSmartctlDiskCollector Create new collector
func NewSmartctlDiskCollector(diskID string, diskN int) *SmartctlDiskCollector {
	// Init labels
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

	// Rerutn Colected metric to ch <-
	// Include labels
	return &SmartctlDiskCollector{
		diskID: diskID,
		diskN:  diskN,
		rawReadErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "rawReadErrorRate"),
			"Smartctl raw read error rate",
			labels,
			nil,
		),
		reallocatedSectorCt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "reallocatedSectorCt"),
			"Smartctl reallocated sector ct",
			labels,
			nil,
		),
		powerOnHours: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "powerOnHours"),
			"Smartctl power on hours",
			labels,
			nil,
		),
		powerCycleCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "powerCycleCount"),
			"Smartctl power cycle down count",
			labels,
			nil,
		),
		runtimeBadBlock: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "runtimeBadBlock"),
			"Smartctl runtime bad block",
			labels,
			nil,
		),
		endToEndError: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "endToEndError"),
			"Smartctl end to end error",
			labels,
			nil,
		),
		reportedUncorrect: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "reportedUncorrect"),
			"Smartctl reported uncorrect",
			labels,
			nil,
		),
		commandTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "commandTimeout"),
			"Smartctl command timeout",
			labels,
			nil,
		),
		hardwareECCRecovered: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "hardwareECCRecovered"),
			"Smartctl hardware ecc recovered",
			labels,
			nil,
		),
		reallocatedEventCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "reallocatedEventCount"),
			"Smartctl reallocated event count",
			labels,
			nil,
		),
		currentPendingSector: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "currentPendingSector"),
			"Smartctl current pending sector",
			labels,
			nil,
		),
		offlineUncorrectable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "offlineUncorrectable"),
			"Smartctl offline uncorrectable",
			labels,
			nil,
		),
		uDMACRCErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "uDMACRCErrorCount"),
			"Smartctl ud macrc error count",
			labels,
			nil,
		),
		unusedRsvdBlkCntTot: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "unusedRsvdBlkCntTot"),
			"Smartctl unused rsvd block Count Total",
			labels,
			nil,
		),
		grownDefects: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "grownDefects"),
			"Smartctl elements in grown defect list",
			labels,
			nil,
		),
		spinUpTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "spinUpTime"),
			"Smartctl spin up time",
			labels,
			nil,
		),
		startStopCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "startStopCount"),
			"Smartctl start stop count",
			labels,
			nil,
		),
		seekErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "seekErrorRate"),
			"Smartctl seek error rate",
			labels,
			nil,
		),
		spinRetryCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "spinRetryCount"),
			"Smartctl spin retry count",
			labels,
			nil,
		),
		airflowTemperature: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "airflowTemperature"),
			"Smartctl airflow temperature",
			labels,
			nil,
		),
		temperatureCelsius: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "temperatureCelsius"),
			"Smartctl temperature celsius",
			labels,
			nil,
		),
		loadCycleCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "loadCycleCount"),
			"Smartctl load cycle count",
			labels,
			nil,
		),
		totalLBAsWritten: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "totalLBAsWritten"),
			"Smartctl total LBAs written",
			labels,
			nil,
		),
		totalLBAsRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "totalLBAsRead"),
			"Smartctl total LBAs read",
			labels,
			nil,
		),
	}
}

// Describe return all description to chanel
func (c *SmartctlDiskCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.rawReadErrorRate,
		c.reallocatedSectorCt,
		c.powerOnHours,
		c.powerCycleCount,
		c.runtimeBadBlock,
		c.endToEndError,
		c.reportedUncorrect,
		c.commandTimeout,
		c.hardwareECCRecovered,
		c.reallocatedEventCount,
		c.currentPendingSector,
		c.offlineUncorrectable,
		c.uDMACRCErrorCount,
		c.unusedRsvdBlkCntTot,
		c.grownDefects,
		c.spinUpTime,
		c.startStopCount,
		c.seekErrorRate,
		c.spinRetryCount,
		c.airflowTemperature,
		c.temperatureCelsius,
		c.loadCycleCount,
		c.totalLBAsWritten,
		c.totalLBAsRead,
	}
	for _, d := range ds {
		ch <- d
	}
}

// Collect create collector
// Get metric
// Handle error
func (c *SmartctlDiskCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		//log.Debugln("[ERROR] failed collecting metric %v: %v", desc, err)
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}
}

func (c *SmartctlDiskCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if c.diskID == "" {
		return nil, nil
	}

	cmd := "smartctl -iA -d cciss," + strconv.Itoa(c.diskN) + " /dev/sda | grep ."
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()

	if err != nil {
		//log.Debugln("[ERROR] smart log: \n%s\n", out)
		return nil, err
	}

	data := parser.ParseSmartctlDisk(string(out))

	if data == nil {
		log.Printf("[FATAL] Unable get data from smartctl exporter")
		return nil, nil
	}

	// Safety check: verify that arrays are not empty to avoid "index out of range" panic
	if len(data.SmartctlDiskDataInfo) == 0 || len(data.SmartctlDiskDataAttr) == 0 {
		return nil, fmt.Errorf("parsed data is empty")
	}

	info := data.SmartctlDiskDataInfo[0]
	attrs := data.SmartctlDiskDataAttr[0]

	var (
		labels = []string{
			c.diskID,
			info.Model,
			info.SN,
			info.RotRate,
			info.FromFact,
		}
	)

	// Check each attribute pointer before sending.
	// If the pointer is not nil, the metric exists on the disk -> send it.
	// If the pointer is nil, the metric is missing -> skip it (do not send 0).

	if attrs.RawReadErrorRate != nil {
		ch <- prometheus.MustNewConstMetric(
			c.rawReadErrorRate,
			prometheus.GaugeValue,
			*attrs.RawReadErrorRate, // Dereference the pointer (*) to get float64
			labels...,
		)
	}
	if attrs.ReallocatedSectorCt != nil {
		ch <- prometheus.MustNewConstMetric(
			c.reallocatedSectorCt,
			prometheus.GaugeValue,
			*attrs.ReallocatedSectorCt,
			labels...,
		)
	}
	if attrs.PowerOnHours != nil {
		ch <- prometheus.MustNewConstMetric(
			c.powerOnHours,
			prometheus.GaugeValue,
			*attrs.PowerOnHours,
			labels...,
		)
	}
	if attrs.PowerCycleCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.powerCycleCount,
			prometheus.GaugeValue,
			*attrs.PowerCycleCount,
			labels...,
		)
	}
	if attrs.RuntimeBadBlock != nil {
		ch <- prometheus.MustNewConstMetric(
			c.runtimeBadBlock,
			prometheus.GaugeValue,
			*attrs.RuntimeBadBlock,
			labels...,
		)
	}
	if attrs.EndToEndError != nil {
		ch <- prometheus.MustNewConstMetric(
			c.endToEndError,
			prometheus.GaugeValue,
			*attrs.EndToEndError,
			labels...,
		)
	}
	if attrs.ReportedUncorrect != nil {
		ch <- prometheus.MustNewConstMetric(
			c.reportedUncorrect,
			prometheus.GaugeValue,
			*attrs.ReportedUncorrect,
			labels...,
		)
	}
	if attrs.CommandTimeout != nil {
		ch <- prometheus.MustNewConstMetric(
			c.commandTimeout,
			prometheus.GaugeValue,
			*attrs.CommandTimeout,
			labels...,
		)
	}
	if attrs.HardwareECCRecovered != nil {
		ch <- prometheus.MustNewConstMetric(
			c.hardwareECCRecovered,
			prometheus.GaugeValue,
			*attrs.HardwareECCRecovered,
			labels...,
		)
	}
	if attrs.ReallocatedEventCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.reallocatedEventCount,
			prometheus.GaugeValue,
			*attrs.ReallocatedEventCount,
			labels...,
		)
	}
	if attrs.CurrentPendingSector != nil {
		ch <- prometheus.MustNewConstMetric(
			c.currentPendingSector,
			prometheus.GaugeValue,
			*attrs.CurrentPendingSector,
			labels...,
		)
	}
	if attrs.OfflineUncorrectable != nil {
		ch <- prometheus.MustNewConstMetric(
			c.offlineUncorrectable,
			prometheus.GaugeValue,
			*attrs.OfflineUncorrectable,
			labels...,
		)
	}
	if attrs.UDMACRCErrorCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.uDMACRCErrorCount,
			prometheus.GaugeValue,
			*attrs.UDMACRCErrorCount,
			labels...,
		)
	}
	if attrs.UnusedRsvdBlkCntTot != nil {
		ch <- prometheus.MustNewConstMetric(
			c.unusedRsvdBlkCntTot,
			prometheus.GaugeValue,
			*attrs.UnusedRsvdBlkCntTot,
			labels...,
		)
	}
	if attrs.GrownDefects != nil {
		ch <- prometheus.MustNewConstMetric(
			c.grownDefects,
			prometheus.GaugeValue,
			*attrs.GrownDefects,
			labels...,
		)
	}

	if attrs.SpinUpTime != nil {
		ch <- prometheus.MustNewConstMetric(
			c.spinUpTime,
			prometheus.GaugeValue,
			*attrs.SpinUpTime,
			labels...,
		)
	}
	if attrs.StartStopCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.startStopCount,
			prometheus.GaugeValue,
			*attrs.StartStopCount,
			labels...,
		)
	}
	if attrs.SeekErrorRate != nil {
		ch <- prometheus.MustNewConstMetric(
			c.seekErrorRate,
			prometheus.GaugeValue,
			*attrs.SeekErrorRate,
			labels...,
		)
	}
	if attrs.SpinRetryCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.spinRetryCount,
			prometheus.GaugeValue,
			*attrs.SpinRetryCount,
			labels...,
		)
	}
	if attrs.AirflowTemperature != nil {
		ch <- prometheus.MustNewConstMetric(
			c.airflowTemperature,
			prometheus.GaugeValue,
			*attrs.AirflowTemperature,
			labels...,
		)
	}
	if attrs.TemperatureCelsius != nil {
		ch <- prometheus.MustNewConstMetric(
			c.temperatureCelsius,
			prometheus.GaugeValue,
			*attrs.TemperatureCelsius,
			labels...,
		)
	}
	if attrs.LoadCycleCount != nil {
		ch <- prometheus.MustNewConstMetric(
			c.loadCycleCount,
			prometheus.GaugeValue,
			*attrs.LoadCycleCount,
			labels...,
		)
	}
	if attrs.TotalLBAsWritten != nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalLBAsWritten,
			prometheus.GaugeValue,
			*attrs.TotalLBAsWritten,
			labels...,
		)
	}
	if attrs.TotalLBAsRead != nil {
		ch <- prometheus.MustNewConstMetric(
			c.totalLBAsRead,
			prometheus.GaugeValue,
			*attrs.TotalLBAsRead,
			labels...,
		)
	}

	return nil, nil
}
