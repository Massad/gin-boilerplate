import type { Meta, StoryObj } from "@storybook/react-vite"
import { RadiusPalette } from "./RadiusPalette"

/**
 * Rayons du contrat Atrium (« Open Design »), lus depuis les variables
 * `--od-radius-*` de `src/index.css` : 6 / 8 / 12 px, jamais au-delà,
 * plus la variante pill pour les pastilles.
 */
const meta = {
  title: "Design System/Rayons",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const Rayons: Story = {
  name: "Rayons",
  parameters: { layout: "padded" },
  render: () => <RadiusPalette />,
}
