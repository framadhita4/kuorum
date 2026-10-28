import type { LLMModel, LLMProvider } from '#/types/llm.type'

export const DEFAULT_FALLBACK_MODELS: Array<LLMModel> = [
  // Google Gemini
  {
    id: 'gemini-2.0-flash',
    displayName: 'Gemini 2.0 Flash',
    provider: 'gemini',
    providerDisplayName: 'Google Gemini',
    description:
      'Ultra-fast multimodal flagship with 1M token context window and high efficiency.',
    contextWindow: 1048576,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'gemini-2.0-pro-exp-02-05',
    displayName: 'Gemini 2.0 Pro Exp',
    provider: 'gemini',
    providerDisplayName: 'Google Gemini',
    description:
      'Frontier reasoning and complex code synthesis model with 2M token context.',
    contextWindow: 2097152,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'gemini-1.5-pro',
    displayName: 'Gemini 1.5 Pro',
    provider: 'gemini',
    providerDisplayName: 'Google Gemini',
    description: 'High-capability multimodal analysis with massive 2M context.',
    contextWindow: 2097152,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  // Anthropic Claude
  {
    id: 'claude-3-7-sonnet-20250219',
    displayName: 'Claude 3.7 Sonnet',
    provider: 'anthropic',
    providerDisplayName: 'Anthropic',
    description:
      'Hybrid reasoning flagship supporting instant responses or deep extended thinking budgets.',
    contextWindow: 200000,
    maxOutputTokens: 64000,
    thinkingCapability: 'optional',
    supportedEfforts: ['low', 'medium', 'high'],
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'claude-3-5-sonnet-20241022',
    displayName: 'Claude 3.5 Sonnet',
    provider: 'anthropic',
    providerDisplayName: 'Anthropic',
    description:
      'Industry-standard text analysis, scientific synthesis, and technical code generation.',
    contextWindow: 200000,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'claude-3-5-haiku-20241022',
    displayName: 'Claude 3.5 Haiku',
    provider: 'anthropic',
    providerDisplayName: 'Anthropic',
    description:
      'High-velocity compact model for quick classification, decomposition, and summarization.',
    contextWindow: 200000,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  // OpenAI
  {
    id: 'gpt-4o-mini',
    displayName: 'GPT-4o Mini',
    provider: 'openai',
    providerDisplayName: 'OpenAI',
    description:
      'Fast, cost-efficient model for lightweight reasoning, multimodal tasks, and vision.',
    contextWindow: 128000,
    maxOutputTokens: 16384,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: true,
  },
  {
    id: 'gpt-4o',
    displayName: 'GPT-4o',
    provider: 'openai',
    providerDisplayName: 'OpenAI',
    description:
      'High-intelligence flagship multimodal model for advanced text and vision tasks.',
    contextWindow: 128000,
    maxOutputTokens: 16384,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'o3-mini',
    displayName: 'o3-mini',
    provider: 'openai',
    providerDisplayName: 'OpenAI',
    description:
      'High-speed STEM, coding, and mathematical reasoning model with selectable reasoning effort.',
    contextWindow: 200000,
    maxOutputTokens: 100000,
    thinkingCapability: 'required',
    supportedEfforts: ['low', 'medium', 'high'],
    allowsTemperature: false,
    isDefault: false,
  },
  {
    id: 'o1',
    displayName: 'o1',
    provider: 'openai',
    providerDisplayName: 'OpenAI',
    description:
      'Deep reasoning frontier model trained with reinforcement learning for complex multi-step problems.',
    contextWindow: 200000,
    maxOutputTokens: 100000,
    thinkingCapability: 'required',
    allowsTemperature: false,
    isDefault: false,
  },
  // DeepSeek
  {
    id: 'deepseek-chat',
    displayName: 'DeepSeek-V3',
    provider: 'deepseek',
    providerDisplayName: 'DeepSeek',
    description:
      'General flagship conversation and coding model with 671B MoE architecture.',
    contextWindow: 64000,
    maxOutputTokens: 8192,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
  {
    id: 'deepseek-reasoner',
    displayName: 'DeepSeek-R1',
    provider: 'deepseek',
    providerDisplayName: 'DeepSeek',
    description:
      'Frontier reasoning model with reinforced multi-step thinking tokens for rigorous proofs.',
    contextWindow: 64000,
    maxOutputTokens: 8192,
    thinkingCapability: 'required',
    allowsTemperature: false,
    isDefault: false,
  },
  // Groq / Meta
  {
    id: 'llama-3.3-70b-versatile',
    displayName: 'Llama 3.3 70B (Groq)',
    provider: 'groq',
    providerDisplayName: 'Groq',
    description:
      'Meta 70B parameter model operating at hundreds of tokens/sec on Groq LPU hardware.',
    contextWindow: 128000,
    maxOutputTokens: 32768,
    thinkingCapability: 'not_supported',
    allowsTemperature: true,
    isDefault: false,
  },
]

export const DEFAULT_FALLBACK_PROVIDERS: Array<LLMProvider> = [
  {
    id: 'gemini',
    displayName: 'Google Gemini',
    description: 'Google official Gemini OpenAI-compatible API',
    defaultModel: 'gemini-2.0-flash',
    modelIds: [
      'gemini-2.0-flash',
      'gemini-2.0-pro-exp-02-05',
      'gemini-1.5-pro',
    ],
  },
  {
    id: 'anthropic',
    displayName: 'Anthropic',
    description:
      'Anthropic Claude models specializing in technical reasoning and coding',
    defaultModel: 'claude-3-7-sonnet-20250219',
    modelIds: [
      'claude-3-7-sonnet-20250219',
      'claude-3-5-sonnet-20241022',
      'claude-3-5-haiku-20241022',
    ],
  },
  {
    id: 'openai',
    displayName: 'OpenAI',
    description: 'OpenAI official API (GPT-4o, o1, o3-mini)',
    defaultModel: 'gpt-4o-mini',
    modelIds: ['gpt-4o-mini', 'gpt-4o', 'o3-mini', 'o1'],
  },
  {
    id: 'deepseek',
    displayName: 'DeepSeek',
    description: 'DeepSeek frontier MoE models & reasoning architecture',
    defaultModel: 'deepseek-chat',
    modelIds: ['deepseek-chat', 'deepseek-reasoner'],
  },
  {
    id: 'groq',
    displayName: 'Groq',
    description: 'Ultra-fast LPU inference hosting open models',
    defaultModel: 'llama-3.3-70b-versatile',
    modelIds: ['llama-3.3-70b-versatile'],
  },
]
