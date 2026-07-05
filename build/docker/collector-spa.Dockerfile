FROM node:22-alpine AS build

WORKDIR /app

COPY collector-spa/package.json collector-spa/package-lock.json ./
RUN npm ci

COPY collector-spa/ ./
RUN npm run build && npm prune --omit=dev

FROM node:22-alpine AS runtime

RUN rm -rf \
    /usr/local/lib/node_modules/npm \
    /usr/local/lib/node_modules/corepack \
    /opt/yarn-v1.22.22 \
    && rm -f \
    /usr/local/bin/npm \
    /usr/local/bin/npx \
    /usr/local/bin/corepack \
    /usr/local/bin/yarn \
    /usr/local/bin/yarnpkg

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

COPY --from=build /app/build ./build
COPY --from=build /app/node_modules ./node_modules
COPY collector-spa/package.json ./package.json

USER 1000:1000

EXPOSE 3000

CMD ["node", "build"]
