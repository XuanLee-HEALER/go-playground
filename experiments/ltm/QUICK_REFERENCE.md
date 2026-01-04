# LTM Lab Quick Reference

This is a minimal guide to explore and extend the lab without digging through configs.

## Stack overview
- Prometheus (metrics): http://localhost:9090
- Loki (logs): http://localhost:3100
- Tempo (traces): http://localhost:3200
- Grafana (UI): http://localhost:3000 (admin/admin)
- Demo service: http://localhost:8080

## Start/stop

PowerShell:
```powershell
.\experiments\ltm\scripts\up.ps1
.\experiments\ltm\scripts\down.ps1
```

Fish:
```fish
./experiments/ltm/scripts/up.fish
./experiments/ltm/scripts/down.fish
```

## Generate traffic

PowerShell:
```powershell
.\experiments\ltm\scripts\load.ps1
```

Fish:
```fish
./experiments/ltm/scripts/load.fish
```

## Demo endpoints
- `GET /ping` quick OK
- `GET /work` variable latency
- `GET /slow` long latency (helps inflight)
- `GET /error` error path

## Query language cheat sheet

PromQL (metrics):
- RPS by route: `sum(rate(ltm_lab_requests_total[1m])) by (route)`
- P95 latency: `histogram_quantile(0.95, sum(rate(ltm_lab_request_latency_ms_milliseconds_bucket[5m])) by (le, route))`
- Inflight: `sum(ltm_lab_inflight_requests)`

LogQL (logs):
- All logs: `{service_name="ltm-lab"}`
- Errors: `{service_name="ltm-lab"} |= "error"`
- By trace: `{service_name="ltm-lab", trace_id="<trace_id>"}`

TraceQL (traces):
- All service traces: `{ resource.service.name = "ltm-lab" }`
- Errors: `{ resource.service.name = "ltm-lab" && span.status = ERROR }`
- Slow spans: `{ resource.service.name = "ltm-lab" && span.duration > 200ms }`

## Where data flows
- App -> OTEL Collector (OTLP gRPC) ->
  - Prometheus (metrics)
  - Tempo (traces)
  - Loki (logs)

## Extend the lab

### Add a new metric
1. Create the instrument in `experiments/ltm/app/telemetry.go`.
2. Record it in `experiments/ltm/app/main.go`.
3. Query in Prometheus/Grafana with `ltm_lab_<metric_name>`.

### Add a new log attribute
1. Add `log.String/Int/...` attributes in `experiments/ltm/app/main.go`.
2. Query in Loki: `{service_name="ltm-lab"} | json | <field>="value"`.

### Add a new trace span
1. Use `telemetry.Tracer.Start` in `experiments/ltm/app/main.go`.
2. Verify in Grafana Tempo with TraceQL.

### Add a new dashboard panel
1. Edit `experiments/ltm/grafana/dashboards/ltm-lab.json`.
2. Grafana auto-reloads dashboards from that folder.

## Common knobs
- OTLP endpoint: `OTEL_EXPORTER_OTLP_ENDPOINT` (set in compose).
- Grafana datasources: `experiments/ltm/grafana/provisioning/datasources/datasources.yaml`.
- Collector pipelines: `experiments/ltm/otel-collector.yaml`.
