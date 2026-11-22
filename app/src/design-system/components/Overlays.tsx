import { useEffect, useState } from "react"
import type { ReactNode } from "react"
import { CheckCircleIcon, CloseIcon } from "./icons"

function useEscape(open: boolean, onClose: () => void) {
  useEffect(() => {
    if (!open) {
      return undefined
    }
    const handleKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose()
      }
    }
    document.addEventListener("keydown", handleKey)
    return () => document.removeEventListener("keydown", handleKey)
  }, [open, onClose])
}

function useBodyOverflow(open: boolean) {
  useEffect(() => {
    if (!open) {
      return undefined
    }
    const previous = document.body.style.overflow
    document.body.style.overflow = "hidden"
    return () => {
      document.body.style.overflow = previous
    }
  }, [open])
}

export interface DrawerProps {
  open: boolean
  onClose: () => void
  title: string
  subtitle: string
  initials: string
  children: ReactNode
  footer: ReactNode
}

export function Drawer({ open, onClose, title, subtitle, initials, children, footer }: DrawerProps) {
  useEscape(open, onClose)
  useBodyOverflow(open)

  return (
    <>
      <div className="scrim" data-show={open ? "true" : "false"} onClick={onClose} aria-hidden="true" />
      <aside
        className="drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="drawer-title"
        aria-hidden={!open}
        data-show={open ? "true" : "false"}
      >
        <div className="drawer-head">
          <span className="avatar" aria-hidden="true">
            {initials}
          </span>
          <div className="drawer-title">
            <h2>{title}</h2>
            <p>{subtitle}</p>
          </div>
          <button type="button" className="icon-btn" aria-label="Fermer le tiroir" onClick={onClose}>
            <CloseIcon aria-hidden="true" />
          </button>
        </div>
        <div className="drawer-body">{children}</div>
        <div className="drawer-foot">{footer}</div>
      </aside>
    </>
  )
}

export interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  description?: string
  labelledBy: string
  children: ReactNode
  footer: ReactNode
}

export function Modal({ open, onClose, title, description, labelledBy, children, footer }: ModalProps) {
  useEscape(open, onClose)
  useBodyOverflow(open)
  if (!open) {
    return null
  }
  return (
    <>
      <div className="scrim" data-show="true" onClick={onClose} aria-hidden="true" />
      <div
        className="modal-wrap"
        onClick={(event) => {
          if (event.target === event.currentTarget) {
            onClose()
          }
        }}
      >
        <div className="modal" role="dialog" aria-modal="true" aria-labelledby={labelledBy}>
          <div className="modal-head">
            <h2>{title}</h2>
            {description ? <p>{description}</p> : null}
          </div>
          <div className="modal-body">{children}</div>
          <div className="modal-foot">{footer}</div>
        </div>
      </div>
    </>
  )
}

export interface ToastProps {
  title: string
  note: string
}

export function Toast({ title, note }: ToastProps) {
  return (
    <div className="toast">
      <CheckCircleIcon aria-hidden="true" />
      <div>
        <b>{title}</b>
        <span>{note}</span>
      </div>
    </div>
  )
}

export interface ToastItem extends ToastProps {
  id: number
}

export function Toasts({ items }: { items: ToastItem[] }) {
  return (
    <div className="toasts" aria-live="polite">
      {items.map((toast) => (
        <Toast key={toast.id} title={toast.title} note={toast.note} />
      ))}
    </div>
  )
}

export function useToasts() {
  const [items, setItems] = useState<ToastItem[]>([])

  const push = (title: string, note: string) => {
    const id = Date.now() + Math.random()
    setItems((current) => [...current, { id, title, note }])
    window.setTimeout(() => {
      setItems((current) => current.filter((toast) => toast.id !== id))
    }, 4200)
  }

  return { items, push }
}
