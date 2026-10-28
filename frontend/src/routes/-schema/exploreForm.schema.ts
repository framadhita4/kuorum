import { z } from 'zod'

export const exploreFormSchema = z.object({
  prompt: z
    .string()
    .min(3, 'Research prompt must be at least 3 characters long')
    .max(500, 'Research prompt must not exceed 500 characters'),
})

export type ExploreFormValues = z.infer<typeof exploreFormSchema>
