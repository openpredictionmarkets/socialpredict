import type * as Types from "../../../types.ts";
import { GeneratedSdkContext, sdkRequest, SdkResponse } from "../../../helpers.ts";

export interface GetMarketsParams {
  status?: "active" | "closed" | "resolved" | "all";
  created_by?: string;
  limit?: number;
  offset?: number;
}

/**
 * List markets
 *
 * Generated from GET /v0/markets
 */
export async function getMarkets(
  ctx: GeneratedSdkContext,
  params?: GetMarketsParams,
  headers?: Record<string, string>,
): Promise<SdkResponse<Types.ListMarketsResponse>> {
  const query: Record<string, string | number | boolean | undefined> = {};

  if (params) {
    if (params.status) {
      query.status = params.status;
    }
    if (params.created_by) {
      query.created_by = params.created_by;
    }
    if (params.limit !== undefined) {
      query.limit = params.limit;
    }
    if (params.offset !== undefined) {
      query.offset = params.offset;
    }
  }

  return sdkRequest<Types.ListMarketsResponse>(ctx, {
    path: "/v0/markets",
    method: "GET",
    query,
    headers
  });
}
