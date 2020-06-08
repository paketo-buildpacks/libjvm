JAVA_OPTS="${JAVA_OPTS} -XX:ActiveProcessorCount=$(nproc)" || exit $?
export JAVA_OPTS
