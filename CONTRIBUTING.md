# Contributing to Safrochain Node

Thank you for considering contributing to Safrochain! This document outlines the process for contributing to this repository.

## Code of Conduct

This project and everyone participating in it is governed by the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository and clone your fork.
2. Add the upstream remote: `git remote add upstream https://github.com/Safrochain-Org/safrochain-node`
3. Create a feature branch from `main`: `git checkout -b feat/your-feature-name`
4. Install dependencies: `make install`

## Development Workflow

1. Make your changes on the feature branch.
2. Run tests: `make test`
3. Format code: `make format`
4. Run lint: `make lint`
5. Commit following [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` - new feature
   - `fix:` - bug fix
   - `docs:` - documentation only
   - `chore:` - maintenance, tooling, dependencies
   - `test:` - adding or updating tests
   - `refactor:` - code change that neither fixes a bug nor adds a feature

## Pull Request Process

1. Ensure all tests pass: `make test`
2. Update documentation if your changes affect public APIs.
3. Update `CHANGELOG.md` if your changes are significant.
4. Submit the PR against the `main` branch.
5. Ensure the PR description clearly describes the problem and solution, and references related issues.

## Coding Standards

- **Go**: Follow [Effective Go](https://go.dev/doc/effective_go) and use `gofmt`.
- **Protobuf**: Follow [Google's Protocol Buffer Style Guide](https://developers.google.com/protocol-buffers/docs/style).
- Keep functions focused and small; write godoc comments for exported symbols.

## Testing

- Write unit tests for new functionality.
- Run the full test suite before submitting: `make test`
- For Cosmos SDK module changes, add integration tests.

## Documentation

- Add godoc comments to all exported functions, types, and constants.
- Keep module READMEs up-to-date.
- Document breaking changes in CHANGELOG.md.

## Questions?

Open a [GitHub Discussion](https://github.com/Safrochain-Org/safrochain-node/discussions) or create an issue.
