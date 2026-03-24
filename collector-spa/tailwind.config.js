/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				brand: {
					primary: 'var(--color-primary)',
					secondary: 'var(--color-secondary)',
					ink: 'var(--color-ink)',
					muted: 'var(--color-ink-muted)',
					surface: 'var(--surface-card)',
					base: 'var(--surface-base)',
					line: 'var(--color-border)'
				}
			}
		}
	},
	plugins: []
};
