package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Registry struct {
	servers []*WASMServer
}

func NewRegistry(dir string) (*Registry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var m []*WASMServer

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".wasm" {
			continue
		}

		m = append(m, &WASMServer{
			Name: strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
			Path: filepath.Join(dir, entry.Name()),
		})
	}

	return &Registry{servers: m}, nil
}

func (r *Registry) Servers() []*WASMServer {
	return r.servers
}

func (r *Registry) Server(hostport string) (*WASMServer, error) {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}

	for _, s := range r.servers {
		if s.Name == host {
			return s, nil
		}
	}

	return nil, fmt.Errorf("server not found: %s", host)
}
