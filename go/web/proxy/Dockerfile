FROM saichler/builder:latest AS build

COPY main /home/src/github.com/saichler/build/main
COPY reverse_proxy.go /home/src/github.com/saichler/build/reverse_proxy.go

RUN go mod init
RUN GOPROXY=direct GOPRIVATE=github.com go mod tidy
RUN go build -o reverse-proxy ./main/main.go

FROM alpine AS final
WORKDIR /home/run
COPY --from=build /home/src/github.com/saichler/build/reverse-proxy /home/run/reverse-proxy
COPY layer8vibe.dev /home/run/layer8vibe.dev
COPY probler.dev /home/run/probler.dev

ENTRYPOINT ["/home/run/reverse-proxy"]