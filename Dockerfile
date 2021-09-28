FROM alpine:3.14

# 'nobody' user in alpine
USER 65534
COPY rbac-manager /

ENTRYPOINT ["/rbac-manager"]
CMD ["--log-level=info"]
