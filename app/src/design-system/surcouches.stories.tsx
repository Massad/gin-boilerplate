import { useState } from "react"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { Button } from "./components/Button"
import { CheckControl, Field, Input } from "./components/Field"
import { Drawer, Modal } from "./components/Overlays"
import { Toasts, useToasts } from "./components/Overlays"

/**
 * Portage des spécimens surcouches du design-system Atrium (Open Design) :
 * tiroir latéral pour l'édition, modale pour la confirmation courte,
 * notification pour le retour d'action — chacune se ferme au clavier.
 */
const meta = {
  title: "Design System/Surcouches",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

function OuvrirLesCouchesDemo() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const toasts = useToasts()

    return (
      <>
        <section className="spec">
          <div className="spec-head">
            <b>Ouvrir</b>
            <span className="spec-code">.drawer · .modal · .toast</span>
          </div>
          <div className="spec-body">
            <Button variant="secondary" onClick={() => setDrawerOpen(true)}>
              Ouvrir le tiroir
            </Button>
            <Button variant="secondary" onClick={() => setModalOpen(true)}>
              Ouvrir la modale
            </Button>
            <Button variant="secondary" onClick={() => toasts.push("Rôle mis à jour", "Le rôle « Administrateur » a été appliqué à Amelie Ndiaye.")}>
              Déclencher une notification
            </Button>
          </div>
          <div className="spec-note">
            Le voile est commun aux trois couches ; il se ferme au clic et à la touche{" "}
            <b>Échap</b>. La notification s'efface seule après 4,2 s.
          </div>
        </section>
        <Drawer
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          title="Amelie Ndiaye"
          subtitle="amelie@studio-nord.fr"
          initials="AN"
          footer={
            <>
              <Button variant="secondary" onClick={() => setDrawerOpen(false)}>
                Annuler
              </Button>
              <Button
                onClick={() => {
                  setDrawerOpen(false)
                  toasts.push("Droits enregistrés", "Les permissions sont appliquées.")
                }}
              >
                Enregistrer
              </Button>
            </>
          }
        >
          <div className="perm-group">
            <h3>Rôle</h3>
            <div className="perm-row">
              <label htmlFor="perm-admin">Administrateur</label>
              <CheckControl id="perm-admin" defaultChecked aria-label="Administrateur" />
            </div>
            <div className="perm-row">
              <label htmlFor="perm-membre">Membre</label>
              <CheckControl id="perm-membre" aria-label="Membre" />
            </div>
            <div className="perm-row">
              <label htmlFor="perm-lecteur">Lecteur</label>
              <CheckControl id="perm-lecteur" aria-label="Lecteur" />
            </div>
          </div>
          <div className="perm-group">
            <h3>Portée</h3>
            <div className="perm-row">
              <label htmlFor="perm-membres">Gérer les membres</label>
              <CheckControl id="perm-membres" defaultChecked aria-label="Gérer les membres" />
            </div>
            <div className="perm-row">
              <label htmlFor="perm-facturation">Voir la facturation</label>
              <CheckControl id="perm-facturation" aria-label="Voir la facturation" />
            </div>
          </div>
        </Drawer>
        <Modal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          title="Inviter un membre"
          description="Un e-mail d'activation partira vers l'adresse saisie."
          labelledBy="modal-title"
          footer={
            <>
              <Button variant="secondary" onClick={() => setModalOpen(false)}>
                Annuler
              </Button>
              <Button
                onClick={() => {
                  setModalOpen(false)
                  toasts.push("Invitation envoyée", "Un e-mail d'activation a été transmis.")
                }}
              >
                Envoyer l'invitation
              </Button>
            </>
          }
        >
          <Field label="Adresse e-mail" htmlFor="modal-email">
            <Input id="modal-email" type="email" placeholder="prenom.nom@studio-nord.fr" />
          </Field>
        </Modal>
        <Toasts items={toasts.items} />
      </>
  )
}

export const OuvrirLesCouches: Story = {
  name: "Ouvrir les couches",
  parameters: { layout: "padded" },
  render: () => <OuvrirLesCouchesDemo />,
}

export const ModeleNotification: Story = {
  name: "Modèle de notification",
  parameters: { layout: "padded" },
  render: () => (
    <section className="spec">
      <div className="spec-head">
        <b>Notification</b>
        <span className="spec-code">.toast</span>
      </div>
      <div className="spec-body stack">
        <div className="toast" style={{ position: "static", animation: "none" }}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10" />
            <path d="m9 12 2 2 4-4" />
          </svg>
          <div>
            <b>Rôle mis à jour</b>
            <span>Le rôle « Administrateur » a été appliqué à Amelie Ndiaye.</span>
          </div>
        </div>
      </div>
      <div className="spec-note">
        Le titre porte l'action, la ligne secondaire porte le détail. La pastille verte confirme
        sans animations superflues.
      </div>
    </section>
  ),
}
