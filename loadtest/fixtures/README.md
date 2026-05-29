# Load Test Fixtures

This directory is ignored by default because it can contain generated credentials, token caches, and market IDs.

Expected files for the initial k6 scripts:

`users.csv`:

```csv
username,password
loaduser000001,loadtest-password
loaduser000002,loadtest-password
```

`markets.csv`:

```csv
market_id,kind
1,hot
2,normal
```

Do not put real user credentials here.
