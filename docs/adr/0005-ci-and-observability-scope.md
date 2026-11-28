# ADR 0005 — CI and observability implementation boundaries

## Status

Accepted

## Context

A portfolio repository should distinguish working automation and telemetry from planned capabilities. The README, dashboard, and CI workflow previously described image signing/SBOM publication and an availability burn-rate alert more broadly than the current implementation supported.

The API currently exports an aggregate `api_requests_total` counter without HTTP status labels, plus pending/completed job counts. The API has a ServiceMonitor; the Worker does not expose a Prometheus metrics endpoint. Therefore, request volume and queue state can be observed, but an HTTP success ratio and availability error budget cannot yet be calculated.

## Decision

- CI verifies Go module consistency, module checksums, `go vet`, race-enabled Go tests, Helm lint, and Terraform formatting/validation for both Terraform roots.
- CI builds API and Worker images on the runner, scans HIGH/CRITICAL vulnerabilities with Trivy, and uploads SARIF to GitHub Code Scanning.
- CI does not currently publish images to GHCR, sign images with cosign, or generate/attach SBOM attestations. Those are Phase 3 work, not completed capabilities.
- Prometheus records aggregate API request rate and raises an operational queue-backlog warning. It does not raise an availability burn-rate alert until the API exports response-status metrics and the SLI is defined.
- The Grafana dashboard reports the request/queue signals that exist and explicitly identifies the availability-measurement gap. A 99% objective is not presented as measured compliance.
- Local Kind is the default execution target. The EKS Terraform root is optional; CI validates its configuration but does not apply it or create AWS resources.

## Consequences

- Workflow permissions are limited to repository read access by default and Code Scanning write access only in the image-scan job.
- Reviewers can distinguish implemented CI/observability from planned supply-chain and SLO capabilities.
- Follow-up work can add status-labelled metrics, a multi-window SLO alert, GHCR publishing, cosign keyless signing, and SBOM attestations without rewriting the current history.
