/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: '#10A37F',
                dark: '#343541',
                surface: '#444654',
            }
        },
    },
    plugins: [],
}
