# Shell Completion for devtools-sync

The `devtools-sync` CLI includes built-in shell completion support for bash, zsh, fish, and PowerShell.

## Features

- Tab completion for all commands and subcommands
- Intelligent completion for profile names when using `profile load`
- Completion for config keys when using `config set`
- Context-aware suggestions with descriptions

## Installation

### Bash

#### Linux
```bash
devtools-sync completion bash | sudo tee /etc/bash_completion.d/devtools-sync
```

#### macOS
```bash
devtools-sync completion bash > $(brew --prefix)/etc/bash_completion.d/devtools-sync
```

#### Per-session (temporary)
```bash
source <(devtools-sync completion bash)
```

### Zsh

#### Oh-My-Zsh
```bash
devtools-sync completion zsh > ~/.oh-my-zsh/completions/_devtools-sync
```

#### Without Oh-My-Zsh
```bash
devtools-sync completion zsh > "${fpath[1]}/_devtools-sync"
```

#### Per-session (temporary)
```bash
source <(devtools-sync completion zsh)
```

**Note:** After installation, you may need to start a new shell session or run:
```bash
compinit
```

### Fish

```bash
devtools-sync completion fish > ~/.config/fish/completions/devtools-sync.fish
```

### PowerShell

```powershell
devtools-sync completion powershell | Out-String | Invoke-Expression
```

To make it permanent, add the above line to your PowerShell profile:
```powershell
notepad $PROFILE
```

## Usage Examples

Once installed, you can use Tab completion:

```bash
# Complete commands
devtools-sync <TAB>
# Shows: completion  config  help  init  profile  sync  version

# Complete profile subcommands
devtools-sync profile <TAB>
# Shows: list  load  save

# Complete profile names when loading
devtools-sync profile load <TAB>
# Shows: work-setup  personal  minimal

# Complete config keys
devtools-sync config set <TAB>
# Shows:
#   server.url             Server URL for syncing profiles
#   profiles.directory     Directory for storing local profiles
#   logging.level          Logging level (info, debug, error)
```

## Troubleshooting

### Completion not working

1. **Verify installation:** Make sure the completion script is in the correct location
2. **Restart shell:** Open a new terminal window or run `exec $SHELL`
3. **Check completion is loaded:** For bash, run `complete -p devtools-sync`
4. **Reinstall:** Remove the completion file and reinstall

### Bash: command not found

Make sure bash-completion is installed:
- **Ubuntu/Debian:** `sudo apt install bash-completion`
- **macOS:** `brew install bash-completion@2`

### Zsh: command not found

Make sure you have completion enabled in your `.zshrc`:
```bash
autoload -Uz compinit
compinit
```

## Advanced Configuration

### Custom Completion Cache

The completion system automatically caches profile names for better performance. The cache is refreshed each time you complete.

### Disable File Completion

By default, most completions disable file completion. If you need to complete file paths for a specific command, use the standard shell file completion (Tab twice).

## Development

To test completion during development:

```bash
# Build the CLI
go build -o devtools-sync ./cmd

# Test bash completion
./devtools-sync completion bash > /tmp/test-completion.sh
source /tmp/test-completion.sh

# Test completion
./devtools-sync profile load <TAB>
```

## See Also

- [Cobra Shell Completions](https://github.com/spf13/cobra/blob/main/shell_completions.md)
- [Bash Completion Guide](https://github.com/scop/bash-completion)
- [Zsh Completion Guide](https://zsh.sourceforge.io/Doc/Release/Completion-System.html)
