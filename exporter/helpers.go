package exporter

import (
	"os/exec"
	"regexp"
	"strings"
)

func getControllerSlots() ([]string, error) {
	out, err := exec.Command("ssacli", "ctrl", "all", "show", "status").CombinedOutput()
	if err != nil {
		return nil, err
	}

	var slots []string
	// Ищем строку вида "Smart Array P420 in Slot 2"
	re := regexp.MustCompile(`Slot\s+(\d+)`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	for _, match := range matches {
		if len(match) > 1 {
			slots = append(slots, match[1])
		}
	}
	return slots, nil
}

func getPhysicalDisksBulk(slotID string) (map[string]string, error) {
	out, err := exec.Command("ssacli", "ctrl", "slot="+slotID, "pd", "all", "show", "detail").CombinedOutput()
	if err != nil {
		return nil, err
	}

	pdMap := make(map[string]string)
	parts := strings.Split(string(out), "physicaldrive ")

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
		diskID := fields[0]
		pdMap[diskID] = "physicaldrive " + part
	}
	return pdMap, nil
}

func getLogicalDrivesBulk(slotID string) (map[string]string, error) {
	out, err := exec.Command("ssacli", "ctrl", "slot="+slotID, "ld", "all", "show", "detail").CombinedOutput()
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
