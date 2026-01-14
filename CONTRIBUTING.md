# Contributing to VS Code Extension Manager

Thank you for your interest in contributing to VS Code Extension Manager! We welcome contributions from the community and are excited to have you here.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. By participating, you are expected to uphold this code.

### Our Standards

- **Be respectful**: Treat everyone with respect and kindness
- **Be collaborative**: Work together and help each other
- **Be inclusive**: Welcome newcomers and diverse perspectives
- **Be constructive**: Provide helpful feedback and accept it graciously
- **Be professional**: Keep discussions focused and productive

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Spam or excessive self-promotion
- Sharing others' private information without permission

## Getting Started

### Prerequisites

- Node.js 18.x or higher
- npm 9.x or higher (or yarn 1.22.x)
- Git
- VS Code (for testing)

### First Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/vscode-ext-manager.git
   cd vscode-ext-manager
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/vscode-ext-manager.git
   ```

4. **Install dependencies**:
   ```bash
   npm install
   ```

5. **Build the project**:
   ```bash
   npm run build
   ```

6. **Run tests**:
   ```bash
   npm test
   ```

## Development Setup

### Project Structure

```
vscode-ext-manager/
├── src/
│   ├── commands/       # CLI command implementations
│   ├── core/           # Core business logic
│   ├── services/       # External service integrations
│   ├── utils/          # Utility functions
│   └── types/          # TypeScript type definitions
├── tests/
│   ├── unit/           # Unit tests
│   ├── integration/    # Integration tests
│   └── fixtures/       # Test data
├── docs/               # Documentation
└── scripts/            # Build and deployment scripts
```

### Development Workflow

1. **Create a branch** for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit frequently:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

3. **Keep your branch updated**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

4. **Run tests** before pushing:
   ```bash
   npm test
   npm run lint
   ```

5. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

### Useful Commands

```bash
# Development build with watch mode
npm run dev

# Run linter
npm run lint

# Fix linting issues
npm run lint:fix

# Run type checking
npm run type-check

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Build for production
npm run build

# Test CLI locally
npm link
vext --help
```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**Include in your bug report:**
- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Environment details (OS, Node version, VS Code version)
- Relevant logs or error messages
- Screenshots if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues.

**Include in your enhancement suggestion:**
- Clear, descriptive title
- Detailed description of the proposed feature
- Use cases and benefits
- Potential implementation approach (optional)
- Examples from similar tools (optional)

### Contributing Code

1. **Pick an issue** to work on or create one
2. **Comment on the issue** to let others know you're working on it
3. **Follow the development workflow** above
4. **Write tests** for your changes
5. **Update documentation** as needed
6. **Submit a pull request**

### Good First Issues

Look for issues labeled `good first issue` or `help wanted` - these are great starting points for new contributors.

## Coding Standards

### TypeScript Guidelines

- Use TypeScript strict mode
- Provide type annotations for function parameters and return types
- Avoid using `any` type - use `unknown` or proper types
- Use interfaces for object shapes
- Use enums or union types for fixed sets of values

### Code Style

We use ESLint and Prettier for code formatting:

```typescript
// Good
export async function loadProfile(name: string): Promise<Profile> {
  const profilePath = getProfilePath(name);
  const data = await fs.readFile(profilePath, 'utf-8');
  return JSON.parse(data) as Profile;
}

// Bad
export async function loadProfile(name) {
  const profilePath = getProfilePath(name)
  const data = await fs.readFile(profilePath, 'utf-8')
  return JSON.parse(data)
}
```

### Best Practices

- **Single Responsibility**: Each function/class should do one thing well
- **Error Handling**: Always handle errors gracefully
- **Async/Await**: Prefer async/await over callbacks or raw promises
- **Immutability**: Avoid mutating input parameters
- **Documentation**: Add JSDoc comments for public APIs
- **DRY**: Don't repeat yourself - extract common functionality

### Naming Conventions

- **Files**: kebab-case (`profile-manager.ts`)
- **Classes**: PascalCase (`ProfileManager`)
- **Functions/Variables**: camelCase (`loadProfile`)
- **Constants**: UPPER_SNAKE_CASE (`DEFAULT_CONFIG`)
- **Types/Interfaces**: PascalCase (`Profile`, `ConfigOptions`)

## Testing Guidelines

### Writing Tests

- Write tests for all new features and bug fixes
- Aim for high test coverage (>80%)
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)

```typescript
describe('ProfileManager', () => {
  describe('loadProfile', () => {
    it('should load profile from file system', async () => {
      // Arrange
      const manager = new ProfileManager();
      const profileName = 'test-profile';
      
      // Act
      const profile = await manager.loadProfile(profileName);
      
      // Assert
      expect(profile.name).toBe(profileName);
      expect(profile.extensions).toBeInstanceOf(Array);
    });
    
    it('should throw error when profile does not exist', async () => {
      // Arrange
      const manager = new ProfileManager();
      
      // Act & Assert
      await expect(
        manager.loadProfile('non-existent')
      ).rejects.toThrow('Profile not found');
    });
  });
});
```

### Test Categories

- **Unit Tests**: Test individual functions/classes in isolation
- **Integration Tests**: Test interactions between components
- **E2E Tests**: Test complete user workflows

### Running Specific Tests

```bash
# Run tests matching a pattern
npm test -- --grep "ProfileManager"

# Run a specific test file
npm test -- tests/unit/profile-manager.test.ts

# Run tests in watch mode
npm run test:watch
```

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, semicolons, etc.)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Build process or tooling changes
- **ci**: CI/CD changes

### Examples

```
feat(profile): add export functionality

Add ability to export profiles to JSON format for sharing with team members.

Closes #123
```

```
fix(sync): resolve GitHub API rate limiting issue

Implement exponential backoff when hitting rate limits.

Fixes #456
```

## Pull Request Process

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated and passing
- [ ] No merge conflicts with main branch
- [ ] Commit messages follow guidelines

### PR Title

Follow the same format as commit messages:

```
feat(profile): add export functionality
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe how you tested your changes

## Checklist
- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] No breaking changes (or documented)

## Related Issues
Closes #123
```

### Review Process

1. At least one maintainer review required
2. All CI checks must pass
3. Address review feedback
4. Squash commits if requested
5. Maintainer will merge when approved

## Community

### Getting Help

- 💬 [GitHub Discussions](https://github.com/yourusername/vscode-ext-manager/discussions)
- 🐛 [Issue Tracker](https://github.com/yourusername/vscode-ext-manager/issues)
- 📖 [Documentation](https://docs.vscode-ext-manager.dev)

### Stay Updated

- Watch the repository for updates
- Follow our [changelog](CHANGELOG.md)
- Join community discussions

## Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- Community showcase

Thank you for contributing to VS Code Extension Manager! 🎉

---

**Questions?** Feel free to open a discussion or reach out to the maintainers.
