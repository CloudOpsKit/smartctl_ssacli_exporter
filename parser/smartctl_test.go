package parser

import (
	"testing"
)

func TestParseSmartctlDiskRealData(t *testing.T) {
	// Your actual output example
	rawOutput := `
=== START OF READ SMART DATA SECTION ===
ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
  1 Raw_Read_Error_Rate     0x010f   083   064   044    Pre-fail  Always       -       0/200164573
  5 Reallocated_Sector_Ct   0x0133   100   100   010    Pre-fail  Always       -       0
  9 Power_On_Hours          0x0032   093   093   000    Old_age   Always       -       6987
194 Temperature_Celsius     0x0022   026   042   000    Old_age   Always       -       26 (0 22 0 0 0)
`

	data := ParseSmartctlDisk(rawOutput)
	attrs := data.SmartctlDiskDataAttr[0]

	// 1. Validate complex format "0/200164573" -> should be 0
	if attrs.RawReadErrorRate == nil {
		t.Fatal("RawReadErrorRate is nil")
	}
	if *attrs.RawReadErrorRate != 0 {
		t.Errorf("RawReadErrorRate: expected 0, got %f", *attrs.RawReadErrorRate)
	}

	// 2. Validate standard zero
	if attrs.ReallocatedSectorCt == nil {
		t.Fatal("ReallocatedSectorCt is nil")
	}
	if *attrs.ReallocatedSectorCt != 0 {
		t.Errorf("ReallocatedSectorCt: expected 0, got %f", *attrs.ReallocatedSectorCt)
	}

	// 3. Validate value with suffix "26 (0 22...)" -> should be 26
	if attrs.TemperatureCelsius == nil {
		t.Fatal("TemperatureCelsius is nil")
	}
	if *attrs.TemperatureCelsius != 26 {
		t.Errorf("TemperatureCelsius: expected 26, got %f", *attrs.TemperatureCelsius)
	}

	// 4. Validate missing attribute (e.g., GrownDefects should be nil)
	if attrs.GrownDefects != nil {
		t.Errorf("GrownDefects should be nil, got %v", attrs.GrownDefects)
	}
}
