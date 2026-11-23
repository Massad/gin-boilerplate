import type { ReactNode } from "react"

export interface AuthLayoutProps {
  children: ReactNode
}

/**
 * Two-pane authentication layout ported from the login maquette: a dark brand
 * pane (desktop only) and the form pane hosting the swapped view cards. The
 * mobile brand bar lives inside each form card, like the source page.
 */
export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="shell">
      <aside className="brand-pane">
        <div className="brand-mark">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <div className="brand-hero">
          <h1>Un accès, tous vos espaces de travail.</h1>
          <p>Atrium regroupe vos organisations, équipes et projets dans une plateforme unique et sécurisée.</p>
        </div>

        <div className="brand-quote">
          <div className="avatar" aria-hidden="true">
            LM
          </div>
          <div>
            <blockquote>
              «&nbsp;Nous avons centralisé six tenants en une seule interface. L'onboarding des nouvelles équipes a
              chuté de trois jours à vingt minutes.&nbsp;»
            </blockquote>
            <cite>Laure Moreau — Directrice des opérations, Studio Nord</cite>
          </div>
        </div>
      </aside>

      <main className="form-pane">{children}</main>
    </div>
  )
}
