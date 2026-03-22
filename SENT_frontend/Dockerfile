FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
# Biến này quan trọng để React biết gọi đi đâu
ENV REACT_APP_API_URL=http://localhost:8000/api/v1
EXPOSE 3000
CMD ["npm", "start"]