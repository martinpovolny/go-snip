FROM redhat/ubi9 AS builder

RUN INSTALL_PKGS="go-toolset" && \
    yum install -y --setopt=tsflags=nodocs $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    yum clean all -y

ENV GOFLAGS="-buildvcs=false"

COPY . /app/
WORKDIR /app

RUN go build

FROM redhat/ubi9-minimal

COPY --from=builder \
    /app/whale \
    /usr/local/bin/

# run rootless
RUN chown 1001:root /usr/local/bin/whale && \
    chmod 550 /usr/local/bin/whale

RUN mkdir -p /app/db && \
    chown -R 1001:root /app && \
    chmod 770 /app/db

ENV APP_DB_PATH /app/db/app.db

USER 1001

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/whale"]
