# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-bookworm AS backend
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    libwebp-dev libjxl-dev libjpeg62-turbo-dev libheif-dev \
    && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist web/dist
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN CGO_ENABLED=1 go build \
    -ldflags "-X github.com/DanielTso/pixshift/internal/version.Version=${VERSION} \
              -X github.com/DanielTso/pixshift/internal/version.Commit=${COMMIT} \
              -X github.com/DanielTso/pixshift/internal/version.Date=${DATE}" \
    -o pixshift ./cmd/pixshift

# Stage 3: Runtime
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libwebp7 libjxl0.7 libjpeg62-turbo libheif1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=backend /app/pixshift /usr/local/bin/pixshift
EXPOSE 8080
ENTRYPOINT ["pixshift"]
CMD ["serve", ":8080"]
