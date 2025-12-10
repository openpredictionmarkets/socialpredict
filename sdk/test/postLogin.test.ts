import { postLogin } from "../api/v0/auth/postLogin.ts";

describe("auth/postLogin", () => {
  it("POSTs credentials and returns the wrapped response", async () => {
    const originalFetch = (global as any).fetch;

    const responseBody = {
      token: "fake-token",
      username: "admin",
      usertype: "admin",
      mustChangePassword: false,
    };

    const fetchMock = jest.fn(async () => ({
      ok: true,
      status: 200,
      headers: {
        get: () => "application/json",
      },
      text: async () => JSON.stringify(responseBody),
      json: async () => responseBody,
      arrayBuffer: async () => new ArrayBuffer(0),
    } as any)) as any;

    (global as any).fetch = fetchMock;

    const ctx = { baseUrl: "http://localhost:8080" };
    const credentials = { username: "admin", password: "password" };
    const result = await postLogin(ctx, credentials);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as [string, any];
    expect(url).toBe("http://localhost:8080/v0/login");
    expect(init.method).toBe("POST");
    expect(init.headers["Content-Type"]).toBe("application/json");
    expect(init.body).toBe(JSON.stringify(credentials));

    expect(result.ok).toBe(true);
    expect(result.status).toBe(200);
    expect(result.result).toEqual(responseBody);

    (global as any).fetch = originalFetch;
  });
});
