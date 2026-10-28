import '@emotion/react'

import type { Theme as CustomTheme } from '#/-components/ui/theme'

declare module '@emotion/react' {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  export interface Theme extends CustomTheme {}
}
