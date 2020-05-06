security-providers-configurer \
  --source "${JAVA_HOME}/{{.source}}" \
  --additional-providers "$(echo "${SECURITY_PROVIDERS}" | tr ' ' ,)"
