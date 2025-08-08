/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Blueprint 특화 색상
        success: {
          DEFAULT: "#22c55e",
          foreground: "#ffffff",
        },
        warning: {
          DEFAULT: "#f59e0b",
          foreground: "#ffffff",
        },
        prediction: {
          high: "#22c55e",
          medium: "#f59e0b",
          low: "#ef4444",
        }
      },
    },
  },
  plugins: [],
}
