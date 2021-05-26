module.exports = {
  purge: [
    './index.html',
    './public/**/*.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {
      colors: {
        cadetBlue: '#4CA1AF',
        blueGrey: '#C4E0E5',
      },
      gridTemplateColumns: {
        '2-1-1': 'minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr)',
      },
    },
  },
  variants: {
    extend: {
      backgroundColor: ['active'],
      ringWidth: ['active'],
      ringColor: ['active'],
      ringOpacity: ['active'],
    }
  },
  plugins: [],
};
