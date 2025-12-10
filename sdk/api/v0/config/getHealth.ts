import { GeneratedSdkContext, sdkRequest, SdkResponse } from "../../../helpers.ts";

/**
 * Backend health check
 *
 * Mirrors GET /health from the OpenAPI specification.
 */
export async function getHealth(
  ctx: GeneratedSdkContext,
): Promise<SdkResponse<string>> {
  return sdkRequest<string>(ctx, { path: "/health", method: "GET" });
}
