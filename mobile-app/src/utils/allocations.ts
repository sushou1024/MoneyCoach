import { useTheme } from '../providers/ThemeProvider'

export function allocationLabelKey(label: string) {
  switch (label) {
    case 'crypto':
      return 'allocation.crypto'
    case 'stock':
      return 'allocation.stock'
    case 'cash':
      return 'allocation.cash'
    case 'manual':
      return 'allocation.manual'
    default:
      return 'allocation.other'
  }
}

export function allocationColor(theme: ReturnType<typeof useTheme>, label: string) {
  switch (label) {
    case 'crypto':
      return theme.colors.accent
    case 'stock':
      return theme.colors.warning
    case 'cash':
      return theme.colors.success
    case 'manual':
      return theme.colors.muted
    default:
      return theme.colors.border
  }
}
