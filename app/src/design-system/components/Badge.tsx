export interface BadgeProps {
  children: React.ReactNode
  status?: "active" | "pending" | "suspended"
  variant?: "solid"
  dot?: boolean
}

export function Badge({ children, status, variant, dot = false }: BadgeProps) {
  return (
    <span
      className="badge"
      data-status={status}
      data-variant={variant === "solid" ? "solid" : undefined}
    >
      {status || dot ? <span className="dot" aria-hidden="true" /> : null}
      {children}
    </span>
  )
}
