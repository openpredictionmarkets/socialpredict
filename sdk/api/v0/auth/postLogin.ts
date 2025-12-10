import type * as Types from "../../../types.ts";
import { GeneratedSdkContext, sdkRequest, SdkResponse } from "../../../helpers.ts";

/**
 * Authenticate user
 *
 * Generated from POST /v0/login
 */
export async function postLogin(
  ctx: GeneratedSdkContext,
  params: Types.LoginRequest,
): Promise<SdkResponse<Types.LoginResponse>> {
  return sdkRequest<Types.LoginResponse, Types.LoginRequest>(ctx, {
    path: "/v0/login",
    method: "POST",
    body: params,
  });
}
