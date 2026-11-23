import { useAuth } from "@/auth/AuthContext"
import { sessionDisplayName } from "@/auth/session"

// « Tenant » by symmetry with « Admin » (identical content: title + welcome
// sentence). Isolated here so the label can change in one line at review.
const HOME_TITLE = "Tenant"

/** Home of the tenant user after the unified login. */
export default function TenantHomePage() {
  const { session } = useAuth()

  return (
    <main className="home-page">
      <h1>{HOME_TITLE}</h1>
      <p>Bienvenue {session === null ? "" : sessionDisplayName(session)}</p>
    </main>
  )
}
