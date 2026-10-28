import styled from '@emotion/styled'
import { zodResolver } from '@hookform/resolvers/zod'
import { ArrowRight } from 'lucide-react'
import { useForm } from 'react-hook-form'

import { Button, Input, ModelSelector } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import type { ExploreFormValues } from '#/routes/-schema/exploreForm.schema'
import { exploreFormSchema } from '#/routes/-schema/exploreForm.schema'

interface ConceptInputFormProps {
  onSubmit: (values: ExploreFormValues) => void
  isLoading: boolean
}

const FormContainer = styled.form`
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 640px;
  margin: 0 auto;
  width: 100%;
`

const InputRow = styled.div`
  display: flex;
  gap: 8px;
  align-items: flex-start;
`

const InputFlex = styled.div`
  flex: 1;
`

const MetaRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  padding-left: 2px;
`

const ExamplesContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.78rem;
  color: ${theme.colors.ink.muted};
  font-family: ${theme.typography.fontFamily.sans};
`

const ExampleTag = styled.button`
  background: transparent;
  border: none;
  color: ${theme.colors.ink.secondary};
  cursor: pointer;
  text-decoration: underline;
  font-family: inherit;
  font-size: inherit;
  padding: 0;
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.accent.purple};
  }

  &:focus-visible {
    outline: 2px solid ${theme.colors.accent.purple};
    outline-offset: 2px;
  }
`

export const ConceptInputForm = ({
  onSubmit,
  isLoading,
}: ConceptInputFormProps) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ExploreFormValues>({
    resolver: zodResolver(exploreFormSchema),
    defaultValues: {
      prompt: 'Self-Attention in Vision Transformers',
    },
  })

  return (
    <FormContainer onSubmit={handleSubmit(onSubmit)}>
      <InputRow>
        <InputFlex>
          <Input
            id="prompt-input"
            inputSize="lg"
            placeholder="Enter research concept (e.g. Graph Neural Networks)..."
            error={errors.prompt?.message}
            {...register('prompt')}
          />
        </InputFlex>
        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={isLoading}
          endIcon={!isLoading && <ArrowRight size={15} />}
          style={{ minWidth: '130px' }}
        >
          Decompose
        </Button>
      </InputRow>

      <MetaRow>
        <ExamplesContainer>
          <span>Examples:</span>
          {[
            'Transformer Attention',
            'Diffusion Models',
            'Federated Learning',
          ].map((tag) => (
            <ExampleTag
              key={tag}
              type="button"
              onClick={() => onSubmit({ prompt: tag })}
            >
              {tag}
            </ExampleTag>
          ))}
        </ExamplesContainer>

        <ModelSelector variant="pill" />
      </MetaRow>
    </FormContainer>
  )
}
