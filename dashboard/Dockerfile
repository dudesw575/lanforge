FROM node:24 AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Production stage - serves static Vue files directly with Caddy as reverse proxy
FROM nginx:latest

COPY --from=builder /app/dist /var/www/

COPY nginx.conf /etc/nginx/nginx.conf


# EXPOSE 80

# CMD ["nginx", "-g", "daemon off;"]

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 80

# Define the entrypoint to run our replacement script first
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# This CMD gets passed as arguments ($@) to the entrypoint script
CMD ["nginx", "-g", "daemon off;"]