FROM alpine:3.17

# 'nobody' user in alpine
USER 65534
COPY rbac-manager /

ENTRYPOINT ["/rbac-manager"]
CMD ["--log-level=info"]
