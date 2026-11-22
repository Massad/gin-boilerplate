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
              <Swatch key={item.name} token={item} />
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
