import type { Meta, StoryObj } from "@storybook/react-vite"
import { ColorPalette } from "./ColorPalette"

/**
 * Palette des jetons couleur, lue à l'exécution depuis les variables CSS de
 * `src/index.css` (`:root` + `.dark`). Le sélecteur de thème de la barre
 * d'outils bascule la classe `.dark` sur `<html>`, exactement comme le
 * mécanisme sombre de l'application (`@custom-variant dark (&:is(.dark *))`).
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
