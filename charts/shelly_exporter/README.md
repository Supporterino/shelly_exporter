# shelly_exporter

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Steps to Use a Helm Chart

### 1. Add a Helm Repository

Helm repositories contain collections of charts. You can add an existing repository using the following command:

```bash
helm repo add shelly_exporter https://supporterino.github.io/shelly_exporter
```

### 2. Install the Helm Chart

To install a chart, use the following command:

```bash
helm install my-shelly_exporter shelly_exporter/shelly_exporter
```

### 3. View the Installation

You can check the status of the release using:

```bash
helm status my-shelly_exporter
```

## Customizing the Chart

Helm charts come with default values, but you can customize them by using the --set flag or by providing a custom values.yaml file.

### 1. Using --set to Override Values
```bash
helm install my-shelly_exporter shelly_exporter/shelly_exporter --set key1=value1,key2=value2
```

### 2. Using a values.yaml File
You can create a custom values.yaml file and pass it to the install command:

```bash
helm install my-shelly_exporter shelly_exporter/shelly_exporter -f values.yaml
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/supporterino/shelly_exporter"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` | This is for the secretes for pulling an image from a private repository more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| livenessProbe.httpGet.path | string | `"/"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| nameOverride | string | `""` | This is to override the chart name. |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| readinessProbe.httpGet.path | string | `"/"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| service.name | string | `"http"` |  |
| service.port | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceMonitor | object | `{"enabled":false,"interval":"30s","labels":{},"metricRelabelings":[],"namespace":"","path":"/metrics","relabelings":[],"scheme":"http","scrapeTimeout":"30s","targetLabels":[],"tlsConfig":{}}` | Configuration for the Prometheus ServiceMonitor |
| serviceMonitor.enabled | bool | `false` | Enable or disable the creation of a ServiceMonitor resource |
| serviceMonitor.interval | string | `"30s"` | Interval at which metrics should be scraped |
| serviceMonitor.labels | object | `{}` | Labels to add to the ServiceMonitor resource |
| serviceMonitor.metricRelabelings | list | `[]` | Relabeling rules for the metrics before ingestion |
| serviceMonitor.namespace | string | `""` | Namespace where the ServiceMonitor resource should be created. Defaults to Release.Namespace |
| serviceMonitor.path | string | `"/metrics"` | Path to scrape for metrics |
| serviceMonitor.relabelings | list | `[]` | Relabeling rules for the scraped metrics |
| serviceMonitor.scheme | string | `"http"` | Scheme to use for scraping metrics (http or https) |
| serviceMonitor.scrapeTimeout | string | `"30s"` | Timeout for scraping metrics |
| serviceMonitor.targetLabels | list | `[]` | Target labels to add to the scraped metrics |
| serviceMonitor.tlsConfig | object | `{}` | TLS configuration for scraping metrics |
| shelly_exporter.debug | bool | `false` | Enable or disable debug mode for the Shelly Exporter. |
| shelly_exporter.devices | list | `[{"host":"1.2.3.4"}]` | List of Shelly devices to monitor. |
| shelly_exporter.devices[0] | object | `{"host":"1.2.3.4"}` | IP address of the Shelly device. |
| shelly_exporter.updateInterval | int | `30` | Interval (in seconds) at which the exporter updates device data. |
| tolerations | list | `[]` |  |
| volumeMounts | list | `[]` | Additional volumeMounts on the output Deployment definition. |
| volumes | list | `[]` | Additional volumes on the output Deployment definition. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
