FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/app /bin/app
EXPOSE 8080
ENTRYPOINT ["/bin/app"]
