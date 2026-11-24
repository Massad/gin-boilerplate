import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterContextProvider, RouterProvider } from "@tanstack/react-router"
import { AuthProvider } from "@/auth/AuthContext"
import { router } from "./router"

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      {/*
        The AuthProvider navigates through the router (useNavigate), so the
        router context must already be available where the provider is
        mounted: RouterContextProvider exposes it above the RouterProvider.
      */}
      <RouterContextProvider router={router}>
        <AuthProvider>
          <RouterProvider router={router} />
        </AuthProvider>
      </RouterContextProvider>
    </QueryClientProvider>
  )
}
