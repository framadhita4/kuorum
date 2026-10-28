import { createRoute, redirect } from '@tanstack/react-router'

import { Route as rootRoute } from '#/routes/__root'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/chat',
  beforeLoad: () => {
    throw redirect({ to: '/' })
  },
})
