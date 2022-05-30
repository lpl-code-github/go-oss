#!/bin/bash

export RABBITMQ_SERVER=amqp://test:test@127.0.0.1:5671,amqp://test:test@127.0.0.1:5672,amqp://test:test@127.0.0.1:5673
export ES_SERVER=127.0.0.1:9200
export REDIS_CLUSTER=127.0.0.1:6371,127.0.0.1:6372,127.0.0.1:6373,127.0.0.1:6374,127.0.0.1:6375,127.0.0.1:6376
export REDIS_PASSWORD=Lpl0618
export MONGO_SERVER=mongodb://127.0.0.1:27017,127.0.0.1:27018,127.0.0.1:27019/?replicaSet=ossSet
export DINGTALK_TOKEN=
export DINGTALK_SECRET=
export LOG_DIRECTORY=/log/

LISTEN_ADDRESS=10.29.1.1:12345 STORAGE_ROOT=/tmp/1 ./$1/dataServer/dataServer &
LISTEN_ADDRESS=10.29.1.2:12345 STORAGE_ROOT=/tmp/2 ./$1/dataServer/dataServer &
LISTEN_ADDRESS=10.29.1.3:12345 STORAGE_ROOT=/tmp/3 ./$1/dataServer/dataServer &
LISTEN_ADDRESS=10.29.1.4:12345 STORAGE_ROOT=/tmp/4 ./$1/dataServer/dataServer &
LISTEN_ADDRESS=10.29.1.5:12345 STORAGE_ROOT=/tmp/5 ./$1/dataServer/dataServer &
LISTEN_ADDRESS=10.29.1.6:12345 STORAGE_ROOT=/tmp/6 ./$1/dataServer/dataServer &


LISTEN_ADDRESS=10.29.2.1:12345 ./$1/apiServer/apiServer &
LISTEN_ADDRESS=10.29.2.2:12345 ./$1/apiServer/apiServer &

sleep 1

netstat -plnt