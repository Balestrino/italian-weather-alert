FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY api/ api/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /iwa ./cmd/iwa \
 && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /iwa-frontend ./cmd/iwa-frontend \
 && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /iwa-inference-probe ./cmd/iwa-inference-probe

FROM alpine:3.22
ARG VCS_REF=unknown
LABEL org.opencontainers.image.revision=$VCS_REF
LABEL org.opencontainers.image.source="https://github.com/Balestrino/italian-weather-alert"
RUN apk add --no-cache ca-certificates poppler-utils postgresql16-client
COPY --from=build /iwa /iwa
COPY --from=build /iwa-frontend /iwa-frontend
COPY --from=build /iwa-inference-probe /iwa-inference-probe
USER 65532:65532
ENTRYPOINT ["/iwa"]
