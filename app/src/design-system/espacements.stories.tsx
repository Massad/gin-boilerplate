import type { Meta, StoryObj } from "@storybook/react-vite"
import { SpacingPalette } from "./SpacingPalette"

/**
 * Échelle d'espacement du contrat Atrium (« Open Design »), lue depuis les
 * variables `--od-space-*` de `src/index.css` : grille de 4 px, 8 paliers
 * de 4 à 48 px.
 */
const meta = {
  title: "Design System/Espacements",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const EchelleDesEspacements: Story = {
  name: "Échelle des espacements",
  parameters: { layout: "padded" },
  render: () => <SpacingPalette />,
}
