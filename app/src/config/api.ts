import { OpenAPI } from '../api'
import { useAuthStore } from '../store/authStore'

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8085'
const BACKEND_BASE_URL = import.meta.env.VITE_BACKEND_BASE_URL || '/api/v1'

export const API_BASE_URL = `${BACKEND_URL}${BACKEND_BASE_URL}`

export function initializeAPI() {
  OpenAPI.BASE = API_BASE_URL

  OpenAPI.TOKEN = async () => {
    const token = useAuthStore.getState().token
    return token || ''
  }

  console.log('API initialized with base URL:', API_BASE_URL)
}
