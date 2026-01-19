import { useMutation } from '@tanstack/react-query'
import type { httpapi_AuthValidateResponse } from '../api'
import { AuthenticationService, OpenAPI } from '../api'

interface AuthValidateRequest {
  token: string
}

export function useAuthValidateMutation() {
  return useMutation<httpapi_AuthValidateResponse, Error, AuthValidateRequest>({
    mutationFn: async ({ token }) => {
      const originalToken = OpenAPI.TOKEN
      OpenAPI.TOKEN = async () => token

      try {
        return await AuthenticationService.getAuthValidate()
      } finally {
        OpenAPI.TOKEN = originalToken
      }
    },
  })
}
