# Contributing to obx

Thank you for your interest in contributing to `obx`! We welcome pull requests from everyone. By participating in this project, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

1. **Fork the Repository**: Clone your fork locally.
   ```bash
   git clone https://github.com/YOUR-USERNAME/obx.git
   cd obx
   ```

2. **Set Up the Environment**: `obx` is built with Go. You should have Go 1.21+ installed. We highly recommend using `mise` to manage tasks.
   ```bash
   mise install
   ```

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

## Development Workflow

1. **Create a Branch**: Create a descriptive branch name from `develop`.
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**: Implement your feature or bug fix. `obx` uses a multiplexed MCP tool architecture located in `internal/vault/`. If adding new tools, ensure you register them in `internal/server/server.go`.

3. **Run Checks**: Before committing, ensure your code passes linting and tests.
   ```bash
   mise run check
   ```
   Or manually:
   ```bash
   golangci-lint run ./...
   go test -race ./...
   ```

4. **Commit Formatting**: We use conventional commits. Try to start your commit message with `feat:`, `fix:`, `docs:`, or `chore:`.

5. **Push and Open a PR**: Push to your fork and submit a Pull Request targeting the `develop` branch of the main repository.

## Adding Documentation
Documentation is built with Astro/Starlight and lives in the `docs/` directory.

```bash
cd docs
pnpm install
pnpm dev
```
If you introduce a new CLI command or MCP action, you **must** include corresponding updates to the `docs/` folder in your Pull Request.
