#!/bin/bash

# 部署脚本

# 设置环境变量
export GO_ENV=production

# 编译项目
make build

# 停止旧的服务（如果存在）
if [ -f ./pid.txt ]; then
    PID=$(cat ./pid.txt)
    echo "Stopping old service with PID: $PID"
    kill -TERM $PID
    sleep 3
    # 强制终止
    if ps -p $PID > /dev/null; then
        kill -KILL $PID
    fi
fi

# 启动新服务
nohup ./bin/shorturl-server > ./logs/app.log 2>&1 &
echo $! > ./pid.txt

# 检查服务是否启动成功
sleep 5
if ps -p $(cat ./pid.txt) > /dev/null; then
    echo "Service started successfully"
else
    echo "Service failed to start"
    exit 1
fi