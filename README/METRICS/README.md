# Repository Traffic Metrics

This directory stores public JSON snapshots backed by GitHub's repository traffic API.

GitHub only exposes repository clone/view traffic for a rolling 14-day window. These values are snapshots of that rolling window, not all-time totals.

The top-level README uses static Shields badge URLs for the rolling clone metrics so badges render correctly before and after branch merges. The scheduled refresh workflow updates both the README badge URLs and these JSON snapshots from `gh api repos/$GITHUB_REPOSITORY/traffic/clones`.

If the default `GITHUB_TOKEN` cannot read traffic metrics, add a repository secret named `TRAFFIC_METRICS_TOKEN` using a fine-grained token with read access to repository administration/traffic metrics.
