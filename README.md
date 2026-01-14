# VS Code Extension Manager

A powerful CLI tool for managing VS Code extensions across multiple profiles and machines. Save, sync, and restore your extension configurations with ease.

## Features

- 📦 **Profile Management**: Create and manage multiple extension profiles
- 🔄 **Sync Extensions**: Export and import extension lists across machines
- 🎯 **Bulk Operations**: Install, uninstall, enable, and disable extensions in batches
- 📊 **Smart Recommendations**: Get intelligent extension suggestions based on your workspace
- 🔍 **Dependency Resolution**: Automatically handle extension dependencies
- 📈 **Usage Analytics**: Track and optimize your extension usage
- 🌐 **Team Collaboration**: Share extension profiles with your team
- 🔒 **Conflict Detection**: Identify and resolve extension conflicts

## Installation

```bash
npm install -g vscode-ext-manager
```

Or with yarn:

```bash
yarn global add vscode-ext-manager
```

## Quick Start

```bash
# Initialize extension manager
vext init

# Save your current extensions to a profile
vext save my-profile

# List all profiles
vext list

# Load a profile
vext load my-profile

# Sync with remote storage
vext sync push
vext sync pull
```

## Usage

### Profile Management

```bash
# Create a new profile from current extensions
vext save <profile-name>

# Load a profile (installs missing extensions)
vext load <profile-name>

# List all profiles
vext list

# Delete a profile
vext delete <profile-name>

# Show profile details
vext show <profile-name>
```

### Extension Operations

```bash
# Install extensions from a list
vext install extension1 extension2 extension3

# Uninstall extensions
vext uninstall extension1 extension2

# Update all extensions
vext update

# Search for extensions
vext search "python"
```

### Synchronization

```bash
# Push profiles to remote storage
vext sync push

# Pull profiles from remote storage
vext sync pull

# Configure sync settings
vext config set sync.provider github
vext config set sync.repo username/repo
```

### Recommendations

```bash
# Get extension recommendations for current workspace
vext recommend

# Get recommendations based on file types
vext recommend --workspace-scan

# Get popular extensions in a category
vext recommend --category "Programming Languages"
```

### Analytics

```bash
# Show extension usage statistics
vext stats

# Identify unused extensions
vext stats --unused

# Show resource usage by extension
vext stats --resources
```

### Team Collaboration

```bash
# Export profile for sharing
vext export my-profile --output team-extensions.json

# Import a shared profile
vext import team-extensions.json --name team-profile

# Generate team profile URL
vext share my-profile
```

## Configuration

Configuration is stored in `~/.vscode-ext-manager/config.json`:

```json
{
  "sync": {
    "enabled": true,
    "provider": "github",
    "repo": "username/vscode-extensions",
    "autoSync": false
  },
  "profiles": {
    "default": "work"
  },
  "recommendations": {
    "enabled": true,
    "autoInstall": false
  },
  "updates": {
    "autoCheck": true,
    "autoInstall": false
  }
}
```

### Sync Providers

- **GitHub**: Store profiles in a GitHub repository
- **GitLab**: Store profiles in a GitLab repository
- **Bitbucket**: Store profiles in a Bitbucket repository
- **Local**: Store profiles locally (default)

## Profile Format

Profiles are stored as JSON files with the following structure:

```json
{
  "name": "my-profile",
  "description": "My development setup",
  "created": "2025-01-13T12:00:00Z",
  "updated": "2025-01-13T12:00:00Z",
  "extensions": [
    {
      "id": "ms-python.python",
      "version": "2024.0.0",
      "enabled": true
    }
  ],
  "settings": {
    "autoUpdate": true,
    "syncSettings": false
  }
}
```

## Advanced Features

### Custom Extension Sources

```bash
# Add a custom extension marketplace
vext source add my-marketplace https://marketplace.example.com

# Install from custom source
vext install extension-name --source my-marketplace
```

### Extension Presets

```bash
# Apply a preset (predefined extension bundle)
vext preset apply web-development
vext preset apply data-science
vext preset apply devops

# List available presets
vext preset list
```

### Workspace Integration

```bash
# Save workspace-specific extensions
vext workspace save

# Load workspace extensions
vext workspace load

# Generate .vscode/extensions.json
vext workspace export
```

## Roadmap

See our [project roadmap](vscode_ext_manager_oss_plan.md) for upcoming features and development plans.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- Code of Conduct
- Development setup
- Submitting pull requests
- Coding standards
- Testing requirements

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- 📖 [Documentation](https://docs.vscode-ext-manager.dev)
- 🐛 [Issue Tracker](https://github.com/yourusername/vscode-ext-manager/issues)
- 💬 [Discussions](https://github.com/yourusername/vscode-ext-manager/discussions)
- 🔔 [Changelog](CHANGELOG.md)

## Acknowledgments

Built with ❤️ by the open-source community. Special thanks to all our contributors!

---

**Note**: This tool is not affiliated with or endorsed by Microsoft or Visual Studio Code.
