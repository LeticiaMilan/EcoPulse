FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY src ./src
ENV PORT=8090
EXPOSE 8090
CMD ["npx", "tsx", "src/index.ts"]