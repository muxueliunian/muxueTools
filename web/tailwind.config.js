/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    darkMode: 'class',
    theme: {
        extend: {
            colors: {
                claude: {
                    bg: '#FAF8F5',
                    sidebar: '#F0EEEB',
                    border: '#E1DFDD',
                    text: '#1F1E1D',
                    secondaryText: '#6F6F78',
                    hover: '#E6E4E1',
                    dark: {
                        bg: '#191919',
                        sidebar: '#111111',
                        border: '#2A2A2E',
                        text: '#FFFFFF',
                        secondaryText: '#9CA3AF',
                        hover: '#212124'
                    }
                }
            },
            fontFamily: {
                serif: ['"Merriweather"', 'serif'],
                sans: ['"Inter"', 'sans-serif'],
            }
        },
    },
    plugins: [],
}
