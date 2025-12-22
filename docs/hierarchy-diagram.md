# Repository Hierarchy (Conceptual)

```text
dev-setup/
├── setup.sh                # macOS bootstrap entrypoint (version-aware)
├── README.md               # Quickstart + post-install steps
├── AGENTS.md               # Contributor rules and workflow
├── docs/
│   ├── guide.md            # Detailed reference (flow, state, policies)
│   └── hierarchy-diagram.md# This file: visual hierarchy map
├── tests/
│   └── test_setup.sh       # Stubbed, no-network regression script
├── zsh/
│   └── dev-setup.zsh       # Standard Zsh profile snippet
└── runtime state (created after running setup.sh)
    └── ~/.local/share/dev-setup/
        ├── version                     # Installed version marker
        ├── POST_INSTALL.txt            # Manual follow-ups
        ├── flutter-wrapper/            # Wrapper repo clone
        ├── git-config/                 # Git settings repo clone
        └── zsh-plugins/                # All cloned Zsh plugins
            ├── zsh-completions/
            ├── zsh-syntax-highlighting/
            ├── zsh-autosuggestions/
            ├── zsh-history-substring-search/
            ├── zsh-interactive-cd/
            └── zsh-you-should-use/
```
