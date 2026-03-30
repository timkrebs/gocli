## Summary

<!--
Explain *what* this PR does and *why*. Link to the relevant issue if one exists.
Example: "Closes #42. Adds UiWriterLevel so callers can route logger output to
ui.Error instead of the hard-coded ui.Info."
-->

Closes #

## Type of change

<!-- Check all that apply. -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that changes existing behaviour)
- [ ] Refactor (no functional change)
- [ ] Documentation update
- [ ] CI / tooling change

## Checklist

- [ ] Tests added or updated to cover the change
- [ ] `go test -race ./...` passes locally
- [ ] `golangci-lint run ./...` passes locally (or lint issues are intentional and explained)
- [ ] Public API additions / changes have godoc comments
- [ ] `CHANGELOG.md` updated under `## [Unreleased]` (for features and bug fixes)
- [ ] PR title follows [Conventional Commits](https://www.conventionalcommits.org/) format
  (`feat: ...`, `fix: ...`, `docs: ...`, etc.)

## Testing notes

<!--
Describe how you tested this. Include the test names that cover the new
behaviour, or explain why no new tests were added.
-->

## Breaking change migration guide

<!--
If this is a breaking change, describe what callers need to update.
Delete this section if not applicable.
-->
