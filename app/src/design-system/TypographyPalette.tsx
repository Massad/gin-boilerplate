import { cn } from "@/lib/utils"
import { FONT_TOKENS, TYPE_RHYTHM, TYPE_STEPS } from "./tokens"
import type { FontToken, TypeStep } from "./tokens"

function FontCard({ item }: { item: FontToken }) {
  const isDisplay = item.name === "--od-font-display"

  return (
    <figure className="flex min-w-72 flex-1 flex-col gap-2">
      <div
        className={cn(
          "flex min-h-16 items-center rounded-md border border-border bg-card px-4",
          isDisplay ? "text-xl" : "text-base",
        )}
        style={{ fontFamily: `var(${item.name})`, ...(isDisplay && { fontWeight: 600 }) }}
      >
        {item.sample}
      </div>
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs font-medium text-foreground">{item.name}</span>
        <span className="text-xs leading-snug text-muted-foreground">{item.label}</span>
      </figcaption>
    </figure>
  )
}

function ScaleRow({ step }: { step: TypeStep }) {
  return (
    <div className="grid grid-cols-[13rem_1fr] items-baseline gap-4">
      <span className="font-mono text-xs text-muted-foreground">
        {step.name} · {step.className}
      </span>
      <span
        className="min-w-0 text-foreground"
        style={{
          fontFamily: step.display ? "var(--od-font-display)" : undefined,
          fontSize: `var(${step.name})`,
          fontWeight: step.display ? 600 : undefined,
          lineHeight: step.leading ?? (step.display ? "var(--od-leading-tight)" : "var(--od-leading-body)"),
          letterSpacing: step.display ? "var(--od-tracking-display)" : undefined,
        }}
      >
        {step.sample}
      </span>
    </div>
  )
}

export function TypographyPalette() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap gap-4">
        {FONT_TOKENS.map((item) => (
          <FontCard key={item.name} item={item} />
        ))}
      </div>
      <section aria-label="Échelle typographique" className="flex flex-col gap-3">
        <h3 className="text-sm font-semibold text-foreground">Échelle</h3>
        <div className="flex flex-col gap-3">
          {TYPE_STEPS.map((step) => (
            <ScaleRow key={step.name} step={step} />
          ))}
        </div>
      </section>
      <section aria-label="Interlignage et chasse" className="flex flex-col gap-3">
        <h3 className="text-sm font-semibold text-foreground">Interlignage et chasse</h3>
        <dl className="flex flex-col gap-1.5">
          {TYPE_RHYTHM.map((item) => (
            <div key={item.name} className="flex items-baseline gap-4">
              <dt className="font-mono text-xs text-foreground">{item.name}</dt>
              <dd className="text-xs text-muted-foreground">{item.value}</dd>
            </div>
          ))}
        </dl>
      </section>
    </div>
  )
}
