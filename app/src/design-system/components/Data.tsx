import { useMemo, useState } from "react"
import type { InputHTMLAttributes, ReactNode } from "react"
import { Badge } from "./Badge"
import { CheckControl, Select } from "./Field"
import { ArrowUpIcon, ChevronLeftIcon, ChevronRightIcon, SearchIcon } from "./icons"

export interface StatProps {
  label: string
  value: string
  note?: ReactNode
  seatsUsedPercent?: number
}

export function Stat({ label, value, note, seatsUsedPercent }: StatProps) {
  return (
    <div className="stat">
      <span className="stat-label">{label}</span>
      <span className="stat-value">{value}</span>
      {note ? <span className="stat-note">{note}</span> : null}
      {seatsUsedPercent !== undefined ? (
        <div className="seat-bar">
          <span style={{ width: `${seatsUsedPercent}%` }} />
        </div>
      ) : null}
    </div>
  )
}

export function Stats({ children }: { children: ReactNode }) {
  return <div className="stats">{children}</div>
}

export interface SearchFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
}

export function SearchField({ label, ...rest }: SearchFieldProps) {
  return (
    <div className="search-field">
      <SearchIcon aria-hidden="true" />
      <input type="search" aria-label={label} {...rest} />
    </div>
  )
}

export interface TableToolbarProps {
  children: ReactNode
  count: string
}

export function TableToolbar({ children, count }: TableToolbarProps) {
  return (
    <div className="toolbar">
      {children}
      <span className="table-count">{count}</span>
    </div>
  )
}

export interface PagerProps {
  pages: number[]
  current: number
  onPageChange?: (page: number) => void
}

export function Pager({ pages, current, onPageChange }: PagerProps) {
  return (
    <div className="pager">
      <button
        type="button"
        className="page-btn"
        aria-label="Page précédente"
        disabled={current === 1}
        onClick={() => onPageChange?.(current - 1)}
      >
        <ChevronLeftIcon aria-hidden="true" />
      </button>
      {pages.map((page) => (
        <button
          key={page}
          type="button"
          className="page-btn"
          aria-current={page === current ? "true" : undefined}
          onClick={() => onPageChange?.(page)}
        >
          {page}
        </button>
      ))}
      <button
        type="button"
        className="page-btn"
        aria-label="Page suivante"
        disabled={current === pages.length}
        onClick={() => onPageChange?.(current + 1)}
      >
        <ChevronRightIcon aria-hidden="true" />
      </button>
    </div>
  )
}

export interface Member {
  id: string
  name: string
  email: string
  role: string
  scopeTitle: string
  scopeDetail: string
  status?: "active" | "pending" | "suspended"
  activity: string
  activityKey: string
}

export interface MembersTableProps {
  members: Member[]
  roles: string[]
  totalCount?: number
  onAction?: (note: string) => void
}

type SortableKey = "name" | "role" | "activity"

function initials(name: string): string {
  return name
    .split(" ")
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2)
    .join("")
    .toUpperCase()
}

export function MembersTable({ members, roles, totalCount = 128, onAction }: MembersTableProps) {
  const [query, setQuery] = useState("")
  const [role, setRole] = useState(roles[0] ?? "")
  const [sortKey, setSortKey] = useState<SortableKey>("name")
  const [sortDir, setSortDir] = useState<"ascending" | "descending">("ascending")
  const [checked, setChecked] = useState<string[]>([])

  const visible = useMemo(() => {
    const needle = query.trim().toLowerCase()
    const filtered = members.filter((member) => {
      const matchesQuery =
        needle.length === 0 ||
        `${member.name} ${member.email} ${member.role}`.toLowerCase().includes(needle)
      const matchesRole = role === roles[0] || member.role === role
      return matchesQuery && matchesRole
    })
    const sorted = [...filtered].sort((a, b) => {
      const av = (sortKey === "activity" ? a.activityKey : sortKey === "role" ? a.role : a.name).toLowerCase()
      const bv = (sortKey === "activity" ? b.activityKey : sortKey === "role" ? b.role : b.name).toLowerCase()
      return av === bv ? 0 : av < bv ? -1 : 1
    })
    return sortDir === "descending" ? [...sorted].reverse() : sorted
  }, [members, query, role, roles, sortKey, sortDir])

  const allCheckedVisible = visible.length > 0 && visible.every((member) => checked.includes(member.id))

  const toggleAll = () => {
    if (allCheckedVisible) {
      setChecked((current) => current.filter((id) => !visible.some((member) => member.id === id)))
    } else {
      setChecked((current) => [...new Set([...current, ...visible.map((member) => member.id)])])
    }
  }

  const toggleMember = (id: string) => {
    setChecked((current) => (current.includes(id) ? current.filter((item) => item !== id) : [...current, id]))
  }

  const clearSelection = () => setChecked([])

  const applySort = (key: SortableKey) => {
    if (key === sortKey) {
      setSortDir((current) => (current === "ascending" ? "descending" : "ascending"))
    } else {
      setSortKey(key)
      setSortDir("ascending")
    }
  }

  const ariaSort = (key: SortableKey) => (key === sortKey ? sortDir : undefined)

  return (
    <section className="spec">
      <div className="spec-head">
        <b>Tableau des membres</b>
        <span className="spec-code">table · .check · .pager</span>
      </div>
      <TableToolbar count={`${visible.length} membres`}>
        <SearchField
          label="Rechercher un membre"
          placeholder="Rechercher un membre"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <Select aria-label="Filtrer par rôle" options={roles} value={role} onChange={(event) => setRole(event.target.value)} />
      </TableToolbar>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th scope="col" style={{ width: "44px" }}>
                <CheckControl aria-label="Tout sélectionner" checked={allCheckedVisible} onChange={toggleAll} />
              </th>
              <th scope="col" className="sortable" data-sort="name" aria-sort={ariaSort("name")}>
                <button type="button" onClick={() => applySort("name")}>
                  Membre
                  <ArrowUpIcon className="sort-ind" aria-hidden="true" />
                </button>
              </th>
              <th scope="col" className="sortable" data-sort="role" aria-sort={ariaSort("role")}>
                <button type="button" onClick={() => applySort("role")}>
                  Rôle
                  <ArrowUpIcon className="sort-ind" aria-hidden="true" />
                </button>
              </th>
              <th scope="col">Portée</th>
              <th scope="col">Statut</th>
              <th scope="col" className="sortable" data-sort="activity" aria-sort={ariaSort("activity")}>
                <button type="button" onClick={() => applySort("activity")}>
                  Dernière activité
                  <ArrowUpIcon className="sort-ind" aria-hidden="true" />
                </button>
              </th>
            </tr>
          </thead>
          <tbody>
            {visible.map((member) => (
              <tr key={member.id}>
                <td>
                  <CheckControl
                    aria-label={`Sélectionner ${member.name}`}
                    checked={checked.includes(member.id)}
                    onChange={() => toggleMember(member.id)}
                  />
                </td>
                <td className="cell-user">
                  <div className="user-cell">
                    <span className="avatar">{initials(member.name)}</span>
                    <div className="who">
                      <b>{member.name}</b>
                      <span>{member.email}</span>
                    </div>
                  </div>
                </td>
                <td>
                  <Badge>{member.role}</Badge>
                </td>
                <td className="scope">
                  <b>{member.scopeTitle}</b>
                  {member.scopeDetail}
                </td>
                <td>
                  {member.status ? (
                    <Badge status={member.status}>
                      {member.status === "active" ? "Actif" : member.status === "pending" ? "En attente" : "Suspendu"}
                    </Badge>
                  ) : null}
                </td>
                <td className="col-activity">{member.activity}</td>
              </tr>
            ))}
            {visible.length === 0 ? (
              <tr>
                <td colSpan={6}>
                  <div className="empty">
                    <h3>Aucun membre</h3>
                    <p>Aucun membre ne correspond à cette recherche.</p>
                  </div>
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
      <div className="table-foot">
        <span className="range">
          1–{visible.length} sur {totalCount} membres
        </span>
        <Pager pages={[1, 2, 3]} current={1} />
      </div>
      {checked.length > 0 ? (
        <div className="bulk-bar">
          <div className="bulk">
            <b>
              <span>{checked.length}</span> sélectionné(s)
            </b>
            <div className="bulk-actions">
              <button
                type="button"
                className="btn-secondary"
                onClick={() => {
                  onAction?.("Invitation envoyée à la sélection")
                  clearSelection()
                }}
              >
                Inviter
              </button>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => {
                  onAction?.("Comptes suspendus")
                  clearSelection()
                }}
              >
                Suspendre
              </button>
              <button type="button" className="btn-secondary" onClick={clearSelection}>
                Annuler
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  )
}
