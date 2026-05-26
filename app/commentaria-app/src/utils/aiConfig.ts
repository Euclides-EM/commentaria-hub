import type { feature_AIProvider } from '@hub-api'

export type AIProviderOption = {
  value: feature_AIProvider
  label: string
}

export type AIModelOption = {
  value: string
  label: string
}

export const aiProviderOptions: AIProviderOption[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'ollama', label: 'Ollama' },
]

export const aiModelOptionsByProvider: Record<
  feature_AIProvider,
  AIModelOption[]
> = {
  openai: [{ value: 'gpt-5-mini', label: 'gpt-5-mini' }],
  ollama: [{ value: 'gpt-oss:120b', label: 'gpt-oss:120b' }],
}
