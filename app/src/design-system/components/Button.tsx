import type { ButtonHTMLAttributes, ReactNode } from "react"
import { cn } from "@/lib/utils"

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary"
  block?: boolean
  large?: boolean
  busy?: boolean
  children?: ReactNode
}

export function Button({
  variant = "primary",
  block = false,
  large = false,
  busy = false,
  className,
  children,
  type = "button",
  ...rest
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        variant === "primary" ? "btn-primary" : "btn-secondary",
        block && "btn-block",
        large && "btn-lg",
        className,
      )}
      aria-busy={busy ? "true" : undefined}
      {...rest}
    >
      {busy ? <span className="spinner" aria-hidden="true" /> : null}
      {children}
    </button>
  )
}
