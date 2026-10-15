/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#f0f7f2",
          100: "#dcefe3",
          700: "#166534",
          800: "#14532d",
          900: "#052e16",
        },
      },
    },
  },
  plugins: [],
};
