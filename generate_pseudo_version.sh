#!/bin/bash

# 获取最新提交的哈希
commit_hash=$(git rev-parse HEAD)
short_commit_hash=${commit_hash:0:12} # 取哈希的前12位

# 获取最新提交的时间戳
timestamp=$(git show -s --format=%ct HEAD)

# 检测操作系统类型
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS 系统使用的命令
    formatted_timestamp=$(date -u -r "$timestamp" +'%Y%m%d%H%M%S')
else
    # Linux 系统使用的命令
    formatted_timestamp=$(date -u -d @"$timestamp" +'%Y%m%d%H%M%S')
fi

# 使用基础版本号 v0.0.0
pseudo_version="v0.0.0-$formatted_timestamp-$short_commit_hash"

echo "Generated pseudo-version: $pseudo_version"
