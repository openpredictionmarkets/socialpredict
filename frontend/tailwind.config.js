/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'green-btn': {
          DEFAULT: '#00bf9a', // Use DEFAULT for the base color
          hover: '#00cca4',
          'border-default': '#00f2c3',
          'border-hover': '#00cca4',
        },
        'red-btn': {
          DEFAULT: '#fd5d93',
          hover: '#ec250d',
          'border-default': '#fd5d93',
          'border-hover': '#ec250d',
        },
        'neutral-btn': {
          DEFAULT: '#ba54f5',
          hover: '#344675',
          active: '#263148',
        },
      },
    },
  },
  plugins: [],
};
