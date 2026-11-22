import { cn } from "@/lib/utils"
import { useCssVariable } from "./tokens"
import type { TokenDefinition } from "./tokens"

export function Swatch({ token }: { token: TokenDefinition }) {
  const value = useCssVariable(token.name)
  const isText = token.mode === "text"

  return (
    <figure className="flex min-w-40 flex-1 flex-col gap-2">
      <div
        role="img"
        aria-label={token.name}
        title={token.name}
        className={cn(
          "flex h-14 w-full items-center justify-center rounded-md border border-border",
          isText && "bg-background text-lg font-semibold",
        )}
        style={
          isText
            ? { color: `var(${token.name})` }
            : { backgroundColor: `var(${token.name})` }
        }
      >
        {isText ? "Aa" : null}
      </div>
      <figcaption className="flex flex-col gap-0.5">
        <span className="font-mono text-xs text-foreground">{token.name}</span>
        <span className="font-mono text-xs text-muted-foreground">{value}</span>
      </figcaption>
    </figure>
  )
}
