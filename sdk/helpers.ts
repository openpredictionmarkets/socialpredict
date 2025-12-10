/**
 * Shared SDK helpers and context for calling the SocialPredict backend.
 *
 * These utilities are intentionally minimal: they know how to build URLs,
 * attach query params, encode JSON bodies, and decode common response types,
 * but they do not impose any retry, auth, or logging behavior.
 */
export interface GeneratedSdkContext {
  /**
   * Base URL for the backend, e.g. "http://localhost:8080".
   * When omitted, the helpers default to this localhost URL.
   */
  baseUrl?: string;
}

export interface SdkRequestOptions<TBody = unknown> {
  /**
   * Path relative to the backend base URL, e.g. "/v0/markets".
   */
  path: string;
  /**
   * HTTP method (e.g. "GET", "POST").
   */
  method: string;
  /**
   * Optional query string parameters appended to the URL.
   */
  query?: Record<string, string | number | boolean | undefined>;
  /**
   * Optional request payload. Non-string bodies are JSON-encoded by default.
   */
  body?: TBody;
  /**
   * Additional HTTP headers to send with the request.
   */
  headers?: Record<string, string>;
}

/**
 * Normalized HTTP response returned by `sdkRequest`.
 *
 * - `ok` mirrors the `Response.ok` flag.
 * - `status` is the numeric HTTP status code.
 * - `result` is the decoded body (JSON, text, or ArrayBuffer).
 */
export interface SdkResponse<TResponse = unknown> {
  ok: boolean;
  status: number;
  result: TResponse;
}

/**
 * Perform an HTTP request against the SocialPredict backend.
 *
 * The function:
 * - Builds the URL from the context `baseUrl` and `options.path`.
 * - Appends any `query` parameters.
 * - JSON-encodes the `body` when it is not already a string.
 * - Decodes JSON responses into `TResponse`, text responses into `string`,
 *   and all other content types into `ArrayBuffer`.
 *
 * Unlike earlier versions, this helper never throws on non-2xx responses;
 * callers should inspect the returned `ok` and `status` fields.
 */
export async function sdkRequest<TResponse = unknown, TBody = unknown>(
  ctx: GeneratedSdkContext,
  options: SdkRequestOptions<TBody>,
): Promise<SdkResponse<TResponse>> {
  const baseUrl = ctx.baseUrl ?? "http://localhost:8080";
  const url = new URL(options.path, baseUrl);

  if (options.query) {
    for (const [key, value] of Object.entries(options.query)) {
      if (value === undefined) continue;
      url.searchParams.append(key, String(value));
    }
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers ?? {}),
  };

  const init: { method: string; headers: Record<string, string>; body?: any } = {
    method: options.method.toUpperCase(),
    headers,
  };

  if (options.body !== undefined && options.body !== null) {
    if (typeof options.body === "string") {
      init.body = options.body;
    } else {
      init.body = JSON.stringify(options.body);
    }
  }

  const response = await fetch(url.toString(), init as any);
  const contentType = response.headers?.get?.("content-type") ?? "";

  let result: unknown;

  if (contentType.includes("application/json")) {
    result = await response.json();
  } else if (contentType.startsWith("text/")) {
    result = await response.text();
  } else {
    // For other content types, return the raw ArrayBuffer.
    result = await response.arrayBuffer();
  }

  return {
    ok: response.ok,
    status: response.status,
    result: result as TResponse,
  };
}
