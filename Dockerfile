FROM golang as builder

ARG SIMPLE
ARG SPECIAL
ARG SPACES

RUN echo "$SIMPLE"
RUN echo "$SPECIAL"
RUN echo "$SPACES"

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64 \
  go build -tags netgo -ldflags '-s -w' -o app github.com/elauffenburger/tcp-echo

FROM alpine:3.10.1
WORKDIR /root/
COPY --from=builder /src .
ENTRYPOINT ["./app"]
