/**
 * Fetch wrapper for the /v1 API. `request` returns the parsed JSON of any 2xx
 * response and throws an ApiError carrying the status and body otherwise.
 *
 * The 401 interception lives here: a tenant session gets one silent refresh
 * through /v1/token/refresh (a single promise shared by every concurrent 401
 * so the endpoint is hit at most once), tokens are replaced through session.ts
 * and the failed request is replayed exactly once. A superadmin session has no
 * refresh: its 401 purges the session and returns to /login.
 */

import { purgeSession, readSession, writeSession } from "@/auth/session"
import type { Session, SessionData } from "@/auth/session"

export class ApiError extends Error {
  status: number
  body: unknown

  constructor(status: number, body: unknown) {
    super(`HTTP ${status}`)
    this.name = "ApiError"
    this.status = status
    this.body = body
  }
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  /** The JSON payload, serialized with the application/json content type. */
  json?: unknown
}

let refreshPromise: Promise<void> | null = null

async function parseBody(response: Response): Promise<unknown> {
  const raw = await response.text()
  if (raw === "") {
    return null
  }
  try {
    return JSON.parse(raw) as unknown
  } catch {
    return raw
  }
}

function navigateToLogin(): void {
  window.location.assign("/login")
}

/**
 * Silently exchanges the tenant refresh token for a fresh access/refresh
 * pair. Concurrent callers share this module's single in-flight promise.
 */
async function refreshTenantSession(): Promise<void> {
  const session = readSession()
  if (session === null || session.kind !== "tenant_user") {
    throw new ApiError(401, null)
  }
  const response = await window.fetch("/v1/token/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: session.token.refresh_token }),
  })
  if (!response.ok) {
    throw new ApiError(response.status, await parseBody(response))
  }
  const tokens = (await response.json()) as { access_token: string; refresh_token: string }
  writeSession({ ...session, token: tokens }, { remember: session.remember })
}

async function requestOnce(path: string, options: RequestOptions): Promise<Response> {
  const session = readSession()
  const headers = new Headers(options.headers)
  if (options.json !== undefined) {
    headers.set("Content-Type", "application/json")
  }
  if (session !== null && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${session.token.access_token}`)
  }
  return window.fetch(`/v1${path}`, {
    ...options,
    body: options.json === undefined ? null : JSON.stringify(options.json),
    headers,
  })
}

export async function request<T = unknown>(path: string, options: RequestOptions = {}): Promise<T> {
  let response = await requestOnce(path, options)

  if (response.status === 401) {
    const session = readSession()
    if (session === null || session.kind === "superadmin") {
      // A superadmin token is a single 8h access token with no refresh; a
      // request without any stored session has nothing to recover either.
      purgeSession()
      navigateToLogin()
      throw new ApiError(401, await parseBody(response))
    }
    // Single shared refresh promise: concurrent 401s trigger the refresh at
    // most once, then every caller replays its request exactly once.
    refreshPromise ??= refreshTenantSession().finally(() => {
      refreshPromise = null
    })
    try {
      await refreshPromise
    } catch (error) {
      // The refresh itself failed (revoked or expired refresh token): the
      // session is unrecoverable.
      purgeSession()
      navigateToLogin()
      throw error
    }
    response = await requestOnce(path, options)
  }

  if (!response.ok) {
    throw new ApiError(response.status, await parseBody(response))
  }
  return (await response.json()) as T
}

/** The discriminated payload of POST /v1/login, persisted by AuthContext. */
export type LoginResponse = SessionData

/** Authenticates through the unified endpoint. */
export function loginRequest(email: string, password: string): Promise<LoginResponse> {
  return request<LoginResponse>("/login", {
    method: "POST",
    json: { email, password },
  })
}

/** The stored session shape (re-exported for consumers of this module). */
export type { Session }
