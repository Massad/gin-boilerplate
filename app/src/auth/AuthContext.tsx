/**
 * Authentication context: the session read at boot, the login that persists it
 * and navigates by kind, and the logout that revokes the server session then
 * purges the browser storage. The silent 401 refresh does not live here — it
 * is owned by api.ts.
 */

import { useNavigate } from "@tanstack/react-router"
import { createContext, useCallback, useContext, useState } from "react"
import type { ReactNode } from "react"
import { request } from "@/lib/api"
import { purgeSession, readSession, writeSession } from "@/auth/session"
import type { Session, SessionData } from "@/auth/session"

export interface AuthContextValue {
  session: Session | null
  login: (payload: SessionData, options: { remember: boolean }) => void
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(() => readSession())
  const navigate = useNavigate()

  const login = useCallback(
    (payload: SessionData, options: { remember: boolean }) => {
      writeSession(payload, options)
      const stored = readSession()
      setSession(stored)
      navigate({ to: stored !== null && stored.kind === "superadmin" ? "/admin" : "/" })
    },
    [navigate],
  )

  const logout = useCallback(async () => {
    const current = readSession()
    try {
      if (current !== null && current.kind === "superadmin") {
        await request("/admin/logout", { method: "GET" })
      } else if (current !== null) {
        await request("/user/logout", { method: "GET" })
      }
    } finally {
      // The session is always purged client-side, even when the server call
      // fails (it may have already expired).
      purgeSession()
      setSession(null)
      navigate({ to: "/login" })
    }
  }, [navigate])

  return <AuthContext.Provider value={{ session, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext)
  if (value === null) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return value
}
