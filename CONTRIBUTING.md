# Contributing to Golwarc

Thank you for your interest in contributing to Golwarc! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/golwarc.git
   cd golwarc
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Running Linter

```bash
make lint
```

### Code Formatting

```bash
make fmt
```

## Pull Request Process

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes and commit with clear messages
3. Ensure all tests pass and linter shows no issues
4. Push to your fork and open a Pull Request

## Code Style

- Follow standard Go conventions
- Add godoc comments for all public functions
- Use interfaces for testability
- Handle errors explicitly

## Reporting Issues

- Use GitHub Issues
- Include Go version and OS
- Provide reproduction steps

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
