import { createRoute } from '@tanstack/react-router'

import { Route as rootRoute } from '#/routes/__root'
import { ResearchWorkspace } from '#/routes/-components/ResearchWorkspace'

export const ChatPage = () => {
  const { chatId } = Route.useParams()
  return <ResearchWorkspace chatId={chatId} />
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/chat/$chatId',
  component: ChatPage,
})
