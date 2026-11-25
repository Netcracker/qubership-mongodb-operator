# AGENTS.md

Guidance for AI coding assistants working in the `qubership-mongodb-operator` repository, a
monorepo that deploys MongoDB on Kubernetes/OpenShift. Repository source and Helm templates
are the source of truth.

## Repository Layout

- `operator/`: main Go operator (`api/v1alpha1/` CRD types, `controllers/`, `pkg/mongodb/`,
  `pkg/dr/`, `pkg/fiber/` HTTP server) and Helm chart `charts/helm/mongodb-operator/`.
- `services/service/`: supplementary operator (backup, DBaaS, exporter, robot-tests) and
  Helm chart `charts/helm/mongodb-services/`.
- `services/dbaas-adapter/`, `services/backup-daemon/`, `services/prometheus-exporter/`:
  adapter and Docker-built services.
- `services/tests/robot/`: Robot Framework integration suites.
- `mongo/`: custom MongoDB 7/8 image.
- `docs/public/`: installation, configuration, operations, DR, backup, and troubleshooting docs.
- `agent-packages/troubleshoot-mongodb/`: APM troubleshooting skill.
- `.github/build-config.cfg` lists image components; `.github/workflows/` is CI.

## Backward Compatibility

- Every change must preserve backward compatibility. Do not propose or implement a breaking
  change; stop and get explicit human approval first, with a migration plan.
- CRD changes must be additive: new optional (`omitempty`) fields only. Do not remove,
  rename, retype, or change the meaning of fields in `operator/api/v1alpha1/` or
  `services/service/api/`.
- Do not remove or rename Helm values. New values need defaults that preserve existing
  deployments and a matching `values.schema.json` entry.
- Do not change routes or request/response bodies in `operator/pkg/fiber/handlers.go`
  (consumed by DBaaS and other services).
- Backup format, data layout, and `pkg/dr/` steps must stay readable and restorable from
  previous releases.

## Safe Operations

- Do not run `docker-push`, `docker-buildx`, `deploy`, `install`, or `undeploy` make targets
  unless explicitly asked.
- Do not delete PVCs, remove data directories, or run `rs.reconfig` on MongoDB without
  approval.
- Never run integration tests against production clusters.
- Credentials come only from Kubernetes Secrets. Never log or print them, and do not loosen
  `securityContext` or TLS settings.
- Do not edit generated files by hand: `zz_generated.deepcopy.go`,
  `charts/helm/*/crds/crd.yaml`, `operator/pkg/fiber/docs/`.
- Do not edit `.github/workflows/*` unless the task is about CI.
- Documentation-only changes must not alter commands, YAML values, or procedures unless a
  technical fix is requested.

## Verification

- After Go changes in `operator/` or `services/service/`, run `go build ./... && go vet ./...`
  in each touched module. Run `go test ./...` on CI or a kind cluster, since tests need a
  cluster/envtest.
- After changes in `services/dbaas-adapter`, run `go build ./...`.
- After API type changes, run `make generate` and `make manifests` in the touched module and
  review the diff. Commit only hunks caused by your change.
- After Helm chart edits, run `helm template` and `helm unittest` for the touched chart, and
  keep `values.schema.json` in sync.
- After Dockerfile or `mongo/` changes, build the touched component image.
- Robot suites run only in kind through `.github/workflows/run_tests.yaml`. If a cluster or
  tool is unavailable, report the check as blocked; never claim it passed.
- Do not bump dependencies without reading the upstream changelog and running tests. Never
  hand-edit `go.sum`.

## Documentation

- Update `docs/public/` (`configuration.md` for Helm values) in the same change as any
  user-facing change.
- Edit `docs/public/troubleshooting.md`, not its copy under `agent-packages/` (CI syncs it).
- Keep relative links current when docs move or are renamed. Do not invent links to
  documents that do not exist.

## Commits

- Use conventional commit prefixes (`fix:`, `feat:`, `docs:`) and target `main`.
- Never create a commit until the user provides a ticket number.
- Use <type>: [<ticket-number>] <meaningful message>.
- Fill in `.github/pull_request_template.md` for every PR.
- Check `git status` first so secrets, kubeconfigs, `cover.out`, `bin/manager`,
  `apm_modules/`, and built binaries are not committed.
- Summarize changed files, check results, and blocked checks before committing.
- Do not push, force-push, merge, or tag a release without an explicit request.
