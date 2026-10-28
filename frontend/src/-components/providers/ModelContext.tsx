import type { ReactNode } from 'react'
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'

import { useGetModels } from '#/services/llm/useGetModels'
import type { LLMModel, LLMProvider, ThinkingLevel } from '#/types/llm.type'

const STORAGE_KEY = 'kuorum_selected_model'
const THINKING_LEVEL_STORAGE_KEY = 'kuorum_thinking_level'

interface ModelContextValue {
  selectedModelId: string
  selectedModel: LLMModel | null
  setSelectedModelId: (id: string) => void
  thinkingLevel: ThinkingLevel
  setThinkingLevel: (level: ThinkingLevel) => void
  models: Array<LLMModel>
  providers: Array<LLMProvider>
  isLoading: boolean
}

const ModelContext = createContext<ModelContextValue | null>(null)

export const ModelProvider = ({ children }: { children: ReactNode }) => {
  const { data, isLoading } = useGetModels()

  // Use only models and providers that have configured API keys from the server
  const models = useMemo(
    () => (data?.models !== undefined ? data.models : []),
    [data?.models],
  )

  const providers = useMemo(
    () => (data?.providers !== undefined ? data.providers : []),
    [data?.providers],
  )

  const defaultId = data?.defaultModelId || (models[0]?.id ?? '')

  // Initialize from localStorage or fallback
  const [selectedModelId, setSelectedModelIdState] = useState<string>(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) return stored
    } catch {
      // localStorage unavailable or restricted
    }
    return defaultId
  })

  const [thinkingLevel, setThinkingLevelState] = useState<ThinkingLevel>(() => {
    try {
      const stored = localStorage.getItem(THINKING_LEVEL_STORAGE_KEY)
      if (stored === 'low' || stored === 'medium' || stored === 'high') {
        return stored
      }
    } catch {
      // localStorage unavailable or restricted
    }
    return 'low'
  })

  // Synchronize selection if stored model is invalid or models list updates
  useEffect(() => {
    if (models.length > 0) {
      if (!models.some((m) => m.id === selectedModelId)) {
        const nextId =
          data?.defaultModelId &&
          models.some((m) => m.id === data.defaultModelId)
            ? data.defaultModelId
            : models[0].id
        setSelectedModelIdState(nextId)
      }
    } else if (!isLoading) {
      setSelectedModelIdState('')
    }
  }, [models, selectedModelId, data?.defaultModelId, isLoading])

  const setSelectedModelId = useCallback((id: string) => {
    setSelectedModelIdState(id)
    try {
      localStorage.setItem(STORAGE_KEY, id)
    } catch {
      // ignore storage error
    }
  }, [])

  const setThinkingLevel = useCallback((level: ThinkingLevel) => {
    setThinkingLevelState(level)
    try {
      localStorage.setItem(THINKING_LEVEL_STORAGE_KEY, level)
    } catch {
      // ignore storage error
    }
  }, [])

  const selectedModel = useMemo(() => {
    if (models.length === 0) return null
    const found = models.find((m) => m.id === selectedModelId)
    return found ?? models[0]
  }, [models, selectedModelId])

  const value = useMemo<ModelContextValue>(
    () => ({
      selectedModelId,
      selectedModel,
      setSelectedModelId,
      thinkingLevel,
      setThinkingLevel,
      models,
      providers,
      isLoading,
    }),
    [
      selectedModelId,
      selectedModel,
      setSelectedModelId,
      thinkingLevel,
      setThinkingLevel,
      models,
      providers,
      isLoading,
    ],
  )

  return <ModelContext.Provider value={value}>{children}</ModelContext.Provider>
}

export const useSelectedModel = (): ModelContextValue => {
  const context = useContext(ModelContext)
  if (!context) {
    throw new Error('useSelectedModel must be used within a ModelProvider')
  }
  return context
}
