import type * as Types from "../../../types.ts";
import { GeneratedSdkContext, sdkRequest, SdkResponse } from "../../../helpers.ts";

export async function postChangePassword(
  ctx: GeneratedSdkContext,
  body: Types.ChangePasswordRequest,
  token: string,
): Promise<SdkResponse<unknown>> {
  return sdkRequest<unknown, Types.ChangePasswordRequest>(ctx, {
    path: "/v0/changepassword",
    method: "POST",
    body,
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}