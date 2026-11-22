import { SPACE_STEPS } from "./tokens"

export function SpacingPalette() {
  return (
    <div className="flex flex-col gap-3">
      <p className="max-w-prose text-sm leading-relaxed text-muted-foreground">
        Grille de 4 px. Le rythme vertical des sections va de 32 à 48 px ; les
        rayons montent de 6 à 12 px, jamais au-delà.
      </p>
      <div className="flex flex-col gap-2">
        {SPACE_STEPS.map((step) => (
          <div key={step.name} className="flex items-center gap-4">
            <span className="w-36 shrink-0 font-mono text-xs text-foreground">
              {step.name}
            </span>
            <span className="w-12 shrink-0 font-mono text-xs text-muted-foreground">
              {step.className}
            </span>
            <span
              className="h-3"
              style={{ width: `var(${step.name})`, backgroundColor: "var(--od-accent)" }}
              aria-hidden="true"
            />
          </div>
        ))}
      </div>
    </div>
  )
}
