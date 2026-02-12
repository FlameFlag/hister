FROM gcr.io/distroless/static-debian13:nonroot

ARG TARGETPLATFORM
ENTRYPOINT ["/usr/bin/hister"]
COPY $TARGETPLATFORM/hister /

CMD ["/hister", "listen"]
