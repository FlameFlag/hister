import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss';
import js from '@eslint/js';
import * as typescriptEslint from 'typescript-eslint';
import svelteEslintParser from 'svelte-eslint-parser';
import globals from 'globals';

export default [
  js.configs.recommended,
  ...typescriptEslint.configs.recommended,
  {
    files: ['**/*.{js,jsx,cjs,mjs,ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2024,
      sourceType: 'module',
      parser: typescriptEslint.parser,
      parserOptions: {
        projectService: {
          allowDefaultProject: ['eslint.config.js', 'svelte.config.js', 'vite.config.ts'],
        },
        extraFileExtensions: ['.svelte'],
      },
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ['**/*.svelte'],
    languageOptions: {
      ecmaVersion: 2024,
      sourceType: 'module',
      parser: svelteEslintParser,
      parserOptions: {
        parser: typescriptEslint.parser,
      },
      globals: {
        ...globals.browser,
        ...globals.node,
        chrome: 'readonly',
      },
    },
    rules: {
      // Allow any type for assignments when TypeScript cannot infer types
      // Useful in scenarios with dynamic typing or when working with external libraries
      '@typescript-eslint/no-unsafe-assignment': 'off',

      // Allow accessing properties on any type
      // Common when working with dynamic data structures or libraries with loose typing
      '@typescript-eslint/no-unsafe-member-access': 'off',

      // Allow calling any type as a function
      // Useful when working with callbacks or dynamic function invocations
      '@typescript-eslint/no-unsafe-call': 'off',

      // Allow returning any type from functions
      // Common when working with APIs that return untyped data or when implementing generic handlers
      '@typescript-eslint/no-unsafe-return': 'off',

      // Allow passing any type as function arguments
      // Useful for generic utility functions or when dealing with external APIs
      '@typescript-eslint/no-unsafe-argument': 'off',

      // Disabled because UI library (bits-ui) uses custom data attribute classes
      // like data-state-open:animate-in that aren't in Tailwind's default registry
      'better-tailwindcss/no-unknown-classes': 'off',
    },
  },

  // Tailwind CSS rules for all files
  {
    files: ['**/*.{ts,tsx,js,jsx,svelte}'],
    plugins: {
      'better-tailwindcss': eslintPluginBetterTailwindcss,
    },
    settings: {
      'better-tailwindcss': {
        entryPoint: 'src/app.css',
        detectComponentClasses: true,
        rootFontSize: 16,
      },
    },
    rules: {
      // Enforce canonical Tailwind classes for better maintainability
      // Converts arbitrary values like mt-[16px] to canonical classes like mt-4
      // Automatically collapses utilities (e.g., mt-2 mr-2 mb-2 ml-2 -> m-2)
      // Uses logical properties where appropriate (e.g., mr-2 + ml-2 -> mx-2)
      'better-tailwindcss/enforce-canonical-classes': [
        'error',
        {
          collapse: true,
          logical: true,
        },
      ],

      // Warn when class strings exceed reasonable line length
      // Automatically wraps long class strings across multiple lines for readability
      'better-tailwindcss/enforce-consistent-line-wrapping': 'warn',

      // Sort Tailwind classes in a consistent order
      // Makes classes easier to scan and reduces merge conflicts
      'better-tailwindcss/enforce-consistent-class-order': 'warn',

      // Remove duplicate classes to reduce bloat
      // Example: "px-4 px-4" becomes "px-4"
      'better-tailwindcss/no-duplicate-classes': 'warn',

      // Warn about deprecated Tailwind classes
      // Helps keep codebase updated with latest Tailwind CSS best practices
      'better-tailwindcss/no-deprecated-classes': 'warn',

      // Remove unnecessary whitespace in class strings
      // Example: "px-4  py-2" becomes "px-4 py-2"
      'better-tailwindcss/no-unnecessary-whitespace': 'warn',

      // Error on classes that produce conflicting styles
      // Example: "flex grid" (can't be both flex and grid)
      'better-tailwindcss/no-conflicting-classes': 'error',

      // Disabled: Covered by enforce-canonical-classes
      // canonical classes already handles shorthand conversions
      'better-tailwindcss/enforce-shorthand-classes': 'off',

      // Enforce consistent position of important modifier (!)
      // Converts !shadow-none to shadow-none! for consistency
      'better-tailwindcss/enforce-consistent-important-position': 'warn',

      // Disabled: Covered by enforce-canonical-classes
      // canonical classes already handles variable syntax
      'better-tailwindcss/enforce-consistent-variable-syntax': 'off',
    },
  },
  {
    ignores: ['.svelte-kit/', 'build/', 'dist/', 'src/lib/components/ui/'],
  },
];
