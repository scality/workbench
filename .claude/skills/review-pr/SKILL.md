---
name: review-pr
description: Review a PR on workbench (Go CLI for S3C Docker Compose dev environments)
argument-hint: <pr-number-or-url>
disable-model-invocation: true
allowed-tools: Read, Bash(gh repo view *), Bash(gh pr view *), Bash(gh pr diff *), Bash(gh pr comment *), Bash(gh api *), Bash(git diff *), Bash(git log *), Bash(git show *)
---

# Review GitHub PR

You are an expert code reviewer. Review this PR: $ARGUMENTS

## Determine PR target

Parse `$ARGUMENTS` to extract the repo and PR number:

- If arguments contain `REPO:` and `PR_NUMBER:` (CI mode), use those values directly.
- If the argument is a GitHub URL (starts with `https://github.com/`), extract `owner/repo` and the PR number from it.
- If the argument is just a number, use the current repo from `gh repo view --json nameWithOwner -q .nameWithOwner`.

## Output mode

- **CI mode** (arguments contain `REPO:` and `PR_NUMBER:`): post inline comments and summary to GitHub.
- **Local mode** (all other cases): output the review as text directly. Do NOT post anything to GitHub.

## Steps

1. **Fetch PR details:**

```bash
gh pr view <number> --repo <owner/repo> --json title,body,headRefOid,author,files
gh pr diff <number> --repo <owner/repo>
```

2. **Read changed files** to understand the full context around each change (not just the diff hunks).

3. **Analyze the changes** against these criteria:

| Area | What to check |
|------|---------------|
| Error wrapping | Use `fmt.Errorf("...: %w", err)`, not `%v` — ensure errors are properly wrapped for `errors.Is`/`errors.As` |
| Context propagation | Pass `context.Context` through call chains, respect cancellation |
| Goroutine leaks | Ensure goroutines have exit conditions, use `errgroup` where appropriate |
| Template correctness | Go text templates in `templates/` must produce valid JSON, YAML, shell scripts, or config files |
| Config defaults | Changes to `EnvironmentConfig` or defaults must not break existing environments |
| Docker Compose validity | Changes to `docker-compose.yaml` template must produce valid Compose v2 syntax |
| Embed consistency | New or renamed template files must be covered by the `embed.go` directives |
| Port conflicts | New services or port changes must not conflict with existing host-network ports |
| Feature flag gating | Optional components must be gated behind the correct Docker Compose profile |
| Security | No credentials or secrets in templates, no command injection in shell script templates |
| Breaking changes | Changes to CLI flags, values.yaml schema, or environment structure that break existing usage |

4. **Deliver your review:**

### If CI mode: post to GitHub

#### Part A: Inline file comments

For each issue, post a comment on the exact file and line. Keep comments short (1-3 sentences), end with `— Claude Code`. Use line numbers from the **new version** of the file.

**Without suggestion block** — single-line command, `<br>` for line breaks:
```bash
gh api -X POST -H "Accept: application/vnd.github+json" "repos/<owner/repo>/pulls/<number>/comments" -f body="Issue description.<br><br>— Claude Code" -f path="file" -F line=42 -f side="RIGHT" -f commit_id="<headRefOid>"
```

**With suggestion block** — use a heredoc (`-F body=@-`) so code renders correctly:
```bash
gh api -X POST -H "Accept: application/vnd.github+json" "repos/<owner/repo>/pulls/<number>/comments" -F body=@- -f path="file" -F line=42 -f side="RIGHT" -f commit_id="<headRefOid>" <<'COMMENT_BODY'
Issue description.

```suggestion
first line of suggested code
second line of suggested code
```

— Claude Code
COMMENT_BODY
```

Only suggest when you can show the exact replacement. For architectural or design issues, just describe the problem.

#### Part B: Summary comment

Single-line command, `<br>` for line breaks. No markdown headings — they render as giant bold text. Flat bullet list only:

```bash
gh pr comment <number> --repo <owner/repo> --body "- file:line — issue<br>- file:line — issue<br><br>Review by Claude Code"
```

If no issues: just say "LGTM". End with: `Review by Claude Code`

### If local mode: output the review as text

Do NOT post anything to GitHub. Instead, output the review directly as text.

For each issue found, output:

```
**<file_path>:<line_number>** — <what's wrong and how to fix it>
```

When the fix is a concrete line change, include a fenced code block showing the suggested replacement.

At the end, output a summary section listing all issues. If no issues: just say "LGTM".

End with: `Review by Claude Code`

## What NOT to do

- Do not comment on markdown formatting preferences
- Do not suggest refactors unrelated to the PR's purpose
- Do not praise code — only flag problems or stay silent
- If no issues are found, post only a summary saying "LGTM"
- Do not flag style issues already covered by the project's linter (eslint, biome, pylint, golangci-lint)
