# Pull Request

> Short summary (one line)  
**Title:** <!-- e.g. feat(auth): add SSO login -->

## Description
Provide a short description of the change and the motivation.  
Explain *why* this change is necessary and what problem it solves.

## Type of change
(Select or remove those that do not apply)
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Chore / refactor / documentation

## Related tickets / issues
- Closes: #<issue-number>  
- Related: #<issue-number>  

## Proposed changes
List the main changes made in this PR (files, modules, behaviour).
- Change 1
- Change 2
- Change 3

## How to test / QA
Step-by-step instructions for testing (include test data / commands / env vars).
1. Checkout branch `...`
2. Run `make test` or `npm run test`
3. Steps to reproduce behavior
4. Expected result

## Checklist (required)
Make sure your PR adheres to the org standards:
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] I have added or updated relevant documentation (README, code comments, CHANGELOG)
- [ ] I have run the test suite locally and all tests pass
- [ ] Linted and formatted code (`prettier`, `golangci-lint`, `rustfmt`, etc.)
- [ ] I updated any necessary config or infra files
- [ ] PR title follows conventional commit / repo convention

## Rollout plan & backwards compatibility
Explain rollout steps, feature flags, and whether this change is backward compatible. If a database migration is present, state whether it is additive or destructive and migration instructions.

## Database migrations
- [ ] Migration script included
- Migration notes:
  - Up: `sql/migrations/xxxx_up.sql`
  - Down: `sql/migrations/xxxx_down.sql`

## Security considerations
If this change affects authentication, authorization, secrets, data exposure, or external dependencies, describe the risk and mitigation.
- Sensitive data included? (Yes / No)
- If yes — how is it protected?

## Performance considerations
Note any performance implications and how you measured them (benchmarks, profiling).

## Screenshots / recordings
If UI changes, include before/after screenshots or a short recording.

## Release notes / changelog entry
Suggested changelog entry (one sentence).

## Reviewers / Assignees
- Requested reviewers: @team-or-user
- Code owners: (will be assigned by CODEOWNERS if present)

---

### Optional: Quick PR template for tiny fixes
For trivial changes (typo, small doc tweak), add `#trivial` in the title to indicate expedited review.

### Releated Issue
#{{ISSUE_NUMBERs}}

