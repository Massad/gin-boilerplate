/**
 * Tests de la vue de connexion (app/src/pages/LoginPage.tsx).
 *
 * Chaque test rend la page dans un routeur neuf, construit avec un
 * historique mémoire positionné sur /login : le routeur de @/router est un
 * singleton qui conserve son emplacement d'un test à l'autre, si bien que les
 * redirections après connexion contamineraient les tests suivants.
 */

import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from "@tanstack/react-router"
import { delay, http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { AuthProvider } from "@/auth/AuthContext"
import { purgeSession } from "@/auth/session"
import AdminHomePage from "@/pages/AdminHomePage"
import LoginPage from "@/pages/LoginPage"
import TenantHomePage from "@/pages/TenantHomePage"

const server = setupServer()

const LOGIN_ERROR_TEXT = "Adresse e-mail ou mot de passe incorrect. Vérifiez vos identifiants et réessayez."
const COMING_SOON_TEXT = "Cette fonctionnalité n'est pas encore disponible."

const tenantPayload = {
  kind: "tenant_user" as const,
  user: {
    id: "u-1",
    email: "camille@studio-nord.fr",
    name: "Camille Laurent",
    tenant_id: "t-1",
    role: "owner",
  },
  tenant: { id: "t-1", slug: "studio-nord", name: "Studio Nord" },
  token: { access_token: "access-1", refresh_token: "refresh-1" },
}

const superadminPayload = {
  kind: "superadmin" as const,
  admin: { id: "a-1", email: "admin@atrium.dev", name: "Laure Moreau" },
  token: { access_token: "admin-access" },
}

beforeAll(() => {
  server.listen({ onUnhandledFrame: "bypass" })
})

afterEach(() => {
  server.resetHandlers()
  purgeSession()
  window.localStorage.clear()
  window.sessionStorage.clear()
  cleanup()
})

afterAll(() => {
  server.close()
})

beforeEach(() => {
  document.title = "untitled"
})

/**
 * Rend la page de connexion dans un routeur neuf : la racine porte le
 * AuthProvider, les trois vues correspondent aux pages de l'application. Les
 * gardes beforeLoad du routeur de production sont volontairement absentes :
 * les pages "/", "/admin" et "/login" doivent rester atteignables pour que
 * les redirections après connexion soient observables.
 */
async function renderLoginApp(): Promise<void> {
  const rootRoute = createRootRoute({
    component: () => (
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    ),
  })
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    component: TenantHomePage,
  })
  const adminRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/admin",
    component: AdminHomePage,
  })
  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/login",
    component: LoginPage,
  })
  const routeTree = rootRoute.addChildren([indexRoute, adminRoute, loginRoute])

  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: ["/login"] }),
  })

  render(
    <QueryClientProvider client={new QueryClient()}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )

  // Le routeur résout la correspondance de façon asynchrone : la page
  // n'apparaît qu'une fois le chargement initial terminé.
  await waitFor(() => {
    expect(screen.getByRole("heading", { name: "Connexion" })).toBeInTheDocument()
  })
}

/** La seule vue visible de la page : les autres cartes sont masquées. */
function visibleCard(): HTMLElement {
  const card = document.querySelector<HTMLElement>("section:not([hidden])")
  if (card === null) {
    throw new Error("no visible card")
  }
  return card
}

function fillLoginFields(email: string, password: string): void {
  fireEvent.change(document.getElementById("email")!, { target: { value: email } })
  fireEvent.change(document.getElementById("password")!, { target: { value: password } })
}

describe("LoginPage", () => {
  it("renders the login card with its fields and links", async () => {
    await renderLoginApp()

    expect(screen.getByRole("heading", { name: "Connexion" })).toBeInTheDocument()
    expect(document.getElementById("email")).toBeInTheDocument()
    expect(document.getElementById("password")).toBeInTheDocument()
    expect(screen.getByRole("checkbox", { name: /Rester connecté/ })).not.toBeChecked()
    expect(screen.getByRole("link", { name: /Mot de passe oublié \?/ })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: /Demander un accès/ })).toBeInTheDocument()
    expect(screen.getByRole("link", { name: /Créer une organisation/ })).toBeInTheDocument()
  })

  it("toggles the password visibility with the eye button", async () => {
    await renderLoginApp()

    fireEvent.click(screen.getByRole("button", { name: "Afficher le mot de passe" }))

    expect(screen.getByPlaceholderText("Votre mot de passe")).toHaveAttribute("type", "text")
    expect(screen.getByRole("button", { name: "Masquer le mot de passe" })).toBeInTheDocument()
  })

  it("blocks an incomplete submission with the maquette banner and no API call", async () => {
    await renderLoginApp()

    fireEvent.click(screen.getByRole("button", { name: "Se connecter" }))

    expect(screen.getByRole("alert")).toHaveTextContent(LOGIN_ERROR_TEXT)
    expect(window.sessionStorage.getItem("atrium.session")).toBeNull()
    expect(window.localStorage.getItem("atrium.session")).toBeNull()
  })

  it("rejects invalid credentials with the visible banner", async () => {
    server.use(
      http.post("/v1/login", () => HttpResponse.json({ message: "Invalid login details" }, { status: 406 })),
    )
    await renderLoginApp()

    fillLoginFields("camille@studio-nord.fr", "wrong-password")
    fireEvent.click(screen.getByRole("button", { name: "Se connecter" }))

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(LOGIN_ERROR_TEXT)
    })
    expect(window.sessionStorage.getItem("atrium.session")).toBeNull()
    expect(window.localStorage.getItem("atrium.session")).toBeNull()
  })

  it("logs a superadmin in and redirects to /admin without remembering", async () => {
    let sent: { email: string; password: string } | null = null
    server.use(
      http.post("/v1/login", async ({ request }) => {
        sent = (await request.json()) as { email: string; password: string }
        await delay(10)
        return HttpResponse.json(superadminPayload)
      }),
    )
    await renderLoginApp()

    fillLoginFields("admin@atrium.dev", "admin-password")
    fireEvent.click(screen.getByRole("button", { name: "Se connecter" }))

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Admin" })).toBeInTheDocument()
      expect(screen.getByText(/Bienvenue Laure Moreau/)).toBeInTheDocument()
    })
    expect(sent).toEqual({ email: "admin@atrium.dev", password: "admin-password" })
    // « Rester connecté » laissé décoché : la session reste en sessionStorage.
    expect(JSON.parse(window.sessionStorage.getItem("atrium.session")!)).toMatchObject(superadminPayload)
    expect(window.localStorage.getItem("atrium.session")).toBeNull()
  })

  it("logs a tenant user in, remembers the session and redirects to /", async () => {
    server.use(
      http.post("/v1/login", async () => {
        await delay(10)
        return HttpResponse.json(tenantPayload)
      }),
    )
    await renderLoginApp()

    fillLoginFields("camille@studio-nord.fr", "tenant-password")
    fireEvent.click(screen.getByRole("checkbox", { name: /Rester connecté/ }))
    fireEvent.click(screen.getByRole("button", { name: "Se connecter" }))

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Tenant" })).toBeInTheDocument()
      expect(screen.getByText(/Bienvenue Camille Laurent/)).toBeInTheDocument()
    })
    // « Rester connecté » coché : la session est écrite dans localStorage.
    expect(JSON.parse(window.localStorage.getItem("atrium.session")!)).toMatchObject(tenantPayload)
    expect(window.sessionStorage.getItem("atrium.session")).toBeNull()
  })

  it("switches to the create-organization view, previews the slug and shows the coming-soon banner", async () => {
    await renderLoginApp()

    fireEvent.click(screen.getByRole("link", { name: /Créer une organisation/ }))

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Créer une organisation" })).toBeInTheDocument()
    })
    expect(document.title).toBe("Créer une organisation — Plateforme multi-tenant")
    expect(screen.getByText(/domain.com\/votre-societe/)).toBeInTheDocument()

    fireEvent.change(within(visibleCard()).getByPlaceholderText("Studio Nord"), {
      target: { value: "Studio Nord !" },
    })
    expect(screen.getByText(/domain.com\/studio-nord/)).toBeInTheDocument()

    fireEvent.click(within(visibleCard()).getByRole("button", { name: "Créer l'organisation" }))
    expect(within(visibleCard()).getByRole("alert")).toHaveTextContent(COMING_SOON_TEXT)
  })

  it("switches to the forgot view and back to the login card", async () => {
    await renderLoginApp()

    fireEvent.click(screen.getByRole("link", { name: /Mot de passe oublié \?/ }))

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Mot de passe oublié" })).toBeInTheDocument()
    })
    expect(document.title).toBe("Mot de passe oublié — Plateforme multi-tenant")

    fireEvent.click(screen.getByRole("button", { name: "Envoyer le lien de réinitialisation" }))
    expect(screen.getByRole("alert")).toHaveTextContent(COMING_SOON_TEXT)

    fireEvent.click(within(visibleCard()).getByRole("link", { name: "Retour à la connexion" }))

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Connexion" })).toBeInTheDocument()
    })
    expect(screen.queryByRole("alert")).toBeNull()
  })
})
