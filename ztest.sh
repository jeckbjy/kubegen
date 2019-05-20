#!/bin/bash
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}")" && pwd )

cd $DIR

if [[ ! -f kubegen ]]; then
  ./zbuild.sh
fi

echo 'rm ./data/game_out'
rm -rf ./data/game_out

echo 'generate game'
./kubegen service.yaml deployment.yaml --expand -l deployment-tencent.yaml=1 -c values.yaml -s tencent -i ./data/game -o ./data/game_out -v APP=commgame -v ENV=alpha

echo 'rm ./data/logstash_out'
rm -rf ./data/logstash_out

echo 'generate logstash'
./kubegen logstash.yaml -v APP=word -v IMAGE=docker.elastic.co/logstash/logstash:7.0.1 -i ./data/logstash -o ./data/logstash_out

echo 'generate logstash1'
./kubegen logstash1.yaml -v APP=word -v IMAGE=docker.elastic.co/logstash/logstash:7.0.1 -v CONF=logstash.conf -i ./data/logstash -o ./data/logstash_out