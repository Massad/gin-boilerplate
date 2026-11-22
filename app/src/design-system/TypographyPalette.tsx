import { useEffect, useRef, useState } from "react"
import { cn } from "@/lib/utils"
import { FONT_STEPS, FONT_TOKENS } from "./tokens"
import type { FontToken } from "./tokens"

function FontCard({ item }: { item: FontToken }) {
  const ref = useRef<HTMLDivElement>(null)
  const [family, setFamily] = useState("")

  useEffect(() => {
    const element = ref.current
    if (element) {
      setFamily(getComputedStyle(element).fontFamily)
    }
  }, [])

  return (
    <figure className="flex min-w-72 flex-1 flex-col gap-2">
      <div
        ref={ref}
        className={cn(
          "flex h-16 items-center rounded-md border border-border bg-card px-4 text-xl text-foreground",
          item.className,
        )}
      >
        Ateliers & saveurs — Aa Bb 0123
      </div>
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs text-foreground">
          {item.name} · {item.label}
        </span>
        <span className="font-mono text-xs text-muted-foreground">{family || "—"}</span>
      </figcaption>
    </figure>
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
      <section aria-label="Échelle typographique" className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-foreground">
          Échelle Tailwind par défaut (non personnalisée)
        </h3>
        <div className="flex flex-col gap-2">
          {FONT_STEPS.map((step) => (
            <div key={step} className="flex items-baseline gap-4">
              <span className="w-24 shrink-0 font-mono text-xs text-muted-foreground">
                {step}
              </span>
              <span className={cn("truncate text-foreground", step)}>
                Ateliers & saveurs — Aa Bb 0123
              </span>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
