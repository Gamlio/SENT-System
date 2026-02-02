/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}", // Quét tất cả file .js, .jsx trong thư mục src
    "./public/index.html",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#0f172a", // Màu nền SENT
      },
    },
  },
  plugins: [],
}