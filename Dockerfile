FROM gcr.io/distroless/static:nonroot

ARG TARGETPLATFORM

WORKDIR /app

COPY $TARGETPLATFORM/wordtris /app/wordtris

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/wordtris"]