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

Create a configuration file in your home directory `~/.aidy.conf.yml` with the following minimal content:

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

Only `deepseek` and `openai` providers are supported. Verify the configuration with:

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

To see all available commands, run:

```bash
aidy help
```

## License

This project is licensed under the [MIT](LICENSE.txt) License.
