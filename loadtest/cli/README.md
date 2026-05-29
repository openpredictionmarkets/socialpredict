# Load Test CLI

`loadtest/cli/loadtest` is a small wrapper around k6 and the dossier summarizer.

Seed fixture CSVs with `./SocialPredict load seed`, or provide files under `loadtest/fixtures/` or with `--users-file` and `--markets-file`.

## Commands

```bash
LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --reset
./loadtest/cli/loadtest check
./loadtest/cli/loadtest run smoke --base-url https://kconfs.com
./loadtest/cli/loadtest run baseline --base-url https://kconfs.com --duration 5m --browse-rate 20 --bet-rate 5
./loadtest/cli/loadtest run hot-market-burst --base-url https://kconfs.com --target-rate 100 --duration 60s
./loadtest/cli/loadtest dossier --summary loadtest/results/<summary>.json --metadata loadtest/dossier/metadata.example.json --out loadtest/dossier/runs/<run>.json
```

## Authentication

k6 logs in with normal fake SocialPredict users from `users.csv` and uses normal bearer tokens for `/v0/bet`.

No DigitalOcean credentials or betting god token are used by this CLI.
