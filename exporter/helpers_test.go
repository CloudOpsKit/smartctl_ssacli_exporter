package exporter

import (
	"strings"
	"testing"
)

func TestSplitPhysicalDisksKeepsOrder(t *testing.T) {
	out := `
Smart Array P420i in Slot 0 (Embedded)

   Array A

      physicaldrive 1I:1:1
         Port: 1I
         Box: 1
         Bay: 1
         Status: OK

      physicaldrive 1I:1:2
         Port: 1I
         Box: 1
         Bay: 2
         Status: OK

   Array B

      physicaldrive 2I:1:5
         Port: 2I
         Box: 1
         Bay: 5
         Status: Failed

      physicaldrive 2I:1:6
         Port: 2I
         Box: 1
         Bay: 6
         Status: OK
`
	want := []string{"1I:1:1", "1I:1:2", "2I:1:5", "2I:1:6"}

	// Repeat to catch nondeterministic ordering (e.g. map iteration)
	for range 20 {
		disks := splitPhysicalDisks(out)
		if len(disks) != len(want) {
			t.Fatalf("got %d disks, want %d", len(disks), len(want))
		}
		for i, d := range disks {
			if d.ID != want[i] {
				t.Fatalf("disk %d: got ID %q, want %q", i, d.ID, want[i])
			}
		}
	}

	if got := splitPhysicalDisks(out)[2].Data; !strings.HasPrefix(got, "physicaldrive 2I:1:5") {
		t.Errorf("disk data should start with its header, got %q", got)
	}
}
