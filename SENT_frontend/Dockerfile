
FROM node:20-bullseye as build
WORKDIR /app

# Khai báo biến môi trường ngay từ đầu
ENV CI=false
ENV GENERATE_SOURCEMAP=false
ENV NODE_OPTIONS="--max-old-space-size=4096"

COPY package*.json ./
RUN npm install --legacy-peer-deps

COPY . .

# Lệnh build chính thức
RUN npm run build -- --no-minify

# Giai đoạn 2: Nginx (Chỉ chạy khi Stage 1 thành công)
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy tệp tĩnh từ thư mục build của bạn
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]