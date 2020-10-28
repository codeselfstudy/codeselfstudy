FROM node:14-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY ./package*.json ./
COPY . .
RUN npm install
CMD ["npm", "run", "production"]
