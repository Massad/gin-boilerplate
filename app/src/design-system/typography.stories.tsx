import type { Meta, StoryObj } from "@storybook/react-vite"
import { TypographyPalette } from "./TypographyPalette"

/**
 * Familles (Geist pour l'affichage et le texte, Fira Code pour les données
 * techniques) et échelle 12 → 48 px du contrat Atrium (« Open Design »),
 * lues depuis les variables `--od-font-*` et `--od-text-*` de `src/index.css`.
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
