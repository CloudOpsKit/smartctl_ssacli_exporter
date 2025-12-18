# smartctl_ssacli_exporter
Export metric from HP enterprise raid card &amp; disk smartctl with auto detect disk

| Flag name   | Default Value | Desc                                                           |
|-------------|---------------|----------------------------------------------------------------|
| listen      |:9633          | Exporter listener port && address                              |
| metricsPath |/metrics       | URL path for surfacing collected metrics                       |
| devicePath  |/dev/sda       | Path to the raid controller device (e.g. /dev/sda or /dev/sg0) |

## Usage

``` bash
./smartctl_ssacli_exporter
```

## Install

### Build from source
``` Bash
git clone https://github.com/CloudOpsKit/smartctl_ssacli_exporter.git
go get
go build
```

## Dashboard
Grafana ID: 12587
https://grafana.com/grafana/dashboards/12587
