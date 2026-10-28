import '#/index.css'

import { StrictMode } from 'react'

import { createRouter, RouterProvider } from '@tanstack/react-router'

import ReactDOM from 'react-dom/client'

import { ChatMessagesProvider } from '#/-components/providers/ChatMessagesContext'
import { ModelProvider } from '#/-components/providers/ModelContext'
import { QueryProvider } from '#/-components/providers/QueryProvider'
import { ThemeProvider } from '#/-components/ui'
import { Route as rootRoute } from '#/routes/__root'
import { Route as chatRoute } from '#/routes/chat/$chatId'
import { Route as chatIndexRoute } from '#/routes/chat/index'
import { Route as indexRoute } from '#/routes/index'

const routeTree = rootRoute.addChildren([indexRoute, chatRoute, chatIndexRoute])

const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const rootElement = document.getElementById('root')
if (rootElement) {
  const root = ReactDOM.createRoot(rootElement)
  root.render(
    <StrictMode>
      <QueryProvider>
        <ThemeProvider>
          <ModelProvider>
            <ChatMessagesProvider>
              <RouterProvider router={router} />
            </ChatMessagesProvider>
          </ModelProvider>
        </ThemeProvider>
      </QueryProvider>
    </StrictMode>,
  )
}
