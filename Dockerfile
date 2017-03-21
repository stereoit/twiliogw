FROM scratch

EXPOSE 9090

ADD ca-certificates.crt /etc/ssl/certs/
ADD twiliogw /

CMD ["/twiliogw"]