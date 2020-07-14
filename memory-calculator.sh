HEAD_ROOM=${BPL_JVM_HEAD_ROOM:=0}

if [[ -z "${BPL_JVM_LOADED_CLASS_COUNT+x}" ]]; then
  LOADED_CLASS_COUNT=$(class-counter --source "{{.source}}" --jvm-class-count "{{.jvmClassCount}}") || exit $?
else
  LOADED_CLASS_COUNT=${BPL_JVM_LOADED_CLASS_COUNT}
fi

THREAD_COUNT=${BPL_JVM_THREAD_COUNT:=250}

TOTAL_MEMORY=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes) || exit $?

if [ "${TOTAL_MEMORY}" -eq 9223372036854771712 ]; then
  printf "Container memory limit unset. Configuring JVM for 1G container.\n"
  TOTAL_MEMORY=1073741824
elif [ ${TOTAL_MEMORY} -gt 70368744177664 ]; then
  printf "Container memory limit too large. Configuring JVM for 64T container.\n"
  TOTAL_MEMORY=70368744177664
fi

MEMORY_CONFIGURATION=$(java-buildpack-memory-calculator \
    --head-room "${HEAD_ROOM}" \
    --jvm-options "${JAVA_OPTS}" \
    --loaded-class-count "${LOADED_CLASS_COUNT}" \
    --thread-count "${THREAD_COUNT}" \
    --total-memory "${TOTAL_MEMORY}") || exit $?

printf "Calculated JVM Memory Configuration: %s (Head Room: %d%%, Loaded Class Count: %d, Thread Count: %d, Total Memory: %s)\n" \
  "${MEMORY_CONFIGURATION}" "${HEAD_ROOM}" "${LOADED_CLASS_COUNT}" "${THREAD_COUNT}" "$(numfmt --to=iec "${TOTAL_MEMORY}")"
export JAVA_OPTS="${JAVA_OPTS} ${MEMORY_CONFIGURATION}"
