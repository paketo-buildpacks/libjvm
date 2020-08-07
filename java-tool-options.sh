[[ -z "${JAVA_OPTS+x}" ]] && return

export JAVA_TOOL_OPTIONS="${JAVA_OPTS} ${JAVA_TOOL_OPTIONS}"
