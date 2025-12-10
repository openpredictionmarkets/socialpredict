/**
 * Integration test for the login flow using the generated SDK.
 *
 * This suite exercises a real backend (pointed to by `SDK_BASE_URL`) and
 * validates the response shape with Zod. It is intentionally higher level
 * than the unit tests under `sdk/test`.
 */
import { postLogin } from "../../api/v0/auth/postLogin.ts";
import { LoginResponseSchema } from "../../schema/auth.ts";

describe("auth/postLogin (integration)", () => {
  const baseUrl = process.env.SDK_BASE_URL ?? "http://localhost:8080";

  it("logs in admin and returns a valid response", async () => {
    const ctx = { baseUrl };
    const credentials = { username: "admin", password: "password" };

    const raw = await postLogin(ctx, credentials);

    const parsed = LoginResponseSchema.parse(raw.result); // throws if invalid

    expect(parsed.username).toBe("admin");
    expect(parsed.token).toBeDefined();
    expect(parsed.usertype).toBeDefined();
    expect(parsed.mustChangePassword).toBeDefined();
  });
});
