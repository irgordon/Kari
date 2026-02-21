// frontend/tailwind.config.ts
import type { Config } from 'tailwindcss';
import defaultTheme from 'tailwindcss/defaultTheme';

export default {
    content: ['./src/**/*.{html,js,svelte,ts}'],
    theme: {
        extend: {
            colors: {
                kari: {
                    teal: '#1BA8A0',       // Primary buttons, highlights, accents
                    'warm-gray': '#8E8F93', // Secondary UI, borders, disabled states
                    'light-gray': '#F4F5F6',// Backgrounds, cards, panels
                    text: '#1A1A1C',        // Headings, body text, navigation
                }
            },
            fontFamily: {
                // SF Pro is an Apple system font. We use a modern system font stack 
                // as the primary UI font, falling back to Inter or standard sans-serif.
                sans: [
                    '-apple-system', 
                    'BlinkMacSystemFont', 
                    '"SF Pro Text"', 
                    '"Segoe UI"', 
                    'Roboto', 
                    'sans-serif',
                    ...defaultTheme.fontFamily.sans
                ],
                // IBM Plex Sans for dense data tables, terminal outputs, and body text
                body: ['"IBM Plex Sans"', 'sans-serif'],
                // Optional: Keep a monospace font for the terminal component
                mono: ['"IBM Plex Mono"', 'monospace', ...defaultTheme.fontFamily.mono],
            }
        }
    },
    plugins: []
} satisfies Config;
