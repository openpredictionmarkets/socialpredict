# Repository Metrics

This directory stores public JSON snapshots backed by GitHub's repository traffic API.

GitHub only exposes repository clone/view traffic for a rolling 14-day window. These values are snapshots of that rolling window, not all-time totals.

The top-level README uses static Shields badge URLs for the rolling clone metrics so badges render correctly before and after branch merges. The scheduled refresh workflow updates both the README badge URLs and these JSON snapshots from `gh api repos/$GITHUB_REPOSITORY/traffic/clones`.

GitHub's traffic API requires access that the default `GITHUB_TOKEN` does not provide. Add a repository secret named `TRAFFIC_METRICS_TOKEN` using a fine-grained token or GitHub App token with read access to repository Administration, read/write access to repository Contents, and read/write access to Pull requests. The write permissions let the scheduled workflow open a PR instead of pushing directly to protected `main`. If this secret is absent, the scheduled workflow exits successfully without updating the traffic badges.

The stargazers-over-time chart is generated from GitHub's stargazers API with `starred_at` timestamps. Run `python3 scripts/readme-stargazers-chart.py` from the repository root to refresh `stargazers-over-time.svg` and `stargazers-over-time.json`. The scheduled README Stargazers Chart workflow uses the default `GITHUB_TOKEN`, writes updated chart artifacts to a docs branch, and opens or updates a pull request.
