import { z } from 'zod'

export const chatSchema = z.object({
  prompt: z
    .string()
    .trim()
    .min(1, 'Query cannot be empty')
    .max(1000, 'Query cannot exceed 1000 characters'),
})

export type ChatValues = z.infer<typeof chatSchema>
