# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen" | judgment-day | /home/chris/.config/opencode/skills/judgment-day/SKILL.md |
| When creating a pull request, opening a PR, or preparing changes for review | branch-pr | /home/chris/.config/opencode/skills/branch-pr/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | /home/chris/.config/opencode/skills/skill-creator/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature | issue-creation | /home/chris/.config/opencode/skills/issue-creation/SKILL.md |
| When writing Go tests, using teatest, or adding test coverage | go-testing | /home/chris/.config/opencode/skills/go-testing/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### judgment-day
- Launch TWO independent blind judges simultaneously for the same target
- Judges must NOT communicate with each other during the review
- Synthesize findings after both judges complete their reviews
- Apply fixes based on consensus findings
- Re-judge after fixes — if both pass, change is approved
- If both judges fail after 2 iterations, escalate to human review
- Use the Skill Resolver Protocol to load relevant skills before launching judges

### branch-pr
- Every PR MUST link an approved issue — no exceptions
- Every PR MUST have exactly one `type:*` label
- Automated checks must pass before merge is possible
- Blank PRs without issue linkage will be blocked by GitHub Actions
- Use issue-first workflow: approved issue → branch → PR → review → merge

### skill-creator
- Create a skill when: pattern used repeatedly, project-specific conventions differ, complex workflows need steps, decision trees help AI
- Don't create when: documentation exists, pattern is trivial, one-off task
- Follow Agent Skills spec structure
- Include allowed-tools in frontmatter

### issue-creation
- Blank issues are disabled — MUST use a template (bug report or feature request)
- Every issue gets `status:needs-review` automatically on creation
- A maintainer MUST add `status:approved` before any PR can be opened
- Questions go to Discussions, not issues

### go-testing
- Use table-driven tests for multiple test cases
- Use teatest for Bubbletea TUI testing (tea.TestT interface)
- Use golden file testing for complex output comparisons
- Use httptest for HTTP integration tests
- Follow Go testing conventions: TestXxx format, t.Errorf for failures
- Run tests with go test -v for verbose output
- Coverage: go test -cover

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| README.md | /home/chris/code/bash/gf/README.md | Project documentation |
| CHANGELOG.md | /home/chris/code/bash/gf/CHANGELOG.md | Version history |
| changelogs/CHANGELOG-2025-08.md | /home/chris/code/bash/gf/changelogs/CHANGELOG-2025-08.md | Archived changelog |

Read the convention files listed above for project-specific patterns and rules. All referenced paths have been extracted — no need to read index files to discover more.
