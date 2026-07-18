import importlib.util
import sys
import unittest
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().parents[1] / "readme-stargazers-chart.py"
SPEC = importlib.util.spec_from_file_location("readme_stargazers_chart", SCRIPT_PATH)
chart = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = chart
SPEC.loader.exec_module(chart)


class ReadmeStargazersChartTests(unittest.TestCase):
    def test_build_monthly_points_returns_cumulative_totals(self):
        points = chart.build_monthly_points(
            [
                "2026-02-03T12:00:00Z",
                "2026-01-01T00:00:00Z",
                "2026-02-20T23:59:59Z",
            ]
        )

        self.assertEqual(
            points,
            [
                chart.ChartPoint(label="2026-01", count=1),
                chart.ChartPoint(label="2026-02", count=3),
            ],
        )

    def test_parse_next_link_finds_next_relation(self):
        link = '<https://api.github.com/repositories/1/stargazers?page=2>; rel="next", <https://api.github.com/repositories/1/stargazers?page=3>; rel="last"'

        self.assertEqual(
            chart.parse_next_link(link),
            "https://api.github.com/repositories/1/stargazers?page=2",
        )

    def test_nice_max_uses_tight_readable_scale(self):
        self.assertEqual(chart._nice_max(212), 250)

    def test_render_svg_contains_chart_metadata_and_polyline(self):
        svg = chart.render_svg(
            "openpredictionmarkets/socialpredict",
            [
                chart.ChartPoint(label="2026-01", count=1),
                chart.ChartPoint(label="2026-02", count=3),
            ],
        )

        self.assertIn("<svg", svg)
        self.assertIn("Stargazers over time", svg)
        self.assertIn("openpredictionmarkets/socialpredict", svg)
        self.assertIn("<polyline", svg)
        self.assertIn(">3<", svg)


if __name__ == "__main__":
    unittest.main()
