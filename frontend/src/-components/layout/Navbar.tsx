import { Link, useRouterState } from '@tanstack/react-router'

import { Activity, ArrowUpRight, BookOpen } from 'lucide-react'

import { Badge, Button, ModelSelector } from '#/-components/ui'
import { theme } from '#/-components/ui/theme'
import { useHealthCheck } from '#/services/exploration/useHealthCheck'

export const Navbar = () => {
  const { isSuccess } = useHealthCheck()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const isWorkspace = pathname === '/' || pathname.startsWith('/chat')

  return (
    <header
      style={{
        borderBottom: theme.colors.border.subtle,
        background: theme.colors.canvas.primary,
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      <div
        className="container"
        style={{
          height: '64px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Link
          to="/"
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            textDecoration: 'none',
          }}
        >
          <div
            style={{
              width: '28px',
              height: '28px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <circle cx="12" cy="6" r="3.6" fill={theme.colors.brand.coral} />
              <circle
                cx="6"
                cy="16.5"
                r="3.6"
                fill={theme.colors.brand.coral}
              />
              <circle
                cx="18"
                cy="16.5"
                r="3.6"
                fill={theme.colors.brand.vibrant}
              />
            </svg>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span
              style={{
                fontWeight: 600,
                fontSize: theme.typography.sizes.sm,
                letterSpacing: '-0.02em',
                color: theme.colors.ink.primary,
              }}
            >
              Research Helper
            </span>
            <Badge variant="neutral" size="sm" isMono>
              v0.1
            </Badge>
          </div>
        </Link>

        {/* Navigation Tabs */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Link
            to="/"
            style={{
              padding: '6px 14px',
              borderRadius: theme.radius.button,
              fontSize: '0.84rem',
              fontWeight: 500,
              textDecoration: 'none',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              backgroundColor: isWorkspace
                ? theme.colors.canvas.neutral
                : 'transparent',
              color: isWorkspace
                ? theme.colors.ink.primary
                : theme.colors.ink.secondary,
              transition: theme.transitions.fast,
            }}
          >
            <BookOpen size={14} />
            <span>Research Canvas</span>
          </Link>
        </nav>

        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <ModelSelector variant="compact" />

          <Badge
            variant={isSuccess ? 'success' : 'warning'}
            dot
            size="sm"
            style={{ textTransform: 'none' }}
          >
            <Activity size={12} style={{ marginRight: 2 }} />
            {isSuccess ? 'Core 8080' : 'Connecting...'}
          </Badge>

          <a
            href="https://github.com"
            target="_blank"
            rel="noreferrer"
            style={{ textDecoration: 'none' }}
          >
            <Button
              variant="outline"
              size="sm"
              endIcon={<ArrowUpRight size={13} />}
            >
              Specifications
            </Button>
          </a>
        </div>
      </div>
    </header>
  )
}
