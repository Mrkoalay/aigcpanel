#!/bin/bash

# 模型调用脚本，用于处理不同模型目录的访问问题

# 检查参数
if [ $# -lt 3 ]; then
    echo "使用方法: ./run_model.sh <模型路径> <配置文件路径> <功能名称> [参数...]"
    echo "示例: ./run_model.sh /path/to/model /path/to/config.json soundTts text=你好，世界 speaker=中文女 speed=1.0"
    exit 1
fi

# 获取参数
MODEL_PATH="$1"
CONFIG_PATH="$2"
FUNCTION_NAME="$3"
shift 3
PARAMS="$*"

# 检查模型路径是否存在
if [ ! -d "$MODEL_PATH" ]; then
    echo "错误: 模型路径不存在或无法访问: $MODEL_PATH"
    echo "请确保模型路径正确且您有访问权限"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "$CONFIG_PATH" ]; then
    echo "错误: 配置文件不存在或无法访问: $CONFIG_PATH"
    echo "请确保配置文件路径正确且您有访问权限"
    exit 1
fi

# 编译模型调用器（如果需要）
if [ ! -f "./model_caller" ]; then
    echo "编译模型调用器..."
    go build -o model_caller model_caller.go config_loader.go
    if [ $? -ne 0 ]; then
        echo "编译失败"
        exit 1
    fi
fi

# 运行模型调用器
echo "运行模型调用器..."
echo "模型路径: $MODEL_PATH"
echo "配置文件: $CONFIG_PATH"
echo "功能名称: $FUNCTION_NAME"
echo "参数: $PARAMS"

./model_caller "$MODEL_PATH" "$CONFIG_PATH" "$FUNCTION_NAME" $PARAMS