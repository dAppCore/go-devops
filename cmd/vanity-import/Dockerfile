FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN go build -trimpath -ldflags="-w -s" -o /vanity-import .

FROM alpine:3.21
RUN adduser -D -h /home/app app
COPY --from=build /vanity-import /vanity-import
USER app
EXPOSE 8080
ENTRYPOINT ["/vanity-import"]
