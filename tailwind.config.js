/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/handlers/templates/**/*.html",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('daisyui'),
  ],
  daisyui: {
    themes: [
      {
        reelscore: {
          "primary": "#667eea",
          "primary-content": "#ffffff",
          "secondary": "#764ba2",
          "accent": "#f59e0b",
          "neutral": "#333333",
          "base-100": "#ffffff",
          "base-200": "#f3f4f6",
          "base-300": "#e5e7eb",
          "info": "#3b82f6",
          "success": "#10b981",
          "warning": "#f59e0b",
          "error": "#ef4444",
        },
      },
      "dark",
    ],
  },
}
