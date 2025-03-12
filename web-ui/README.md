# Web UI

Adds a React + TypeScript + shadcn/ui web application.

## Get Started
```bash
yarn # (install) dependencies
yarn dev # run dev server at http://localhost:5173
# yarn build # build files
# yarn preview # run a "production" preview of built files at http://localhost:4173
# yarn lint # run eslint and fix any obvious issues
```

## Stuff to Know

**Notable packages in use**

- `React` - the baseline UI system
- `TypeScript` - you know, for types
- `Vite` - for building, bundling, and dev server-ing
- `React Router` - for client-side routing (debatably needed)
    - `Generouted` - file system routing in `/pages`, so routes are determined by filenames/dirs
- `shadcn/ui` - the UI library for look and feel
    - `Radix` - the headless UI library underneath shadcn/ui
- `Tailwind CSS` - for easier inline styling compared to raw CSS
- `Zod` - to make run-time schemas for parsing data and generating TS types
