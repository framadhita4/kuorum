import type { ReactNode } from 'react'
import { createContext, useContext, useState } from 'react'

import type { ChatMessage } from '#/routes/-types/research.type'

interface ChatMessagesContextValue {
  messagesMap: Record<string, Array<ChatMessage>>
  setMessagesMap: React.Dispatch<
    React.SetStateAction<Record<string, Array<ChatMessage>>>
  >
  splitRatio: number
  setSplitRatio: React.Dispatch<React.SetStateAction<number>>
  isRightPanelOpen: boolean
  setIsRightPanelOpen: React.Dispatch<React.SetStateAction<boolean>>
  rightPanelTab: 'report' | 'references' | 'graph'
  setRightPanelTab: React.Dispatch<
    React.SetStateAction<'report' | 'references' | 'graph'>
  >
}

const ChatMessagesContext = createContext<ChatMessagesContextValue | null>(null)

export const ChatMessagesProvider = ({ children }: { children: ReactNode }) => {
  const [messagesMap, setMessagesMap] = useState<
    Record<string, Array<ChatMessage>>
  >({ new: [] })
  const [splitRatio, setSplitRatio] = useState<number>(0.5)
  const [isRightPanelOpen, setIsRightPanelOpen] = useState<boolean>(true)
  const [rightPanelTab, setRightPanelTab] = useState<
    'report' | 'references' | 'graph'
  >('report')

  return (
    <ChatMessagesContext.Provider
      value={{
        messagesMap,
        setMessagesMap,
        splitRatio,
        setSplitRatio,
        isRightPanelOpen,
        setIsRightPanelOpen,
        rightPanelTab,
        setRightPanelTab,
      }}
    >
      {children}
    </ChatMessagesContext.Provider>
  )
}

export const useChatMessages = (): ChatMessagesContextValue => {
  const context = useContext(ChatMessagesContext)
  if (!context) {
    throw new Error(
      'useChatMessages must be used within a ChatMessagesProvider',
    )
  }
  return context
}
