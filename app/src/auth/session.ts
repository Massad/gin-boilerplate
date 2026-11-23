/**
 * Session persistence for the Atrium authentication. One storage key holds the
 * discriminated payload returned by POST /v1/login: localStorage when
 * « Rester connecté » is checked, sessionStorage otherwise. Reads fall back
 * from sessionStorage to localStorage so a remembered session survives a
 * browser restart while a non-remembered one dies with the tab.
 */

export type SessionKind = "tenant_user" | "superadmin"

export interface TenantUserSessionData {
  kind: "tenant_user"
  user: {
    id: string
    email: string
    name: string
    tenant_id: string
    role: string
  }
  tenant: {
    id: string
    slug: string
    name: string
  }
  token: {
    access_token: string
    refresh_token: string
  }
}

export interface SuperadminSessionData {
  kind: "superadmin"
  admin: {
    id: string
    email: string
    name: string
  }
  token: {
    access_token: string
  }
}

/** The login payload as returned by POST /v1/login. */
export type SessionData = TenantUserSessionData | SuperadminSessionData

/** The stored payload plus the « Rester connecté » choice it was written with. */
export type Session = SessionData & { remember: boolean }

const SESSION_KEY = "atrium.session"

export function writeSession(session: SessionData, options: { remember: boolean }): void {
  const payload: Session = { ...session, remember: options.remember }
  const raw = JSON.stringify(payload)
  if (options.remember) {
    window.localStorage.setItem(SESSION_KEY, raw)
    window.sessionStorage.removeItem(SESSION_KEY)
  } else {
    window.sessionStorage.setItem(SESSION_KEY, raw)
    window.localStorage.removeItem(SESSION_KEY)
  }
}

export function readSession(): Session | null {
  const raw = window.sessionStorage.getItem(SESSION_KEY) ?? window.localStorage.getItem(SESSION_KEY)
  if (raw === null) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as Session
    if (!parsed || (parsed.kind !== "tenant_user" && parsed.kind !== "superadmin")) {
      return null
    }
    return { ...parsed, remember: parsed.remember === true }
  } catch {
    return null
  }
}

export function purgeSession(): void {
  window.sessionStorage.removeItem(SESSION_KEY)
  window.localStorage.removeItem(SESSION_KEY)
}

/** The display name of the authenticated account, for the salutation. */
export function sessionDisplayName(session: Session): string {
  return session.kind === "superadmin" ? session.admin.name : session.user.name
}
