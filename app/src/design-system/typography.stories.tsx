import type { Meta, StoryObj } from "@storybook/react-vite"
import { TypographyPalette } from "./TypographyPalette"

/**
 * Fontes déclarées dans `src/index.css` (`--font-sans`, `--font-heading`)
 * et échelle de corps de texte Tailwind par défaut.
 */
const meta = {
  title: "Design System/Typographie",
  tags: ["autodocs"],
} satisfies Meta

export default meta

type Story = StoryObj

export const FontesEtCorpsDeTexte: Story = {
  name: "Fontes et corps de texte",
  parameters: { layout: "padded" },
  render: () => <TypographyPalette />,
}
