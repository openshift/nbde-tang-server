FROM scratch

WORKDIR /
COPY manager /nbde-tang-server-manager

USER "root"

ENTRYPOINT ["/nbde-tang-server-manager"]
