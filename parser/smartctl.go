package parser

import (
	"strings"
)

// SmartctlDisk data structure for output
type SmartctlDisk struct {
	SmartctlDiskDataInfo []SmartctlDiskDataInfo
	SmartctlDiskDataAttr []SmartctlDiskDataAttr
}

// SmartctlDiskDataInfo comment
type SmartctlDiskDataInfo struct {
	Model    string
	SN       string
	RotRate  string
	FromFact string
}

// SmartctlDiskDataAttr comment
type SmartctlDiskDataAttr struct {
	RawReadErrorRate      *float64
	ReallocatedSectorCt   *float64
	PowerOnHours          *float64
	PowerCycleCount       *float64
	RuntimeBadBlock       *float64
	EndToEndError         *float64
	ReportedUncorrect     *float64
	CommandTimeout        *float64
	HardwareECCRecovered  *float64
	ReallocatedEventCount *float64
	CurrentPendingSector  *float64
	OfflineUncorrectable  *float64
	UDMACRCErrorCount     *float64
	UnusedRsvdBlkCntTot   *float64
	GrownDefects          *float64
	SpinUpTime            *float64
	StartStopCount        *float64
	SeekErrorRate         *float64
	SpinRetryCount        *float64
	AirflowTemperature    *float64
	TemperatureCelsius    *float64
	LoadCycleCount        *float64
	TotalLBAsWritten      *float64
	TotalLBAsRead         *float64
}

// ParseSmartctlDisk return specific metric
func ParseSmartctlDisk(s string) *SmartctlDisk {

	dataAtr := SmartctlDiskDataAttr{}
	dataInfo := SmartctlDiskDataInfo{}
	for _, section := range strings.Split(s, "=== START OF ") {
		if strings.Contains(section, "INFORMATION SECTION ===") {
			dataInfo = parseSmartctlDiskInfo(section)
		} else if strings.Contains(section, "READ SMART DATA SECTION ===") {
			dataAtr = parseSmartctlDiskAtr(section)
		}
	}

	data := SmartctlDisk{
		SmartctlDiskDataAttr: []SmartctlDiskDataAttr{
			dataAtr,
		},
		SmartctlDiskDataInfo: []SmartctlDiskDataInfo{
			dataInfo,
		},
	}

	return &data
}

func parseSmartctlDiskInfo(s string) SmartctlDiskDataInfo {

	var (
		tmp SmartctlDiskDataInfo
	)

	for _, line := range strings.Split(s, "\n") {
		kvs := strings.Trim(line, " \t")
		kv := strings.Split(kvs, ": ")

		if len(kv) == 2 {
			switch kv[0] {
			case "Device Model":
				tmp.Model = trim(kv[1])
			case "Serial Number":
				tmp.SN = trim(kv[1])
			case "Rotation Rate":
				tmp.RotRate = trim(kv[1])
			case "Form Factor":
				tmp.FromFact = trim(kv[1])
			}
		}
	}

	return tmp
}

func parseSmartctlDiskAtr(s string) SmartctlDiskDataAttr {
	var tmp SmartctlDiskDataAttr

	lines := strings.Split(s, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		vals := strings.Fields(line)
		// If the line looks like this:
		// ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
		// 1 Raw_Read_Error_Rate     0x000f   092   092   006    Pre-fail  Always       -       0/104646032
		if len(vals) >= 10 {
			// Attribute name is in the second field (index 1)
			attrName := vals[1]

			// Raw value is in the tenth field (index 9).
			// Even if the value is "26 (Min/Max...)", Fields splits this into ["26", "(Min/Max..."].
			// Therefore, vals[9] will contain only the number or "number/number".
			rawValPtr := parseSmartRawValue(vals[9])

			switch attrName {
			case "Raw_Read_Error_Rate":
				tmp.RawReadErrorRate = rawValPtr
			case "Reallocated_Sector_Ct":
				tmp.ReallocatedSectorCt = rawValPtr
			case "Power_On_Hours":
				tmp.PowerOnHours = rawValPtr
			case "Power_Cycle_Count":
				tmp.PowerCycleCount = rawValPtr
			case "Runtime_Bad_Block":
				tmp.RuntimeBadBlock = rawValPtr
			case "End-to-End_Error":
				tmp.EndToEndError = rawValPtr
			case "Reported_Uncorrect":
				tmp.ReportedUncorrect = rawValPtr
			case "Command_Timeout":
				tmp.CommandTimeout = rawValPtr
			case "Hardware_ECC_Recovered":
				tmp.HardwareECCRecovered = rawValPtr
			case "Reallocated_Event_Count":
				tmp.ReallocatedEventCount = rawValPtr
			case "Current_Pending_Sector":
				tmp.CurrentPendingSector = rawValPtr
			case "Offline_Uncorrectable":
				tmp.OfflineUncorrectable = rawValPtr
			case "UDMA_CRC_Error_Count":
				tmp.UDMACRCErrorCount = rawValPtr
			case "Unused_Rsvd_Blk_Cnt_Tot":
				tmp.UnusedRsvdBlkCntTot = rawValPtr
			case "Spin_Up_Time":
				tmp.SpinUpTime = rawValPtr
			case "Start_Stop_Count":
				tmp.StartStopCount = rawValPtr
			case "Seek_Error_Rate":
				tmp.SeekErrorRate = rawValPtr
			case "Spin_Retry_Count":
				tmp.SpinRetryCount = rawValPtr
			case "Airflow_Temperature_Cel":
				tmp.AirflowTemperature = rawValPtr
			case "Temperature_Celsius":
				tmp.TemperatureCelsius = rawValPtr
			case "Load_Cycle_Count":
				tmp.LoadCycleCount = rawValPtr
			case "Total_LBAs_Written":
				tmp.TotalLBAsWritten = rawValPtr
			case "Total_LBAs_Read":
				tmp.TotalLBAsRead = rawValPtr
			}

		} else {
			// Handle special lines, e.g.: "Elements in grown defect list: 71"
			// The separator here is ": "
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				switch parts[0] {
				case "Elements in grown defect list":
					tmp.GrownDefects = parseSmartRawValue(parts[1])
				}
			}
		}
	}

	return tmp
}
