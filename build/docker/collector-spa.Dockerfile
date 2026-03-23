FROM node:22-alpine AS build

WORKDIR /app

COPY collector-spa/package.json collector-spa/package-lock.json ./
RUN npm ci

COPY collector-spa/ ./
RUN npm run build && npm prune --omit=dev

FROM node:22-alpine AS runtime

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

COPY --from=build /app/build ./build
COPY --from=build /app/node_modules ./node_modules
COPY collector-spa/package.json ./package.json

USER node

EXPOSE 3000

CMD ["node", "build"]
