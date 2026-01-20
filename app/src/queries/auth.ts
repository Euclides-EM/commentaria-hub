import { useMutation } from '@tanstack/react-query'
import { AuthenticationService } from '../api'
import { useAuthStore } from '../store/authStore.ts'

interface AuthValidateRequest {
  token: string
}

export function useAuthValidateMutation() {
  const [setAuth, clearAuth] = useAuthStore((store) => [
    store.setAuth,
    store.clearAuth,
  ])
  return useMutation<boolean, Error, AuthValidateRequest>({
    mutationFn: async ({ token }) => {
      try {
        setAuth(token, '')
        const res = await AuthenticationService.getAuthValidate()
        setAuth(token, res.username || res.email || '')
        return true
      } catch (error) {
        console.warn('Auth validation failed:', error)
        clearAuth()
        throw error
      }
    },
  })
}
