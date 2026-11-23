# Repository Metrics

This directory stores public JSON snapshots backed by GitHub's repository traffic API.

GitHub only exposes repository clone/view traffic for a rolling 14-day window. These values are snapshots of that rolling window, not all-time totals.

The top-level README uses static Shields badge URLs for the rolling clone metrics so badges render correctly before and after branch merges. The scheduled refresh workflow updates both the README badge URLs and these JSON snapshots from `gh api repos/$GITHUB_REPOSITORY/traffic/clones`.

GitHub's traffic API requires access that the default `GITHUB_TOKEN` does not provide. Add a repository secret named `TRAFFIC_METRICS_TOKEN` using a fine-grained token or GitHub App token with read access to repository Administration. This token is used only to fetch traffic; the workflow uses its automatic `GITHUB_TOKEN` with Contents and Pull requests write permissions to publish updates through a PR into the default branch. Enable "Allow GitHub Actions to create and approve pull requests" in the repository's Actions settings for both metrics workflows.

PR checks created by the automatic token require a user with write access to select "Approve workflows to run" in the PR merge box. See [GitHub's token workflow behavior](https://docs.github.com/en/actions/concepts/security/github_token#when-github_token-triggers-workflow-runs).

If `TRAFFIC_METRICS_TOKEN` is absent, the traffic workflow exits successfully without updating the badges. If GitHub rejects the token with HTTP 401 or 403, the workflow emits a warning and records the skipped refresh in its job summary, preserves the last successful metrics, and skips PR creation. Replace an expired, revoked, or insufficiently scoped token to resume traffic collection. Other API failures still fail the job.

The stargazers-over-time chart is generated from GitHub's stargazers API with `starred_at` timestamps. Run `python3 scripts/readme-stargazers-chart.py` from the repository root to refresh `stargazers-over-time.svg` and `stargazers-over-time.json`. The scheduled README Stargazers Chart workflow uses the default `GITHUB_TOKEN` for fetching and publishing, writes updated chart artifacts to a docs branch, and opens or updates a pull request. It does not depend on `TRAFFIC_METRICS_TOKEN`.
