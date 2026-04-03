const { withAndroidManifest } = require('@expo/config-plugins')

const BLOCKED_PERMISSIONS = [
  'android.permission.READ_EXTERNAL_STORAGE',
  'android.permission.WRITE_EXTERNAL_STORAGE',
  'android.permission.READ_MEDIA_IMAGES',
  'android.permission.READ_MEDIA_VIDEO',
  'android.permission.READ_MEDIA_VISUAL_USER_SELECTED',
  'android.permission.READ_MEDIA_AUDIO',
  'android.permission.RECORD_AUDIO',
  'android.permission.SYSTEM_ALERT_WINDOW',
]

module.exports = function withAndroidPlayPolicy(config) {
  return withAndroidManifest(config, (config) => {
    const manifest = config.modResults.manifest
    manifest.$ = manifest.$ || {}
    if (!manifest.$['xmlns:tools']) {
      manifest.$['xmlns:tools'] = 'http://schemas.android.com/tools'
    }

    const existing = manifest['uses-permission'] || []
    const byName = new Map(
      existing
        .map((entry) => entry?.$?.['android:name'])
        .filter(Boolean)
        .map((name, index) => [name, index])
    )

    for (const permission of BLOCKED_PERMISSIONS) {
      const index = byName.get(permission)
      if (index != null) {
        const current = existing[index].$
        existing[index] = {
          $: {
            ...current,
            'android:name': permission,
            'tools:node': 'remove',
          },
        }
        continue
      }
      existing.push({
        $: {
          'android:name': permission,
          'tools:node': 'remove',
        },
      })
    }

    manifest['uses-permission'] = existing

    const applications = manifest.application || []
    if (applications.length > 0 && applications[0]?.$?.['android:requestLegacyExternalStorage'] != null) {
      delete applications[0].$['android:requestLegacyExternalStorage']
    }

    return config
  })
}
