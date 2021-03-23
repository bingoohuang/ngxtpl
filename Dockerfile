FROM golang:1.16.2-alpine3.12
# workspace directory
WORKDIR /app
# copy `go.mod` and `go.sum`
ADD go.mod go.sum ./
# install dependencies
RUN go mod download -x

# copy source code
COPY . .
RUN go install ./...

FROM openresty/openresty:1.19.3.1-2-alpine
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata  && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
COPY --from=0 /go/bin/ngxtpl /app/ngxtpl
CMD ["sh", "-c", "/usr/local/openresty/bin/openresty; /app/ngxtpl install -c /etc/app/docker.hcl; /app/ngxtpl start"]
