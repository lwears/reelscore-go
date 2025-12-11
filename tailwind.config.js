/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/handlers/templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        reelscore: {
          primary: "#6366f1", // Indigo-500
          "primary-content": "#ffffff",
          secondary: "#d946ef", // Fuchsia-500
          "secondary-content": "#ffffff",
          accent: "#f43f5e", // Rose-500
          neutral: "#1e293b", // Slate-800
          "base-100": "#f8fafc", // Slate-50
          "base-200": "#f1f5f9", // Slate-100
          "base-300": "#e2e8f0", // Slate-200
          info: "#3b82f6",
          success: "#10b981",
          warning: "#f59e0b",
          error: "#ef4444",
        },
      },
      "dark",
    ],
  },
};
