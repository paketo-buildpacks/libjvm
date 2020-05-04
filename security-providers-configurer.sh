security-providers-configurer \
  --jdk-source "${JDK_HOME}/{{.jdkSource}}" \
  --jre-source "${JAVA_HOME}/{{.jreSource}}" \
  --additional-providers "$(echo "${SECURITY_PROVIDERS}" | tr ' ' ,)"
