FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY docs/prerequisiti-mvp/embed.go docs/prerequisiti-mvp/embed.go
COPY docs/prerequisiti-mvp/mcp-tools.json docs/prerequisiti-mvp/mcp-tools.json
COPY docs/prerequisiti-mvp/openapi.json docs/prerequisiti-mvp/openapi.json
COPY docs/prerequisiti-mvp/contratto.schema.json docs/prerequisiti-mvp/contratto.schema.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /iwa ./cmd/iwa \
 && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /iwa-inference-probe ./cmd/iwa-inference-probe

FROM alpine:3.22
ARG VCS_REF=unknown
LABEL org.opencontainers.image.revision=$VCS_REF
RUN apk add --no-cache ca-certificates poppler-utils postgresql16-client
COPY --from=build /iwa /iwa
COPY --from=build /iwa-inference-probe /iwa-inference-probe
USER 65532:65532
ENTRYPOINT ["/iwa"]
