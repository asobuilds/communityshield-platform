/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        tactical: {
          darkest: '#0A0E17',   // Deep command room backdrop
          card: '#121824',      // Protected administrative panels
          border: '#1E293B',    // Steel dividing lines
          glow: '#38BDF8',      // Information stream accent
        }
      }
    },
  },
  plugins: [],
}
