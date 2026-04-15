# vaultdiff

> CLI tool to diff and audit changes between HashiCorp Vault secret versions across environments.

---

## Installation

```bash
go install github.com/yourusername/vaultdiff@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultdiff.git
cd vaultdiff
go build -o vaultdiff .
```

---

## Usage

Compare two versions of a secret within the same Vault path:

```bash
vaultdiff --path secret/myapp/config --v1 3 --v2 5
```

Diff secrets across environments:

```bash
vaultdiff --src secret/staging/myapp --dst secret/production/myapp
```

Output audit log to a file:

```bash
vaultdiff --path secret/myapp/config --v1 1 --v2 2 --audit --output audit.log
```

### Flags

| Flag | Description |
|------|-------------|
| `--path` | Vault secret path |
| `--v1` | First version to compare |
| `--v2` | Second version to compare |
| `--src` | Source environment path |
| `--dst` | Destination environment path |
| `--audit` | Enable audit logging |
| `--output` | Write output to file |
| `--token` | Vault token (defaults to `VAULT_TOKEN` env var) |
| `--addr` | Vault address (defaults to `VAULT_ADDR` env var) |

---

## Requirements

- Go 1.21+
- HashiCorp Vault with KV v2 secrets engine

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)