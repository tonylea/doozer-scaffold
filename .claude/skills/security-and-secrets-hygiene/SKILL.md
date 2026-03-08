---
name: security-and-secrets-hygiene
description: Use this skill whenever writing code that touches configuration, environment variables, API keys, credentials, authentication, or external service connections. Also use when setting up a new project, adding dependencies, creating Docker configurations, or writing CI pipelines. This skill prevents accidental exposure of secrets and enforces dependency safety. It should be consulted alongside any language-specific coding skill.
---

# Security and Secrets Hygiene

This skill prevents the most common security mistakes in codebases: leaking secrets and shipping vulnerable dependencies. These rules apply to every language, framework, and project.

## Secrets: Absolute Rules

These rules have zero exceptions. Do not rationalise around them. "Just for testing", "I'll rotate it later", "it's only local" — none of these are valid reasons.

1. **Never commit secrets.** No API keys, tokens, passwords, private keys, database connection strings, or credentials of any kind in any file that is tracked by git. No exceptions. No "temporary" commits. No "I'll remove it later." If you are about to type or paste a real credential into a tracked file, stop.
2. **Never hardcode secrets in source code.** Not in config files, not in comments, not in variable defaults, not in test fixtures. Secrets come from environment variables or a secrets manager at runtime. "Making it work quickly" by hardcoding a key is creating a security incident.
3. **Use `.env.example`, never `.env`.** The `.env` file must be in `.gitignore`. Provide `.env.example` with placeholder values documenting every required variable:
   ```
   DATABASE_URL=postgresql://user:password@localhost:5432/dbname
   API_KEY=your-api-key-here
   ```
4. **Check before every commit.** Before staging files, verify that no secrets are present in the diff. If you accidentally commit a secret, rotating the credential is mandatory — removing it from git history is not sufficient, as the secret has already been exposed.

## Common Secret Leak Patterns to Watch For

- Inline connection strings with embedded passwords
- API keys passed as default function parameters
- Private keys or certificates copied into the repo for "convenience"
- Docker Compose files with hardcoded credentials (use environment variable substitution instead)
- CI pipeline configs with secrets in plaintext (use the platform's secrets management)
- Jupyter notebooks or REPL history with authenticated API calls
- Screenshots or logs that contain tokens or session IDs

## Configuration Files

- **Docker:** Never use `ENV` in a Dockerfile for secrets. Pass them at runtime via `docker run -e` or Docker Compose `environment`/`env_file` with a `.env` file that is gitignored.
- **CI pipelines:** Use GitHub Actions secrets (`${{ secrets.MY_SECRET }}`), never plaintext in workflow files.
- **Application config:** Use environment variables for anything sensitive. Config files checked into the repo should contain only non-sensitive defaults.

## .gitignore Requirements

Every project's `.gitignore` must include at minimum:

```
# Secrets and environment
.env
.env.local
.env.*.local
*.pem
*.key
*.p12
*.pfx

# IDE and OS
.idea/
.vscode/
*.swp
.DS_Store
Thumbs.db

# Dependencies (language-specific — add the appropriate patterns)
node_modules/
__pycache__/
*.pyc
vendor/

# Build outputs (language-specific — add the appropriate patterns)
dist/
build/
*.o
*.class
```

Tailor the language-specific sections to the actual stack. The secrets section is non-negotiable for all projects.

## Dependencies

1. **Review before adding.** Before installing a new dependency, check: is it actively maintained? Does it have known vulnerabilities? Is the scope appropriate, or are you pulling in a large library for one function?
2. **Pin versions in production.** Use lock files (`package-lock.json`, `poetry.lock`, `Gemfile.lock`, etc.) and commit them. This ensures reproducible builds and prevents supply chain attacks via version drift.
3. **Keep dependencies updated.** Regularly check for security advisories. Use tools like `npm audit`, `pip-audit`, or Dependabot/Renovate for automated vulnerability alerts. Don't ignore security warnings.
4. **Minimise dependency surface.** Fewer dependencies mean fewer attack vectors. If you can implement something in a few lines rather than adding a package, do so. Every dependency is code you don't control.

## What To Do If a Secret Is Leaked

If a secret is committed to any branch, even locally:

1. **Rotate the credential immediately.** Generate a new key/password/token and update wherever it's used.
2. **Remove from git history** using `git filter-repo` or BFG Repo-Cleaner — but treat this as damage limitation, not a fix. The old credential must be considered compromised.
3. **Check access logs** for the affected service to see if the credential was used by an unauthorised party.
4. **Document the incident** so the team can learn from it.

The rotation step is the only one that actually secures the system. History rewriting is cosmetic.
