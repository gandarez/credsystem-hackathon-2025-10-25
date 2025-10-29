#!/usr/bin/env bash

startContainer() {
    pushd ../participantes/$1 > /dev/null
        services=$(docker compose config --services | wc -l)
        echo "" > docker-compose.logs
        docker compose up --build --wait -d >> docker-compose.logs
    popd > /dev/null
}

stopContainer() {
    pushd ../participantes/$1 > /dev/null
        docker compose down -v --remove-orphans
        docker compose rm -s -v -f
    popd > /dev/null
}

for directory in ../participantes/*; do
(
    git pull
    participant=$(echo $directory | sed -e 's/..\/participantes\///g' -e 's/\///g')
    echo ""
    echo ""
    echo "========================================"
    echo "  Participant $participant starting..."
    echo "========================================"

    # testedFile="$directory/93.json"

    # if ! test -f $testedFile; then
    # touch $testedFile

    if [ ! -d "$directory/results" ]; then
        mkdir -p "$directory/results"
    fi

    echo "executing test for $participant..."
    stopContainer $participant
    startContainer $participant

    success=1
    max_attempts=5
    attempt=1
    while [ $success -ne 0 ] && [ $max_attempts -ge $attempt ]; do
        curl -f -s --max-time 5 localhost:18020/api/healthz
        success=$?
        echo "tried $attempt out of $max_attempts..."
        sleep 10
        ((attempt++))
    done

    if [ $success -eq 0 ]; then
        # echo "" > $directory/k6.logs
        # k6 run -e PARTICIPANT=$participant --log-output=file=$directory/k6.logs k6_runner.js
        echo "" > $directory/results/test.logs
        
        echo "Running initial test for $participant..."
        go run main.go ../assets/intents_pre_loaded.csv http://localhost:18020/api/find-service $directory/results/93.json > $directory/results/test.logs 2>&1

        echo "Running extra test for $participant..."
        go run main.go ../assets/extra_intents.csv http://localhost:18020/api/find-service $directory/results/80.json > $directory/results/test.logs 2>&1    

        stopContainer $participant
        echo "======================================="
        echo "working on $participant"
        
        # TODO: revisitar
        # sed -n "1,100p" file.txt
        # sed -i '1001,$d' $directory/docker-compose.logs
        # sed -i '1001,$d' $directory/k6.logs
        # echo "log truncated at line 1000" >> $directory/docker-compose.logs
        # echo "log truncated at line 1000" >> $directory/k6.logs
    else
        stopContainer $participant
        echo "[$(date)] Seu backend não respondeu nenhuma das $max_attempts tentativas de GET para http://localhost:18020/api/healthz. Teste abortado." > $directory/error.logs
        echo "[$(date)] Inspecione o arquivo docker-compose.logs para mais informações." >> $directory/error.logs
        echo "Could not get a successful response from backend... aborting test for $participant"
    fi

    # git add $directory
    # git commit -m "add $participant's partial result"
    # git push

    echo "================================="
    echo "  Finished testing $participant!"
    echo "================================="

    sleep 5
    # else
    #     echo "================================="
    #     echo "  Skipping $participant"
    #     echo "================================="
    # fi
)
done

date

# echo "generating results preview..."

# PREVIA_RESULTADOS=../PREVIA_RESULTADOS.md