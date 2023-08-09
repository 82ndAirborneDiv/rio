FROM alpine:3.16.7
RUN apk -U --no-cache add ca-certificates
COPY bin/rio-controller bin/rio /usr/bin/
CMD ["rio-controller"]
