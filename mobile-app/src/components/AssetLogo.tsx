import React, { useEffect, useMemo, useState } from 'react'
import { Image, Text, View } from 'react-native'
import { SvgXml } from 'react-native-svg'

import { useTheme } from '../providers/ThemeProvider'

type AssetLogoProps = {
  uri?: string | null
  label: string
  size?: number
}

export function AssetLogo({ uri, label, size = 40 }: AssetLogoProps) {
  const theme = useTheme()
  const [failed, setFailed] = useState(false)
  const [svgXml, setSvgXml] = useState<string | null>(null)

  const initial = useMemo(() => {
    const trimmed = label.trim()
    if (!trimmed) return '?'
    return trimmed.slice(0, 1).toUpperCase()
  }, [label])

  const resolvedUri = typeof uri === 'string' ? uri.trim() : ''
  const isSvg = resolvedUri.toLowerCase().endsWith('.svg')
  const showLogo = resolvedUri !== '' && !failed

  useEffect(() => {
    if (!resolvedUri) return
    setFailed(false)
  }, [resolvedUri])

  useEffect(() => {
    if (!showLogo || !isSvg) {
      setSvgXml(null)
      return
    }
    let active = true
    setSvgXml(null)
    fetch(resolvedUri)
      .then((resp) => {
        if (!resp.ok) {
          throw new Error('Failed to load svg')
        }
        return resp.text()
      })
      .then((text) => {
        if (!active) return
        setSvgXml(normalizeSvgViewBox(text))
      })
      .catch(() => {
        if (active) setFailed(true)
      })
    return () => {
      active = false
    }
  }, [isSvg, resolvedUri, showLogo])

  return (
    <View
      style={{
        width: size,
        height: size,
        borderRadius: size / 2,
        backgroundColor: theme.colors.accentSoft,
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
      }}
    >
      {showLogo ? (
        isSvg ? (
          svgXml ? (
            <SvgXml xml={svgXml} width={size} height={size} />
          ) : (
            <View style={{ width: size, height: size }} />
          )
        ) : (
          <Image
            source={{ uri: resolvedUri }}
            style={{ width: size, height: size }}
            resizeMode="contain"
            onError={() => setFailed(true)}
          />
        )
      ) : (
        <Text style={{ color: theme.colors.accent, fontFamily: theme.fonts.bodyBold }}>{initial}</Text>
      )}
    </View>
  )
}

function normalizeSvgViewBox(source: string) {
  const svgTagMatch = source.match(/<svg[^>]*>/i)
  if (!svgTagMatch) return source
  const svgTag = svgTagMatch[0]
  if (/viewBox=/i.test(svgTag)) return source
  // Some vendor SVGs omit viewBox; infer from width/height so scaling works.

  const widthMatch = svgTag.match(/width=["']?([\d.]+)(px)?["']?/i)
  const heightMatch = svgTag.match(/height=["']?([\d.]+)(px)?["']?/i)
  if (!widthMatch || !heightMatch) return source
  if (/%/.test(widthMatch[0]) || /%/.test(heightMatch[0])) return source

  const width = Number.parseFloat(widthMatch[1])
  const height = Number.parseFloat(heightMatch[1])
  if (!Number.isFinite(width) || !Number.isFinite(height)) return source

  const withViewBox = svgTag.replace('<svg', `<svg viewBox="0 0 ${width} ${height}"`)
  return source.replace(svgTag, withViewBox)
}
