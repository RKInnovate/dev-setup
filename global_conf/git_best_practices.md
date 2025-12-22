# **GitHub Naming & Commit Standards — Unified & Consistent**

**Scope:** Branch names, commit message format, PR titles, issue references, and versioning references across all RK Innovate repos.  
 **Goal:** Have one consistent pattern that is human-readable, machine-enforceable, and interoperable with CI/release tools.  
---

## **1\) Global principles**

* Use **short type tokens** (feat, fix, chore, docs, refactor, perf, test, ci, style, deploy, hotfix) consistently for branches and commits.  
* Branch names follow:  
  \<type\>/\<ISSUE\_NUMS\>-\<short-kebab-description\>  
* Commit messages follow **Conventional Commits**:  
  \<type\>(\<scope\>)?: \<short summary\> \[optional body\] \[footer(s)\]  
  and include issue refs in the footer or subject as (\#NN) or Fixes \#NN.

---

## **2\) Branch naming (canonical)**

Format:  
`<type>/<ISSUE_NUMS>-<short-kebab-description>`

* \<type\>: feat, fix, chore, docs, refactor, perf, test, ci, style, deploy, hotfix  
* \<ISSUE\_NUMS\>: single or multiple issue numbers joined by hyphens (no \#), e.g. 3 or 3-4-7  
* \<short-kebab-description\>: short descriptive phrase in kebab-case (avoid generic words like update)

Examples:

* feat/3-add-login-api  
* fix/7-null-pointer-dashboard  
* chore/12-update-deps  
* feat/3-4-add-reporting-endpoints (works for a branch addressing multiple related issues)

**Notes:**

* Keep branch names reasonably short — prioritize clarity over length.  
* Use hotfix/ for urgent changes that must go in immediately.

---

## **3\) Commit message format (canonical — Conventional Commits)**

Format:  
`<type>(<scope>)?: <subject>`  
`<body> (optional)`  
`<footer(s)> (optional)`

* type: same list as branch types (feat, fix, ...). Use lowercase.  
* scope: optional, short identifier of area/module (e.g., auth, api, ui). Use no spaces.  
* subject: imperative, present tense, max \~72 characters.  
* body: explain what and why (not how). Wrap at \~72 chars.  
* footer: include issue references and breaking changes:  
  * Fixes \#3 (or Fixes \#3, \#4) — will auto-close issues on merge.  
  * BREAKING CHANGE: \<description\> — indicating major breaking change.

**Examples**  
`feat(auth): add JWT login endpoint`

`Adds /auth/login endpoint using JWT for stateless sessions.`  
`This implements the initial user login flow described in the spec.`

`Fixes #3`  
---

`fix(ui): prevent crash on missing avatar`

`Check for null avatar and fallback to initials.`  
`Fixes #7, #8`  
---

`refactor(api): split user service into user & account modules`

`BREAKING CHANGE: user service endpoints changed from /user/v1/* to /user/v2/*`

**Important**: Prefer referencing issues in the footer with Fixes \#N for auto-closing, and also add (\#N) to the subject if desired for quick scanning:  
`feat(auth): add JWT login endpoint (#3)`

---

## **4\) PR title & description**

**Title format**:  
`<type>: <short summary> [ (#ISSUE)]`

Examples:

* feat: add JWT login endpoint (\#3)  
* fix: prevent null avatar crash (\#7)

**PR description template** (copy into PR body when creating):  
`### Summary`  
`Short description of the changes.`

`### Changes`  
`- Bullet list of main changes`  
`- Any migrations / DB changes`

`### Related Issues`  
`Fixes #3`

`### QA / Testing`  
`Steps to reproduce / test plan.`

`### Screenshots / Notes`  
`Attach screenshots or important notes here.`

---

## **5\) Issues naming**

Issue title format:  
`[type]: <short summary>`

Examples:

* bug: login fails for users with special characters  
* feature: export user data to CSV

Label issues with bug, feature, enhancement, etc., and include acceptance criteria inside the issue body.  
---

## **6\) Multi-issue branches & commits**

If a branch addresses multiple issues:

* Join issue numbers with hyphens in the branch: feat/3-4-5-add-bulk-export  
* In commit footer and PR use: Fixes \#3, \#4, \#5 (comma-separated)  
* Keep scope and subject clear about the primary purpose.

---

## **7\) Mapping from our current flow → recommended canonical flow**

We’re using variants like feat/3, feat/3-4-5, bug/7. That's fine — the canonical version simply expands those with a readable description:

| Current | Canonical (recommended) |
| ----- | ----- |
| feat/3 | feat/3-add-login-api |
| feat/3-4 | feat/3-4-add-reporting |
| bug/7 | fix/7-null-pointer-dashboard |

**Action:** Developers to stop creating branches with only feat/3 — add a short description so branches are self-describing.  
---

## **8\) Standard commit types (your template aligned to Conventional Commits)**

Use the following types and meanings:

* feat: a new feature  
* fix: a bug fix  
* docs: documentation only changes  
* style: formatting, missing semi colons, white-space, no code change  
* refactor: code-change that neither fixes bug nor adds feature  
* perf: code change that improves performance  
* test: adding or updating tests  
* chore: build process or auxiliary tools, no production code change  
* ci: CI-related config changes  
* deploy: deployment-only changes  
* debug: temporary debug changes (prefer not to commit to main branches)  
* BREAKING CHANGE: placed in footer for breaking changes

**Recommended commit template example** (use in editors or git commit):  
`<type>(<scope>)?: <subject>`

`<body>`

`Fixes #<issue-number>`  
`BREAKING CHANGE: <description> (if any)`
