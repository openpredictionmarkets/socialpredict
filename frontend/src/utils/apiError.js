/**
 * parseApiError — standard convention for extracting backend error messages.
 *
 * The backend (WriteJSONError / WriteInternalError from errors/apierror.go) always
 * returns { "error": "<message>" } with the appropriate HTTP status code.
 * This function handles that format and falls back gracefully for plain-text
 * or empty bodies.
 *
 * Usage:
 *   if (!response.ok) {
 *     const msg = await parseApiError(response);
 *     throw new Error(msg);
 *   }
 *
 * @param {Response} response - A non-ok fetch Response
 * @returns {Promise<string>} Human-readable error message
 */
export async function parseApiError(response) {
  let text;
  try {
    text = await response.text();
  } catch {
    return `HTTP ${response.status}`;
  }

  if (!text) {
    return `HTTP ${response.status}`;
  }

  try {
    const data = JSON.parse(text);
    // Backend convention: { "error": "..." }
    if (data.error) return data.error;
    // Tolerate legacy shapes
    if (data.message) return data.message;
  } catch {
    // Not JSON — return as plain text
  }

  return text;
}
