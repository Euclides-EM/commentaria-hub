import { OpenAPI } from '@hub-api';

export const configureHubApi = (authToken: string | null) => {
  OpenAPI.BASE = `${import.meta.env.VITE_BACKEND_URL}${import.meta.env.VITE_BACKEND_BASE_URL}`;
  OpenAPI.TOKEN = authToken ?? "";
};
