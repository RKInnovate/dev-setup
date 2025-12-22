## Claude Code Instructions

This document defines **mandatory engineering standards** that Claude must follow when generating, modifying, or committing code in this repository. These rules apply **regardless of code size or complexity**.

---

## 📁Files to Always Consider for Commits and PRs

Before creating **any commits or pull requests**, Claude must review and comply with the rules defined in the following files:

1. **`<ROOT>/.git/hooks/commit-msg*`**  
     
   * Enforces commit message format and validation rules  
   * Commit messages must pass this hook without modification

   

2. **`<ROOT>/.github/workflow/pr_checks.yml`**  
     
   * Defines CI checks, linting, tests, and validation required for PRs  
   * All generated code must pass these checks

Failure to comply with these files means the code is **not ready for commit**.

---

## 🧠 Supported Languages (Primary)

The current primary languages used in this repository are:

* **JavaScript (JS)**  
* **TypeScript (TS)**  
* **Python**

All rules below apply equally to these languages unless explicitly stated otherwise.

---

## 📦 Package Manager Policy (Strict)

**Only the defined package managers are allowed. This rule is non-negotiable.**

### Allowed Package Managers

| Language | Required Package Manager |
| :---- | :---- |
| Python | **`uv`** |
| Node.js (JS/TS) | **`pnpm`** |

### Rules

1. **Developers and AI must use only the defined package manager**  
     
   * Python → `uv`  
   * Node.js → `pnpm`

   

2. **Using any other package manager is incorrect**, including but not limited to:  
     
   * Python: `pip`, `pipenv`, `poetry`  
   * Node.js: `npm`, `yarn`, `bun`

   

3. **If a different package manager is detected or requested**:  
     
   * Claude **must explicitly warn** that this is **not allowed**  
   * Claude **must not generate commands or configuration** for the incorrect tool  
   * Claude should guide the user to the correct package manager instead

   

4. Lockfiles must match the chosen package manager:  
     
   * `uv.lock` (or equivalent uv lockfile)  
   * `pnpm-lock.yaml`

⚠️ **Never mix package managers in the same project or commit.**

---

## 🧩 Code Documentation & Commenting Rules (Mandatory)

Documentation is **not optional**, even for simple or self-explanatory code.

### 1\. File-Level Documentation

Every file **must begin with a detailed DocString or header comment** that explains:

* The **purpose of the file**  
* The **problem it solves**  
* The **role it plays in the overall architecture**  
* How and **when it should be used**  
* Any important design decisions or assumptions

---

### 2\. Function / Method Documentation

Every function or method **must include a DocString** that clearly defines:

* **What the function does**  
* **Why it exists**  
* **Parameters** (name, type, purpose)  
* **Return value**  
* **Example usage** (where meaningful)  
* Edge cases or constraints (if applicable)

This applies even to:

* Utility functions  
* Private/internal helpers  
* Small or seemingly obvious methods

---

### 3\. Inline Comments

* Use **clear, concise inline comments** to explain:  
    
  * Non-obvious logic  
  * Complex conditions  
  * Important side effects


* Comments should explain **why**, not just **what**

---

## 🧪 Linting & Testing (Non-Negotiable)

**Claude must always ensure code quality before committing any piece of code.**

### Required Before Any Commit or PR:

1. **Run all relevant linters**  
     
   * JS/TS (e.g., ESLint, Prettier)  
   * Python (e.g., Ruff, Flake8, Black, pylint — as configured)

   

2. **Run all applicable test suites**  
     
   * Unit tests  
   * Integration tests (if present)  
   * Any repo-specific test commands

   

3. **Fix all linting errors and test failures**  
     
   * Do not suppress errors unless explicitly allowed  
   * Do not commit failing or flaky tests

⚠️ **Never commit code that introduces breaking changes, fialeing tests, or lint violations.**

---

## 🔒 Commit & PR Quality Standards

* Commits must be:  
    
  * **Small**  
  * **Focused**  
  * **Logically grouped**


* Commit messages must comply with the `commit-msg` hook rules  
    
* PRs should:  
    
  * Clearly describe **what changed and why**  
  * Reference relevant files, modules, or architectural concerns  
  * Avoid unrelated changes

---

## ✅ Summary of Non-Negotiable Rules

Claude **must always**:

* Follow commit hooks and PR workflow rules  
* Use **only `uv` for Python** and **`pnpm` for Node.js**  
* Warn explicitly if any other package manager is used or suggested  
* Add extensive file-level and method-level documentation  
* Use DocStrings consistently across JS, TS, and Python  
* Comment code clearly and intentionally  
* Run linting and test cases before committing  
* Ensure zero breaking changes or failing checks
