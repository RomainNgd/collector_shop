# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project
npx sv create my-app
```

To recreate this project with the same configuration:

```sh
# recreate this project
npx sv create --template demo --types ts --add prettier eslint --install npm collector-spa
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## Runtime configuration

The SvelteKit server expects:

- `API_BASE_URL`: internal API URL used by SSR requests.
- `API_PUBLIC_BASE_URL`: public API URL used to build browser-visible image URLs.
- `JWT_SECRET`: same value as `go-api`, used to verify JWT cookies before populating `locals.user`.

Generated coverage reports are ignored by Prettier via `.prettierignore`; lint should cover source files, not generated HTML reports.
