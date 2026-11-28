# Grafana dashboards

`slo-api.json` shows the API request rate and the Postgres job-queue state:

- Aggregate API request rate.
- Pending jobs.
- Completed jobs.
- Job-drain throughput.
- Request-rate and queue-depth trends.

## Metrics and current limits

The API exports `api_requests_total` as an aggregate counter and exposes pending/completed job counts. The request counter does **not** currently include an HTTP status label. As a result, this dashboard cannot calculate a trustworthy successful-request ratio, 99% availability compliance, or error-budget burn rate.

The 99% objective is a target, not a measured result. The dashboard displays this instrumentation gap explicitly rather than plotting a fabricated compliance value. The current Prometheus alert is an operational queue-backlog warning, not an SLO burn-rate alert. The Worker does not currently expose a Prometheus metrics endpoint.

## Import

The Grafana sidecar (configured in `observability/prometheus/values.yaml`) auto-imports any ConfigMap labeled `grafana_dashboard=1`. The `make observe` target creates and labels this ConfigMap automatically. To load it manually:

~~~bash
kubectl -n monitoring create configmap api-slo-dashboard \
  --from-file=slo-api.json=observability/grafana/dashboards/slo-api.json \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n monitoring label configmap api-slo-dashboard grafana_dashboard=1 --overwrite
~~~

## Open Grafana

~~~bash
kubectl -n monitoring port-forward svc/kps-grafana 3000:80
# http://localhost:3000 — admin / admin (local demo only)
~~~

Change the demo Grafana password before using the chart in a shared environment.
