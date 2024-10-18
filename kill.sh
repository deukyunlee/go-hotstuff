for id in {1..4}
do
    pid=$(pgrep -f "./hotstuff -id=$id")

    if [ -n "$pid" ]; then
        echo "Killing hotstuff process with -id=$id and PID: $pid"
        kill -9 $pid
    else
        echo "No hotstuff process with -id=$id is running."
    fi
done
