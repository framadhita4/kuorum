import { createRootRoute, Outlet } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: function Root() {
    return <Outlet />
  },
})
