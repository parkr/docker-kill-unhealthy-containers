FROM alpine
RUN which ping || apk add --no-cache ping
HEALTHCHECK --interval=1s --retries=3 --timeout=30s CMD [ "ping", "blahblahblahhh.com" ]
CMD [ "/bin/sh" ]
