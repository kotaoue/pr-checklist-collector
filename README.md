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

To save files with a date-based name (e.g. `results/2026-03-08.json`), wrap a date pattern in `{}`:

```yaml
- uses: kotaoue/pr-checklist-collector@v1
  with:
    checks: |
      dog
      cat
      bird
    output_file: results/{yyyy-mm-dd}.json
```

Supported date tokens inside `{}`:

| Token  | Example output | Description   |
|--------|----------------|---------------|
| `yyyy` | `2026`         | 4-digit year  |
| `yy`   | `26`           | 2-digit year  |
| `mm`   | `03`           | 2-digit month |
| `dd`   | `08`           | 2-digit day   |

Tokens can be combined freely: `{yyyymmdd}` → `20260308`, `{yyyy/mm/dd}` → `2026/03/08`, etc.
Paths without `{}` (e.g. `results/results.json`) are used as-is.

This will:
1. Create a new branch `checklist/<timestamp>`.
2. Commit the output file at the resolved `output_file` path with the initial checklist state (all items unchecked).
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
| `output_file`  | yes      | —                    | Repository-relative path for the output file. Wrap date tokens in `{}` for date-based filenames (e.g. `results/{yyyy-mm-dd}.json`). |
| `assignee`     | no       | `""` (no assignment) | GitHub username to assign the pull request to. |
| `github_token` | no       | `github.token`       | Token used to create branches, commits, and pull requests. |

## Supported formats

| Extension | Status |
|-----------|--------|
| `.json`   | ✅ Supported |

## Release management

Releases use [kotaoue/major-tag-floater](https://github.com/kotaoue/major-tag-floater) to keep the floating major-version tag (e.g. `v1`) pointing at the latest release.

## License

MIT
