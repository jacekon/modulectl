# CLAUDE.md

modulectl is a **CLI tool for Kyma module developers**. It provides commands to scaffold a new module structure and to prepare data for the `ocm` command that further builds the module as an Open Component Model (OCM) artifact and pushes it to an OCI registry, ready for consumption by [Lifecycle Manager](https://github.com/kyma-project/lifecycle-manager).

It is a [Cobra](https://github.com/spf13/cobra)-based CLI written in Go. There is no operator, no controller, and no Kubernetes runtime dependency — it is a developer toolchain binary. It builds cross-platform (`CGO_ENABLED=0`, no FIPS requirement).

To build modulectl, run `make build`.

## modulectl provides the following commands

| Command | What it does |
|---|---|
| `modulectl create` | Creates input files for the module as required by the `ocm` tool |
| `modulectl scaffold` | Generates the files needed to start a new module (`module-config.yaml`, default CR, etc.) |
| `modulectl version` | Prints the current version |

Full flag reference: `docs/gen-docs/` (auto-generated — do not edit manually, run `make docs` to regenerate).

## modulectl follows a layered architecture

```
cmd/modulectl/      ← CLI entrypoint, Cobra command definitions
internal/service/   ← Business logic, one package per concern
tools/              ← Shared low-level utilities (filesystem, yaml, io)
```

The composition root is in `cmd/modulectl/cmd.go` — all dependencies are wired there, not inside the service packages.

## modulectl uses a unit and e2e test setup

Run `make test` for unit tests with race detector.

E2E tests require a local registry — set it up first with `./scripts/re-create-test-registry.sh && ./scripts/build-modulectl.sh`, then run `make test-create-cmd` or `make test-scaffold-cmd`. See `docs/contributor/local-test-setup.md` for the full walkthrough.

## modulectl uses the following code conventions

Go nolint and import ordering rules load automatically when editing `.go` files — see [`.claude/rules/go-conventions.md`](.claude/rules/go-conventions.md).

Key rules from `.golangci.yaml`:
- **All linters enabled by default** — check `.golangci.yaml` before adding `//nolint`
- **`//nolint` requires explanation**: e.g., `//nolint:funlen // command wiring`
- **Import ordering** (gci): standard → third-party → project (`github.com/kyma-project/modulectl`)
- **Import aliases strictly enforced** — key ones: `iotools`, `commonerrors`, `scaffoldcmd`, `createcmd`, `moduleconfiggenerator`, `moduleconfigreader`, `ocmv1`, `ociartifacttypes`
- **Line length**: 120 chars | **Function length**: 80 lines / 50 statements | **Cyclomatic complexity**: 20
- `fmt.Print` is forbidden — use `fmt.Println` for any console output

## Commits and Pull Requests

- PRs are usually created from a **fork branch** against `main`.
- PRs are merged with **squash merge** — the PR title and description form the commit message.
- Follow [conventional commits](https://www.conventionalcommits.org/), enforced by `.github/workflows/lint-conventional-prs.yml`.
- PR title format: `<type>: <title>` where the title is one sentence explaining the reason for the changeset.
- Ask what type to use when creating a PR: `deps`, `chore`, `docs`, `feat`, `fix`, `refactor`, `test`.
- PR description should contain a short summary of the changes and, if applicable, a reference to the issue using the `closes` or `resolves` keyword.
- Never mention Claude or any AI agent in commits or PRs (no author attribution, no `Co-Authored-By`, no references in commit messages).
