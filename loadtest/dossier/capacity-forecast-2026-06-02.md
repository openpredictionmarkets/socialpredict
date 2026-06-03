# SocialPredict Capacity Forecast Dossier

Date: 2026-06-02

Status: draft planning dossier; June 3 fresh-state sustained test added

Related dossiers:

- `loadtest/dossier/staging-capacity-2026-05-29.md`
- `loadtest/dossier/general-purpose-capacity-experiment-2026-06-02.md`

## Executive Summary

SocialPredict has enough evidence to make a cautious near-term capacity forecast, but not enough evidence yet to claim sustained large-scale hot-market throughput.

Current evidence supports three practical conclusions:

- The current small staging shape, DigitalOcean Basic `1 vCPU / 1 GiB RAM`, is useful for functional staging and light model-office traffic, not for large hot-market events.
- A larger DigitalOcean Basic AMD shared-CPU `8 vCPU / 32 GiB RAM` single-node host produced clean one-minute hot-market bursts through about `300` bets/sec, and a clean five-minute sustained run at `250` bets/sec.
- A June 3 repeat run started from zero fixture bets and still failed the strict clean-run standard at `300` bets/sec, with Postgres CPU peaking at `568%` and host CPU idle falling to `0.12%`.
- Sustained hot-market capacity is currently limited more by Postgres/write-path CPU behavior than by RAM. The leading software finding remains the missing composite bet-history index for the placement-time `market_id + username` lookup.

The user-capacity forecast depends on how many users act inside a short event window. If `50,000` registered users produce only routine traffic, the system is likely fine at modest machine sizes. If `25-50%` of those users place bets inside the same one-minute hot window, the workload becomes a database scaling problem.

Recommended next decision:

- Treat `300` bets/sec for `1m` on the `8 vCPU / 32 GiB` Basic AMD host as the best current burst datapoint.
- Treat `250` bets/sec for `5m` on the same host as the best current clean sustained datapoint.
- Treat `300` bets/sec for `5m` as not cleanly supported on the current single-node shape and current write path.
- Add the missing bet-history indexes through migrations, then rerun five-minute confirmation tests.
- For a serious `50,000` user target, plan around split app/database architecture rather than a colocated single-Droplet topology.

## Evidence Base

### Staging Host, 2026-05-29

Host shape:

- DigitalOcean Basic
- `1 vCPU`
- `1 GiB RAM`
- `25 GiB SSD`
- App, Traefik, nginx, frontend, backend, and Postgres colocated in Docker
- Approximate observed price: `$0.00893/hr`, `$6/mo`

Best supported result:

- `50` pre-authenticated hot-market bets/sec for `1m`
- `3001/3001` bets succeeded
- `0%` HTTP failures
- HTTP p95 around `704ms`

Ceiling evidence:

- `75` bets/sec for `1m` passed functionally but p95 rose above `1s` and k6 dropped iterations.
- `75` bets/sec for `5m` degraded/fell over under stress.
- `100` bets/sec for `1m` was beyond this host.

Interpretation:

- This is a staging/model-office class machine, not a large-event machine.
- Safe hot-market planning target on this host should be materially below `75` bets/sec.

### Large Basic AMD Host, 2026-06-02

Host shape:

- DigitalOcean Basic AMD shared CPU
- Size slug: `s-8vcpu-32gb-amd`
- `8 vCPU`
- `32 GiB RAM`
- CPU/RAM-only resize; root disk remained about `25 GiB`
- App, Traefik, nginx, frontend, backend, and Postgres colocated in Docker
- Approximate observed price from `doctl`: `$0.250000/hr`, `$168/mo`
- Approximate 48-hour experiment cost: `$12.00`

One-minute burst evidence:

| Target | Result | Notes |
| ---: | --- | --- |
| `100` bets/sec | clean pass | p95 `39.15ms`; no failures |
| `200` bets/sec | clean pass | p95 `40.76ms`; no failures |
| `300` bets/sec | clean pass | p95 `72.65ms`; no failures; CPU pressure visible |
| `350` bets/sec | degraded | dropped iterations; p95 `1.3s`; CPU saturation |
| `400` bets/sec | degraded | dropped iterations; p95 `2.3s`; CPU saturation |

Sustained five-minute evidence:

- `300` bets/sec for `5m` degraded.
- `200` bets/sec for `5m` degraded.
- `100` bets/sec for `5m` later produced timeout failures after cumulative bet history grew.
- A June 3 repeat at `300` bets/sec began from a verified zero-bet fixture state on a fresh temporary host and still produced `288` failed bets, `563` dropped iterations, host idle CPU as low as `0.12%`, and Postgres CPU as high as `568.27%`.
- A June 3 fresh-state `250` bets/sec run passed for the full `5m`: `75001` successful bets, `0` failed bets, `0` dropped iterations, HTTP p95 `153.72ms`, Postgres CPU up to `407.78%`, and host idle CPU as low as `0.86%`.

Important diagnostic:

- After cumulative tests, the `bets` table had about `258,785` rows concentrated across `10` hot markets.
- The placement path checks whether a user has already bet in a market using `WHERE market_id = ? AND username = ?`.
- The observed indexes did not include a composite `(market_id, username)` index.
- Postgres CPU saturated during degraded sustained runs while RAM remained plentiful.

Interpretation:

- The larger host showed strong burst capacity.
- Sustained `250/sec` capacity is cleanly supported by the June 3 fresh-state run on this single-node shared-CPU shape.
- Sustained `300/sec` capacity is not cleanly supported on this single-node shared-CPU shape.
- Sustained capacity cannot be responsibly forecast upward until the bet-history indexing issue is fixed and retested.
- Because the fresh-state repeat saturated CPU while leaving more than `31 GiB` RAM available, a CPU-optimized or dedicated-CPU Droplet is a more relevant next hardware experiment than a RAM-optimized Droplet.

## Rate-Limit Model

Normal/model-office rate limits are intentionally conservative:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=0.1
RATE_LIMIT_LOGIN_BURST=3
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1
RATE_LIMIT_GENERAL_BURST=10
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

Load-test/staging rate limits are intentionally permissive for single-source k6 testing:

```env
RATE_LIMIT_LOGIN_RATE_PER_SECOND=100
RATE_LIMIT_LOGIN_BURST=200
RATE_LIMIT_GENERAL_RATE_PER_SECOND=1000
RATE_LIMIT_GENERAL_BURST=2000
RATE_LIMIT_CLEANUP_INTERVAL=5m
```

The normal runtime limiter is currently best interpreted as a per-client-identity limiter, effectively per IP/client identity in practical external traffic. Therefore:

- One normal client identity can sustain about `1` general API action/sec.
- A single NAT, proxy, school, office, VPN, or bot source may become capped even if many human users sit behind it.
- Aggregate platform throughput can exceed `1` request/sec if traffic comes from many client identities.

## User Forecast Formula

For one hot-market event window:

```text
bets_per_second = active_bettors_in_window / window_seconds
active_bettors_in_window = platform_users * fraction_betting_in_window
```

For active-user equivalents at a known bet rate:

```text
active_bettors = bets_per_second * seconds_between_bets_per_user
```

For client-identity requirements under normal rate limits:

```text
required_client_identities = ceil(bets_per_second / 1)
```

This assumes one bet request uses the full `1` general API action/sec sustained allowance for that client identity. Real page behavior can include additional reads, so this is a lower-bound identity estimate.

## User Forecast Table

This table translates user-event scenarios into required hot-market bet rates.

| Platform users | Fraction betting in `1m` | Active bettors in window | Required bets/sec |
| ---: | ---: | ---: | ---: |
| `10,000` | `5%` | `500` | `8.3` |
| `10,000` | `10%` | `1,000` | `16.7` |
| `10,000` | `25%` | `2,500` | `41.7` |
| `10,000` | `50%` | `5,000` | `83.3` |
| `30,000` | `5%` | `1,500` | `25.0` |
| `30,000` | `10%` | `3,000` | `50.0` |
| `30,000` | `25%` | `7,500` | `125.0` |
| `30,000` | `50%` | `15,000` | `250.0` |
| `50,000` | `5%` | `2,500` | `41.7` |
| `50,000` | `10%` | `5,000` | `83.3` |
| `50,000` | `25%` | `12,500` | `208.3` |
| `50,000` | `50%` | `25,000` | `416.7` |

Interpretation against current evidence:

- `10,000` users with `25%` betting in one minute requires about `42` bets/sec. This fits within staging's observed `50/sec` burst evidence, but staging is still too small for production comfort.
- `30,000` users with `25%` betting in one minute requires about `125` bets/sec. This is plausible on the larger host as a one-minute burst, but sustained proof is still missing.
- `50,000` users with `25%` betting in one minute requires about `208` bets/sec. This is within the larger host's one-minute clean evidence, but not yet within sustained proof.
- `50,000` users with `50%` betting in one minute requires about `417` bets/sec. This exceeded the larger host's clean envelope in one-minute tests.

## Normal Rate-Limit Identity Matrix

Under the normal `1` general API action/sec per client identity policy, the minimum number of distinct client identities is numerically similar to required bets/sec. Real clients may need more headroom because browsing, market detail refreshes, portfolio reads, and other API actions share the same general limit.

| Required bets/sec | Minimum client identities at `1` action/sec | Client identities if each places `1` bet every `10s` | User interpretation |
| ---: | ---: | ---: | --- |
| `50` | `50` | `500` | Roughly `30,000` users with `10%` betting in `1m`, or `10,000` with `30%` in `1m`. |
| `100` | `100` | `1,000` | A moderate hot-window event; above small staging comfort. |
| `200` | `200` | `2,000` | Roughly `50,000` users with `24%` betting in `1m`; plausible burst target on larger host, not yet sustained. |
| `300` | `300` | `3,000` | Current best one-minute large-host burst datapoint; needs indexing and sustained retest. |
| `400` | `400` | `4,000` | Degraded on the large Basic AMD host; do not claim yet. |
| `500` | `500` | `5,000` | Ambitious target; not supported by current evidence. |

## Cost And Machine Recommendation

Observed/tested costs:

| Host shape | Observed role | Approximate price | Evidence-supported use |
| --- | --- | ---: | --- |
| Basic `1 vCPU / 1 GiB / 25 GiB` | staging/model-office floor | `$0.00893/hr`, `$6/mo` | Functional staging and light traffic; clean `50/sec` burst, not large hot events. |
| Basic AMD `8 vCPU / 32 GiB / 25 GiB root disk observed` | temporary load-test host | `$0.250000/hr`, `$168/mo` | Clean `300/sec` one-minute burst; clean `250/sec` five-minute sustained run; fresh-state `300/sec` five-minute repeat failed strict clean criteria under Postgres CPU saturation. |

Recommendation by target:

| Target | Recommendation | Rationale |
| --- | --- | --- |
| Functional staging | Keep current small staging shape. | Cheapest useful deploy validation target. |
| Model office with modest public usage | Use at least `2 vCPU / 4 GiB` or `4 vCPU / 8 GiB`, ideally with headroom for deploys/logs. | The `1 GiB` host works but is memory-tight under stress. |
| `10,000` users with normal traffic | `4 vCPU / 8 GiB` single node may be enough for early validation; test before launch. | One-minute `25%` event requires about `42/sec`, below current burst evidence. |
| `30,000` users with hot events | Larger app host plus serious Postgres tuning; consider separating DB. | `25%` event requires `125/sec`, which should not rely on a tiny colocated DB. |
| `50,000` users with `25%` one-minute hot participation | Larger app tier plus managed or separately tuned Postgres. | Requires about `208/sec`; current burst evidence says plausible, sustained evidence does not. |
| `50,000` users with `50%` one-minute hot participation | Treat as a scale project, not a single-Droplet deployment. | Requires about `417/sec`; current large-host test degraded by `350-400/sec`. |

Hardware interpretation:

- The June 3 repeat indicates CPU, not RAM, is the binding resource for concentrated hot-market writes.
- A CPU-optimized or dedicated-CPU Droplet is therefore the better next hardware comparison than a RAM-optimized Droplet.
- Bigger hardware should not replace the schema work. The bet-placement write path still needs index/migration hardening before using larger-host results as production guidance.

## Recommendations

### Product/Launch Recommendation

Use the current evidence to say:

> SocialPredict has demonstrated external hot-market burst capacity up to `300` bets/sec for one minute on a single DigitalOcean Basic AMD `8 vCPU / 32 GiB RAM` host, but sustained large-event capacity is not yet validated. The next release should not claim `500` bets/sec or `50,000`-user hot-window capacity until the bet-history indexing issue is fixed and retested.

### Engineering Recommendation

Prioritize the database write-path fix:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_username ON bets (market_id, username);
```

Then consider:

```sql
CREATE INDEX CONCURRENTLY idx_bets_market_id_placed_at_id ON bets (market_id, placed_at, id);
```

These must be implemented through SocialPredict's timestamped migration system and tested with the existing API/load-test tooling.

### Testing Recommendation

After the index migration:

1. Create a fresh temporary load-test Droplet.
2. Deploy with the load-test workflow.
3. Seed `10000` users, `100` moderators, `1000` markets, and `10` hot markets.
4. Verify local fixture IDs match server market IDs.
5. Run `300/sec for 5m`.
6. If clean, run `350/sec for 5m`.
7. Only attempt `400/sec` or `500/sec` after `350/sec for 5m` is clean.

### Architecture Recommendation

For serious `30,000-50,000` user planning, move the thinking away from one all-in-one Docker host:

- Keep app containers horizontally scalable.
- Separate Postgres from the app host or use managed Postgres.
- Add connection pooling if DB connection pressure appears.
- Keep external load tests, not only internal container tests.
- Add fixture integrity checks to the load-test CLI.
- Preserve raw host telemetry CSVs and k6 summaries for cited runs, but do not commit raw summaries containing bearer tokens.

## Current Planning Position

The practical forecast today is:

- `50/sec`: proven on small staging for `1m`; safe target for small/model-office traffic with headroom on larger hosts.
- `100-200/sec`: likely feasible as sustained hot-market traffic on the larger host, but still should be retested after index work.
- `250/sec`: proven as a clean five-minute sustained hot-market run on the larger host.
- `300/sec`: proven as a one-minute burst on the larger host; fresh-state five-minute repeat failed strict clean criteria.
- `400/sec`: degraded on the larger host; not currently supported.
- `500/sec`: not currently supported.

The most important next work is not buying a bigger machine. It is fixing the bet-history query path, then repeating five-minute sustained tests.
