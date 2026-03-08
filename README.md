# pr-checklist-collector

A GitHub Action that, when a pull request is merged, collects the checklist state from the PR body and saves it as a JSON file committed directly to the base branch.

## Usage

Add this action to a workflow triggered on `pull_request` closed events:

```yaml
on:
  pull_request:
    types: [closed]

permissions:
  contents: write
  pull-requests: read

jobs:
  collect-checklist:
    if: github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    steps:
      - uses: kotaoue/pr-checklist-collector@v1
        with:
          output_file: results/{yyyy-mm-dd}.json
          github_token: ${{ secrets.GITHUB_TOKEN }}
```

The action reads the merged PR body, parses all GitHub-flavored markdown checkboxes (`- [x]` / `- [ ]`), and commits the result as a JSON file to the base branch.

**Example PR body:**
```
- [x] dog
- [ ] cat
- [x] bird
```

**Produces** `results/2026-03-08.json`:
```json
{
  "date": "2026-03-08",
  "checks": {
    "dog": true,
    "cat": false,
    "bird": true
  }
}
```

To save files with a date-based name (e.g. `results/2026-03-08.json`), wrap a date pattern in `{}`:

```yaml
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

To reproduce a custom key name (e.g. `"exercises"` like [FitnessStreak](https://github.com/kotaoue/FitnessStreak/blob/main/results/2026-03-06.json)), set `checks_key`:

```yaml
- uses: kotaoue/pr-checklist-collector@v1
  with:
    output_file: results/{yyyy-mm-dd}.json
    checks_key: exercises
```

This produces:
```json
{
  "date": "2026-03-08",
  "exercises": {
    "ラジオ体操": false,
    "ストレッチ": false
  }
}
```

To filter by PR title and customize the commit identity (matching the [FitnessStreak workflow](https://github.com/kotaoue/FitnessStreak/blob/main/.github/workflows/save-results.yml)):

```yaml
- uses: kotaoue/pr-checklist-collector@v1
  with:
    output_file: results/{yyyy-mm-dd}.json
    checks_key: exercises
    pr_title_pattern: '\d{4}-\d{2}-\d{2}'   # only process PRs whose title contains a date
    commit_user_name: 'github-actions[bot]'
    commit_user_email: 'github-actions[bot]@users.noreply.github.com'
```

When `pr_title_pattern` is set and the merged PR's title does not match, the action exits successfully without writing any file.

## Inputs

| Name                 | Required | Default                                          | Description |
|----------------------|----------|--------------------------------------------------|-------------|
| `output_file`        | yes      | —                                                | Repository-relative path for the output file. Wrap date tokens in `{}` for date-based filenames (e.g. `results/{yyyy-mm-dd}.json`). |
| `checks_key`         | no       | `checks`                                         | Name of the JSON object key that holds the checklist map (e.g. `exercises`). |
| `pr_title_pattern`   | no       | *(accept all)*                                   | Regular expression matched against the merged PR title. When set, PRs whose title does not match are silently skipped (action exits with success, no file written). |
| `commit_user_name`   | no       | `github-actions[bot]`                            | Git committer name recorded on the result commit. |
| `commit_user_email`  | no       | `github-actions[bot]@users.noreply.github.com`   | Git committer email recorded on the result commit. |
| `github_token`       | no       | `github.token`                                   | Token used to commit the result file. |

## Supported formats

| Extension | Status |
|-----------|--------|
| `.json`   | ✅ Supported |

## Release management

Releases use [kotaoue/major-tag-floater](https://github.com/kotaoue/major-tag-floater) to keep the floating major-version tag (e.g. `v1`) pointing at the latest release.

## License

MIT
