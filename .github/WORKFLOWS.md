# GitHub workflows

This repository uses GitHub Actions to test the provider, check its quality, build and deploy the documentation site, and publish releases. This page is for contributors who need to know what runs on a pull request and how a release is made.

## Index workflow

Every run starts from [`.index.yml`](./workflows/.index.yml) (`index`), on a pull request, a push to `main`, a manual run, the `auto-merged` dispatch the merge pipeline sends after merging a Dependabot pull request, and a weekly schedule. Each of its jobs calls one pipeline, so the run page reads `<pipeline> / <stage>`:

| Job | Pipeline | Stages | When |
| :--- | :--- | :--- | :--- |
| `changes` | | | Every run but the weekly one: on a pull request, works out whether the provider's code, its tooling, a shell script or a workflow changed. Every other run checks everything. |
| `version` | forgego's [`version.yml`](https://github.com/apollogeddon/forgego/blob/main/.github/workflows/version.yml) | `release` | On `main` (see [Release process](#release-process)). |
| `provider` | [`provider.yml`](./workflows/provider.yml) | `lint`, `scan`, `test`, `test (gateway)`, `build` | On a pull request when the provider changed, and on `main` when `version` created a release. |
| `publish` | forgego's [`publish.yml`](https://github.com/apollogeddon/forgego/blob/main/.github/workflows/publish.yml) | `assets` | When a release was created and `provider` passed. |
| `website` | forgejs's [`website.yml`](https://github.com/apollogeddon/forgejs/blob/main/.github/workflows/website.yml) | `lint`, `scan`, `build`, `check`, and this repository's `lint (markdown)` | Every run but the weekly one; after a release, once `publish` has attached its files (see [Documentation](#documentation)). |
| `pages` | forgejs's [`pages.yml`](https://github.com/apollogeddon/forgejs/blob/main/.github/workflows/pages.yml) | `deploy` | On `main`, once `website` passed. |
| `patch (provider)` | forgego's [`patch.yml`](https://github.com/apollogeddon/forgego/blob/main/.github/workflows/patch.yml) | `update` | Weekly, on a manual run and after a Dependabot merge: refreshes the files forgego manages and upgrades the modules govulncheck reports as vulnerable, committing to `main`. |
| `patch (webpage)` | forgejs's [`patch.yml`](https://github.com/apollogeddon/forgejs/blob/main/.github/workflows/patch.yml) | `update` | After `patch (provider)`: commits OSV-Scanner's security patches to the site's dependencies. |
| `merge` | forgego's [`merge.yml`](https://github.com/apollogeddon/forgego/blob/main/.github/workflows/merge.yml) | `review`, `squash` | A Dependabot pull request: requests a review, and squash-merges it once every check above has passed or been skipped. |

A new push to a pull request cancels its previous run. A run on `main` is never cancelled, so a release is never created without being published. The forge pipelines are called at `@main`.

## Quality and testing

The provider's tooling comes from [forgego](https://github.com/apollogeddon/forgego). Each tool is pinned in its own module under `.forgego/`, and `Taskfile.yml` runs them. Run a task with:

```bash
go tool -modfile=.forgego/task/go.mod task <name>
```

| Task | What it does |
| :--- | :--- |
| `hooks` | Installs the git hooks. Run it once per clone. |
| `lint` | Formats the code, then lints it and fixes what it can. |
| `format` | Formats the code. |
| `type` | Runs `go vet` on every package. |
| `test` | Runs the unit tests with the race detector and coverage, against OpenTofu. |
| `test:acc` | Runs the acceptance tests against a running gateway, with OpenTofu. |
| `build` | Builds the provider into `dist/`. |
| `release:snapshot` | Builds every release binary into `dist/` without publishing or signing. |
| `security` | Reports known vulnerabilities in code the provider calls (govulncheck). |
| `sync`, `sync-check` | Refreshes, or checks, the files forgego manages. |

The git hooks format staged Go files, check that forgego's files are current and run shellcheck on staged scripts (when it is installed) before a commit, lint before a push, and check commit messages against Conventional Commits.

[`provider.yml`](./workflows/provider.yml) runs:

- **lint**: a check that forgego's files are current (on a Dependabot pull request, they are refreshed instead, and the patch pipeline commits them), `go mod tidy -diff`, the golangci-lint format check and lint (forgego's base configuration merged with `.golangci.local.yml` into `.golangci.yml`), and shellcheck on every shell script.
- **scan**: OSV-Scanner (`osv-scanner.toml` lists the findings accepted until a set date, each with its reason; a high or critical finding fails a pull request and is annotated elsewhere), govulncheck, and Trivy on the provider's Go modules, which fails on high or critical vulnerabilities.
- **test**: `task test`.
- **test (gateway)**: the acceptance tests (see [Acceptance tests](#acceptance-tests)).
- **build**: `task build`.

## Acceptance tests

The `provider / test (gateway)` stage tests the provider against a real gateway:

- **Gateway**: starts an Ignition gateway (`bootstrap/Dockerfile`) from `docker-compose.yml`, which restores the seed backup in `assets/ignition.gwbk`.
- **Readiness**: waits until the gateway responds.
- **Tests**: runs `task test:acc`, which creates, reads, updates and deletes each resource through the live API with OpenTofu.

The unit tests, the acceptance tests and the documentation build all install OpenTofu with `.github/scripts/install-tofu.sh`, which verifies the download against the release's checksums.

## Release process

- **Version**: `version / release` runs forgego's release-please, which manages version bumps and `CHANGELOG.md` from Conventional Commits, and requests a review of the release pull request it opens. When that pull request merges, it creates the release as a draft.
- **Checks**: the `provider` pipeline runs again on the release.
- **Publish**: `publish / assets` runs GoReleaser, which builds the provider in the layout the Terraform Registry expects (`.goreleaser.yaml`): a zip per platform, a `SHA256SUMS` file signed with the release GPG key, and the provider manifest (`terraform-registry-manifest.json`). With `gpg_sign: true`, forgego imports the key from the `GPG_PRIVATE_KEY` secret (its passphrase is the `PASSPHRASE` secret) and exports its fingerprint as `GPG_FINGERPRINT`, which the `signs` section of `.goreleaser.yaml` reads.
- **Site**: `website` then rebuilds the documentation site, and `pages` deploys it, so its provider mirror lists the new release.

release-please creates each release as a draft; GoReleaser attaches the files and then publishes it, so the flow also works with immutable releases. Releases are published on GitHub only: the provider is not yet listed on the OpenTofu or Terraform registry. To try the build locally, run `task release:snapshot`, which skips signing.

## Documentation

The `website` pipeline builds the [Astro](https://astro.build/) documentation site in `webpage/` with forgejs's [`website.yml`](https://github.com/apollogeddon/forgejs/blob/main/.github/workflows/website.yml) and `working_directory: webpage`:

- **lint (markdown)**: a stage of this repository's own, in `.index.yml`, which lints every Markdown file in the repository with `markdownlint-cli2`.
- **Generated content**: before each install, forgejs installs Go and runs the pre-build command, which generates the provider reference with `tfplugindocs` from the schema OpenTofu reports and from `examples/`, copies it into the site with `migrate-docs.sh`, and writes the provider network mirror (`generate-mirror.sh`) from the releases.
- **lint** and **scan**: Biome and the type check; Gitleaks over the whole repository and OSV-Scanner on the site's dependencies.
- **build** and **check**: builds the site, then checks its internal links with lychee and every page against WCAG 2.2 A and AA with axe, so a broken or inaccessible site fails the pull request.
- **Deploy**: on `main`, `pages / deploy` publishes the site to GitHub Pages.

## Configuration

| File | Purpose |
| :--- | :--- |
| `release.json` | release-please configuration. |
| `.release.json` | release-please manifest: the current version. |
| `dependabot.yml` | Keeps GitHub Actions, Go modules and npm packages up to date, with a 3-day cooldown. |
