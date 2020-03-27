[[ -z "${SECURITY_PROVIDERS_CLASSPATH+x}" ]] && return

EXT_DIRS="${JAVA_HOME}/lib/ext"

for I in ${SECURITY_PROVIDERS_CLASSPATH//:/$'\n'}; do
  EXT_DIRS="${EXT_DIRS}:$(dirname "${I}")"
done

export JAVA_OPTS="${JAVA_OPTS} -Djava.ext.dirs=${EXT_DIRS}"
