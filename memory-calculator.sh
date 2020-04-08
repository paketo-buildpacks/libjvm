HEAD_ROOM=${BPL_HEAD_ROOM:=0}

if [[ -z "${BPL_LOADED_CLASS_COUNT+x}" ]]; then
  LOADED_CLASS_COUNT=$(class-counter --source "{{.source}}" --jvm-class-count "{{.jvmClassCount}}")
else
  LOADED_CLASS_COUNT=${BPL_LOADED_CLASS_COUNT}
fi

THREAD_COUNT=${BPL_THREAD_COUNT:=250}

TOTAL_MEMORY=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)

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
    --total-memory "${TOTAL_MEMORY}")

printf "Calculated JVM Memory Configuration: ${MEMORY_CONFIGURATION} (Head Room: ${HEAD_ROOM}%%, Loaded Class Count: ${LOADED_CLASS_COUNT}, Thread Count: ${THREAD_COUNT}, Total Memory: ${TOTAL_MEMORY})\n"
export JAVA_OPTS="${JAVA_OPTS} ${MEMORY_CONFIGURATION}"
