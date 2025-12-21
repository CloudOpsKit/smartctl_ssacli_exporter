package parser

import (
	"strings"
)

type SsacliLogDisk struct {
	SsacliLogDiskData []SsacliLogDiskData
}

type SsacliLogDiskData struct {
	ID             string
	Size           string
	Cylinders      float64
	Status         string
	Caching        string
	UID            string
	LName          string
	LID            string
	FaultTolerance string
	UME            string
}

func ParseSsacliLogDisk(s string) *SsacliLogDisk {
	var (
		data []SsacliLogDiskData
		tmp  SsacliLogDiskData
	)

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Logical Drive:") {
			if tmp.ID != "" {
				data = append(data, tmp)
			}
			tmp = SsacliLogDiskData{}
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				tmp.ID = parts[1]
			}
			continue
		}

		kv := strings.SplitN(line, ": ", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])

			switch key {
			case "Size":
				tmp.Size = val
			case "Cylinders":
				tmp.Cylinders = toFLO(val)
			case "Status":
				tmp.Status = val
			case "Caching":
				tmp.Caching = val
			case "Unique Identifier":
				tmp.UID = val
			case "Disk Name":
				tmp.LName = val
			case "Logical Drive Label":
				tmp.LID = val
			case "Fault Tolerance":
				tmp.FaultTolerance = val
			case "Unrecoverable Media Errors":
				tmp.UME = val
			}
		}

		if i == len(lines)-1 && tmp.ID != "" {
			data = append(data, tmp)
		}
	}

	if len(data) == 0 && tmp.Status != "" {
		data = append(data, tmp)
	}

	return &SsacliLogDisk{SsacliLogDiskData: data}
}
