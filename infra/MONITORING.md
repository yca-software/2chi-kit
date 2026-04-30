# Monitoring and Observability Guide

This repo ships a local monitoring stack with:

- Prometheus (metrics) at `http://localhost:9090`
- Loki (log store) at `http://localhost:3100`
- Promtail (Docker log shipper)
- Grafana (dashboards + Explore) at `http://localhost:3030` (`admin` / `admin`)

## 1) Start and stop

From repo root:

```bash
make all-up
```

This starts infra + monitoring and initializes RabbitMQ job topology.

Useful commands:

```bash
make monitoring-up
make monitoring-down
make monitoring-logs
make infra-grafana-recreate
make infra-down
```

Use `make infra-grafana-recreate` after changing Grafana datasource/dashboard provisioning files.

## 2) What is pre-provisioned

Grafana is preconfigured with:

- Datasource `Prometheus` (UID: `prometheus`)
- Datasource `Loki` (UID: `loki`)
- Dashboard `Go API — Observability` (UID: `go-api-observability`)

## 3) Health checks (quick verification)

Run these from repo root:

```bash
docker compose -f infra/docker-compose.yml ps
curl -sf http://localhost:9090/-/ready
curl -sf http://localhost:3100/ready
curl -sf http://localhost:3030/api/health
curl -s -u admin:admin http://localhost:3030/api/datasources
curl -s -u admin:admin "http://localhost:3030/api/search?query=go-api-observability"
curl -sG --data-urlencode 'query=up{job="go-api-template"}' http://localhost:9090/api/v1/query
```

Expected:

- all compose services `Up`
- ready endpoints return success
- Grafana API returns both `Prometheus` and `Loki`
- dashboard is found in search API
- Prometheus `up{job="go-api-template"} == 1` when API is running

## 4) Metrics workflow (Prometheus + dashboard)

### In Grafana

1. Open `http://localhost:3030`
2. Go to **Dashboards -> Browse**
3. Open **Go API — Observability**

This dashboard includes:

- HTTP traffic / errors / latency
- DB query rate + p95 latency
- Job publish/outcome/duration
- Rate limit hits
- Go runtime basics

### In Grafana Explore (metrics)

1. Go to **Explore**
2. Choose datasource **Prometheus**
3. Try:

```promql
up
```

```promql
sum(rate(go_api_starter_http_requests_total[5m]))
```

```promql
histogram_quantile(0.95, sum by (le) (rate(go_api_starter_http_request_duration_seconds_bucket[5m])))
```

## 5) Logs workflow (Loki + Promtail)

Promtail scrapes Docker container logs via `/var/run/docker.sock` and pushes to Loki.

### In Grafana Explore (logs)

1. Go to **Explore**
2. Choose datasource **Loki**
3. Run queries such as:

```logql
{container="rabbitmq"}
```

```logql
{compose_service="redis"}
```

```logql
{compose_service="go-api-prometheus"} |= "error"
```

```logql
{compose_service="rabbitmq"} |~ "warn|error|failed"
```

### Useful label filters

- `container`
- `compose_service`
- `container_id`

## 6) Important note for Cursor debugger / host-run API

If you run the API with Cursor/VSCode debugger on host (not in Docker):

- Prometheus can still scrape API metrics at `localhost:1337/metrics` (configured target: `host.docker.internal:1337`)
- Loki/Promtail will **not** automatically capture that API process stdout

Why: Promtail is currently configured for Docker log discovery only.

If you want host-run API logs in Loki too, add one of:

- a file target in Promtail (`static_configs` + `__path__`) and write API logs to file
- run API in Docker so logs are discoverable via Docker socket

## 7) Common troubleshooting

### Loki datasource missing in Grafana

```bash
make infra-grafana-recreate
```

Then refresh Grafana and re-check datasource list.

### Dashboard missing or stale

```bash
make infra-grafana-recreate
```

Check `infra/grafana/dashboards/go-api-observability.json`.

### Prometheus has no API metrics

- Verify API is running: `curl -sf http://localhost:1337/metrics`
- Verify target: `up{job="go-api-template"}`
- Check Prometheus targets UI: `http://localhost:9090/targets`

### Loki returns no logs

- Generate a log event (example: `docker restart redis`)
- Query recent range in Explore
- Check promtail logs:

```bash
docker logs go-api-promtail --tail 200
```
