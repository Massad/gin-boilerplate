import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from "react"
import { useState } from "react"
import { cn } from "@/lib/utils"
import { AlertCircleIcon, ChevronDownIcon, EyeIcon, EyeOffIcon, LockIcon } from "./icons"

export interface FieldProps {
  label: string
  htmlFor: string
  /** Help line under the control; may embed styled spans (e.g. slug preview). */
  help?: ReactNode
  error?: string
  children: ReactNode
}

export function Field({ label, htmlFor, help, error, children }: FieldProps) {
  return (
    <div className="field" data-invalid={error ? "true" : undefined}>
      <label htmlFor={htmlFor}>{label}</label>
      {children}
      {help ? <p className="field-help">{help}</p> : null}
      {error ? <p className="field-error">{error}</p> : null}
    </div>
  )
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  icon?: ReactNode
}

export function Input({ icon, className, ...rest }: InputProps) {
  if (!icon) {
    return <input className={cn("no-icon", className)} {...rest} />
  }
  return (
    <div className="input-wrap">
      <span className="input-icon">{icon}</span>
      <input className={className} {...rest} />
    </div>
  )
}

export function PasswordInput(props: InputHTMLAttributes<HTMLInputElement>) {
  const [visible, setVisible] = useState(false)

  return (
    <div className="input-wrap">
      <span className="input-icon">
        <LockIcon aria-hidden="true" />
      </span>
      <input type={visible ? "text" : "password"} {...props} />
      <button
        type="button"
        className="pwd-toggle"
        aria-label={visible ? "Masquer le mot de passe" : "Afficher le mot de passe"}
        aria-pressed={visible ? "true" : "false"}
        onClick={() => setVisible((value) => !value)}
      >
        {visible ? <EyeOffIcon aria-hidden="true" /> : <EyeIcon aria-hidden="true" />}
      </button>
    </div>
  )
}

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  options: string[]
  fullWidth?: boolean
}

export function Select({ options, fullWidth, ...rest }: SelectProps) {
  return (
    <span className="select-wrap" style={fullWidth ? { width: "100%" } : undefined}>
      <select {...rest}>
        {options.map((option) => (
          <option key={option}>{option}</option>
        ))}
      </select>
      <ChevronDownIcon className="chev" aria-hidden="true" />
    </span>
  )
}

export function TextArea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea {...props} />
}

export function Checkbox({
  label,
  defaultChecked = false,
}: {
  label: string
  defaultChecked?: boolean
}) {
  const [checked, setChecked] = useState(defaultChecked)

  return (
    <label className="checkbox">
      <input type="checkbox" checked={checked} onChange={(event) => setChecked(event.target.checked)} />
      <span>{label}</span>
    </label>
  )
}

export function CheckControl({
  defaultChecked = false,
  checked,
  ...rest
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      type="checkbox"
      className="check"
      {...(checked === undefined ? { defaultChecked } : { checked })}
      {...rest}
    />
  )
}

export function Divider({ children }: { children: ReactNode }) {
  return <p className="divider">{children}</p>
}

export function ErrorBanner({ children }: { children: ReactNode }) {
  return (
    <div className="error-banner" role="alert">
      <AlertCircleIcon aria-hidden="true" />
      <span>{children}</span>
    </div>
  )
}
