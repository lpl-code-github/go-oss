#!/bin/bash
cd /home/lpl/Desktop/oss/
go build -o apiServer apiServer/apiServer.go
go build -o dataServer dataServer/dataServer.go
go build -o deleteOldMetadata deleteOldMetadata/deleteOldMetadata.go
go build -o deleteOrphanObject deleteOrphanObject/deleteOrphanObject.go
go build -o objectScanner objectScanner/objectScanner.go

#关闭服务
killall apiServer
killall dataServer


#开启服务
for i in `seq 1 6`
do
    mkdir -p /tmp/$i/objects
    mkdir -p /tmp/$i/temp
    mkdir -p /tmp/$i/garbage
done

for i in `seq 1 6`
do
    rm -rf /tmp/$i/objects/*
    rm -rf /tmp/$i/temp/*
    rm -rf /tmp/$i/garbage/*
done

sudo chmod 777 /log/

sudo ifconfig lo:1 10.29.1.1/16
sudo ifconfig lo:2 10.29.1.2/16
sudo ifconfig lo:3 10.29.1.3/16
sudo ifconfig lo:4 10.29.1.4/16
sudo ifconfig lo:5 10.29.1.5/16
sudo ifconfig lo:6 10.29.1.6/16
sudo ifconfig lo:7 10.29.1.7/16
sudo ifconfig lo:8 10.29.2.1/16
sudo ifconfig lo:9 10.29.2.2/16
sudo ifconfig lo:10 10.29.3.1/16

chmod 777 apiServer/apiServer
chmod 777 dataServer/dataServer
chmod 777 deleteOldMetadata/deleteOldMetadata
chmod 777 deleteOrphanObject/deleteOrphanObject
chmod 777 objectScanner/objectScanner