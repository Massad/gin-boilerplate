import { RADIUS_STEPS, useCssVariable } from "./tokens"
import type { RadiusStep } from "./tokens"

function RadiusCard({ step }: { step: RadiusStep }) {
  const value = useCssVariable(step.name)
  const isPill = step.name === "--od-radius-pill"

  return (
    <figure className="flex min-w-44 flex-col gap-2">
      <div
        role="img"
        aria-label={step.name}
        title={step.name}
        className={isPill ? "h-8 w-full" : "h-16 w-28"}
        style={{
          border: "1px solid var(--od-border)",
          backgroundColor: "var(--od-surface)",
          borderRadius: `var(${step.name})`,
        }}
      />
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs font-medium text-foreground">{step.name}</span>
        <span className="font-mono text-xs text-muted-foreground">{value}</span>
        <span className="text-xs leading-snug text-muted-foreground">{step.role}</span>
      </figcaption>
    </figure>
  )
}

export function RadiusPalette() {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap gap-4">
        {RADIUS_STEPS.map((step) => (
          <RadiusCard key={step.name} step={step} />
        ))}
      </div>
      <p className="max-w-prose text-sm leading-relaxed text-muted-foreground">
        <strong className="font-semibold">sm</strong> pour les champs et boutons,{" "}
        <strong className="font-semibold">md</strong> pour les cartes,{" "}
        <strong className="font-semibold">lg</strong> pour les conteneurs
        surélevés, <strong className="font-semibold">pill</strong> pour les
        pastilles.
      </p>
    </div>
  )
}
