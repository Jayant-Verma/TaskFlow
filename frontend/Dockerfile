# ==========================================
# Stage 1: Build the React/Vite application
# ==========================================
FROM node:20-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy package files first to leverage Docker layer caching
COPY package*.json ./

# Install dependencies 
# (Using --legacy-peer-deps since we are forcing Chakra on React 19)
RUN npm ci --legacy-peer-deps

# Copy the rest of the frontend source code
COPY . .

# Build the project for production (outputs to /app/dist)
RUN npm run build

# ==========================================
# Stage 2: Serve the app with Nginx
# ==========================================
FROM nginx:alpine AS production

# Remove the default Nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Copy the built static files from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy our custom Nginx config to handle React Router
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 to the outside world
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]