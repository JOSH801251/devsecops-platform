# devsecops-platform 🚩

A GitOps-driven, security-hardened Kubernetes platform for a small API + Worker workload. The local path uses **Kind**, provisioned by **Terraform**, deployed with **Helm/Argo CD**, secured with **Kyverno/Vault**, and observed with **Prometheus/Grafana**.

This is a portfolio project for demonstrating platform engineering decisions, delivery automation, Kubernetes security, and operational visibility. The implementation status and current limitations are called out below rather than implying unfinished capabilities are already live.

[![ci](https://github.com/JOSH801251/devsecops-platform/actions/workflows/ci.yml/badge.svg)](https://github.com/JOSH801251/devsecops-platform/actions/workflows/ci.yml)
[![license: Apache 2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

## Local-first and cost boundaries

The default path is designed to run on a developer workstation with Docker and Kind. It does not require an AWS account or cloud infrastructure.

`infra/terraform/cloud/` is an optional EKS configuration, not a prerequisite for the local project. It has **not** been applied as part of the current local validation. AWS charges depend on region, resource sizes, networking, and runtime; estimate costs before applying and destroy resources when finished. No fixed total cloud-cost claim is made here.

## Architecture

~~~mermaid
flowchart TB
    PR[Pull Request / push to main]

    subgraph CI["GitHub Actions"]
      Go[Go module checks, vet, tests]
      Infra[Helm lint + Terraform fmt/validate]
      Build[Build API and Worker images on runner]
      Scan[Trivy HIGH/CRITICAL image scan]
      Report[Upload SARIF to GitHub Code Scanning]
    end

    PR --> Go
    PR --> Infra
    Go --> Build --> Scan --> Report

    subgraph GitOps["GitOps"]
      Repo[deploy/argocd manifests]
      Argo[Argo CD app-of-apps]
    end

    Repo --> Argo
    Argo -->|automated sync| K8s

    subgraph K8s["Kubernetes: Kind by default / optional EKS"]
      Admission[Kyverno admission]
      API[API service]
      Worker[Worker service]
      PG[(Postgres)]
      Vault[Optional Vault injector]
      Prom[Prometheus]
      Graf[Grafana]
    end

    Admission --> API
    Admission --> Worker
    API --> PG
    Worker --> PG
    Vault -.->|inject DB URL file when enabled| API
    API -->|ServiceMonitor /metrics| Prom
    Prom --> Graf
~~~

CI builds images for scanning on the runner; it does not currently publish them to GHCR. The GitOps flow is a separate deployment path: Argo CD watches the configured repository after the root Application is installed.

## What is implemented

| Capability | Current implementation |
|---|---|
| **Infrastructure as Code** | Terraform local root provisions a Kind cluster. An optional cloud root describes EKS; validation does not create cloud resources. |
| **GitOps** | Argo CD app-of-apps watches `deploy/argocd/apps/` and is configured for automated sync/prune/self-heal. |
| **Secrets** | The basic local install uses a development Kubernetes Secret. The secure-mode path can enable Vault Agent injection and let the API read the database URL from a file. Vault dev credentials and the default Grafana password are demo-only. |
| **Image scanning** | CI builds API/Worker images locally and scans HIGH/CRITICAL findings with Trivy; SARIF is uploaded to GitHub Code Scanning. Image publishing, signing, and SBOM attestation are not implemented in this workflow yet. |
| **Runtime hardening** | Distroless/nonroot images, read-only root filesystem, dropped capabilities, and an application NetworkPolicy are configured in the charts. |
| **Observability** | Prometheus scrapes the API `/metrics` endpoint and tracks aggregate request volume plus pending/completed job counts. The Worker has no metrics endpoint yet. |
| **Availability SLO** | A 99% objective is documented, but a status-labelled availability SLI and error-budget burn-rate alert are not implemented yet. The current Prometheus alert is an operational queue-backlog warning. |
| **Autoscaling** | The API HPA targets CPU utilization; it is not driven by request rate. |

## Repository layout

~~~
services/                Go API + Worker sharing a Postgres job queue
deploy/
  helm/                  Charts for API, Worker, and Postgres
  argocd/                Argo CD Application + AppProject manifests
  kind/                  Kind cluster configuration
infra/
  terraform/local/       Local cluster provisioning (Kind provider)
  terraform/cloud/       Optional EKS configuration — not required locally
observability/
  prometheus/            kube-prometheus-stack values and operational rules
  grafana/dashboards/    API request/queue dashboard
security/
  vault/                 Vault config, policies, Kubernetes auth setup
  policies/              Kyverno and OPA/Gatekeeper policies
docs/adr/                Architecture decision records
scripts/                 Local smoke test
.github/workflows/       Go checks, Helm/Terraform validation, image scan + SARIF
~~~

## Quickstart (local)

Prerequisites: Docker, Kind, kubectl, Helm, Terraform, and Go 1.25+.

~~~bash
# Create the three-node Kind cluster
make cluster

# Build/load local images and install Postgres, API, and Worker
make deploy

# Post 5 jobs and assert that the Worker drains them
make smoke
~~~

For the optional security and observability components:

~~~bash
make policies          # Kyverno admission policies
make vault-bootstrap   # Vault dev mode + Kubernetes auth + API secret injection
make observe           # kube-prometheus-stack + API ServiceMonitor/PrometheusRule + dashboard
make argocd            # Argo CD app-of-apps
~~~

`make secure-mode` combines the local bring-up sequence. For cleanup:

~~~bash
make destroy
~~~

The default Grafana `admin/admin` credential and Vault dev-mode root token are strictly for a disposable local demonstration. Override credentials and use an appropriate production Vault configuration before sharing or reusing the cluster.

## Optional cloud variant

The EKS root is optional and requires valid AWS credentials. It was not applied during the current project validation. Review the Terraform plan and current AWS pricing before creating anything:

~~~bash
cd infra/terraform/cloud
terraform init
terraform plan
terraform apply
# When the demo is complete:
terraform destroy
~~~

Do not use `-auto-approve` for a first run. Confirm the resources and estimated charges in the plan before applying. See [infra/terraform/cloud/README.md](infra/terraform/cloud/README.md) for the module's design notes.

## Roadmap

- [x] **Phase 1a** — Scaffold: services, Helm, Kind, Terraform local
- [x] **Phase 1b** — Functional: API persists jobs to Postgres; Worker drains via `FOR UPDATE SKIP LOCKED`
- [x] **Phase 1c** — Optional cloud variant: EKS Terraform configuration
- [x] **Phase 2a** — Vault Agent sidecar injection path (ADR 0004)
- [x] **Phase 2b** — Kyverno admission policies: trusted registry + resource requirements
- [x] **Phase 2c** — Argo CD app-of-apps wiring
- [x] **Phase 2d** — kube-prometheus-stack, API ServiceMonitor, queue alert, Grafana dashboard
- [ ] **Phase 3** — HTTP status metrics and a measured availability SLI; image publishing, cosign signing, and SBOM attestation
- [ ] **Stretch** — Chaos lab, canary deployments, and further policy automation

See [docs/adr/](docs/adr/) for architecture decisions, including the current implementation boundaries in [ADR 0005](docs/adr/0005-ci-and-observability-scope.md).

## License

Apache 2.0 — see [LICENSE](LICENSE).
