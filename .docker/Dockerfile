FROM golang:1.22.3 AS builder

WORKDIR /src/app
ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY main.go ./main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /app

FROM node:18 AS static

WORKDIR /src/app
COPY package.json ./
COPY static ./static
COPY templates ./templates
COPY ple.json ./
COPY gulpfile.js ./

RUN npm install
RUN ./node_modules/.bin/gulp

FROM alpine:3.15
WORKDIR /src/app
COPY --from=builder /src/app/ . 
COPY --from=static /src/app/dist ./dist
COPY --from=static /src/app/templates ./templates

EXPOSE 8080

CMD ["/app"]
