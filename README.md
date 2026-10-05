# smartctl_ssacli_exporter
Export metric from HP enterprise raid card &amp; disk smartctl with auto detect disk

| Flag name   | Default Value | Desc                                                           |
|-------------|---------------|----------------------------------------------------------------|
| listen      |:9633          | Exporter listener port && address                              |
| metricsPath |/metrics       | URL path for surfacing collected metrics                       |
| devicePath  |/dev/sda       | Path to the raid controller device (e.g. /dev/sda or /dev/sg0) |
| timeout     |30s            | Timeout for each ssacli/smartctl call                          |

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

### Container image
Images are built with [ko](https://ko.build) and published to `ghcr.io/cloudopskit/smartctl_ssacli_exporter` (linux/amd64 only, since `ssacli` is amd64-only).
The exporter needs access to the RAID controller, so run it privileged:
``` Bash
docker run -d --privileged -p 9633:9633 ghcr.io/cloudopskit/smartctl_ssacli_exporter:latest
```

The runtime base image with `smartctl` and `ssacli` is defined in `Dockerfile.base` and published as `ghcr.io/cloudopskit/smartctl_ssacli_exporter/base:latest`.

To build locally:
``` Bash
KO_DOCKER_REPO=ghcr.io/cloudopskit/smartctl_ssacli_exporter ko build . --bare --push=false --tarball=image.tar
```

## Dashboard
Grafana ID: 12587
https://grafana.com/grafana/dashboards/12587
