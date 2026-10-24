# Copyright (c) 2026 Jody Scott
# SPDX-License-Identifier: Apache-2.0
#
# Release image: the pre-built static tenantecho binary on a minimal base. The binary
# is built by the fleet's image pipeline against the fleet's merged protos.
ARG BASE_IMAGE=alpine:3.22
FROM ${BASE_IMAGE}
ARG VERSION=0.0.0
ARG VCS_REF=local
LABEL org.opencontainers.image.title="tenantecho" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/tacplus-dev/tenantecho"
RUN apk add --no-cache ca-certificates && adduser -D -u 65532 app
COPY bin/tenantecho /usr/local/bin/tenantecho
USER 65532
ENTRYPOINT ["tenantecho"]
