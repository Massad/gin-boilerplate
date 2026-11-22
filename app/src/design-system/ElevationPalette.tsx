import { CONTAINER_TOKENS, ELEVATION_STEPS, MOTION_TOKENS, useCssVariable } from "./tokens"
import type { ElevationStep } from "./tokens"

function ElevationCard({ step }: { step: ElevationStep }) {
  return (
    <figure className="flex w-44 flex-col gap-2">
      <div
        role="img"
        aria-label={step.name}
        title={step.name}
        className="flex h-20 items-center justify-center rounded-lg"
        style={{
          border: step.style === "flat" ? "1px solid var(--od-border-soft)" : undefined,
          boxShadow: `var(${step.name})`,
          backgroundColor: "var(--od-surface)",
        }}
      />
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs font-medium text-foreground">{step.name}</span>
        <span className="text-xs leading-snug text-muted-foreground">{step.role}</span>
      </figcaption>
    </figure>
  )
}

export function ElevationPalette() {
  const focusRing = useCssVariable("--od-focus-ring")

  const motionAndContainer = [...MOTION_TOKENS, ...CONTAINER_TOKENS]

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap gap-6">
        {ELEVATION_STEPS.map((step) => (
          <ElevationCard key={step.name} step={step} />
        ))}
        <figure className="flex w-44 flex-col gap-2">
          <div
            role="img"
            aria-label="--od-focus-ring"
            title="--od-focus-ring"
            className="flex h-20 items-center justify-center rounded-lg bg-background"
            style={{ boxShadow: focusRing }}
          />
          <figcaption className="flex flex-col gap-0.5">
            <span className="font-mono text-xs font-medium text-foreground">--od-focus-ring</span>
            <span className="text-xs leading-snug text-muted-foreground">
              Anneau de focus — 2 px canvas puis 2 px accent
            </span>
          </figcaption>
        </figure>
      </div>
      <p className="max-w-prose text-sm leading-relaxed text-muted-foreground">
        L'élévation se limite à deux couches — un filet plus une ombre diffuse.
        Focus visible sur chaque contrôle : anneau de 2 px canvas puis 2 px
        accent. Aucun effet de motion au-delà de 250 ms.
      </p>
      <dl className="flex flex-col gap-1.5">
        {motionAndContainer.map((item) => (
          <div key={item.name} className="flex items-baseline gap-4">
            <dt className="w-72 shrink-0 font-mono text-xs text-foreground">{item.name}</dt>
            <dd className="font-mono text-xs text-muted-foreground">{item.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}
