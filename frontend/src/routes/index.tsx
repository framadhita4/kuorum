import { createRoute } from '@tanstack/react-router'

import { Route as rootRoute } from '#/routes/__root'
import { ResearchWorkspace } from '#/routes/-components/ResearchWorkspace'

export const IndexPage = () => {
  return <ResearchWorkspace />
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: IndexPage,
})
