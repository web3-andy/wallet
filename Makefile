.PHONY: help build run test clean

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译所有程序
	@echo "编译 server..."
	go build -o bin/server ./cmd/server
	@echo "编译 worker..."
	go build -o bin/worker ./cmd/worker
	@echo "编译 cli..."
	go build -o bin/cli ./cmd/cli
	@echo "✓ 编译完成"

run-server: ## 运行 API 服务器
	go run ./cmd/server/main.go

run-worker: ## 运行后台 Worker
	go run ./cmd/worker/main.go

run-gas: ## 运行 Gas 估算示例
	go run ./pkg/gas/cmd/main.go

test: ## 运行测试
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## 代码检查
	golangci-lint run ./...

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .

mod: ## 整理依赖
	go mod tidy
	go mod verify

clean: ## 清理编译产物
	rm -rf bin/
	rm -f coverage.out coverage.html

install: ## 安装工具
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

docker-build: ## 构建 Docker 镜像
	docker build -t wallet:latest .

docker-run: ## 运行 Docker 容器
	docker-compose up -d
