import { useWindowDimensions } from 'react-native'

const BASE_WIDTH = 390
const BASE_HEIGHT = 844

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value))

export function useResponsiveScale() {
  const { width, height } = useWindowDimensions()
  const widthScale = width / BASE_WIDTH
  const heightScale = height / BASE_HEIGHT
  const scale = clamp(Math.min(widthScale, heightScale), 0.86, 1.04)
  const compact = width < 360 || height < 700

  const font = (size: number, min = 12) => Math.max(min, Math.round(size * scale))

  return { width, height, scale, compact, font }
}
