
FROM golang:1.17 as builder
# Create a directory for the app
RUN apt-get update -y && apt-get install -y apt-transport-https
RUN  apt install -y openjdk-8-jdk
RUN apt-get install -y  elasticsearch
RUN mkdir /app
 
# Copy all files from the current directory to the app directory
 
# Set working directory
WORKDIR /app

COPY / /app/
 
# Run command as described:
# go build will build an executable file named server in the current directory
RUN make build
# Run the server
# EXPOSE 3000
# CMD [ "/app/server" ]

FROM debian:11
WORKDIR /
COPY --from=builder /app/Service-Binding-ElasticSearch .
EXPOSE 3000
ENTRYPOINT ["/Service-Binding-ElasticSearch"]