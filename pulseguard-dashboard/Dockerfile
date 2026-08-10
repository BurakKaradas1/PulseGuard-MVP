# ---- Build stage ----
FROM node:20-bookworm AS builder
WORKDIR /app

# IMPORTANT: Vite bakes VITE_* variables into the JS bundle at BUILD
# time, not at container start time. This value is used by the
# browser (which runs OUTSIDE the Docker network), so it must be an
# address the user's browser can actually reach - e.g.
# http://localhost:8080 - NOT an internal Docker service name like
# http://pulseguard-collector:8080 (the browser cannot resolve that).
ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# ---- Runtime stage ----
FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
