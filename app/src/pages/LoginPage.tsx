/**
 * Login view ported from the maquette's login.html: four locally switched view
 * cards (login | create | request | forgot), the SSO block removed, and a
 * single backend call on the login card — the unified POST /v1/login mutation.
 * The secondary views are display-only: their submission shows the « à venir »
 * banner and never reaches the API.
 */

import { useMutation } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import type { FormEvent } from "react"
import { z } from "zod"
import { AuthLayout } from "@/design-system/components/AuthLayout"
import { Button } from "@/design-system/components/Button"
import { CheckControl, ErrorBanner, Field, Input, PasswordInput, TextArea } from "@/design-system/components/Field"
import { BuildingIcon, ChevronLeftIcon, MailIcon, UsersIcon } from "@/design-system/components/icons"
import { useAuth } from "@/auth/AuthContext"
import { ApiError, loginRequest } from "@/lib/api"

type LoginView = "login" | "create" | "request" | "forgot"

const VIEW_TITLES: Record<LoginView, string> = {
  login: "Connexion — Plateforme multi-tenant",
  create: "Créer une organisation — Plateforme multi-tenant",
  request: "Demander un accès — Plateforme multi-tenant",
  forgot: "Mot de passe oublié — Plateforme multi-tenant",
}

/** The maquette's login-page banner text, used for every rejected attempt. */
const LOGIN_ERROR_TEXT = "Adresse e-mail ou mot de passe incorrect. Vérifiez vos identifiants et réessayez."
const COMING_SOON_TEXT = "Cette fonctionnalité n'est pas encore disponible."

// Client-side scope of the unified login: both fields are required, the email
// must be well-formed (server-side validation remains authoritative).
const loginSchema = z.object({
  email: z.string().trim().min(1).email(),
  password: z.string().min(1),
})

// The maquette's slug preview rules.
function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-+|-+$/g, "")
}

function stopNavigation(event: { preventDefault(): void }) {
  event.preventDefault()
}

export default function LoginPage() {
  const auth = useAuth()
  const [view, setView] = useState<LoginView>("login")

  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [remember, setRemember] = useState(false)
  const [loginError, setLoginError] = useState<string | null>(null)

  const [orgName, setOrgName] = useState("")
  const [orgEmail, setOrgEmail] = useState("")
  const [orgPassword, setOrgPassword] = useState("")
  const [createBanner, setCreateBanner] = useState<string | null>(null)

  const [reqName, setReqName] = useState("")
  const [reqEmail, setReqEmail] = useState("")
  const [reqOrg, setReqOrg] = useState("")
  const [reqMessage, setReqMessage] = useState("")
  const [requestBanner, setRequestBanner] = useState<string | null>(null)

  const [forgotEmail, setForgotEmail] = useState("")
  const [forgotBanner, setForgotBanner] = useState<string | null>(null)

  // The maquette switches the document title and focuses the first field of
  // the secondary views on every view change.
  useEffect(() => {
    document.title = VIEW_TITLES[view]
    const selector = view === "create" ? "#orgName" : view === "request" ? "#reqName" : view === "forgot" ? "#forgotEmail" : null
    if (selector !== null) {
      document.querySelector<HTMLInputElement>(selector)?.focus()
    }
  }, [view])

  const loginMutation = useMutation({
    mutationFn: (variables: { email: string; password: string; remember: boolean }) =>
      loginRequest(variables.email, variables.password),
    onSuccess: (payload, variables) => {
      auth.login(payload, { remember: variables.remember })
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 403) {
        const body = error.body as { message?: string } | null
        setLoginError(body?.message ?? LOGIN_ERROR_TEXT)
        return
      }
      setLoginError(LOGIN_ERROR_TEXT)
    },
  })

  function handleLoginSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const parsed = loginSchema.safeParse({ email, password })
    if (!parsed.success) {
      setLoginError(LOGIN_ERROR_TEXT)
      return
    }
    loginMutation.mutate({ email: parsed.data.email, password: parsed.data.password, remember })
  }

  const slug = slugify(orgName) || "votre-societe"

  return (
    <AuthLayout>
      {/* ── Connexion ─────────────────────────────────────────── */}
      <section className="form-card" hidden={view !== "login"}>
        <div className="mobile-brand">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <header className="form-head">
          <h2>Connexion</h2>
          <p>Signez avec vos identifiants — vous serez redirigé vers votre espace.</p>
        </header>

        {loginError !== null ? <ErrorBanner>{loginError}</ErrorBanner> : null}

        <form noValidate onSubmit={handleLoginSubmit}>
          <Field label="Adresse e-mail" htmlFor="email">
            <Input
              icon={<MailIcon aria-hidden="true" />}
              id="email"
              type="email"
              placeholder="vous@exemple.com"
              autoComplete="email"
              value={email}
              onChange={(event) => {
                setEmail(event.target.value)
                setLoginError(null)
              }}
            />
          </Field>

          <Field label="Mot de passe" htmlFor="password">
            <PasswordInput
              id="password"
              placeholder="Votre mot de passe"
              autoComplete="current-password"
              value={password}
              onChange={(event) => {
                setPassword(event.target.value)
                setLoginError(null)
              }}
            />
          </Field>

          <div className="row-between">
            <label className="checkbox">
              <CheckControl
                id="remember"
                checked={remember}
                onChange={(event) => {
                  setRemember(event.target.checked)
                }}
              />
              <span>Rester connecté</span>
            </label>
            <a href="#" onClick={(event) => { stopNavigation(event); setView("forgot") }}>
              Mot de passe oublié ?
            </a>
          </div>

          <Button type="submit" block busy={loginMutation.isPending}>
            <span className="btn-label">Se connecter</span>
          </Button>
        </form>

        <p className="form-foot">
          Pas encore de compte ?{" "}
          <a href="#" onClick={(event) => { stopNavigation(event); setView("request") }}>
            Demander un accès
          </a>
          <br />
          <a href="#" onClick={(event) => { stopNavigation(event); setView("create") }}>
            Créer une organisation
          </a>
        </p>
      </section>

      {/* ── Créer une organisation ────────────────────────────── */}
      <section className="form-card" hidden={view !== "create"}>
        <div className="mobile-brand">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <header className="form-head">
          <a href="#" className="back-link" onClick={(event) => { stopNavigation(event); setView("login") }}>
            <ChevronLeftIcon aria-hidden="true" />
            <span>Retour à la connexion</span>
          </a>
          <h2>Créer une organisation</h2>
          <p>Créez l'espace de travail de votre société sur Atrium.</p>
        </header>

        {createBanner !== null ? <ErrorBanner>{createBanner}</ErrorBanner> : null}

        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            setCreateBanner(COMING_SOON_TEXT)
          }}
        >
          <Field label="Nom de la société" htmlFor="orgName" help={<>Votre espace sera accessible sur <span>domain.com/{slug}</span></>}>
            <Input
              icon={<BuildingIcon aria-hidden="true" />}
              id="orgName"
              type="text"
              placeholder="Studio Nord"
              autoComplete="organization"
              value={orgName}
              onChange={(event) => {
                setOrgName(event.target.value)
                setCreateBanner(null)
              }}
            />
          </Field>

          <Field label="E-mail de l'administrateur" htmlFor="adminEmail">
            <Input
              icon={<MailIcon aria-hidden="true" />}
              id="adminEmail"
              type="email"
              placeholder="vous@exemple.com"
              autoComplete="email"
              value={orgEmail}
              onChange={(event) => {
                setOrgEmail(event.target.value)
                setCreateBanner(null)
              }}
            />
          </Field>

          <Field label="Mot de passe" htmlFor="adminPassword">
            <PasswordInput
              id="adminPassword"
              placeholder="Créez un mot de passe"
              autoComplete="new-password"
              value={orgPassword}
              onChange={(event) => {
                setOrgPassword(event.target.value)
                setCreateBanner(null)
              }}
            />
          </Field>

          <label className="checkbox terms">
            <CheckControl />
            <span>
              J'accepte les{" "}
              <a href="#" onClick={stopNavigation}>
                conditions d'utilisation
              </a>{" "}
              et la{" "}
              <a href="#" onClick={stopNavigation}>
                politique de confidentialité
              </a>
              .
            </span>
          </label>

          <Button type="submit" block>
            <span className="btn-label">Créer l'organisation</span>
          </Button>
        </form>

        <p className="form-foot">
          Vous avez déjà une organisation ?{" "}
          <a href="#" onClick={(event) => { stopNavigation(event); setView("login") }}>
            Se connecter
          </a>
        </p>
      </section>

      {/* ── Demander un accès ─────────────────────────────────── */}
      <section className="form-card" hidden={view !== "request"}>
        <div className="mobile-brand">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <header className="form-head">
          <a href="#" className="back-link" onClick={(event) => { stopNavigation(event); setView("login") }}>
            <ChevronLeftIcon aria-hidden="true" />
            <span>Retour à la connexion</span>
          </a>
          <h2>Demander un accès</h2>
          <p>Sollicitez une invitation auprès de l'administrateur d'une organisation existante.</p>
        </header>

        {requestBanner !== null ? <ErrorBanner>{requestBanner}</ErrorBanner> : null}

        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            setRequestBanner(COMING_SOON_TEXT)
          }}
        >
          <Field label="Nom complet" htmlFor="reqName">
            <Input
              icon={<UsersIcon aria-hidden="true" />}
              id="reqName"
              type="text"
              placeholder="Camille Laurent"
              autoComplete="name"
              value={reqName}
              onChange={(event) => {
                setReqName(event.target.value)
                setRequestBanner(null)
              }}
            />
          </Field>

          <Field label="Adresse e-mail professionnelle" htmlFor="reqEmail">
            <Input
              icon={<MailIcon aria-hidden="true" />}
              id="reqEmail"
              type="email"
              placeholder="vous@entreprise.com"
              autoComplete="email"
              value={reqEmail}
              onChange={(event) => {
                setReqEmail(event.target.value)
                setRequestBanner(null)
              }}
            />
          </Field>

          <Field
            label="Nom de l'organisation"
            htmlFor="reqOrg"
            help="Indiquez l'organisation à laquelle vous souhaitez joindre. L'administrateur recevra votre demande."
          >
            <Input
              icon={<BuildingIcon aria-hidden="true" />}
              id="reqOrg"
              type="text"
              placeholder="Studio Nord"
              autoComplete="organization"
              value={reqOrg}
              onChange={(event) => {
                setReqOrg(event.target.value)
                setRequestBanner(null)
              }}
            />
          </Field>

          <Field label="Message à l'administrateur (facultatif)" htmlFor="reqMessage">
            <TextArea
              id="reqMessage"
              rows={3}
              placeholder="Précisez votre rôle et le motif de votre demande…"
              value={reqMessage}
              onChange={(event) => {
                setReqMessage(event.target.value)
              }}
            />
          </Field>

          <Button type="submit" block>
            <span className="btn-label">Envoyer la demande</span>
          </Button>
        </form>

        <p className="form-foot">
          Vous administrez une organisation ?{" "}
          <a href="#" onClick={(event) => { stopNavigation(event); setView("login") }}>
            Se connecter
          </a>
          <br />
          <a href="#" onClick={(event) => { stopNavigation(event); setView("create") }}>
            Créer une organisation
          </a>
        </p>
      </section>

      {/* ── Mot de passe oublié ───────────────────────────────── */}
      <section className="form-card" hidden={view !== "forgot"}>
        <div className="mobile-brand">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <header className="form-head">
          <a href="#" className="back-link" onClick={(event) => { stopNavigation(event); setView("login") }}>
            <ChevronLeftIcon aria-hidden="true" />
            <span>Retour à la connexion</span>
          </a>
          <h2>Mot de passe oublié</h2>
          <p>Saisissez votre adresse e-mail — nous vous enverrons un lien de réinitialisation.</p>
        </header>

        {forgotBanner !== null ? <ErrorBanner>{forgotBanner}</ErrorBanner> : null}

        <form
          noValidate
          onSubmit={(event) => {
            event.preventDefault()
            setForgotBanner(COMING_SOON_TEXT)
          }}
        >
          <Field label="Adresse e-mail" htmlFor="forgotEmail">
            <Input
              icon={<MailIcon aria-hidden="true" />}
              id="forgotEmail"
              type="email"
              placeholder="vous@exemple.com"
              autoComplete="email"
              value={forgotEmail}
              onChange={(event) => {
                setForgotEmail(event.target.value)
                setForgotBanner(null)
              }}
            />
          </Field>

          <Button type="submit" block>
            <span className="btn-label">Envoyer le lien de réinitialisation</span>
          </Button>
        </form>

        <p className="form-foot">
          Vous vous souvenez de votre mot de passe ?{" "}
          <a href="#" onClick={(event) => { stopNavigation(event); setView("login") }}>
            Se connecter
          </a>
        </p>
      </section>
    </AuthLayout>
  )
}

// The unified endpoint answers with a discriminated payload; the SSO block of
// the maquette (« ou continuer avec ») is deliberately absent.
