import type { Meta, StoryObj } from "@storybook/react-vite"
import { ElevationPalette } from "./ElevationPalette"

/**
 * Élévation, anneau de focus, motion et conteneur du contrat Atrium
 * (« Open Design »), lus depuis les variables `--od-elev-*`, `--od-focus-*`,
 * `--od-motion-*` et `--od-container-*` de `src/index.css`.
 */
const meta = {
  title: "Design System/Élévation et mouvement",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const ElevationsEtMouvement: Story = {
  name: "Élévations et mouvement",
  parameters: { layout: "padded" },
  render: () => <ElevationPalette />,
}
