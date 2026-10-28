// For more info, see https://github.com/storybookjs/eslint-plugin-storybook#configuration-flat-config-format
import storybook from 'eslint-plugin-storybook'

//  @ts-check
import { tanstackConfig } from '@tanstack/eslint-config'
import reactPlugin from 'eslint-plugin-react'
import prettierRecommended from 'eslint-plugin-prettier/recommended'
import simpleImportSort from 'eslint-plugin-simple-import-sort'
import pluginQuery from '@tanstack/eslint-plugin-query'
import tseslint from 'typescript-eslint'

export default [
  ...tanstackConfig,
  ...tseslint.configs.recommended,
  {
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.app.json', './tsconfig.node.json'],
      },
    },
    settings: {
      react: {
        version: '19.0',
      },
    },
    plugins: {
      react: reactPlugin,
      'simple-import-sort': simpleImportSort,
      '@tanstack/query': pluginQuery,
    },
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'warn',
        {
          args: 'all',
          argsIgnorePattern: '^_',
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          destructuredArrayIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          ignoreRestSiblings: true,
        },
      ],

      'import/no-cycle': 'off',
      'import/order': 'off',
      'sort-imports': 'off',
      'pnpm/json-enforce-catalog': 'off',
      'react/function-component-definition': [
        'error',
        {
          namedComponents: 'arrow-function',
          unnamedComponents: 'arrow-function',
        },
      ],
      'implicit-arrow-linebreak': 'off',
      'prettier/prettier': 'error',
      'linebreak-style': 'off',

      // --- Import Sorting ---
      'import/order': 'off',
      'sort-imports': 'off',
      'simple-import-sort/exports': 'error',
      'simple-import-sort/imports': [
        'error',
        {
          groups: [
            ['^\\u0000'],
            ['^react'],
            ['^@tanstack'],
            ['^@?\\w', '^react-', '^react-icons'],
            ['^#/'],
          ],
        },
      ],

      // --- Prettier Overrides ---
      'prettier/prettier': [
        'error',
        {
          endOfLine: 'auto',
        },
      ],
    },
  },
  {
    files: ['src/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              regex: '^#themes/',
              message:
                'Application code must import only from the public #themes package entry point.',
            },
            {
              regex: '^(?:\\.\\.?/)+themes(?:/|$)',
              message:
                'Application code must import only from the public #themes package entry point.',
            },
          ],
        },
      ],
    },
  },
  {
    ignores: [
      'eslint.config.js',
      'prettier.config.js',
      'node_modules/',
      'dist/',
      'build/',
      'coverage/',
      '*.min.js',
    ],
  },
  ...storybook.configs['flat/recommended'],

  prettierRecommended,
]
