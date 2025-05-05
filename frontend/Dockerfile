# Stage 1: Build the frontend (if needed)
# Uncomment if you need to build your frontend with npm/yarn
# FROM node:16-alpine as build
# WORKDIR /app
# COPY ./frontend/package*.json ./
# RUN npm install
# COPY ./frontend ./
# RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:alpine

# Copy static files to nginx html directory
COPY . /usr/share/nginx/html

# Copy custom nginx configuration if needed
# COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]