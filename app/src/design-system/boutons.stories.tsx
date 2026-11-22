import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button } from "./components/Button"
import { UsersIcon } from "./components/icons"

/**
 * Portage des spécimens boutons du design-system Atrium (Open Design) :
 * deux variantes seulement, l'action primaire noir étant réservée à
 * l'action décisive de l'écran.
 */
const meta = {
  title: "Design System/Boutons",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const Variantes: Story = {
  name: "Variantes",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Variantes</b>
        <span className="spec-code">.btn-primary · .btn-secondary</span>
      </div>
      <div className="spec-body">
        <Button>Action principale</Button>
        <Button variant="secondary">Action secondaire</Button>
        <Button>
          <UsersIcon aria-hidden="true" /> Inviter
        </Button>
      </div>
      <div className="spec-note">
        Le survol du primaire éclaircit le fond vers <b>--od-accent-hover</b> (mélange 10 % de
        blanc) — le texte blanc reste inchangé, le contraste ne baisse jamais. Ce catalogue
        présente la variante primaire plusieurs fois : un écran produit n'admet qu'un seul
        appel primaire.
      </div>
    </section>
  ),
}

export const Etats: Story = {
  name: "États",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>États</b>
        <span className="spec-code">hover · focus-visible · active · busy · disabled</span>
      </div>
      <div className="spec-body">
        <span>
          <span className="spec-label">Focus visible</span>
          <Button className="focus-demo">Focus visible</Button>
        </span>
        <span>
          <span className="spec-label">Chargement</span>
          <Button busy>Enregistrement…</Button>
        </span>
        <span>
          <span className="spec-label">Désactivé</span>
          <Button disabled>Indisponible</Button>
        </span>
        <span className="w-64">
          <span className="spec-label">Plein largeur</span>
          <Button block large>
            Se connecter
          </Button>
        </span>
      </div>
      <div className="spec-note">
        Anneau de focus : <b>0 0 0 2px var(--od-bg), 0 0 0 4px var(--od-accent)</b> — un halo
        de canvas puis l'anneau d'accent.
      </div>
    </section>
  ),
}
