# Contributing

Thanks for contributing. Keep changes focused and consistent with the existing thin-layer package structure.

## Quick Checklist

1. Run tests: `make test`
2. Format code: `gofmt -w .`
3. Run lint: `golangci-lint run ./...`
4. Update docs if behavior or structure changes

## Commit Style

Use Conventional Commits where possible:

- `feat:` new features
- `fix:` bug fixes
- `chore:` tooling or maintenance
- `docs:` documentation changes

## Project Structure

- `main.go` is the entry point.
- `internal/` holds application code — see `CLAUDE.md` for package responsibilities.
