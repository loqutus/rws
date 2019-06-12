FROM arm32v6/alpine
COPY build/package/server /server
ADD web /web
VOLUME /data /data
CMD ["/server"]