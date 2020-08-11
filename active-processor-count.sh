JAVA_TOOL_OPTIONS="${JAVA_TOOL_OPTIONS} -XX:ActiveProcessorCount=$(nproc)" || exit $?
export JAVA_TOOL_OPTIONS
