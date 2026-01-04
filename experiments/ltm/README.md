# LTM Lab

This lab is isolated under `experiments/ltm`. It spins up Prometheus, Tempo, Loki, Grafana, and an OTEL Collector.

## 1) Start the stack (all services in Docker)

```powershell
docker compose -f experiments\ltm\docker-compose.yml up -d --build
```

Grafana: http://localhost:3000 (admin / admin)

## 2) Demo service

The Gin demo service runs in Docker and is exposed at `localhost:8080`.

Endpoints:
- http://localhost:8080/ping
- http://localhost:8080/work
- http://localhost:8080/slow
- http://localhost:8080/error

## 3) Basic query examples

PromQL (Prometheus):
- RPS by route: `sum(rate(ltm_lab_requests_total[1m])) by (route)`
- P95 latency: `histogram_quantile(0.95, sum(rate(ltm_lab_request_latency_ms_milliseconds_bucket[5m])) by (le, route))`
- Inflight: `sum(ltm_lab_inflight_requests)`

LogQL (Loki):
- By service: `{service_name="ltm-lab"}`
- Filter error logs: `{service_name="ltm-lab"} |= "error"`
- Filter by trace: `{service_name="ltm-lab", trace_id="<trace_id>"}`

TraceQL (Tempo):
- All traces for service: `{ resource.service.name = "ltm-lab" }`
- Errors only: `{ resource.service.name = "ltm-lab" && span.status = ERROR }`
- Slow spans: `{ resource.service.name = "ltm-lab" && span.duration > 200ms }`

## 4) Notes

- OTLP endpoint is `localhost:4317` by default; override via `OTEL_EXPORTER_OTLP_ENDPOINT`.
- Logs include `trace_id` and `span_id` in the message for easy linking.

## 5) Shortcuts

- PowerShell: `.\experiments\ltm\scripts\up.ps1` and `.\experiments\ltm\scripts\down.ps1`
- Fish: `./experiments/ltm/scripts/up.fish` and `./experiments/ltm/scripts/down.fish`
- Traffic generator: `.\experiments\ltm\scripts\load.ps1` or `./experiments/ltm/scripts/load.fish`
- Quick reference: `experiments/ltm/QUICK_REFERENCE.md`
