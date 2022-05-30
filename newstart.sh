#!/bin/bash
# 启动一个新的节点
mkdir -p /tmp/7/objects
mkdir -p /tmp/7/temp
mkdir -p /tmp/7/garbage
rm -rf /tmp/7/objects/*
rm -rf /tmp/7/temp/*
rm -rf /tmp/7/garbage/*

export RABBITMQ_SERVER=amqp://test:test@127.0.0.1:5671,amqp://test:test@127.0.0.1:5672,amqp://test:test@127.0.0.1:5673
export ES_SERVER=127.0.0.1:9200
export REDIS_CLUSTER=127.0.0.1:6371,127.0.0.1:6372,127.0.0.1:6373,127.0.0.1:6374,127.0.0.1:6375,127.0.0.1:6376
export REDIS_PASSWORD=Lpl0618
export MONGO_SERVER=mongodb://127.0.0.1:27017,127.0.0.1:27018,127.0.0.1:27019/?replicaSet=ossSet
export DINGTALK_TOKEN=
export DINGTALK_SECRET=
export LOG_DIRECTORY=/log/

LISTEN_ADDRESS=10.29.1.7:12345 STORAGE_ROOT=/tmp/6 ./$1/dataServer/dataServer &

sleep 1

netstat -plnt