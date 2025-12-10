import { postLogin } from "../../api/v0/auth/postLogin.ts";
import { postChangePassword } from "../../api/v0/users/postChangepassword.ts";
import { LoginResponseSchema } from "../../schema/auth.ts";
import type { GeneratedSdkContext } from "../../helpers.ts";

let cachedToken: string | null = null;

/**
 * Ensure the admin account has the desired password and return a fresh JWT.
 * Idempotent across test runs: if the password is already changed, it just logs in.
 */
export async function ensureAdminToken(ctx: GeneratedSdkContext): Promise<string> {
  if (cachedToken) return cachedToken;

  const username = process.env.SDK_ADMIN_USERNAME ?? "admin";
  const initialPassword = process.env.SDK_ADMIN_INITIAL_PASSWORD ?? "password";
  const currentPassword = process.env.SDK_ADMIN_PASSWORD ?? initialPassword;

  // 1) Try logging in with the current/desired password.
  {
    const res = await postLogin(ctx, { username, password: currentPassword });
    if (res.ok) {
      const parsed = LoginResponseSchema.parse(res.result);
      cachedToken = parsed.token;
      return cachedToken;
    }
  }

  // 2) Fall back to initial password and run the changePassword flow.
  const initialLogin = await postLogin(ctx, { username, password: initialPassword });
  if (!initialLogin.ok) {
    throw new Error("Cannot log in with either current or initial admin credentials");
  }
  const initialParsed = LoginResponseSchema.parse(initialLogin.result);

  const changeRes = await postChangePassword(
    ctx,
    { currentPassword: initialPassword, newPassword: currentPassword },
    initialParsed.token,
  );
  if (!changeRes.ok) {
    throw new Error(`changePassword failed with status ${changeRes.status}`);
  }

  // 3) Log in again with the desired password.
  const finalLogin = await postLogin(ctx, { username, password: currentPassword });
  if (!finalLogin.ok) {
    throw new Error("Login failed after changePassword ran");
  }
  const finalParsed = LoginResponseSchema.parse(finalLogin.result);
  cachedToken = finalParsed.token;
  return cachedToken;
}
