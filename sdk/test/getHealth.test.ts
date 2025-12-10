import { getHealth } from "../api/v0/config/getHealth.ts";

describe("config/getHealth", () => {
  it("calls /health and returns the wrapped response", async () => {
    const originalFetch = (global as any).fetch;

    const fetchMock = jest.fn(async () => ({
      ok: true,
      status: 200,
      headers: {
        get: () => "text/plain",
      },
      text: async () => "ok",
      json: async () => ({}),
      arrayBuffer: async () => new ArrayBuffer(0),
    } as any));

    (global as any).fetch = fetchMock;

    const ctx = { baseUrl: "http://localhost:8080" };
    const result = await getHealth(ctx);

    expect(fetchMock).toHaveBeenCalled();
    expect(result.ok).toBe(true);
    expect(result.status).toBe(200);
    expect(result.result).toBe("ok");

    (global as any).fetch = originalFetch;
  });
});
