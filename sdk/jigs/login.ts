import { postLogin } from "../api/v0/auth/postLogin.ts";

async function main() {
  const ctx = {
    baseUrl: "http://localhost:8080",
  };

  const credentials = {
    username: "admin",
    password: "password",
  };

  try {
    const response = await postLogin(ctx, credentials);
    // eslint-disable-next-line no-console
    console.log(JSON.stringify(response, null, 2));
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error("Login failed:", error);
    process.exitCode = 1;
  }
}

void main();
