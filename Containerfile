FROM golang:1.23-alpine AS go-build
WORKDIR /NeighBot
COPY . /NeighBot
RUN go build -o NeighBot

FROM golang:1.23-alpine
COPY --from=go-build /NeighBot/NeighBot /NeighBot/NeighBot
WORKDIR /NeighBot

ENV CONFIG_DIR=""

ENTRYPOINT ["/NeighBot/NeighBot"]
CMD "--config-dir=${CONFIG_DIR}"
