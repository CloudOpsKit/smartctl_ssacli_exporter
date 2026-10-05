package exporter

import (
	"github.com/CloudOpsKit/smartctl_ssacli_exporter/command"
	"regexp"
	"strings"
)

var controllerSlotRe = regexp.MustCompile(`in Slot\s+(\d+)`)

// parseControllerSlots extracts slot numbers from "ssacli ctrl all show detail"
// output, where each controller starts with a header like "Smart Array P420 in Slot 2".
func parseControllerSlots(out string) []string {
	var slots []string
	seen := make(map[string]bool)
	for _, match := range controllerSlotRe.FindAllStringSubmatch(out, -1) {
		if slot := match[1]; !seen[slot] {
			seen[slot] = true
			slots = append(slots, slot)
		}
	}
	return slots
}

// physicalDisk holds the raw "show detail" block of a single physical drive.
type physicalDisk struct {
	ID   string
	Data string
}

// getPhysicalDisksBulk returns physical drives in the order ssacli reports them.
// The order matters: a drive's position is its smartctl "-d cciss,N" index.
func getPhysicalDisksBulk(slotID string) ([]physicalDisk, error) {
	out, err := command.Run("ssacli", "ctrl", "slot="+slotID, "pd", "all", "show", "detail")
	if err != nil {
		return nil, err
	}
	return splitPhysicalDisks(string(out)), nil
}

func splitPhysicalDisks(out string) []physicalDisk {
	var disks []physicalDisk
	parts := strings.Split(out, "physicaldrive ")

	for i, part := range parts {
		if i == 0 {
			continue
		}
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) < 1 {
			continue
		}
		disks = append(disks, physicalDisk{ID: fields[0], Data: "physicaldrive " + part})
	}
	return disks
}

func getLogicalDrivesBulk(slotID string) (map[string]string, error) {
	out, err := command.Run("ssacli", "ctrl", "slot="+slotID, "ld", "all", "show", "detail")
	if err != nil {
		return nil, err
	}

	ldMap := make(map[string]string)
	parts := strings.Split(string(out), "Logical Drive: ")

	for i, part := range parts {
		if i == 0 {
			continue
		}
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) < 1 {
			continue
		}
		ldID := fields[0]
		ldMap[ldID] = "Logical Drive: " + part
	}
	return ldMap, nil
}
