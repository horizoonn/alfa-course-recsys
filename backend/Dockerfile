FROM golang:1.27.0-bookworm AS build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY api/ ../api/
ARG OGEN_VERSION=v1.24.0
RUN go run github.com/ogen-go/ogen/cmd/ogen@${OGEN_VERSION} \
    --clean --config ../api/ogen.yaml \
    --target internal/transport/http/generated --package openapi ../api/openapi.yaml
COPY backend/ ./
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -o /app ./cmd/app
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -o /catalog-export ./cmd/catalog-export

FROM alpine:3.22 AS runtime
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 1000:1000
WORKDIR /app

FROM runtime AS catalog-export
COPY --from=build /catalog-export /catalog-export
COPY --from=build /src/backend/prepare-catalog.sh /prepare-catalog.sh
ENTRYPOINT ["/bin/sh", "/prepare-catalog.sh"]

FROM runtime AS app
COPY --from=build /app /usr/local/bin/app
EXPOSE 8080
ENTRYPOINT ["app"]
