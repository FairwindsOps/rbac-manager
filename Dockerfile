FROM alpine:3.23

# 'nobody' user in alpine
USER 65534
COPY rbac-manager /

ENTRYPOINT ["/rbac-manager"]
CMD ["--log-level=info"]
