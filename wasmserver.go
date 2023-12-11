package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/stealthrocket/wasi-go/imports"
	"github.com/stealthrocket/wasi-go/imports/wasi_http"
	"github.com/tetratelabs/wazero"
)

type WASMServer struct {
	Name    string
	Path    string
	Address string
}

func (s *WASMServer) Serve(ctx context.Context, wg *sync.WaitGroup) error {
	evaluatorWasm, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}

	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	wasmModule, err := runtime.CompileModule(ctx, evaluatorWasm)
	if err != nil {
		return fmt.Errorf("error compiling module: %w", err)
	}
	defer wasmModule.Close(ctx)

	if err := wasi_http.MakeWasiHTTP().Instantiate(ctx, runtime); err != nil {
		return fmt.Errorf("error instantiating module: %w", err)
	}

	address, err := freeAddress()
	if err != nil {
		return err
	}

	s.Address = address

	wg.Done()

	builder := imports.NewBuilder().
		WithDirs("/").
		WithArgs(address).
		WithSocketsExtension("wasmedgev2", wasmModule)

	ctx, system, err := builder.Instantiate(ctx, runtime)
	if err != nil {
		return fmt.Errorf("error instantiating module: %w", err)
	}
	defer system.Close(ctx)

	if _, err := runtime.InstantiateModule(ctx, wasmModule, wazero.NewModuleConfig()); err != nil {
		return fmt.Errorf("error instantiating module: %w", err)
	}

	return nil
}

func freeAddress() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()

	return fmt.Sprintf("localhost:%d", l.Addr().(*net.TCPAddr).Port), nil
}
