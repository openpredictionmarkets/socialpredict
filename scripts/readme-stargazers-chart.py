#!/usr/bin/env python3
"""Generate the README stargazers-over-time chart from GitHub stargazer data."""

from __future__ import annotations

import argparse
import html
import json
import os
import shutil
import subprocess
from collections import Counter
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Iterable
from urllib.error import HTTPError
from urllib.request import Request, urlopen


GITHUB_API = "https://api.github.com"


@dataclass(frozen=True)
class ChartPoint:
    label: str
    count: int


def parse_month(timestamp: str) -> str:
    return datetime.fromisoformat(timestamp.replace("Z", "+00:00")).strftime("%Y-%m")


def build_monthly_points(starred_at_values: Iterable[str]) -> list[ChartPoint]:
    month_counts = Counter(parse_month(value) for value in starred_at_values)
    if not month_counts:
        return []

    months = sorted(month_counts)
    running_total = 0
    points: list[ChartPoint] = []
    for month in months:
        running_total += month_counts[month]
        points.append(ChartPoint(label=month, count=running_total))
    return points


def parse_next_link(link_header: str | None) -> str | None:
    if not link_header:
        return None
    for part in link_header.split(","):
        pieces = part.strip().split(";")
        if len(pieces) < 2:
            continue
        url_part = pieces[0].strip()
        rels = [piece.strip() for piece in pieces[1:]]
        if 'rel="next"' in rels and url_part.startswith("<") and url_part.endswith(">"):
            return url_part[1:-1]
    return None


def fetch_stargazers(repo: str, token: str | None = None) -> list[dict]:
    url = f"{GITHUB_API}/repos/{repo}/stargazers?per_page=100"
    stargazers: list[dict] = []

    while url:
        request = Request(
            url,
            headers={
                "Accept": "application/vnd.github.star+json",
                "X-GitHub-Api-Version": "2022-11-28",
                "User-Agent": "socialpredict-readme-stargazers-chart",
            },
        )
        if token:
            request.add_header("Authorization", f"Bearer {token}")

        try:
            with urlopen(request) as response:
                stargazers.extend(json.loads(response.read().decode("utf-8")))
                url = parse_next_link(response.headers.get("Link"))
        except HTTPError as error:
            detail = error.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"GitHub stargazers request failed with HTTP {error.code}: {detail}") from error

    return stargazers


def resolve_token(cli_token: str | None = None) -> str | None:
    if cli_token:
        return cli_token
    env_token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
    if env_token:
        return env_token
    if not shutil.which("gh"):
        return None
    try:
        return subprocess.check_output(["gh", "auth", "token"], text=True, stderr=subprocess.DEVNULL).strip() or None
    except (OSError, subprocess.CalledProcessError):
        return None


def _nice_max(value: int) -> int:
    if value <= 10:
        return 10
    magnitude = 10 ** (len(str(value)) - 1)
    for multiplier in (1, 2, 2.5, 5, 10):
        candidate = multiplier * magnitude
        if candidate >= value:
            return int(candidate)
    return value


def render_svg(repo: str, points: list[ChartPoint]) -> str:
    width = 760
    height = 320
    margin_left = 62
    margin_right = 28
    margin_top = 44
    margin_bottom = 54
    plot_width = width - margin_left - margin_right
    plot_height = height - margin_top - margin_bottom
    max_count = _nice_max(max((point.count for point in points), default=0))

    def x_for(index: int) -> float:
        if len(points) <= 1:
            return margin_left
        return margin_left + (index / (len(points) - 1)) * plot_width

    def y_for(count: int) -> float:
        return margin_top + plot_height - (count / max_count) * plot_height

    polyline = " ".join(f"{x_for(index):.1f},{y_for(point.count):.1f}" for index, point in enumerate(points))
    area = ""
    if points:
        area = (
            f"{margin_left},{margin_top + plot_height} "
            f"{polyline} "
            f"{margin_left + plot_width},{margin_top + plot_height}"
        )

    y_ticks = [0, max_count // 4, max_count // 2, (max_count * 3) // 4, max_count]
    y_grid = "\n".join(
        f'<line x1="{margin_left}" y1="{y_for(tick):.1f}" x2="{width - margin_right}" y2="{y_for(tick):.1f}" stroke="#243244" stroke-width="1" />'
        for tick in y_ticks
    )
    y_labels = "\n".join(
        f'<text x="{margin_left - 12}" y="{y_for(tick) + 4:.1f}" text-anchor="end" class="axis">{tick}</text>'
        for tick in y_ticks
    )

    x_labels = ""
    if points:
        label_indexes = sorted(set([0, len(points) // 2, len(points) - 1]))
        x_labels = "\n".join(
            f'<text x="{x_for(index):.1f}" y="{height - 22}" text-anchor="middle" class="axis">{html.escape(points[index].label)}</text>'
            for index in label_indexes
        )

    total = points[-1].count if points else 0
    escaped_repo = html.escape(repo)

    return f"""<svg width="{width}" height="{height}" viewBox="0 0 {width} {height}" fill="none" xmlns="http://www.w3.org/2000/svg" role="img" aria-labelledby="title desc">
  <title id="title">SocialPredict stargazers over time</title>
  <desc id="desc">Cumulative GitHub stargazers for {escaped_repo}, generated from GitHub stargazer timestamps.</desc>
  <style>
    .title {{ fill: #f8fafc; font: 700 22px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }}
    .subtitle {{ fill: #9fb2c8; font: 500 13px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }}
    .axis {{ fill: #9fb2c8; font: 500 12px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }}
    .total {{ fill: #f8fafc; font: 700 28px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }}
  </style>
  <rect width="{width}" height="{height}" rx="8" fill="#0d1117" />
  <text x="{margin_left}" y="30" class="title">Stargazers over time</text>
  <text x="{width - margin_right}" y="28" text-anchor="end" class="total">{total:,}</text>
  <text x="{width - margin_right}" y="47" text-anchor="end" class="subtitle">stars</text>
  <text x="{margin_left}" y="{height - 8}" class="subtitle">Generated from GitHub stargazer timestamps</text>
  {y_grid}
  <line x1="{margin_left}" y1="{margin_top + plot_height}" x2="{width - margin_right}" y2="{margin_top + plot_height}" stroke="#3a4658" stroke-width="1.5" />
  <line x1="{margin_left}" y1="{margin_top}" x2="{margin_left}" y2="{margin_top + plot_height}" stroke="#3a4658" stroke-width="1.5" />
  {y_labels}
  {x_labels}
  <polygon points="{area}" fill="#62c3f8" opacity="0.16" />
  <polyline points="{polyline}" fill="none" stroke="#62c3f8" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />
  {f'<circle cx="{x_for(len(points) - 1):.1f}" cy="{y_for(total):.1f}" r="5" fill="#2ea043" stroke="#f8fafc" stroke-width="2" />' if points else ''}
</svg>
"""


def write_outputs(repo: str, stargazers: list[dict], svg_path: Path, json_path: Path) -> None:
    starred_at_values = [item["starred_at"] for item in stargazers if item.get("starred_at")]
    points = build_monthly_points(starred_at_values)

    svg_path.parent.mkdir(parents=True, exist_ok=True)
    json_path.parent.mkdir(parents=True, exist_ok=True)
    svg_path.write_text(render_svg(repo, points), encoding="utf-8")
    json_path.write_text(
        json.dumps(
            {
                "repo": repo,
                "source": "GitHub stargazers API with starred_at timestamps",
                "total_stargazers": points[-1].count if points else 0,
                "points": [{"month": point.label, "total": point.count} for point in points],
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default=os.environ.get("GITHUB_REPOSITORY", "openpredictionmarkets/socialpredict"))
    parser.add_argument("--output", default="README/METRICS/stargazers-over-time.svg")
    parser.add_argument("--json-output", default="README/METRICS/stargazers-over-time.json")
    parser.add_argument("--token", default=None)
    args = parser.parse_args()

    stargazers = fetch_stargazers(args.repo, resolve_token(args.token))
    write_outputs(args.repo, stargazers, Path(args.output), Path(args.json_output))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
