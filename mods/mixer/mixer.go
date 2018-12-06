package mixer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kuuland/kuu"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ModServer 模块gRPC服务器实例
var ModServer = grpc.NewServer()

// Call 服务调用
// 1.初始化连接
// 2.初始化上下文
func Call(address string, handler func(*grpc.ClientConn, context.Context)) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		kuu.Error("did not connect: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if cancel != nil {
			cancel()
		}
	}()
	handler(conn, ctx)
}

// Run 封装启动函数（gRPC和HTTP复用端口）
func Run(k *kuu.Kuu, addr ...string) (err error) {
	address := resolveAddress(addr)
	port := resolvePort(address)

	// 初始化监听
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		kuu.Error("failed to listen: %v", err)
		return err
	}
	// 初始化链路复用器
	m := cmux.New(lis)
	// 定义匹配优先级：gRPC > HTTP
	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())
	// 初始化gRPC服务器
	grpcS := ModServer
	reflection.Register(grpcS)
	// 初始化HTTP服务器
	httpS := &http.Server{
		Addr:    address,
		Handler: k,
	}
	kuu.Emit("BeforeRun", k)
	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)
	err = m.Serve()
	if err != nil {
		kuu.Error("failed to serve: %v", err)
	}
	return err
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			return ":" + port
		}
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too much parameters")
	}
}

func resolvePort(address string) int {
	split := strings.Split(address, ":")
	port := "8080"
	switch len(split) {
	case 0:
		if p := os.Getenv("PORT"); p != "" {
			port = p
		}
	case 1:
		port = split[0]
	case 2:
		port = split[1]
	default:
		panic("too much parameters")
	}
	p, _ := strconv.Atoi(port)
	return p
}
