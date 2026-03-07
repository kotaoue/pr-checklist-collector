# pr-checklist-collector

A GitHub Action that creates a pull request containing a markdown checklist and saves the initial checklist state to a file in a configurable format.

## Usage

```yaml
- uses: kotaoue/pr-checklist-collector@v1
  with:
    checks: |
      dog
      cat
      bird
    output_file: results/results.json
    assignee: kotaoue          # optional
    github_token: ${{ secrets.GITHUB_TOKEN }}  # optional, defaults to github.token
```

This will:
1. Create a new branch `checklist/<timestamp>`.
2. Commit `results/results.json` with the initial checklist state (all items unchecked).
3. Open a pull request whose body contains:
   ```
   - [ ] dog
   - [ ] cat
   - [ ] bird
   ```
4. Assign the pull request to `kotaoue` (skipped when `assignee` is empty).

## Inputs

| Name           | Required | Default              | Description |
|----------------|----------|----------------------|-------------|
| `checks`       | yes      | —                    | Newline-separated list of checklist items. |
| `output_file`  | yes      | —                    | Repository-relative path for the output file (e.g. `results/results.json`). |
| `assignee`     | no       | `""` (no assignment) | GitHub username to assign the pull request to. |
| `github_token` | no       | `github.token`       | Token used to create branches, commits, and pull requests. |

## Supported formats

| Extension | Status |
|-----------|--------|
| `.json`   | ✅ Supported |

Additional formats can be added by implementing the `Formatter` interface in the `formatter` package.

## Release management

Releases use [kotaoue/major-tag-floater](https://github.com/kotaoue/major-tag-floater) to keep the floating major-version tag (e.g. `v1`) pointing at the latest release.

## License

MIT
