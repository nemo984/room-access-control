FROM golang:1.22-alpine as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY *.go ./

RUN go build -mod=readonly -v -o server

# Use a gcloud image based on debian:buster-slim for a lean production container.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM gcr.io/google.com/cloudsdktool/cloud-sdk:slim

WORKDIR /app

COPY --from=builder /app/server /app/server

#COPY *.sh /app/
#RUN chmod +x /app/*.sh

CMD ["/app/server"]
