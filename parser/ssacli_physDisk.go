package parser

import (
	"regexp"
	"strings"
)

// SsacliPhysDisk data structure for output
type SsacliPhysDisk struct {
	SsacliPhysDiskData []SsacliPhysDiskData
}

// SsacliPhysDiskData data structure for output
type SsacliPhysDiskData struct {
	ID        string
	Bay       string
	Status    string
	DriveType string
	IntType   string
	Size      string
	BlockSize string
	Speed     string
	Firmware  string
	SN        string
	WWID      string
	CurTemp   float64
	MaxTemp   float64
	Model     string
}

// ParseSsacliPhysDisk return specific metric
func ParseSsacliPhysDisk(s string) *SsacliPhysDisk {
	return parseSsacliPhysDisk(s)
}

func parseSsacliPhysDisk(s string) *SsacliPhysDisk {
	var (
		disks []SsacliPhysDiskData
		tmp   SsacliPhysDiskData
	)

	re := regexp.MustCompile(`(.+?)\: (.+)`)
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		kvs := strings.Trim(line, " \t\r")

		if strings.HasPrefix(kvs, "physicaldrive ") {
			if tmp.ID != "" {
				disks = append(disks, tmp)
			}
			tmp = SsacliPhysDiskData{}
			parts := strings.Split(kvs, " ")
			if len(parts) > 1 {
				tmp.ID = parts[1]
			}
			continue
		}

		kv := re.FindStringSubmatch(kvs)
		if len(kv) == 3 {
			key := kv[1]
			value := kv[2]
			switch key {
			case "Bay":
				tmp.Bay = value
			case "Serial Number":
				tmp.SN = value
			case "Status":
				tmp.Status = value
			case "Drive Type":
				tmp.DriveType = value
			case "Interface Type":
				tmp.IntType = value
			case "Size":
				tmp.Size = value
			case "Logical/Physical Block Size":
				tmp.BlockSize = value
			case "Rotational Speed":
				tmp.Speed = value
			case "Firmware Revision":
				tmp.Firmware = value
			case "WWID":
				tmp.WWID = value
			case "Model":
				tmp.Model = value
			case "Current Temperature (C)":
				tmp.CurTemp = toFLO(value)
			case "Maximum Temperature (C)":
				tmp.MaxTemp = toFLO(value)
			}
		}

		if i == len(lines)-1 && tmp.ID != "" {
			disks = append(disks, tmp)
		}
	}

	if len(disks) == 0 && tmp.Status != "" {
		disks = append(disks, tmp)
	}

	return &SsacliPhysDisk{SsacliPhysDiskData: disks}
}
