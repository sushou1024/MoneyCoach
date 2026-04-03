import Constants, { ExecutionEnvironment } from 'expo-constants'

export function isExpoGo() {
  const ownership = (Constants.appOwnership ?? '') as string
  return (
    Constants.executionEnvironment === ExecutionEnvironment.StoreClient || ownership === 'expo' || ownership === 'guest'
  )
}
