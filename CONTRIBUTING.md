# Contributing to cssgen

Thank you for your interest in contributing!

## Reporting Issues

When reporting bugs, please include:
- cssgen version (`cssgen -version` or commit hash)
- Go version (`go version`)
- Operating system
- Minimal reproduction case
- Expected vs actual behavior

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yacobolo/cssgen.git
   cd cssgen
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   task test
   # or
   go test ./...
   ```

4. Run linter:
   ```bash
   task lint
   # or
   golangci-lint run
   ```

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure tests pass (`task check`)
6. Commit with clear messages
7. Push to your fork
8. Open a Pull Request

## Code Style

- Follow standard Go formatting (`go fmt`)
- Add comments for exported functions/types
- Keep functions focused and testable
- Use table-driven tests where appropriate

## Questions?

Open an issue for discussion before starting major work.
