# ============================================================
# Codex Rules – Global Engineering & Code Quality Standards
# ============================================================

These rules are mandatory. All code generation, modification,
and commits must comply fully. Violations are not allowed.

------------------------------------------------------------
FILES TO ALWAYS CONSIDER
------------------------------------------------------------

Before creating any commit or pull request, always review and
comply with:

- <ROOT>/.git/hooks/commit-msg
- <ROOT>/.github/workflows/pr_checks.yml

If generated code does not pass these validations, it is NOT
ready to be committed.

------------------------------------------------------------
SUPPORTED LANGUAGES
------------------------------------------------------------

Primary languages used in this repository:

- JavaScript
- TypeScript
- Python

All rules apply equally to these languages unless stated
otherwise.

------------------------------------------------------------
PACKAGE MANAGER POLICY (STRICT)
------------------------------------------------------------

Only the defined package managers are allowed.

Allowed package managers:
- Python: uv
- Node.js (JavaScript / TypeScript): pnpm

Rules:
- Do NOT use any other package manager.
- Do NOT generate commands, configs, or lockfiles for:
  - Python: pip, pipenv, poetry
  - Node.js: npm, yarn, bun
- If a different package manager is detected or requested:
  - Explicitly warn that it is NOT allowed
  - Guide the user to the correct package manager
- Never mix package managers in the same project or commit.
- Lockfiles must match the package manager:
  - uv lockfile for Python
  - pnpm-lock.yaml for Node.js

------------------------------------------------------------
DOCUMENTATION RULES (MANDATORY)
------------------------------------------------------------

Documentation is required even for simple code.

File-level documentation:
- Every file must start with a detailed DocString or header
  comment explaining:
  - Purpose of the file
  - Problem it solves
  - Architectural role
  - How and when to use it
  - Important design decisions or assumptions

Function / method documentation:
- Every function or method must include a DocString defining:
  - What it does
  - Why it exists
  - Parameters (name, type, purpose)
  - Return value
  - Example usage where meaningful
  - Edge cases or constraints if applicable

Inline comments:
- Use clear comments to explain WHY logic exists
- Especially required for complex or non-obvious logic

------------------------------------------------------------
LINTING & TESTING (NON-NEGOTIABLE)
------------------------------------------------------------

Before committing ANY code:

- Run all relevant linters
  - JS/TS: ESLint, Prettier (as configured)
  - Python: Ruff / Flake8 / Black / pylint (as configured)
- Run all applicable test suites
  - Unit tests
  - Integration tests if present
- Fix all linting errors and test failures
- Do NOT suppress errors unless explicitly allowed

Never commit:
- Failing tests
- Lint violations
- Breaking changes

------------------------------------------------------------
COMMIT & PR STANDARDS
------------------------------------------------------------

Commits must be:
- Small
- Focused
- Logically grouped

Commit messages:
- Must pass commit-msg hook validation

Pull requests:
- Clearly describe what changed and why
- Reference relevant files or architectural concerns
- Avoid unrelated or mixed changes

------------------------------------------------------------
NON-NEGOTIABLE SUMMARY
------------------------------------------------------------

Always:
- Follow commit hooks and PR workflows
- Use ONLY uv (Python) and pnpm (Node.js)
- Warn explicitly if any other package manager is used
- Add extensive file-level and method-level documentation
- Use DocStrings consistently
- Comment code clearly and intentionally
- Run linting and tests before committing
- Ensure zero breaking changes or failing checks
