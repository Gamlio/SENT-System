# Stage 1: Build
FROM node:20-bullseye as build
WORKDIR /app
COPY package*.json ./
RUN rm -f package-lock.json && npm cache clean --force && npm install --legacy-peer-deps
COPY . .

ENV VITE_API_URL=/api/v1
ENV VITE_WS_URL=ws://192.168.2.1:8000/ws
RUN npm run build

# Stage 2: Nginx
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]