
FROM node:20-bullseye as build
ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=$VITE_API_URL
WORKDIR /app
COPY package*.json ./

RUN rm -f package-lock.json && npm cache clean --force && npm install --legacy-peer-deps
COPY . .

RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]