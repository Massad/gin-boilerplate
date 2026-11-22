import type { Meta, StoryObj } from "@storybook/react-vite"
import { RadiusPalette } from "./RadiusPalette"

/**
 * Échelle des rayons dérivée du jeton `--radius` de `src/index.css`
 * (facteurs 0.6 à 2.6, utilitaires Tailwind `rounded-*`).
 */
const meta = {
  title: "Design System/Rayons",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const EchelleDesRayons: Story = {
  name: "Échelle des rayons",
  parameters: { layout: "padded" },
  render: () => <RadiusPalette />,
}
