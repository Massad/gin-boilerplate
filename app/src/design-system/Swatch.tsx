import { cn } from "@/lib/utils"
import { useCssVariable } from "./tokens"
import type { TokenDefinition } from "./tokens"

export function Swatch({ tokenItem }: { tokenItem: TokenDefinition }) {
  const value = useCssVariable(tokenItem.name)
  const isText = tokenItem.mode === "text"

  return (
    <figure className="flex min-w-44 flex-1 flex-col gap-2">
      <div
        role="img"
        aria-label={tokenItem.name}
        title={tokenItem.name}
        className={cn(
          "flex h-14 w-full items-center justify-center rounded-md border border-border",
          isText && "bg-background text-lg font-semibold",
        )}
        style={
          isText
            ? { color: `var(${tokenItem.name})` }
            : { backgroundColor: `var(${tokenItem.name})` }
        }
      >
        {isText ? "Aa" : null}
      </div>
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs font-medium text-foreground">
          {tokenItem.name}
        </span>
        <span className="font-mono text-xs text-muted-foreground">{value}</span>
        <span className="text-xs leading-snug text-muted-foreground">
          {tokenItem.role}
        </span>
      </figcaption>
    </figure>
  )
}
