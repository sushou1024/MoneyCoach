const { getDefaultConfig } = require('expo/metro-config')
const { resolve } = require('metro-resolver')

const config = getDefaultConfig(__dirname)

config.resolver.sourceExts.push('cjs')
const defaultResolveRequest = config.resolver.resolveRequest

const mapZustandToCjs = (moduleName) => {
  if (moduleName === 'zustand') {
    return 'zustand/index.js'
  }
  if (moduleName.startsWith('zustand/')) {
    if (moduleName.endsWith('.js')) {
      return moduleName
    }
    return `${moduleName}.js`
  }
  return null
}

config.resolver.resolveRequest = (context, moduleName, platform) => {
  const mapped = mapZustandToCjs(moduleName)
  if (mapped) {
    return resolve(context, mapped, platform)
  }
  if (defaultResolveRequest) {
    return defaultResolveRequest(context, moduleName, platform)
  }
  return resolve(context, moduleName, platform)
}

module.exports = config
