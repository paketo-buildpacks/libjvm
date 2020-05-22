openssl-certificate-loader \
  --ca-certificates="/etc/ssl/certs/ca-certificates.crt" \
  --keystore-path="${JAVA_HOME}/{{.source}}" \
  --keystore-password="changeit"
