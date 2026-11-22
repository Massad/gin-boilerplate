import type { Meta, StoryObj } from "@storybook/react-vite"
import { Checkbox, Divider, ErrorBanner, Field, Input, PasswordInput, Select, TextArea } from "./components/Field"
import { MailIcon } from "./components/icons"

/**
 * Portage des spécimens champs de formulaire du design-system Atrium
 * (Open Design) : étiquette au-dessus, aide en dessous, erreur sous le
 * champ invalide, bascule d'affichage du mot de passe.
 */
const meta = {
  title: "Design System/Champs de formulaire",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const ChampTexte: Story = {
  name: "Champ texte",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Champ texte</b>
        <span className="spec-code">.field · .input-wrap · .input-icon</span>
      </div>
      <div className="spec-body stack">
        <Field
          label="Adresse e-mail"
          htmlFor="champ-email"
          help="L'adresse sert à vous rediriger vers votre organisation après la première connexion."
        >
          <Input
            icon={<MailIcon aria-hidden="true" />}
            id="champ-email"
            type="email"
            placeholder="vous@exemple.com"
            autoComplete="email"
          />
        </Field>
      </div>
      <div className="spec-note">
        L'étiquette reste toujours visible, jamais remplacée par le placeholder. L'icône est
        décorative : le sens vient de l'étiquette.
      </div>
    </section>
  ),
}

export const MotDePasse: Story = {
  name: "Mot de passe",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Champ mot de passe</b>
        <span className="spec-code">.input-wrap · .pwd-toggle</span>
      </div>
      <div className="spec-body stack">
        <Field
          label="Mot de passe"
          htmlFor="champ-motdepasse"
          help="Le bouton d'affichage porte une étiquette accessible et conserve un focus visible."
        >
          <PasswordInput
            id="champ-motdepasse"
            placeholder="12 caractères minimum"
            autoComplete="new-password"
          />
        </Field>
      </div>
    </section>
  ),
}

export const ListesEtZoneDeTexte: Story = {
  name: "Liste déroulante et zone de texte",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Liste déroulante &amp; zone de texte</b>
        <span className="spec-code">.select-wrap · textarea</span>
      </div>
      <div className="spec-body stack">
        <Field label="Rôle par défaut" htmlFor="champ-role">
          <Select id="champ-role" fullWidth options={["Membre", "Administrateur", "Lecteur"]} defaultValue="Membre" />
        </Field>
        <Field label="Note de service" htmlFor="champ-note">
          <TextArea id="champ-note" placeholder="Visible par les membres de l'organisation" />
        </Field>
      </div>
    </section>
  ),
}

export const EtatInvalideEtCases: Story = {
  name: "État invalide et cases à cocher",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>État invalide &amp; cases à cocher</b>
        <span className="spec-code">.field[data-invalid] · .checkbox · .divider</span>
      </div>
      <div className="spec-body stack">
        <Field label="Domaine" htmlFor="champ-domaine" error="Le domaine doit se terminer par .atrium.app">
          <Input id="champ-domaine" type="text" defaultValue="studio-nord" />
        </Field>
        <Checkbox label="Se souvenir de cet appareil pendant 30 jours" defaultChecked />
        <Divider>ou</Divider>
        <Checkbox label="Utiliser le SSO de l'organisation" />
      </div>
    </section>
  ),
}

export const BanniereErreur: Story = {
  name: "Bannière d'erreur",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Bannière d'erreur</b>
        <span className="spec-code">.error-banner</span>
      </div>
      <div className="spec-body stack">
        <ErrorBanner>
          Adresse e-mail ou mot de passe incorrect. Vérifiez vos identifiants et réessayez.
        </ErrorBanner>
      </div>
      <div className="spec-note">
        Fond et bordure dérivés de <b>--od-danger</b> par <b>color-mix()</b> ; le texte reste au
        contraste AA sur fond clair.
      </div>
    </section>
  ),
}
