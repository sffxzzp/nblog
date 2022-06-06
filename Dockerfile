FROM alpine:edge as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN apk add --no-cache bash git go gcc musl-dev
RUN go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
RUN go build -o ./bin/nblog -ldflags="-s -w"

ARG PORT
FROM alpine:edge
WORKDIR /opt/nblog
COPY . .
COPY --from=builder /app/bin/nblog ./
EXPOSE ${PORT}
CMD [ "./nblog" ]