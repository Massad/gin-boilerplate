import { useState } from "react"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Crumbs, NavItem, OrgSwitch, Tabs } from "./components/Navigation"
import { BuildingIcon, CardIcon, UsersIcon } from "./components/icons"

/**
 * Portage des spécimens navigation du design-system Atrium (Open Design) :
 * rail latéral, fil d'Ariane, sélecteur d'organisation et onglets partagent
 * la même grisaille — le noir ne sert qu'à l'élément courant.
 */
const meta = {
  title: "Design System/Navigation",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const RailLateral: Story = {
  name: "Rail latéral",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Rail latéral</b>
        <span className="spec-code">.nav-item</span>
      </div>
      <div className="spec-body stack">
        <NavItem icon={<UsersIcon aria-hidden="true" />} label="Utilisateurs" count={128} current />
        <NavItem icon={<BuildingIcon aria-hidden="true" />} label="Organisations" />
        <NavItem icon={<CardIcon aria-hidden="true" />} label="Facturation" />
      </div>
      <div className="spec-note">
        Un seul élément porte <b>aria-current=&quot;page&quot;</b> ; les autres restent en gris et gagnent le
        noir au survol, sans changer de contraste.
      </div>
    </section>
  ),
}

export const FilArianeEtSelecteur: Story = {
  name: "Fil d'Ariane et sélecteur d'organisation",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Fil d'Ariane &amp; sélecteur d'organisation</b>
        <span className="spec-code">.crumbs · .org-switch</span>
      </div>
      <div className="spec-body stack">
        <Crumbs orgGlyph="SN" orgName="Studio Nord" current="Utilisateurs" />
        <OrgSwitch glyph="SN" name="Studio Nord" domain="studio-nord.atrium.app" />
      </div>
    </section>
  ),
}

const TAB_ITEMS = [
  { id: "all", label: "Tous", count: 128 },
  { id: "active", label: "Actifs", count: 121 },
  { id: "pending", label: "Invitations", count: 5 },
  { id: "suspended", label: "Suspendus", count: 2 },
]

const TAB_PANELS: Record<string, string> = {
  all: "Les 128 membres de l'organisation, tous rôles confondus.",
  active: "121 comptes actifs, dont 94 % se connectent chaque semaine.",
  pending: "5 invitations en attente d'acceptation, la plus ancienne date de 6 jours.",
  suspended: "2 comptes suspendus ; les données restent conservées 90 jours.",
}

function OngletsDemo() {
  const [value, setValue] = useState("all")

  return (
    <section className="spec">
      <div className="spec-head">
        <b>Onglets</b>
        <span className="spec-code">.tabs · .tab[aria-selected]</span>
      </div>
      <div className="spec-body stack" style={{ padding: 0 }}>
        <Tabs ariaLabel="Filtrer les membres" items={TAB_ITEMS} value={value} onChange={setValue} panels={TAB_PANELS} />
      </div>
      <div className="spec-note">
        Le composant est contrôlé : la valeur sélectionnée vient de l'état applicatif, comme
        l'onglet mémorisé de la page source.
      </div>
    </section>
  )
}

export const Onglets: Story = {
  name: "Onglets",
  parameters: { layout: "padded" },
  render: () => <OngletsDemo />,
}
