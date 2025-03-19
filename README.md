# Aidy

Aidy is a command-line tool that generates GitHub pull request commands with AI-generated titles and body messages.

## Features

- Generate concise and structured pull request titles and bodies using OpenAI.
- Retrieve current branch name and diffs from Git.
- Mock implementations for testing purposes.

## Installation

### Prerequisites

- Go 1.24.1 or later installed on your system.
- OpenAI API key stored in `~/.aidy.conf.yml`.

### Build and Install

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

## Usage

Run the `aidy` command to generate a pull request command:

```bash
./aidy
```

This will output a GitHub pull request command with an AI-generated title and body.

## Configuration

Ensure you have a configuration file at `~/.aidy.conf.yml` with your OpenAI API key:

```yaml
openai-api-key: your_openai_api_key_here
```

## License

This project is licensed under the MIT License.
