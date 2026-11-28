<!-- Animated header: GitHub renders externally hosted SVG images; it does not execute custom JavaScript or CSS. -->
<p align="center">
  <img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=700&size=25&duration=2800&pause=900&color=38BDF8&center=true&vCenter=true&width=850&lines=DEVSECOPS+PLATFORM;GitOps+%7C+Kubernetes+%7C+Platform+Engineering;Secure+delivery+%7C+Observability+%7C+Automation" alt="Animated DevSecOps Platform heading" />
</p>

<h1 align="center">DevSecOps Platform</h1>

<p align="center">
  A local-first, GitOps-driven Kubernetes platform for a Go API and background Worker backed by PostgreSQL.
  <br />
  Built to practice platform engineering, secure workload delivery, infrastructure automation and operational troubleshooting.
</p>

<p align="center">
  <a href="https://github.com/JOSH801251/devsecops-platform/actions/workflows/ci.yml"><img src="https://github.com/JOSH801251/devsecops-platform/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <img src="https://img.shields.io/badge/Kubernetes-Min ikube-326CE5?logo=kubernetes&logoColor=white" alt="Kubernetes" />
  <img src="https://img.shields.io/badge/GitOps-Argo%20CD-EF7B4D?logo=argo" alt="Argo CD" />
  <img src="https://img.shields.io/badge/Packaging-Helm-0F1689?logo=helm" alt="Helm" />
  <img src="https://img.shields.io/badge/Infrastructure-Terraform-7B42BC?logo=terraform" alt="Terraform" />
  <img src="https://img.shields.io/badge/Container-Docker-2496ED?logo=docker&logoColor=white" alt="Docker" />
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache--2.0-blue.svg" alt="Apache 2.0 license" /></a>
</p>

<p align="center">
  <strong>Local deployment validated on Minikube</strong> · No cloud account required for the demonstrated workflow
</p>

---

## Table of contents

- [Project overview](#project-overview)
- [Validated deployment](#validated-deployment)
- [Architecture](#architecture)
- [Technology stack](#technology-stack)
- [GitOps workflow](#gitops-workflow)
- [Security and supply chain](#security-and-supply-chain)
- [Observability and autoscaling](#observability-and-autoscaling)
- [Screenshots](#screenshots)
- [Repository layout](#repository-layout)
- [Run locally](#run-locally)
- [Optional infrastructure paths](#optional-infrastructure-paths)
- [Implementation status and limitations](#implementation-status-and-limitations)
- [Troubleshooting notes](#troubleshooting-notes)
- [Roadmap](#roadmap)
- [License](#license)

## Project overview

This repository contains a small cloud-native workload and the platform configuration around it:

- **API:** accepts work and persists jobs in PostgreSQL.
- **Worker:** polls the database queue and marks pending jobs as completed.
- **PostgreSQL:** stores job state on a persistent volume.
- **Helm:** packages PostgreSQL, API and Worker Kubernetes resources.
- **Argo CD:** manages the applications using an app-of-apps GitOps structure.
- **Kubernetes controls:** resource requests/limits, container security settings, health probes and API CPU-based autoscaling.
- **CI and security tooling:** GitHub Actions validation and Trivy image scanning are defined in the repository.

The project is intentionally local-first. The current hands-on deployment was validated on **Minikube**. The repository also contains a Kind/Terraform local infrastructure path and an optional EKS Terraform configuration; those are separate paths and should not be confused with the Minikube deployment demonstrated below.

## Validated deployment

The API, Worker and PostgreSQL were deployed to a local Minikube cluster using their Helm charts and then reconciled by Argo CD.

| Component | Validated configuration / result |
|---|---|
| Kubernetes | Local Minikube cluster |
| GitOps | Argo CD root Application plus API, PostgreSQL and Worker child Applications |
| PostgreSQL | StatefulSet with a 1 Gi persistent volume claim |
| API | 2 replicas; ClusterIP Service; `/healthz` and `/readyz` probes |
| Worker | 1 replica; database-backed queue polling |
| Autoscaling | API HPA: minimum 2, maximum 6, target CPU utilization 70% |
| Metrics | `kubectl top pods` returned CPU and memory usage |
| GitOps reconciliation | Applications reported Synced / Healthy; a manual API scale change from 2 to 3 replicas was reverted to 2 by Argo CD self-healing |

The queue workflow was exercised: a job was inserted into PostgreSQL and the Worker changed its status from `pending` to `done`, recording `processed_at`. This demonstrates database-backed job handoff; it is not a claim of a production-grade queue with retry/dead-letter handling.

## Architecture

~~~mermaid
flowchart TB
    Dev[Developer] --> Git[GitHub repository]
    Git --> CI[GitHub Actions CI]
    Git --> Root[Argo CD root Application]
    Root --> ApiApp[API Application]
    Root --> PgApp[PostgreSQL Application]
    Root --> WorkerApp[Worker Application]

    subgraph Cluster[Local Kubernetes - Minikube validated]
      ApiApp --> API[API Deployment - 2 replicas]
      PgApp --> PG[(PostgreSQL StatefulSet + PVC)]
      WorkerApp --> Worker[Worker Deployment - 1 replica]
      API -->|enqueue job| PG
      Worker -->|poll and process jobs| PG
      HPA[Horizontal Pod Autoscaler] -. CPU target .-> API
      Metrics[Metrics Server] -. resource metrics .-> HPA
    end

    CI --> Scan[Trivy image scan]
    API --> Health[healthz / readyz]
~~~

The repository also contains optional configuration for Kind/Terraform, EKS, Vault, Kyverno and Prometheus/Grafana. Their presence in the repository does not mean every optional component was enabled in the Minikube validation.

## Technology stack

| Area | Tools |
|---|---|
| Application | Go, PostgreSQL |
| Containers | Docker |
| Kubernetes | Kubernetes, Minikube; Kind configuration also present |
| Packaging | Helm |
| GitOps | Argo CD |
| Infrastructure as Code | Terraform |
| CI | GitHub Actions |
| Image vulnerability scanning | Trivy |
| Policy and secrets configurations | Kyverno, Vault |
| Monitoring configuration | Prometheus, Grafana |
| Source control | Git, GitHub |

## GitOps workflow

The root Argo CD Application watches `deploy/argocd/apps/` and manages the API, PostgreSQL and Worker child Applications.

The child applications point to their respective Helm charts:

- `deploy/helm/api` — Helm release `api`
- `deploy/helm/postgres` — Helm release `app`
- `deploy/helm/worker` — Helm release `worker`

Automated sync, pruning and self-healing are configured. During validation, the API Deployment was manually scaled from 2 to 3 replicas. Argo CD reconciled the live state back to the Git-defined value of 2 replicas. The applications then reported `Synced` and `Healthy`.

## Security and supply chain

The repository includes security-oriented configuration and CI checks:

- Trivy scans API and Worker images for HIGH and CRITICAL vulnerabilities in the CI workflow.
- CI uploads SARIF results to GitHub Code Scanning when the workflow and repository settings support it.
- Helm charts configure non-root execution, read-only root filesystems, dropped Linux capabilities and resource requests/limits.
- Kubernetes NetworkPolicy and Kyverno policy manifests are present in the repository.
- A Vault Agent injection path and related configuration are documented in the repository.

**Scope note:** the validated Minikube session described above focused on the API, Worker, PostgreSQL, Argo CD, probes, persistence, metrics and HPA. Do not interpret the presence of Vault/Kyverno manifests as proof that those optional controls were enabled in that session. The basic PostgreSQL deployment uses a development Kubernetes Secret; replace demo credentials and configure production-grade secret management before any production use.

The current CI workflow builds images for scanning on the runner. It does not currently publish images to GHCR, sign them with Cosign or produce SBOM attestations.

## Observability and autoscaling

The API exposes health endpoints:

- `GET /healthz` — process liveness.
- `GET /readyz` — application readiness.

Both endpoints returned successful JSON responses during validation:

~~~json
{"status":"live"}
{"status":"ready"}
~~~

The API HPA is configured for CPU utilization with a 70% target, 2 minimum replicas and 6 maximum replicas. Metrics Server was available during validation: `kubectl top pods -n app` returned CPU and memory readings, and `kubectl get hpa -n app` reported a current CPU value against the target.

Prometheus/Grafana configuration exists in the repository. The Worker does not currently expose its own application metrics endpoint. A measured availability SLI and error-budget burn-rate alert remain future work.

## Screenshots

The following image paths are reserved for the project evidence. Add the corresponding PNG files under `docs/screenshots/` for them to render on GitHub.

<p align="center"><strong>Argo CD application health</strong></p>
<p align="center"><img src="docs/screenshots/argocd-applications.png" alt="Argo CD root and child application health" width="850" /></p>

<p align="center"><strong>Kubernetes workloads and persistent storage</strong></p>
<p align="center"><img src="docs/screenshots/kubernetes-workloads.png" alt="API, Worker, PostgreSQL pods, services and persistent volume claim" width="850" /></p>

<p align="center"><strong>API health checks</strong></p>
<p align="center"><img src="docs/screenshots/api-health-checks.png" alt="API liveness and readiness endpoint responses" width="850" /></p>

<p align="center"><strong>Horizontal Pod Autoscaler</strong></p>
<p align="center"><img src="docs/screenshots/hpa-autoscaling.png" alt="API Horizontal Pod Autoscaler status" width="850" /></p>

<p align="center"><strong>Argo CD self-healing demonstration</strong></p>
<p align="center"><img src="docs/screenshots/argocd-self-healing.png" alt="Argo CD restoring the API Deployment to its Git-defined replica count" width="850" /></p>

## Repository layout

~~~text
services/                  Go API and Worker
deploy/
  helm/                    API, Worker and PostgreSQL charts
  argocd/                  Argo CD AppProject, root and child Applications
  kind/                    Kind cluster configuration
infra/
  terraform/
    local/                 Local Kind provisioning configuration
    cloud/                 Optional EKS configuration
observability/
  prometheus/              Prometheus values and operational rules
  grafana/dashboards/      API and queue dashboard
security/
  vault/                   Vault configuration, policies and auth setup
  policies/                Kyverno and OPA/Gatekeeper policies
docs/
  adr/                     Architecture decision records
  screenshots/             Project validation screenshots
scripts/                   Local smoke-test scripts
.github/workflows/         CI and image scanning workflows
~~~

## Run locally

### Validated Minikube deployment

The following describes the deployment path exercised during the project work. It assumes Docker Desktop, Minikube, kubectl, Helm and Git are installed and Minikube is running.

Build the API and Worker images using the repository's documented build targets or Dockerfiles, then make the images available to Minikube. The image references used in the validated Helm releases were:

- `ghcr.io/josh801251/devsecops-platform-api:0.1.0`
- `ghcr.io/josh801251/devsecops-platform-worker:0.1.0`

Deploy the charts with distinct Helm release names (important: the PostgreSQL release is `app`, the API release is `api`, and the Worker release is `worker`). The Argo CD root Application is defined in `deploy/argocd/application.yaml`.

~~~bash
# Check the active Kubernetes context
kubectl config current-context

# Inspect the Argo CD applications
kubectl get applications -n argocd

# Inspect application workloads and persistent storage
kubectl get pods,svc,pvc -n app

# Check API autoscaling
kubectl get hpa -n app

# Test API health endpoints after port-forwarding the api-api Service
kubectl port-forward -n app svc/api-api 8080:80
# In another terminal:
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
~~~

### Repository Makefile path

The repository also provides a Kind-based Makefile workflow:

~~~bash
make cluster
make deploy
make smoke
~~~

Review the Makefile and the relevant Terraform/Kind configuration before using this path. The Makefile workflow is distinct from the Minikube + Argo CD deployment validated above; do not run both paths against the same cluster without understanding their resource names and lifecycle.

## Optional infrastructure paths

### Kind with Terraform

The repository includes a local Terraform root for provisioning a Kind cluster. This is a separate local infrastructure option from the Minikube cluster used for the validation screenshots.

### Optional EKS

`infra/terraform/cloud/` contains an optional EKS configuration. It was **not applied** as part of the local validation. AWS credentials and billable infrastructure are required to apply it. Review the plan, region, networking, node sizes and current AWS pricing before applying. Do not use `-auto-approve` for a first run.

~~~bash
cd infra/terraform/cloud
terraform init
terraform plan
# Apply only after reviewing the plan and expected charges:
terraform apply
# Destroy resources when the experiment is complete:
terraform destroy
~~~

See [the cloud Terraform README](infra/terraform/cloud/README.md) for design notes.

## Implementation status and limitations

| Capability | Status |
|---|---|
| Go API + PostgreSQL-backed job creation | Implemented and exercised locally |
| Worker queue polling and job completion | Implemented and exercised locally |
| Helm charts for API, Worker and PostgreSQL | Deployed in Minikube |
| Argo CD app-of-apps and child Applications | Deployed; project Applications reported Synced / Healthy |
| Argo CD automated self-healing | Demonstrated by manually changing API replicas and observing reconciliation |
| PostgreSQL persistent volume | PVC Bound; existing job record verified after probe correction |
| API health probes | `/healthz` and `/readyz` verified |
| API CPU HPA | Configured and metrics available during validation |
| Terraform Kind path | Configuration present; separate from Minikube validation |
| EKS | Optional configuration only; not applied |
| Trivy CI scanning | Workflow configuration present; review latest Actions run for current result |
| GHCR publishing, Cosign signing, SBOM attestation | Not implemented in current workflow |
| Vault and Kyverno | Configuration present; not claimed as enabled in the validated Minikube deployment |
| Prometheus/Grafana | Configuration present; full monitoring stack not claimed as part of the validated Minikube deployment |
| Worker metrics endpoint | Not implemented |
| Availability SLI and burn-rate alert | Not implemented |
| Production queue retries / dead-letter handling | Not implemented |

## Troubleshooting notes

### PostgreSQL health-probe correction

During local validation, PostgreSQL restarted because the Kubernetes exec probe passed strings such as `$(POSTGRES_USER)` literally to `pg_isready`. Kubernetes exec probes do not automatically run commands through a shell, so shell-style variable expansion did not occur.

The probe was corrected to call `pg_isready` directly against `127.0.0.1:5432`, with an explicit timeout. After the change was committed and synchronized by Argo CD:

- PostgreSQL returned `accepting connections`.
- The PostgreSQL pod was `1/1 Running` with zero restarts after the rollout.
- The existing job record remained in the database.
- API readiness and liveness checks returned success.
- The Worker remained running.

### HPA metrics

The HPA initially reported missing resource metrics. Metrics Server became available and `kubectl top pods -n app` returned CPU and memory readings. The HPA then reported current CPU utilization against its 70% target.

These are useful examples of diagnosing Kubernetes probe behavior, GitOps reconciliation and resource-metrics dependencies.

## Roadmap

- [x] Go API and PostgreSQL-backed job queue
- [x] Worker polling and job completion
- [x] Helm charts for PostgreSQL, API and Worker
- [x] Argo CD root Application and child Applications
- [x] Automated sync and self-healing demonstration
- [x] PostgreSQL persistence and API health-check validation
- [x] API CPU HPA and resource metrics validation
- [ ] Tighten Argo CD AppProject permissions to least privilege
- [ ] Complete and document the Kind/Terraform workflow end-to-end
- [ ] Validate Vault and Kyverno paths in a dedicated local run
- [ ] Validate Prometheus/Grafana dashboards and alerts in a dedicated local run
- [ ] Add Worker metrics and a measured availability SLI
- [ ] Add image publishing, Cosign signing and SBOM attestation
- [ ] Explore chaos testing and progressive delivery

## License

Apache 2.0 — see [LICENSE](LICENSE).
