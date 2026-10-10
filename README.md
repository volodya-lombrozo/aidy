# Aidy

[![codecov](https://codecov.io/gh/volodya-lombrozo/aidy/branch/main/graph/badge.svg)](https://codecov.io/gh/volodya-lombrozo/aidy)

Aidy is a command-line tool designed to enhance your [GitHub flow](https://docs.github.com/en/get-started/using-github/github-flow) with AI-powered assistance.
It helps generate commit messages, issues, pull requests, releases, and more.

## Installation

### Releases

Download the latest stable version from the [releases page](https://github.com/volodya-lombrozo/aidy/releases). Pre-built binaries are available for MacOS, Windows, and Linux.

### Using Go

If you have Go 1.24.1 or later installed, you can run:

```bash
go install github.com/volodya-lombrozo/aidy@latest
```

To install a specific version, use:

```bash
go install github.com/volodya-lombrozo/aidy@v0.1.0
```

### From Sources 

You need to have Go 1.24.1 or later installed on your system.

1. Clone the repository:

   ```bash
   git clone https://github.com/volodya-lombrozo/aidy.git
   cd aidy
   ```

2. Build the binary:

   ```bash
   go build -o aidy
   ```

3. (Optional) Install the binary to your `$GOPATH/bin`:

   ```bash
   go install
   ```

## Configuration

The quickest way to get started is to run:

```bash
aidy init
```

This walks you through choosing an AI provider, a model, and an API key, then writes `~/.aidy.conf.yml` for you. It also asks for a GitHub API key, which you can skip if you don't have one yet. If a configuration file already exists, you'll be asked before it's overwritten.

Alternatively, create the configuration file in your home directory `~/.aidy.conf.yml` yourself with the following minimal content:

```yaml
default-model: deepseek

api-keys:
  github: <github-key>
  deepseek: <deepseek-key>

models:
  deepseek:
    provider: deepseek
    model-id: deepseek-chat
```

For OpenAI:

```yaml
default-model: 4o

api-keys:
  github: <github-key>
  openai: <openai-key>

models:
  4o:
    provider: openai
    model-id: gpt-4o 
```

For Anthropic:

```yaml
default-model: claude-sonnet

api-keys:
  github: <github-key>
  anthropic: <anthropic-key>

models:
  claude-sonnet:
    provider: anthropic
    model-id: claude-sonnet-4-6
```

Supported providers: `deepseek`, `openai`, and `anthropic`. Verify the configuration with:

```bash
aidy conf
```

### Commit

Make a commit with a human-readable message:

```bash
aidy commit
```

This command will stage all current changes in your Git repository, generate a readable message based on the changes made, and create a commit.
Example output:

```
Commit created with message: 'docs(#209): Update README with installation and configuration details'
```

or if you a using JIRA-sytle branches:

```
Commit created with message: 'docs(PROJ-209): Update README with installation and configuration details'
```

Shorter version is `aidy ci`.

### Pull Request

You can create a pull request for the implemented feature using:

```bash
aidy pull-request
```

or the shorter version:

```bash
aidy pr
```

This command will generate a meaningful title and body for your pull request and provide the corresponding `gh` command:

```
gh pr create
  --title "docs(#209): Update README to reflect current features and installation"
  --body "This PR updates the `README.md` to better reflect the current functionality of `aidy`, including installation options, configuration details, and usage examples.

Closes #209"
```

You can then run this command to create the pull request. For that, you will need the [GitHub CLI](https://cli.github.com)

### Issue

You can also create an issue quickly using:

```bash
aidy issue "update the documentation regarding all recent changes"
```

or the short form:

```bash
aidy i "update the documentation regarding all recent changes"
```

This command generates a title and body for the issue, suggests appropriate labels, and prints a `gh` command:

```
gh issue create
  --title "Update documentation to reflect recent changes"
  --body "The `README.md` and other documentation files need updating to reflect recent changes in the codebase, including new features, API modifications, and deprecated functionality."
  --label "documentation,enhancement"
  --repo volodya-lombrozo/aidy
```

### Start

Create a branch for an issue:

```bash
aidy start 42
```

For a GitHub issue, the branch name is generated from the issue description, e.g. `42-fix-login-error`.

You can also pass a Jira issue key or link:

```bash
aidy start https://company.atlassian.net/browse/PROJ-123
```

The branch prefix depends on the Jira issue type: `Bug` becomes `fix/PROJ-123`, `Test` becomes `test/PROJ-123`, and any other type becomes `feat/PROJ-123`. To let `aidy` read issue types, add a Jira connection to `.aidy.conf.yml`:

```yaml
jira:
  url: https://company.atlassian.net
  email: <your-email>
  token: <jira-api-token>
```

Leave out `email` to use a personal access token (Jira Data Center). Without a Jira connection, `aidy` always uses `feat/PROJ-123`.

### Release

`aidy` also helps generate releases. It creates a new tag with release notes (AI-generated) and bumps the version number according to the [SemVer](https://semver.org/) specification.

Patch release:

```bash
aidy release patch
```

Updates version from `0.1.0` to `0.1.1`.

Minor release:

```bash
aidy release minor
```

Updates version from `0.1.0` to `0.2.0`.

Major release:

```bash
aidy release major
```

Updates version from `0.1.0` to `1.0.0`.

> **Note:** This command only creates a tag with release notes. To push it to the remote repository, run:

```bash
git push --tags
```

### Last

If a generated command fails (network outage, expired token, etc.), you don't have to generate and edit it again. Repeat the last command, including your edits:

```bash
aidy last
```

### Stats

Every run is recorded locally in `~/.aidy/metrics.json` with the command name, a timestamp, the outcome, and the duration. The file never leaves your machine. See which commands you use and which fail often:

```bash
aidy stats
```

Example output:

```
command        runs   failed   last used
commit         412    3        2026-09-29
pull-request   97     11       2026-09-28
squash         0      -        never
```

To stop recording, add the following line to `.aidy.conf.yml`:

```yaml
metrics: false
```

To see all available commands, run:

```bash
aidy help
```

## License

This project is licensed under the [MIT](LICENSE.txt) License.
