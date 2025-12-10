/**
 * Integration test for listing markets using the generated SDK.
 *
 * This suite talks to a real backend (pointed to by `SDK_BASE_URL`) and
 * verifies that the `/v0/markets` endpoint responds successfully and returns
 * a collection of markets.
 */
import { ensureAdminToken } from "./authBootstrap.ts";
import { getMarkets } from "../../api/v0/markets/getMarkets.ts";

describe("markets/getMarkets (integration)", () => {
  const baseUrl = process.env.SDK_BASE_URL ?? "http://localhost:8080";
  const ctx = { baseUrl };

  let token: string;
  beforeAll(async () => {
    token = await ensureAdminToken(ctx);
  });

  it("returns a successful response with a markets array", async () => {
    const headers = { Authorization: `Bearer ${token}` };

    const res = await getMarkets(ctx, { status: "all", limit: 5 }, headers);

    expect(res.ok).toBe(true);
    expect(res.status).toBe(200);
    expect(res.result).toBeDefined();
    expect(Array.isArray(res.result.markets)).toBe(true);
  });
});
