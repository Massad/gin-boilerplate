import type { Meta, StoryObj } from "@storybook/react-vite"
import { MembersTable, Stat, Stats } from "./components/Data"
import type { Member } from "./components/Data"
import { Toasts, useToasts } from "./components/Overlays"

/**
 * Portage des spécimens données du design-system Atrium (Open Design) :
 * indicateurs de synthèse puis tableau des membres — recherche, filtre,
 * tri par attributs, sélection et barre d'actions groupées.
 */
const meta = {
  title: "Design System/Données et tableaux",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

const MEMBERS: Member[] = [
  {
    id: "ndiaye",
    name: "Amelie Ndiaye",
    email: "amelie@studio-nord.fr",
    role: "Administrateur",
    scopeTitle: "Toute l'organisation",
    scopeDetail: "Membres, projets, facturation",
    status: "active",
    activity: "Il y a 12 min",
    activityKey: "2026-11-21",
  },
  {
    id: "petit",
    name: "Jonas Petit",
    email: "jonas@studio-nord.fr",
    role: "Membre",
    scopeTitle: "Équipe Produit",
    scopeDetail: "Projets et documents",
    status: "active",
    activity: "Hier, 18:40",
    activityKey: "2026-11-20",
  },
  {
    id: "ribeiro",
    name: "Salomé Ribeiro",
    email: "salome@studio-nord.fr",
    role: "Lecteur",
    scopeTitle: "Projet Hélios",
    scopeDetail: "Lecture seule",
    status: "pending",
    activity: "Il y a 3 jours",
    activityKey: "2026-11-18",
  },
  {
    id: "keller",
    name: "Tom Keller",
    email: "tom@studio-nord.fr",
    role: "Membre",
    scopeTitle: "Équipe Design",
    scopeDetail: "Projets et documents",
    status: "suspended",
    activity: "Il y a 8 jours",
    activityKey: "2026-11-13",
  },
]

const ROLES = ["Tous les rôles", "Administrateur", "Membre", "Lecteur"]

export const Indicateurs: Story = {
  name: "Indicateurs",
  parameters: { layout: "padded" },
  render: () => (
    <Stats>
      <Stat label="Membres" value="128" note="+6 ce mois-ci" />
      <Stat label="Actifs" value="121" note="94 % de l'effectif" />
      <Stat label="Invitations" value="5" note={<>dont <b>2</b> expirées</>} />
      <Stat label="Sièges" value="140" note="91 % utilisés" seatsUsedPercent={91} />
    </Stats>
  ),
}

function TableauDesMembresDemo() {
  const toasts = useToasts()

  return (
    <>
      <MembersTable members={MEMBERS} roles={ROLES} totalCount={128} onAction={(note) => toasts.push("Action groupée", note)} />
      <Toasts items={toasts.items} />
    </>
  )
}

export const TableauDesMembres: Story = {
  name: "Tableau des membres",
  parameters: { layout: "padded" },
  render: () => <TableauDesMembresDemo />,
}
