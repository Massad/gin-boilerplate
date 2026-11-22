import type { Meta, StoryObj } from "@storybook/react-vite"
import { Badge } from "./components/Badge"

/**
 * Portage des spécimens badges du design-system Atrium (Open Design) :
 * les pastilles qualifient un état, jamais une action — trois statuts
 * métier, une variante neutre et une variante pleine.
 */
const meta = {
  title: "Design System/Badges et statuts",
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
        <span className="spec-code">.badge[data-status] · .badge[data-variant]</span>
      </div>
      <div className="spec-body">
        <Badge dot>Neutre</Badge>
        <Badge variant="solid">Prioritaire</Badge>
        <Badge status="active">Actif</Badge>
        <Badge status="pending">En attente</Badge>
        <Badge status="suspended">Suspendu</Badge>
      </div>
      <div className="spec-note">
        Texte et bordure sont dérivés du token sémantique par <b>color-mix()</b>, le fond reste
        à 95 % de blanc. Le point de couleur est le seul signal chromatique du statut.
      </div>
    </section>
  ),
}

export const Usage: Story = {
  name: "Usage",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Usage</b>
        <span className="spec-code">règle</span>
      </div>
      <div className="spec-body stack">
        <p className="max-w-prose text-sm leading-relaxed text-muted-foreground">
          Un badge ne porte jamais plus de trois mots et n'est pas cliquable : c'est la ligne
          entière qui reçoit l'action. Les statuts se lisent dans un tableau, jamais dans un
          titre.
        </p>
      </div>
    </section>
  ),
}
