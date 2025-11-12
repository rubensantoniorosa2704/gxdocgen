# Contributing to GXDocGen

Thank you for your interest in contributing to **GXDocGen**! 🎉  
Your help is greatly appreciated — whether it’s fixing bugs, improving docs, or adding new features.

---

## 🧭 Project Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/gxdocgen.git
   cd gxdocgen
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run tests:**
   ```bash
   go test ./...
   ```

4. **Build the CLI:**
   ```bash
   go build -o bin/gxdocgen ./cmd/gxdocgen
   ```

---

## 🧱 Branch Naming Convention

| Type       | Prefix Example           | Description                                  |
| ----------- | ------------------------ | -------------------------------------------- |
| Feature     | `feature/parser-improve` | For new features or enhancements.            |
| Fix         | `fix/xpz-unzip-bug`      | For bug fixes.                               |
| Docs        | `docs/readme-update`     | For documentation-only changes.              |
| Refactor    | `refactor/model-update`  | For internal refactors or optimizations.     |

---

## Code Guidelines

- Follow Go’s official [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Use clear, self-explanatory names for functions and variables.
- Keep functions small and focused.
- Write unit tests for new logic — use `testing` package.
- Run `go fmt ./...` before committing.

---

## Testing

To ensure stability, every module should have tests under the same path, e.g.:
```
internal/parser/parser_test.go
internal/xpz/xpz_test.go
```

Run all tests with:
```bash
go test ./...
```

---

## Submitting Pull Requests

1. Fork the repository and create your feature branch.
2. Write clear commit messages.
3. Add or update documentation when necessary.
4. Open a Pull Request with a concise description of your change.

---

## Code of Conduct

Please be respectful and professional in all discussions and code reviews.  
We aim to maintain a welcoming and collaborative environment.

---

Thank you for contributing to **GXDocGen** 💙