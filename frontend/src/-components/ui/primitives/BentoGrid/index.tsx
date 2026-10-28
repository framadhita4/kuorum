import type { HTMLAttributes, ReactNode } from 'react'

import styled from '@emotion/styled'

import { theme } from '../../theme'

export interface BentoGridProps extends HTMLAttributes<HTMLDivElement> {
  columns?: 1 | 2 | 3 | 4
  gap?: 'sm' | 'md' | 'lg'
  children: ReactNode
}

export interface BentoItemProps extends HTMLAttributes<HTMLDivElement> {
  colSpan?: 1 | 2 | 3 | 4
  rowSpan?: 1 | 2
  children: ReactNode
}

const StyledBentoGrid = styled.div<{
  $columns: number
  $gap: 'sm' | 'md' | 'lg'
}>`
  display: grid;
  width: 100%;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: ${({ $gap }) => {
    switch ($gap) {
      case 'sm':
        return theme.spacing['4']
      case 'lg':
        return theme.spacing['8']
      case 'md':
      default:
        return theme.spacing['6']
    }
  }};

  @media (min-width: 768px) {
    grid-template-columns: repeat(
      ${({ $columns }) => Math.min($columns, 2)},
      minmax(0, 1fr)
    );
  }

  @media (min-width: 1024px) {
    grid-template-columns: repeat(
      ${({ $columns }) => $columns},
      minmax(0, 1fr)
    );
  }
`

const StyledBentoItem = styled.div<{
  $colSpan?: number
  $rowSpan?: number
}>`
  display: flex;
  flex-direction: column;
  width: 100%;

  @media (min-width: 1024px) {
    ${({ $colSpan }) => $colSpan && `grid-column: span ${$colSpan};`}
    ${({ $rowSpan }) => $rowSpan && `grid-row: span ${$rowSpan};`}
  }
`

export const BentoGrid = ({
  columns = 3,
  gap = 'md',
  children,
  ...props
}: BentoGridProps) => {
  return (
    <StyledBentoGrid $columns={columns} $gap={gap} {...props}>
      {children}
    </StyledBentoGrid>
  )
}

export const BentoItem = ({
  colSpan,
  rowSpan,
  children,
  ...props
}: BentoItemProps) => {
  return (
    <StyledBentoItem $colSpan={colSpan} $rowSpan={rowSpan} {...props}>
      {children}
    </StyledBentoItem>
  )
}
