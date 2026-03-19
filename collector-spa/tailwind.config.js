/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				brand: {
					// On utilise la syntaxe var() pour lier au CSS
					primary: 'var(--color-primary)',
					secondary: 'var(--color-secondary)',
					accent: 'var(--color-accent)'
				}
			}
		}
	},
	plugins: []
};
