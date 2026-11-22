import { COLOR_GROUPS } from "./tokens"
import { Swatch } from "./Swatch"

export function ColorPalette() {
  return (
    <div className="flex flex-col gap-8">
      {COLOR_GROUPS.map((group) => (
        <section key={group.title} aria-label={group.title} className="flex flex-col gap-3">
          <h3 className="text-sm font-semibold text-foreground">{group.title}</h3>
          <div className="flex flex-wrap gap-4">
            {group.tokens.map((item) => (
              <Swatch key={item.name} tokenItem={item} />
            ))}
          </div>
        </section>
      ))}
      <p className="max-w-prose text-sm leading-relaxed text-muted-foreground">
        Palette monochrome : un seul accent — le noir — réservé aux actions
        primaires. Les couleurs sémantiques n'apparaissent que sur les pastilles
        de statut et les messages d'erreur. Toute couleur dérivée passe par{" "}
        <code className="font-mono text-xs">color-mix()</code> à partir d'un
        jeton.
      </p>
    </div>
  )
}
