if [[ -f /etc/ssl/certs/ca-certificates.crt ]]; then
  export JAVA_OPTS="${JAVA_OPTS} -Dio.paketo.openssl.ca-certificates=/etc/ssl/certs/ca-certificates.crt"
fi
