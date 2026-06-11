---
paths:
  - "**/*.go"
---

# Go code conventions — modulectl

`make lint` is the authoritative check. The full config is in `.golangci.yaml`.

## nolint policy

Every `//nolint` directive **must** include an explanation:

```go
//nolint:funlen // command wiring — acceptable exception
```

Bare suppressions fail CI. Check `.golangci.yaml` before adding any.

## Import aliases — strictly enforced

| Package | Alias |
|---|---|
| `github.com/kyma-project/modulectl/tools/io` | `iotools` |
| `github.com/kyma-project/modulectl/internal/common/errors` | `commonerrors` |
| `github.com/kyma-project/modulectl/cmd/modulectl/scaffold` | `scaffoldcmd` |
| `github.com/kyma-project/modulectl/cmd/modulectl/create` | `createcmd` |
| `github.com/kyma-project/modulectl/internal/service/moduleconfig/generator` | `moduleconfiggenerator` |
| `github.com/kyma-project/modulectl/internal/service/moduleconfig/reader` | `moduleconfigreader` |
| `ocm.software/ocm/api/ocm/compdesc/meta/v1` | `ocmv1` |
| `ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact` | `ociartifacttypes` |

## Import ordering

gci enforces: standard → third-party → project (`github.com/kyma-project/modulectl`).

## Console output

`fmt.Print` is forbidden — use `fmt.Println` for any console output (enforced by forbidigo linter).

## Composition root

All dependency wiring belongs in `cmd/modulectl/cmd.go`. Service packages must not import other service packages directly — dependencies are injected as interfaces.
