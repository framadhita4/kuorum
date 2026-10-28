import { Dialog as BaseDialog } from '@base-ui/react'
import { keyframes } from '@emotion/react'
import styled from '@emotion/styled'

import { theme } from '../../theme'

const fadeIn = keyframes`
  from { opacity: 0; }
  to { opacity: 1; }
`

const scaleIn = keyframes`
  from {
    opacity: 0;
    transform: translate(-50%, -49%) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
`

export const DialogRoot = BaseDialog.Root
export const DialogTrigger = BaseDialog.Trigger
export const DialogPortal = BaseDialog.Portal

export const DialogBackdrop = styled(BaseDialog.Backdrop)`
  position: fixed;
  inset: 0;
  background: ${theme.colors.canvas.overlay};
  z-index: 100;
  animation: ${fadeIn} 0.2s ease-out;

  &[data-closed] {
    opacity: 0;
  }
`

export const DialogPopup = styled(BaseDialog.Popup)`
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 101;
  background: ${theme.colors.canvas.primary};
  border: ${theme.colors.border.subtle};
  border-radius: ${theme.radius.card};
  padding: ${theme.spacing['6']};
  box-shadow: ${theme.shadows.lg};
  width: 90vw;
  max-width: 520px;
  outline: none;
  animation: ${scaleIn} 0.2s cubic-bezier(0.16, 1, 0.3, 1);

  &[data-closed] {
    opacity: 0;
    transform: translate(-50%, -49%) scale(0.98);
  }
`

export const DialogTitle = styled(BaseDialog.Title)`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.h4};
  font-weight: ${theme.typography.weights.medium};
  letter-spacing: -0.01em;
  color: ${theme.colors.ink.primary};
  margin: 0 0 6px 0;
`

export const DialogDescription = styled(BaseDialog.Description)`
  font-family: ${theme.typography.fontFamily.sans};
  font-size: ${theme.typography.sizes.sm};
  color: ${theme.colors.ink.secondary};
  line-height: ${theme.typography.lineHeights.normal};
  margin: 0 0 20px 0;
`

export const DialogClose = styled(BaseDialog.Close)`
  position: absolute;
  top: 16px;
  right: 16px;
  background: transparent;
  border: 1px solid rgba(100, 111, 121, 0.2);
  border-radius: ${theme.radius.sm};
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.colors.ink.muted};
  cursor: pointer;
  transition: ${theme.transitions.fast};

  &:hover {
    color: ${theme.colors.ink.primary};
    background: ${theme.colors.canvas.neutral};
  }
`

export const Dialog = {
  Root: DialogRoot,
  Trigger: DialogTrigger,
  Portal: DialogPortal,
  Backdrop: DialogBackdrop,
  Popup: DialogPopup,
  Title: DialogTitle,
  Description: DialogDescription,
  Close: DialogClose,
}
