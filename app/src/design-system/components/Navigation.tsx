import type { ReactNode } from "react"

export interface NavItemProps {
  icon: ReactNode
  label: string
  count?: string | number
  current?: boolean
  href?: string
}

export function NavItem({ icon, label, count, current = false, href = "#" }: NavItemProps) {
  return (
    <a href={href} className="nav-item" aria-current={current ? "page" : undefined}>
      {icon}
      {label}
      {count !== undefined ? <span className="count">{count}</span> : null}
    </a>
  )
}

export interface CrumbsProps {
  orgGlyph: string
  orgName: string
  current: string
  ariaLabel?: string
}

export function Crumbs({ orgGlyph, orgName, current, ariaLabel = "Fil d'Ariane" }: CrumbsProps) {
  return (
    <nav className="crumbs" aria-label={ariaLabel}>
      <span className="org-chip">
        <span className="glyph">{orgGlyph}</span>
        {orgName}
      </span>
      <span className="sep">/</span>
      <span className="current">{current}</span>
    </nav>
  )
}

export interface OrgSwitchProps {
  glyph: string
  name: string
  domain: string
  onClick?: () => void
}

export function OrgSwitch({ glyph, name, domain, onClick }: OrgSwitchProps) {
  return (
    <button type="button" className="org-switch" onClick={onClick}>
      <span className="org-glyph">{glyph}</span>
      <span className="org-info">
        <span className="org-name">{name}</span>
        <span className="org-domain">{domain}</span>
      </span>
      <ChevronDownGlyph />
    </button>
  )
}

function ChevronDownGlyph() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="6 9 12 15 18 9" />
    </svg>
  )
}

export interface TabItem {
  id: string
  label: string
  count?: string | number
}

export interface TabsProps {
  ariaLabel: string
  items: TabItem[]
  value: string
  onChange: (id: string) => void
  panels: Record<string, ReactNode>
}

export function Tabs({ ariaLabel, items, value, onChange, panels }: TabsProps) {
  const active = items.find((item) => item.id === value)

  return (
    <>
      <div className="tabs" role="tablist" aria-label={ariaLabel}>
        {items.map((item) => {
          const selected = item.id === value

          return (
            <button
              key={item.id}
              type="button"
              className="tab"
              role="tab"
              aria-controls={`panel-${item.id}`}
              aria-selected={selected ? "true" : "false"}
              tabIndex={selected ? 0 : -1}
              onClick={() => onChange(item.id)}
            >
              {item.label}
              {item.count !== undefined ? <span className="tab-count">{item.count}</span> : null}
            </button>
          )
        })}
      </div>
      <div role="tabpanel" id={`panel-${active?.id ?? value}`} aria-labelledby={`tab-${active?.id ?? value}`}>
        {panels[active?.id ?? value]}
      </div>
    </>
  )
}
