# Graph Report - shelly_exporter  (2026-09-26)

## Corpus Check
- Corpus is ~36,352 words - fits in a single context window. You may not need a graph.

## Summary
- 185 nodes · 303 edges · 16 communities (12 shown, 4 thin omitted)
- Extraction: 91% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 25 edges (avg confidence: 0.91)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Exporter Core & Config
- Helm Chart & Deployment
- Metrics Registration
- OpenSpec Workflow
- Shelly API Client
- Device Manager & RPC
- Shelly Config Response
- Switch Config Metrics
- Switch Status Metrics
- Release Automation
- Shelly Status Response
- graphify Plugin Runtime
- OpenCode Plugin Config
- Config JSON Unmarshal
- Cover Config Response
- Go Module

## God Nodes (most connected - your core abstractions)
1. `APIClient` - 12 edges
2. `ShellyGetConfigResponse` - 11 edges
3. `Register()` - 11 edges
4. `fetchAndUpdateMetrics()` - 11 edges
5. `opsx-apply Command` - 9 edges
6. `opsx-archive Command` - 8 edges
7. `opsx-continue Command` - 8 edges
8. `opsx-onboard Command` - 8 edges
9. `openspec-onboard Skill` - 8 edges
10. `ShellyGetStatusResponse` - 7 edges

## Surprising Connections (you probably didn't know these)
- `config.yaml Configuration` --semantically_similar_to--> `OpenSpec spec-driven Schema`  [AMBIGUOUS] [semantically similar]
  README.md → openspec/config.yaml
- `config.yaml Configuration` --semantically_similar_to--> `k8s-config.yaml Runtime Config`  [INFERRED] [semantically similar]
  README.md → charts/shelly-exporter/templates/configmap.yaml
- `Register()` --calls--> `RegisterShellyGetDeviceInfoMetrics()`  [EXTRACTED]
  metrics/metrics.go → rpc/Shelly.GetDeviceInfo/ShellyGetDeviceInfoMetrics.go
- `Register()` --calls--> `RegisterSwitchGetConfigMetrics()`  [EXTRACTED]
  metrics/metrics.go → rpc/Switch.GetConfig/SwitchGetConfigMetrics.go
- `Register()` --calls--> `RegisterSwitchGetStatusMetrics()`  [EXTRACTED]
  metrics/metrics.go → rpc/Switch.GetStatus/SwitchGetStatusMetrics.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **OpenSpec Workflow Command/Skill Family** — _opencode_commands_opsx_apply_command, _opencode_commands_opsx_archive_command, _opencode_commands_opsx_continue_command, _opencode_commands_opsx_explore_command, _opencode_commands_opsx_new_command, _opencode_commands_opsx_onboard_command, _opencode_commands_opsx_propose_command, _opencode_commands_opsx_sync_command, _opencode_commands_opsx_update_command, _opencode_commands_opsx_verify_command, _opencode_skills_openspec_apply_change_skill_skill, _opencode_skills_openspec_archive_change_skill_skill, _opencode_skills_openspec_continue_change_skill_skill, _opencode_skills_openspec_explore_skill_skill, _opencode_skills_openspec_new_change_skill_skill, _opencode_skills_openspec_onboard_skill_skill, _opencode_skills_openspec_propose_skill_skill, _opencode_skills_openspec_sync_specs_skill_skill, _opencode_skills_openspec_update_change_skill_skill, _opencode_skills_openspec_verify_change_skill_skill [EXTRACTED 1.00]
- **Tag-Triggered Release Automation Pipeline** — _github_workflows_chart_release_workflow, _github_workflows_docker_release_workflow, _github_workflows_release_workflow, _github_cr_config, _goreleaser_config [INFERRED 0.75]
- **OpenSpec Archive and Spec-Sync Flow** — _opencode_commands_opsx_archive_command, _opencode_commands_opsx_sync_command, _opencode_skills_openspec_archive_change_skill_skill, _opencode_skills_openspec_sync_specs_skill_skill, _opencode_skills_openspec_verify_change_skill_skill [INFERRED 0.85]
- **Helm Chart Kubernetes Templates** — charts_shelly_exporter_templates_deployment_deployment, charts_shelly_exporter_templates_configmap_configmap, charts_shelly_exporter_templates_service_service, charts_shelly_exporter_templates_servicemonitor_servicemonitor [INFERRED 0.95]
- **Shelly Exporter Deployment Methods** — readme_standalone_binary, readme_docker_image, readme_helm_chart [EXTRACTED 1.00]

## Communities (16 total, 4 thin omitted)

### Community 0 - "Exporter Core & Config"
Cohesion: 0.10
Nodes (24): healthHandler(), main(), serverIsHealthy(), YamlConfig, NewConfig(), ParseFlags(), ValidateConfigPath(), DeviceYamlConfig (+16 more)

### Community 1 - "Helm Chart & Deployment"
Cohesion: 0.10
Nodes (25): shelly-exporter Helm Chart, Custom Values Override, Helm Install Workflow, ConfigMap Template, k8s-config.yaml Runtime Config, Deployment Template, Service Template, ServiceMonitor Template (+17 more)

### Community 2 - "Metrics Registration"
Cohesion: 0.12
Nodes (18): CoverGetStatusResponse, CoverGetStatusMetrics, go_pkg_github_com_prometheus_client_golang_prometheus, go_pkg_github_com_supporterino_shelly_exporter_client, github.com/prometheus/client_golang/prometheus.GaugeVec, Register(), boolToFloat64(), RegisterCoverGetStatusMetrics() (+10 more)

### Community 3 - "OpenSpec Workflow"
Cohesion: 0.19
Nodes (23): opsx-apply Command, opsx-archive Command, opsx-continue Command, opsx-explore Command, opsx-new Command, opsx-onboard Command, opsx-propose Command, opsx-sync Command (+15 more)

### Community 4 - "Shelly API Client"
Cohesion: 0.14
Nodes (15): APIClient, NewAPIClient(), ShellyGetDeviceInfoResponse, WiFiGetStatusResponse, net/http.Client, time.Duration, DeviceConfig, fetchAndUpdateMetrics() (+7 more)

### Community 5 - "Device Manager & RPC"
Cohesion: 0.16
Nodes (14): go_pkg_context, go_pkg_github_com_supporterino_shelly_exporter_rpc, go_pkg_github_com_supporterino_shelly_exporter_rpc_cover_getstatus, go_pkg_github_com_supporterino_shelly_exporter_rpc_shelly_getconfig, go_pkg_github_com_supporterino_shelly_exporter_rpc_shelly_getdeviceinfo, go_pkg_github_com_supporterino_shelly_exporter_rpc_shelly_getstatus, go_pkg_github_com_supporterino_shelly_exporter_rpc_switch_getconfig, go_pkg_github_com_supporterino_shelly_exporter_rpc_switch_getstatus (+6 more)

### Community 6 - "Shelly Config Response"
Cohesion: 0.22
Nodes (9): BLE, Cloud, Eth, MQTT, ShellyGetConfigResponse, ShellyGetConfigResponseInput, ShellyGetConfigResponseSwitch, ShellyGetConfigResponseSys (+1 more)

### Community 7 - "Switch Config Metrics"
Cohesion: 0.38
Nodes (5): SwitchGetConfigResponse, boolToFloat64(), RegisterSwitchGetConfigMetrics(), UpdateSwitchGetConfigMetrics(), SwitchGetConfig

### Community 8 - "Switch Status Metrics"
Cohesion: 0.38
Nodes (5): SwitchGetStatusResponse, boolToFloat64(), RegisterSwitchGetStatusMetrics(), UpdateSwitchGetStatusMetrics(), SwitchGetStatusMetrics

### Community 9 - "Release Automation"
Cohesion: 0.40
Nodes (5): Chart Releaser Config (cr.yaml), Chart Release Workflow, Docker Release Workflow, Go Release Workflow, GoReleaser Configuration

### Community 10 - "Shelly Status Response"
Cohesion: 0.40
Nodes (5): Energy, ShellyGetStatusResponse, Sys, Temperature, Wifi

### Community 11 - "graphify Plugin Runtime"
Cohesion: 0.40
Nodes (3): IMPORTANT: keep the reminder string free of backticks and $(...) constructs., ref_fs, ref_path

## Ambiguous Edges - Review These
- `config.yaml Configuration` → `OpenSpec spec-driven Schema`  [AMBIGUOUS]
  README.md · relation: semantically_similar_to

## Knowledge Gaps
- **30 isolated node(s):** `$schema`, `plugin`, `ShellyGetConfigResponse`, `BLE`, `Cloud` (+25 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 54 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `config.yaml Configuration` and `OpenSpec spec-driven Schema`?**
  _Edge tagged AMBIGUOUS (relation: semantically_similar_to) - confidence is low._
- **Why does `fetchAndUpdateMetrics()` connect `Shelly API Client` to `Switch Status Metrics`, `Metrics Registration`, `Device Manager & RPC`, `Switch Config Metrics`?**
  _High betweenness centrality (0.075) - this node is a cross-community bridge._
- **Why does `ShellyGetConfigResponse` connect `Shelly Config Response` to `Metrics Registration`?**
  _High betweenness centrality (0.054) - this node is a cross-community bridge._
- **Why does `UpdateShellyGetConfigMetrics()` connect `Metrics Registration` to `Shelly API Client`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **What connects `$schema`, `plugin`, `ShellyGetConfigResponse` to the rest of the system?**
  _30 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Exporter Core & Config` be split into smaller, more focused modules?**
  _Cohesion score 0.10052910052910052 - nodes in this community are weakly interconnected._
- **Should `Helm Chart & Deployment` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._