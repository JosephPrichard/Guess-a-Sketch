FROM node:18.18 as client
WORKDIR /usr/app
COPY ./client . 
RUN npm install
RUN npm run build

FROM golang:1.21.1
WORKDIR /usr/app
COPY --from=client /usr/app/dist /usr/app/dist
COPY ./server . 
RUN CGO_ENABLED=1 go build guessthesketch
EXPOSE 8080
CMD ["./guessthesketch"]