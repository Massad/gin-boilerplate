import { useEffect, useRef, useState } from "react"
import { cn } from "@/lib/utils"
import { RADIUS_STEPS, useCssVariable } from "./tokens"
import type { RadiusStep } from "./tokens"

function RadiusCard({ step, base }: { step: RadiusStep; base: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const [direct, setDirect] = useState("")

  useEffect(() => {
    const element = ref.current
    if (element) {
      setDirect(getComputedStyle(element).getPropertyValue(step.name).trim())
    }
  }, [step.name])

  const value = direct || (base ? `${base} × ${step.factor}` : "—")

  return (
    <figure className="flex flex-col gap-2">
      <div
        role="img"
        aria-label={step.name}
        title={step.name}
        className={cn("h-16 w-28 border border-border bg-accent", step.className)}
      />
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs text-foreground">{step.className}</span>
        <span className="font-mono text-xs text-muted-foreground">{value}</span>
      </figcaption>
    </figure>
  )
}

export function RadiusPalette() {
  const base = useCssVariable("--radius", "")

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-4">
        {RADIUS_STEPS.map((step) => (
          <RadiusCard key={step.name} step={step} base={base} />
        ))}
      </div>
      <p className="font-mono text-xs text-muted-foreground">--radius = {base || "—"}</p>
    </div>
  )
}
