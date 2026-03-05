/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{vue,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        border: 'hsl(215 28% 17%)',
        background: 'hsl(222 47% 11%)',
        foreground: 'hsl(210 20% 98%)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}
