import type { Meta, StoryObj } from "@storybook/react-vite"
import { ColorPalette } from "./ColorPalette"

/**
 * Palette des jetons couleur du contrat Atrium (« Open Design »), lue à
 * l'exécution depuis les variables `--od-*` de `src/index.css`. Accent unique
 * noir, couleurs sémantiques réservées aux pastilles et erreurs.
 */
const meta = {
  title: "Design System/Couleurs",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const Palette: Story = {
  name: "Palette",
  parameters: { layout: "padded" },
  render: () => <ColorPalette />,
}
