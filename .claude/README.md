# Claude Skills for Terraform Provider Polytomic

This directory contains Claude Code skills specific to this project.

## Available Skills

### compare-api-versions

Compares two versions of the Polytomic OpenAPI specification to identify changes needed for Terraform provider updates.

**Usage:**
```bash
.claude/skills/compare-api-versions go@1.11.0 go@1.12.0
```

**What it does:**
- Fetches OpenAPI specs from the polytomic/fern-config repository
- Identifies schema changes (additions, removals, modifications)
- Detects breaking changes (required fields, type changes, etc.)
- Maps changes to Terraform resources
- Generates priority-ranked implementation checklists

**Output:**
- `/tmp/openapi-diff-summary.md` - Comprehensive analysis
- `/tmp/implementation-checklist.md` - Actionable task list

**When to use:**
- Before updating to a new polytomic-go SDK version
- To assess impact of API changes
- To plan provider feature additions

**Priority levels:**
- **CRITICAL**: Breaking changes requiring immediate attention
- **HIGH**: Major new features (3+ new fields)
- **MEDIUM**: Minor additions (1-2 new fields)
- **LOW**: Documentation or minor updates

## Adding New Skills

To add a new skill to this project:

1. Create a directory in `.claude/skills/` for your skill
2. Add a `SKILL.md` file with documentation (metadata and usage)
3. Create the executable script (Python, Bash, etc.)
4. Make it executable: `chmod +x .claude/skills/your-skill-name/script-name`
5. Update this README with usage information

## Skill Development Tips

- Keep skills focused on specific tasks
- Provide clear usage examples
- Generate output files in `/tmp` for easy review
- Use descriptive error messages
- Make scripts idempotent (safe to run multiple times)
