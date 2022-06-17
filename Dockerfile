FROM golang as builder
ENV GOPROXY=goproxy.cn
WORKDIR /src
COPY . /src
RUN go build -o tencent-cdn-cert-manager

FROM ubuntu
RUN apt update
RUN apt install ca-certificates -y
RUN update-ca-certificates
COPY --from=builder /src/tencent-cdn-cert-manager /usr/local/bin/tencent-cdn-cert-manager
WORKDIR /app
ENTRYPOINT /usr/local/bin/tencent-cdn-cert-manager
