#!/bin/bash
go build

# 이전 프로세스 종료
sh kill.sh

# 로그 디렉토리 생성
mkdir -p logs

# 로그 파일 초기화
for id in {1..4}; do
    > "./logs/hotstuff_$id.log"
done

# 순서대로 노드 시작 (1부터 4까지)
for id in $(seq 1 4)
do
    ./hotstuff -id=$id >> "./logs/hotstuff_$id.log" 2>&1 &
    echo "Started hotstuff node $id"
    # 각 노드가 완전히 시작될 때까지 충분히 기다림
    sleep 5
done

echo "All nodes started. Check logs for details."