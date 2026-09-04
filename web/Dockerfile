FROM node:26-alpine

RUN npm install -g pnpm@12

COPY package.json .
COPY pnpm-lock.yaml .

RUN pnpm install --frozen-lockfile

CMD [ "pnpm", "run", "dev" ]
