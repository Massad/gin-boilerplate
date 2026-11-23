import { useAuth } from "@/auth/AuthContext"
import { sessionDisplayName } from "@/auth/session"

/** Home of the platform superadmin after the unified login. */
export default function AdminHomePage() {
  const { session } = useAuth()

  return (
    <main className="home-page">
      <h1>Admin</h1>
      <p>Bienvenue {session === null ? "" : sessionDisplayName(session)}</p>
    </main>
  )
}
