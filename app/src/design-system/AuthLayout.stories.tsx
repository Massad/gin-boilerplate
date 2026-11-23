import type { Meta, StoryObj } from "@storybook/react-vite"
import { AuthLayout } from "./components/AuthLayout"
import { Button } from "./components/Button"
import { Field, Input, PasswordInput } from "./components/Field"
import { MailIcon } from "./components/icons"

/**
 * Portage du volet d'authentification de la maquette login : volet marque
 * sombre sur desktop + volet formulaire hébergeant une carte d'exemple.
 */
const meta = {
  title: "Design System/AuthLayout",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const VoletConnexion: Story = {
  name: "Volet de connexion",
  parameters: { layout: "fullscreen" },
  render: () => (
    <AuthLayout>
      <section className="form-card">
        <div className="mobile-brand">
          <span className="logo-dot" aria-hidden="true"></span>
          <span>Atrium</span>
        </div>

        <header className="form-head">
          <h2>Connexion</h2>
          <p>Signez avec vos identifiants — vous serez redirigé vers votre espace.</p>
        </header>

        <form noValidate onSubmit={(event) => event.preventDefault()}>
          <Field label="Adresse e-mail" htmlFor="story-auth-email">
            <Input icon={<MailIcon aria-hidden="true" />} id="story-auth-email" type="email" placeholder="vous@exemple.com" autoComplete="email" />
          </Field>
          <Field label="Mot de passe" htmlFor="story-auth-password">
            <PasswordInput id="story-auth-password" placeholder="Votre mot de passe" autoComplete="current-password" />
          </Field>
          <div className="row-between">
            <label className="checkbox">
              <input type="checkbox" />
              <span>Rester connecté</span>
            </label>
            <a href="#" onClick={(event) => event.preventDefault()}>
              Mot de passe oublié ?
            </a>
          </div>
          <Button type="submit" block>
            <span className="btn-label">Se connecter</span>
          </Button>
        </form>

        <p className="form-foot">
          Pas encore de compte ?{" "}
          <a href="#" onClick={(event) => event.preventDefault()}>
            Demander un accès
          </a>
          <br />
          <a href="#" onClick={(event) => event.preventDefault()}>
            Créer une organisation
          </a>
        </p>
      </section>
    </AuthLayout>
  ),
}
