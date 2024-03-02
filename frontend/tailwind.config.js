/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'primary-background': {
          DEFAULT: '#0e121d',
        },
        'custom-gray-light': {
          DEFAULT: '#708090',
        },
        'custom-gray-dark': {
          DEFAULT: '#303030',
        },
        'green-btn': {
          DEFAULT: '#00bf9a',
          hover: '#00cca4',
          'border-default': '#00f2c3',
          'border-hover': '#00cca4',
        },
        'red-btn': {
          DEFAULT: '#ec250d', // #dc3545
          hover: '#fd5d93', // #ff69b4
          'border-default': '#fd5d93',
          'border-hover': '#ec250d',
        },
        'neutral-btn': {
          DEFAULT: '#ba54f5',
          hover: '#344675',
          active: '#263148',
        },
        'primary-pink': {
          DEFAULT: '#ff69b4',
        },
        'info-blue': {
          DEFAULT: '#17a2b8',
        },
        'warning-orange': {
          DEFAULT: '#ffc107',
        },
      },
      spacing: {
        'sidebar': '8rem', // more rem means sidebar thicker
      },
      zIndex: {
        'sidebar': 40, // higher number means more on top
      },
    },
  },
  plugins: [],
};