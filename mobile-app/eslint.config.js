const { FlatCompat } = require('@eslint/eslintrc')
const nPlugin = require('eslint-plugin-n')

const compat = new FlatCompat({
  baseDirectory: __dirname,
  resolvePluginsRelativeTo: __dirname,
})

module.exports = [
  {
    ignores: ['node_modules/**', 'dist/**', 'build/**', '.expo/**', 'android/**', 'ios/**'],
  },
  ...compat.extends('universe/native'),
  {
    plugins: {
      n: nPlugin,
    },
    rules: {
      'node/handle-callback-err': 'off',
      'node/no-new-require': 'off',
      'n/handle-callback-err': ['warn', '^(e|err|error|.+Error)$'],
      'n/no-new-require': 'warn',
    },
  },
  {
    files: ['eslint.config.js', 'metro.config.js', 'babel.config.js', 'app.config.js'],
    languageOptions: {
      globals: {
        __dirname: 'readonly',
        module: 'readonly',
        require: 'readonly',
      },
      sourceType: 'commonjs',
    },
  },
  {
    files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx'],
    languageOptions: {
      globals: {
        jest: 'readonly',
      },
    },
  },
]
