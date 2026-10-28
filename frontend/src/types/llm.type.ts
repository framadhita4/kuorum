export type ThinkingCapability = 'not_supported' | 'optional' | 'required'
export type ThinkingLevel = 'low' | 'medium' | 'high'

export interface LLMModel {
  id: string
  displayName: string
  provider: string
  providerDisplayName: string
  description: string
  contextWindow: number
  maxOutputTokens: number
  thinkingCapability: ThinkingCapability
  supportedEfforts?: Array<string>
  allowsTemperature: boolean
  isDefault: boolean
}

export interface LLMProvider {
  id: string
  displayName: string
  description: string
  defaultModel: string
  modelIds: Array<string>
}

export interface GetModelsResult {
  defaultModelId: string
  models: Array<LLMModel>
  providers: Array<LLMProvider>
}
