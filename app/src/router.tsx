/**
 * Code-based TanStack Router tree. The guards read the persisted session: a
 * route needs an authenticated session of its expected kind, a missing session
 * falls back to /login, and a session of the other kind is sent to its own
 * home (an admin on the tenant home is redirected to /admin and vice versa).
 */

import { createRootRoute, createRoute, createRouter, Outlet, redirect } from "@tanstack/react-router"
import { AuthProvider } from "@/auth/AuthContext"
import { readSession } from "@/auth/session"
import AdminHomePage from "@/pages/AdminHomePage"
import LoginPage from "@/pages/LoginPage"
import TenantHomePage from "@/pages/TenantHomePage"

const rootRoute = createRootRoute({
  component: () => {
    return (
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    )
  },
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: () => {
    const session = readSession()
    if (session === null) {
      throw redirect({ to: "/login" })
    }
    if (session.kind === "superadmin") {
      throw redirect({ to: "/admin" })
    }
  },
  component: TenantHomePage,
})

const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/admin",
  beforeLoad: () => {
    const session = readSession()
    if (session === null) {
      throw redirect({ to: "/login" })
    }
    if (session.kind === "tenant_user") {
      throw redirect({ to: "/" })
    }
  },
  component: AdminHomePage,
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
})

const routeTree = rootRoute.addChildren([indexRoute, adminRoute, loginRoute])

export const router = createRouter({ routeTree })

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
