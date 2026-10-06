package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

const fakeSsacli = `#!/bin/sh
echo "$*" >> "$(dirname "$0")/calls.log"
case "$*" in
  "ctrl all show detail") printf 'Smart Array P420i in Slot 0 (Embedded)\n   Slot: 0\n   Controller Status: OK\n' ;;
  *"pd all show detail") printf 'Smart Array P420i in Slot 0\n\n   physicaldrive 1I:1:1\n      Status: OK\n\n   physicaldrive 1I:1:2\n      Status: OK\n' ;;
  *"ld all show detail") printf 'Smart Array P420i in Slot 0\n\n   Logical Drive: 1\n      Status: OK\n' ;;
esac
`

// installFakeTools puts fake ssacli and smartctl first in PATH and returns
// a function reporting how many times ssacli was called.
func installFakeTools(t *testing.T, ssacli string) func() int {
	t.Helper()
	dir := t.TempDir()
	for name, script := range map[string]string{
		"ssacli":   ssacli,
		"smartctl": "#!/bin/sh\nexit 1\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return func() int {
		data, err := os.ReadFile(filepath.Join(dir, "calls.log"))
		if err != nil {
			return 0
		}
		return strings.Count(string(data), "\n")
	}
}

func gather(t *testing.T, e *Exporter) map[string][]*dto.Metric {
	t.Helper()
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(e)
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	out := make(map[string][]*dto.Metric)
	for _, f := range families {
		out[f.GetName()] = f.GetMetric()
	}
	return out
}

func TestCollectServesCachedMetrics(t *testing.T) {
	ssacliCalls := installFakeTools(t, fakeSsacli)
	e := New("/dev/sda")

	if got := gather(t, e); len(got) != 0 {
		t.Errorf("expected no metrics before first collection, got %d families", len(got))
	}

	e.refresh()
	calls := ssacliCalls()
	if calls != 3 {
		t.Errorf("refresh made %d ssacli calls, want 3", calls)
	}

	for range 3 {
		metrics := gather(t, e)
		if n := len(metrics["ssacli_phys_disk_status"]); n != 2 {
			t.Errorf("got %d phys disk metrics, want 2", n)
		}
		if v := metrics["smartctl_ssacli_exporter_last_collect_success"][0].GetGauge().GetValue(); v != 1 {
			t.Errorf("last_collect_success = %v, want 1", v)
		}
	}

	if got := ssacliCalls(); got != calls {
		t.Errorf("scrapes ran ssacli %d times, want 0", got-calls)
	}
}

func TestCollectReportsFailure(t *testing.T) {
	installFakeTools(t, "#!/bin/sh\nexit 1\n")
	e := New("/dev/sda")

	e.refresh()
	metrics := gather(t, e)

	if v := metrics["smartctl_ssacli_exporter_last_collect_success"][0].GetGauge().GetValue(); v != 0 {
		t.Errorf("last_collect_success = %v, want 0", v)
	}
	if _, ok := metrics["ssacli_hw_raid_controller_slot"]; ok {
		t.Error("controller metrics should not be exported when ssacli fails")
	}
}
