import { useEffect, useState } from "react"

export type SwatchMode = "fill" | "text"

export interface TokenDefinition {
  name: string
  mode: SwatchMode
}

export interface TokenGroup {
  title: string
  tokens: TokenDefinition[]
}

function token(name: string): TokenDefinition {
  return { name, mode: name.endsWith("-foreground") ? "text" : "fill" }
}

const surfaces = [
  "--background",
  "--foreground",
  "--card",
  "--card-foreground",
  "--popover",
  "--popover-foreground",
]

const states = [
  "--primary",
  "--primary-foreground",
  "--secondary",
  "--secondary-foreground",
  "--muted",
  "--muted-foreground",
  "--accent",
  "--accent-foreground",
  "--destructive",
]

const strokes = ["--border", "--input", "--ring"]

const charts = ["--chart-1", "--chart-2", "--chart-3", "--chart-4", "--chart-5"]

const sidebar = [
  "--sidebar",
  "--sidebar-foreground",
  "--sidebar-primary",
  "--sidebar-primary-foreground",
  "--sidebar-accent",
  "--sidebar-accent-foreground",
  "--sidebar-border",
  "--sidebar-ring",
]

export const COLOR_GROUPS: TokenGroup[] = [
  { title: "Socle", tokens: surfaces.map(token) },
  { title: "Actions et états", tokens: states.map(token) },
  { title: "Bordures et saisie", tokens: strokes.map(token) },
  { title: "Graphiques", tokens: charts.map(token) },
  { title: "Barre latérale", tokens: sidebar.map(token) },
]

export interface RadiusStep {
  name: string
  className: string
  factor: string
}

export const RADIUS_STEPS: RadiusStep[] = [
  { name: "--radius-sm", className: "rounded-sm", factor: "0.6" },
  { name: "--radius-md", className: "rounded-md", factor: "0.8" },
  { name: "--radius-lg", className: "rounded-lg", factor: "1" },
  { name: "--radius-xl", className: "rounded-xl", factor: "1.4" },
  { name: "--radius-2xl", className: "rounded-2xl", factor: "1.8" },
  { name: "--radius-3xl", className: "rounded-3xl", factor: "2.2" },
  { name: "--radius-4xl", className: "rounded-4xl", factor: "2.6" },
]

export interface FontToken {
  name: string
  className: string
  label: string
}

export const FONT_TOKENS: FontToken[] = [
  { name: "--font-sans", className: "font-sans", label: "texte courant" },
  { name: "--font-heading", className: "font-heading", label: "titres, hérite de --font-sans" },
]

export const FONT_STEPS = [
  "text-xs",
  "text-sm",
  "text-base",
  "text-lg",
  "text-xl",
  "text-2xl",
] as const

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
