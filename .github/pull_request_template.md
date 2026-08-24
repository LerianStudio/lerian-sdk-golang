<table border="0" cellspacing="0" cellpadding="0">
  <tr>
    <td><img src="https://github.com/LerianStudio.png" width="72" alt="Lerian" /></td>
    <td><h1>Lerian Go SDK</h1></td>
  </tr>
</table>

---

## Description

<!-- Summarize what this PR changes and why. Mention the product client(s)
     affected (midaz, matcher, tracer, reporter, fees). -->

## Type of Change

- [ ] `feat`: New feature or capability
- [ ] `fix`: Bug fix
- [ ] `perf`: Performance improvement
- [ ] `refactor`: Internal restructuring with no behavior change
- [ ] `docs`: Documentation only (README, docs/, inline comments)
- [ ] `style`: Formatting, whitespace, naming (no logic change)
- [ ] `test`: Adding or updating tests
- [ ] `ci`: CI pipeline or workflow changes
- [ ] `build`: Build system, GoReleaser, Go module dependencies
- [ ] `chore`: Maintenance, config, tooling
- [ ] `revert`: Reverts a previous commit
- [ ] `BREAKING CHANGE`: Consumers must update their integration

## Breaking Changes

<!-- If applicable, describe exactly what breaks in the public SDK surface
     (exported types, method signatures, client interfaces) and how downstream
     consumers should migrate. Remove this section if not applicable. -->

None.

## Testing

- [ ] `make test` passes
- [ ] `make coverage` unaffected or improved
- [ ] `make lint` passes
- [ ] `make gosec` passes
- [ ] `make verify-sdk` passes if the public API surface changed

**Test evidence / Actions run:** <!-- Optional: link to a CI run or screenshot -->

## Architectural Checklist

- [ ] No `panic()` in exported SDK paths
- [ ] Errors wrapped with `%w`
- [ ] Exported types/methods documented

## Related Issues

Closes #
