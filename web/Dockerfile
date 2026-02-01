# Stage 1: Build
FROM node:18-alpine as builder

WORKDIR /app

# Install dependencies (use cache if package.json hasn't changed)
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code
COPY . .

# Build for production
RUN npm run build

# Stage 2: Serve
FROM nginx:alpine

# Copy build artifacts
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
