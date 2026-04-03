import { TranslationKey } from './i18n'

type Translator = (key: TranslationKey, params?: Record<string, string | number>) => string

export function getStrategyDisplayName(t: Translator, strategyId?: string | null) {
  if (!strategyId) return t('strategy.name.unknown')
  const key = `strategy.name.${strategyId}`
  const translated = t(key as TranslationKey)
  if (translated === key) {
    return t('strategy.name.unknown')
  }
  return translated
}
