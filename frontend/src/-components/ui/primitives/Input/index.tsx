import type { ReactNode } from 'react'
import * as React from 'react'

import { Input as BaseInput } from '@base-ui/react'
import { css } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

export type InputSize = 'sm' | 'md' | 'lg'

export interface InputProps extends Omit<
  React.ComponentProps<typeof BaseInput>,
  'size' | 'className'
> {
  label?: string
  error?: string
  helperText?: string
  inputSize?: InputSize
  startAdornment?: ReactNode
  endAdornment?: ReactNode
  className?: string
}

const Wrapper = styled.div`
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
`

const Label = styled.label`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  font-weight: ${theme.typography.weights.medium};
  color: ${theme.colors.ink.primary};
`

const InputContainer = styled.div`
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
`

const Adornment = styled.div<{ $position: 'start' | 'end' }>`
  position: absolute;
  ${({ $position }) => ($position === 'start' ? 'left: 14px;' : 'right: 14px;')}
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.ink.muted};
  pointer-events: none;
  z-index: 2;
`

const StyledBaseInput = styled(BaseInput)<{
  $hasError?: boolean
  $inputSize: InputSize
  $hasStart?: boolean
  $hasEnd?: boolean
}>`
  width: 100%;
  box-sizing: border-box;
  font-family: ${theme.typography.fontFamily.sans};
  background-color: ${theme.colors.canvas.primary};
  color: ${theme.colors.ink.primary};
  border: 1px solid
    ${({ $hasError }) =>
      $hasError ? theme.colors.status.error : theme.colors.border.subtleColor};
  border-radius: ${theme.radius.button};
  outline: none;
  box-shadow: none;
  transition: ${theme.transitions.fast};

  ${({ $inputSize }) => {
    switch ($inputSize) {
      case 'sm':
        return css`
          padding: 7px 12px;
          font-size: ${theme.typography.sizes.sm};
          height: 36px;
        `
      case 'lg':
        return css`
          padding: 12px 18px;
          font-size: ${theme.typography.sizes.body};
          height: 48px;
        `
      case 'md':
      default:
        return css`
          padding: 9px 14px;
          font-size: ${theme.typography.sizes.sm};
          height: 42px;
        `
    }
  }}

  ${({ $hasStart }) => $hasStart && 'padding-left: 40px;'}
  ${({ $hasEnd }) => $hasEnd && 'padding-right: 40px;'}

  &::placeholder {
    color: ${theme.colors.ink.muted};
  }

  &:focus {
    border-color: ${({ $hasError }) =>
      $hasError ? theme.colors.status.error : theme.colors.accent.purple};
    outline: 2px solid
      ${({ $hasError }) =>
        $hasError ? 'rgba(239, 68, 68, 0.2)' : 'rgba(121, 110, 255, 0.2)'};
  }

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
`

const MessageText = styled.span<{ $isError?: boolean }>`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.caption};
  color: ${({ $isError }) =>
    $isError ? theme.colors.status.error : theme.colors.ink.muted};
`

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      label,
      error,
      helperText,
      inputSize = 'md',
      startAdornment,
      endAdornment,
      id,
      className,
      ...props
    },
    ref,
  ) => {
    const inputId =
      id ||
      (label ? `input-${label.toLowerCase().replace(/\s+/g, '-')}` : undefined)

    return (
      <Wrapper className={className}>
        {label && <Label htmlFor={inputId}>{label}</Label>}
        <InputContainer>
          {startAdornment && (
            <Adornment $position="start">{startAdornment}</Adornment>
          )}
          <StyledBaseInput
            ref={ref}
            id={inputId}
            $hasError={!!error}
            $inputSize={inputSize}
            $hasStart={!!startAdornment}
            $hasEnd={!!endAdornment}
            {...props}
          />
          {endAdornment && (
            <Adornment $position="end">{endAdornment}</Adornment>
          )}
        </InputContainer>
        {(error || helperText) && (
          <MessageText $isError={!!error}>{error || helperText}</MessageText>
        )}
      </Wrapper>
    )
  },
)

Input.displayName = 'Input'
