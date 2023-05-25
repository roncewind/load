# -----------------------------------------------------------------------------
# Stages
# -----------------------------------------------------------------------------

ARG IMAGE_GO_BUILDER=golang:1.20.4
ARG IMAGE_FINAL=senzing/senzingapi-runtime:staging
# ARG IMAGE_FINAL=senzing/senzingapi-runtime:3.4.2

# -----------------------------------------------------------------------------
# Stage: go_builder
# -----------------------------------------------------------------------------

FROM ${IMAGE_FINAL} as senzing-runtime
FROM ${IMAGE_GO_BUILDER} as go_builder
ENV REFRESHED_AT 2023-02-28
LABEL Name="roncewind/load" \
      Maintainer="dad@lynntribe.net" \
      Version="0.0.0"

# Build arguments.

ARG PROGRAM_NAME="load"
ARG BUILD_VERSION=0.0.0
ARG BUILD_ITERATION=0
ARG GO_PACKAGE_NAME="github.com/roncewind/load"

# Copy local files from the Git repository.

# COPY ./rootfs /
COPY . ${GOPATH}/src/${GO_PACKAGE_NAME}

# Copy remote files from DockerHub to build

COPY --from=senzing-runtime  "/opt/senzing/g2/lib/"   "/opt/senzing/g2/lib/"
COPY --from=senzing-runtime  "/opt/senzing/g2/sdk/c/" "/opt/senzing/g2/sdk/c/"

# Set path to Senzing libs.

ENV LD_LIBRARY_PATH=/opt/senzing/g2/lib/

# Build go program.

WORKDIR ${GOPATH}/src/${GO_PACKAGE_NAME}
RUN make build

# --- Test go program ---------------------------------------------------------

# Run unit tests.

# RUN go get github.com/jstemmer/go-junit-report \
#  && mkdir -p /output/go-junit-report \
#  && go test -v ${GO_PACKAGE_NAME}/... | go-junit-report > /output/go-junit-report/test-report.xml

# Copy binaries to /output.

RUN mkdir -p /output \
      && cp -R ${GOPATH}/src/${GO_PACKAGE_NAME}/target/*  /output/

# -----------------------------------------------------------------------------
# Stage: final
# -----------------------------------------------------------------------------

FROM ${IMAGE_FINAL} as final
ENV REFRESHED_AT 2023-02-28
LABEL Name="roncewind/load" \
      Maintainer="dad@lynntribe.net" \
      Version="0.0.0"

# Copy files from prior step.

COPY --from=go_builder "/output/linux/load" "/app/load"

# Runtime environment variables.

ENV LD_LIBRARY_PATH=/opt/senzing/g2/lib/

RUN apt-get update  \
 && apt-get install -y \
      procps \
&& rm -rf /var/lib/apt/lists/*
# Runtime execution.

WORKDIR /app
ENTRYPOINT ["/app/load"]
