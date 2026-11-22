import { useEffect, useState } from "react"

export type SwatchMode = "fill" | "text"

export interface TokenDefinition {
  name: string
  role: string
  mode: SwatchMode
}

export interface TokenGroup {
  title: string
  tokens: TokenDefinition[]
}

function token(name: string, role: string): TokenDefinition {
  return { name, role, mode: name.endsWith("-on") ? "text" : "fill" }
}

const surfaces: TokenDefinition[] = [
  token("--od-bg", "Canvas de page"),
  token("--od-surface", "Cartes, panneaux, modales"),
  token("--od-surface-warm", "Alias de --od-surface"),
  token("--od-fg", "Texte principal, titres"),
  token("--od-fg-2", "Alias de --od-fg"),
  token("--od-muted", "Légendes, métadonnées"),
  token("--od-meta", "Alias de --od-muted"),
  token("--od-border", "Filets, bordures de champs"),
  token("--od-border-soft", "Alias de --od-border"),
]

const accent: TokenDefinition[] = [
  token("--od-accent", "Action primaire, anneau de focus"),
  token("--od-accent-on", "Texte sur accent"),
  token("--od-accent-hover", "Survol de l'action primaire"),
  token("--od-accent-active", "Action primaire active"),
]

const states: TokenDefinition[] = [
  token("--od-success", "Pastille « actif », confirmation"),
  token("--od-warn", "Pastille « en attente »"),
  token("--od-danger", "Pastille « suspendu », erreur"),
]

export const COLOR_GROUPS: TokenGroup[] = [
  { title: "Surfaces et texte", tokens: surfaces },
  { title: "Accent — noir monochrome", tokens: accent },
  { title: "États sémantiques", tokens: states },
]

export interface FontToken {
  name: string
  label: string
  sample: string
}

export const FONT_TOKENS: FontToken[] = [
  {
    name: "--od-font-display",
    label: "Geist · titres et chiffres clés",
    sample: "Un accès, tous vos espaces",
  },
  {
    name: "--od-font-body",
    label: "Geist · texte courant et interface",
    sample: "Atrium regroupe vos organisations, équipes et projets.",
  },
  {
    name: "--od-font-mono",
    label: "Fira Code · données techniques : domaines, e-mails, URL",
    sample: "studio-nord.atrium.app",
  },
]

export interface TypeStep {
  name: string
  className: string
  usage: string
  sample: string
  display?: boolean
  /** Interlignage spécifique, sinon --od-leading-body ou --od-leading-tight. */
  leading?: string
}

export const TYPE_STEPS: TypeStep[] = [
  {
    name: "--od-text-4xl",
    className: "48px",
    usage: "Titre d'affichage",
    sample: "Un accès, tous vos espaces",
    display: true,
    leading: "1.1",
  },
  {
    name: "--od-text-3xl",
    className: "40px",
    usage: "Titre de page majeur",
    sample: "Écrans de la console",
    display: true,
  },
  {
    name: "--od-text-2xl",
    className: "32px",
    usage: "Titre de section",
    sample: "Utilisateurs & droits",
    display: true,
  },
  {
    name: "--od-text-xl",
    className: "24px",
    usage: "Titre de dialogue",
    sample: "Modifier les droits",
    display: true,
  },
  {
    name: "--od-text-lg",
    className: "20px",
    usage: "Sous-titre, phrase d'accroche",
    sample: "Atrium regroupe vos organisations",
  },
  {
    name: "--od-text-base",
    className: "16px",
    usage: "Texte mis en avant",
    sample: "Atrium regroupe vos organisations, équipes et projets.",
  },
  {
    name: "--od-text-sm",
    className: "14px",
    usage: "Corps de texte par défaut de l'interface",
    sample: "Le corps de texte par défaut de l'interface.",
  },
  {
    name: "--od-text-xs",
    className: "12px",
    usage: "Légendes, métadonnées, horodatages",
    sample: "Légendes, métadonnées, horodatages",
  },
]

export type SimpleToken = { name: string; value: string }

export const TYPE_RHYTHM: SimpleToken[] = [
  { name: "--od-leading-body", value: "1.5 — corps de texte" },
  { name: "--od-leading-tight", value: "1.2 — titres" },
  { name: "--od-tracking-display", value: "-0.02em — titres d'affichage" },
]

export interface SpaceStep {
  name: string
  className: string
}

export const SPACE_STEPS: SpaceStep[] = [
  { name: "--od-space-1", className: "4px" },
  { name: "--od-space-2", className: "8px" },
  { name: "--od-space-3", className: "12px" },
  { name: "--od-space-4", className: "16px" },
  { name: "--od-space-5", className: "20px" },
  { name: "--od-space-6", className: "24px" },
  { name: "--od-space-8", className: "32px" },
  { name: "--od-space-12", className: "48px" },
]

export interface RadiusStep {
  name: string
  role: string
}

export const RADIUS_STEPS: RadiusStep[] = [
  { name: "--od-radius-sm", role: "sm · champs et boutons" },
  { name: "--od-radius-md", role: "md · cartes, panneaux" },
  { name: "--od-radius-lg", role: "lg · conteneurs surélevés" },
  { name: "--od-radius-pill", role: "pill · pastilles" },
]

export interface ElevationStep {
  name: string
  role: string
  style: "flat" | "ring" | "raised"
}

export const ELEVATION_STEPS: ElevationStep[] = [
  { name: "--od-elev-flat", role: "Repère plat", style: "flat" },
  { name: "--od-elev-ring", role: "Filet 1 px", style: "ring" },
  { name: "--od-elev-raised", role: "Surface flottante", style: "raised" },
]

export const MOTION_TOKENS: SimpleToken[] = [
  { name: "--od-motion-fast", value: "150 ms" },
  { name: "--od-motion-base", value: "200 ms" },
  { name: "--od-ease-standard", value: "cubic-bezier(0.2, 0, 0, 1)" },
]

export const CONTAINER_TOKENS: SimpleToken[] = [
  { name: "--od-container-max", value: "1280 px" },
  { name: "--od-container-gutter-desktop", value: "24 px" },
  { name: "--od-container-gutter-tablet", value: "16 px" },
  { name: "--od-container-gutter-phone", value: "16 px" },
]

export function useCssVariable(name: string, fallback = "—"): string {
  const [value, setValue] = useState(fallback)

  useEffect(() => {
    const root = document.documentElement
    const read = () =>
      setValue(getComputedStyle(root).getPropertyValue(name).trim() || fallback)
    read()
    const observer = new MutationObserver(read)
    observer.observe(root, { attributes: true, attributeFilter: ["class"] })
    return () => observer.disconnect()
  }, [name, fallback])

  return value
}
