<div align="center">

[![Test](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml/badge.svg)](https://github.com/artus-engineering/terraform-provider-n8n/actions/workflows/test.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Terraform Registry](https://img.shields.io/github/v/release/artus-engineering/terraform-provider-n8n?label=Latest%Version)](https://registry.terraform.io/providers/artus-engineering/n8n/latest)

</div>

# Terraform Provider for n8n

A Terraform provider for managing n8n resources.

**Documentation**: [Terraform Registry](https://registry.terraform.io/providers/artus-engineering/n8n/latest)

> **Note**: 🔨 This is an early and minimal provider. It does not support all n8n resources but is under development and will be extended step by step. 🔨

## Features

- **Credential Management**: Manage n8n credentials

## Development

### Prerequisites

- Go 1.25 or later
- Terraform 1.0 or later
- Make (for using the Makefile)
- [golangci-lint](https://golangci-lint.run/) (for `make lint` and `make check`)
- [pre-commit](https://pre-commit.com/) (for `make pre-commit`, optional)
- A running n8n instance (for `make testacc`, optional)

### Building and Testing

```bash
make build    # Build the provider
make test     # Run unit tests
make check    # Run all checks (lint, format, test)
```

## Releasing

The release process is automated via GitHub Actions. To create a new release:

1. **Go to the Actions tab** in your GitHub repository
2. **Select "Release Provider"** workflow from the left sidebar
3. **Click "Run workflow"** button
4. **Enter the version number** (e.g., `0.1.0` or `v0.1.0` - both formats are accepted)
5. **Click "Run workflow"** to start the release process

The GitHub Actions workflow (`.github/workflows/release.yml`) will automatically:

- Create and push a version tag (e.g., `v0.1.0`)
- Build the provider for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64, arm64)
- Generate SHA256 checksums for all binaries
- Generate a Terraform Registry manifest
- Sign checksums (if GPG keys are configured)
- Create a GitHub Release with all artifacts

The release will be available on GitHub and automatically published to the Terraform Registry.

**Note**: The workflow will fail if you try to create a release with a version that already has a tag, preventing duplicate releases.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## References

- [n8n API Documentation](https://docs.n8n.io/api/)
- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- Check out the [examples](./examples/) directory for more use cases
