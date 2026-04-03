import fs from 'node:fs'

const gradlePath = 'android/app/build.gradle'

if (!fs.existsSync(gradlePath)) {
  console.error(`Missing ${gradlePath}. Run expo prebuild --platform android first.`)
  process.exit(1)
}

let source = fs.readFileSync(gradlePath, 'utf8')

const helperNeedle = 'def projectRoot = rootDir.getAbsoluteFile().getParentFile().getAbsolutePath()\n'
const helperBlock = `def envOrProperty = { String name ->
    def value = findProperty(name)
    if (value != null && value.toString().trim()) {
        return value.toString().trim()
    }
    def envValue = System.getenv(name)
    if (envValue != null && envValue.trim()) {
        return envValue.trim()
    }
    return null
}
def releaseStoreFileName = envOrProperty("MONEYCOACH_UPLOAD_STORE_FILE")
def releaseStorePassword = envOrProperty("MONEYCOACH_UPLOAD_STORE_PASSWORD")
def releaseKeyAlias = envOrProperty("MONEYCOACH_UPLOAD_KEY_ALIAS")
def releaseKeyPassword = envOrProperty("MONEYCOACH_UPLOAD_KEY_PASSWORD")
def releaseSigningConfigured = [releaseStoreFileName, releaseStorePassword, releaseKeyAlias, releaseKeyPassword].every { it != null }
def releaseTaskRequested = gradle.startParameter.taskNames.any { it.toLowerCase().contains("release") }
def resolvedVersionCode = envOrProperty("MONEYCOACH_VERSION_CODE")
def resolvedVersionName = envOrProperty("MONEYCOACH_VERSION_NAME")
`

if (!source.includes('def envOrProperty = { String name ->')) {
  if (!source.includes(helperNeedle)) {
    console.error('Unable to locate projectRoot definition in android/app/build.gradle')
    process.exit(1)
  }
  source = source.replace(helperNeedle, `${helperNeedle}${helperBlock}`)
}

const replaceOrFail = (from, to, label) => {
  if (source.includes(to)) {
    return
  }
  if (!source.includes(from)) {
    console.error(`Unable to patch ${label}. Template drift detected in android/app/build.gradle.`)
    process.exit(1)
  }
  source = source.replace(from, to)
}

replaceOrFail(
  '        versionCode 1\n',
  '        versionCode resolvedVersionCode != null ? resolvedVersionCode.toInteger() : 1\n',
  'versionCode'
)
replaceOrFail(
  '        versionName "0.1.0"\n',
  '        versionName resolvedVersionName != null ? resolvedVersionName : "0.1.0"\n',
  'versionName'
)
replaceOrFail(
  `        debug {
            storeFile file('debug.keystore')
            storePassword 'android'
            keyAlias 'androiddebugkey'
            keyPassword 'android'
        }
`,
  `        debug {
            storeFile file('debug.keystore')
            storePassword 'android'
            keyAlias 'androiddebugkey'
            keyPassword 'android'
        }
        release {
            if (releaseSigningConfigured) {
                storeFile file(releaseStoreFileName)
                storePassword releaseStorePassword
                keyAlias releaseKeyAlias
                keyPassword releaseKeyPassword
            }
        }
`,
  'release signingConfigs block'
)
replaceOrFail(
  `        release {
            // Caution! In production, you need to generate your own keystore file.
            // see https://reactnative.dev/docs/signed-apk-android.
            signingConfig signingConfigs.debug
`,
  `        release {
            if (releaseTaskRequested && !releaseSigningConfigured) {
                throw new GradleException("Missing release signing configuration. Set MONEYCOACH_UPLOAD_STORE_FILE, MONEYCOACH_UPLOAD_STORE_PASSWORD, MONEYCOACH_UPLOAD_KEY_ALIAS, and MONEYCOACH_UPLOAD_KEY_PASSWORD.")
            }
            signingConfig releaseSigningConfigured ? signingConfigs.release : signingConfigs.debug
`,
  'release buildType signing'
)

fs.writeFileSync(gradlePath, source)
console.log(`Patched ${gradlePath} for release signing and version overrides.`)
