bytes=$(
  docker volume ls -q | while read -r v; do
    docker run --rm -v "$v":/data alpine sh -c "du -sk /data 2>/dev/null | awk '{print \$1*1024}' || echo 0"
  done | awk '{s+=$1} END{print s+0}'
); numfmt --to=iec --suffix=B "$bytes"
