# macOS Dev Setup Bootstrap

Automates a fast baseline dev environment for macOS with Homebrew, Node + pnpm, Python via uv, Flutter wrapper, Git helpers, Zed, Starship prompt + Hack Nerd Font, AI CLIs (when available via brew), and a couple of Zsh plugins.

## What gets installed
- Homebrew (if missing) with formulas: git, node, pnpm, git-lfs, wget
- uv via the official installer
- Zed editor (brew cask)
- Starship prompt (brew) with package.display_private enabled
- Hack Nerd Font (brew cask from homebrew/cask-fonts tap)
- Flutter wrapper from `BadRat-in/flutter-wrapper` with a `flutterw` shim in `~/.local/bin`
- Git config repo from `BadRat-in/git-config` (left for you to review/apply)
- Zsh plugins: `zsh-interactive-cd` and `zsh-you-should-use`
- Best-effort Homebrew installs for Codex, Gemini CLI, and Claude Code (skipped gracefully if formulas/casks missing)

## Usage
```bash
./setup.sh
```

The script targets macOS only. It will prompt during the first Homebrew install. State is stored under `~/.local/share/dev-setup` and shims under `~/.local/bin`.

## After running
- Ensure `~/.local/bin` is on your `PATH` (add to `~/.zprofile`)
- Run `flutterw doctor` to download Flutter via the wrapper
- Switch terminal font to **Hack Nerd Font** (installed). Example iTerm2 path: Profiles → Text → Font.
- Enable Starship in `~/.zshrc` after PATH setup: `eval "$(starship init zsh)"` (nerd-font-symbols preset applied; package.display_private already set).
- Source the Zsh plugins in `~/.zshrc`:
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-completions/src"`
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"`
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"`
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-history-substring-search/zsh-history-substring-search.zsh"`
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-interactive-cd/zsh-interactive-cd.zsh"`
  - `source "${HOME}/.local/share/dev-setup/zsh-plugins/zsh-you-should-use/zsh-you-should-use.zsh"`
- Review `~/.local/share/dev-setup/git-config` before applying to `~/.gitconfig`
- If Codex, Gemini CLI, or Claude Code were installed by brew, run their auth/init flows (`codex login`, `gemini login`, `claude auth`)
- Installed dev-setup version is recorded at `~/.local/share/dev-setup/version`.

## Notes
- Package manager policy: pnpm for Node.js, uv for Python (per repository rules). npm/yarn/pip/poetry are intentionally not used.

## Tests (safe, no-network)
- Run stubbed regression checks without touching your environment:
  ```bash
  bash tests/test_setup.sh
  ```
  This uses temporary HOME and stubbed tools to verify version recording and Starship config patching.

## Zsh profile template
- To replicate the standard environment on any macOS host, source the provided profile snippet from your `~/.zshrc`:
  ```zsh
  source /path/to/repo/zsh/dev-setup.zsh
  ```
- The snippet configures Homebrew shellenv, Starship, aliases, PATH, and Zsh plugins using the locations installed by `setup.sh` (prefers `~/.local/share/dev-setup/...` for plugins).
