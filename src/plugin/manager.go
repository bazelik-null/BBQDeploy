package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/bazelik-null/BBQDeploy/src/pluginapi"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Handler is the signature for any plugin callback
// Variable can be any type
type Handler func(payload interface{})

// Manager loads plugins and dispatches Entry calls
type Manager struct {
	interpreter *interp.Interpreter
	hooks       map[string][]Handler
}

// Global is the shared plugin manager instance
var Global = NewManager()

// NewManager creates a Yaegi interpreter and prepares hook storage
func NewManager() *Manager {
	i := interp.New(interp.Options{})

	_ = i.Use(stdlib.Symbols)

	_ = i.Use(map[string]map[string]reflect.Value{
		"github.com/bazelik-null/BBQDeploy/src/pluginapi": {
			"Page0Package":        reflect.ValueOf((*pluginapi.Page0Package)(nil)),
			"PageInstallPackage":  reflect.ValueOf((*pluginapi.PageInstallPackage)(nil)),
			"PageEndPackage":      reflect.ValueOf((*pluginapi.PageEndPackage)(nil)),
			"AfterInstallPayload": reflect.ValueOf((*pluginapi.AfterInstallPayload)(nil)),
		},
	})

	return &Manager{
		interpreter: i,
		hooks:       make(map[string][]Handler),
	}
}

// RegisterEntry is called by plugins in their Init() to bind a handler to a named entry point.
func (m *Manager) RegisterEntry(name string, h Handler) {
	m.hooks[name] = append(m.hooks[name], h)
}

// LoadPlugins scans dir for “.go” files, Eval’s each as package main, and if it defines Init(), runs it so the plugin can register entries.
func (m *Manager) LoadPlugins(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("[ERROR]: Reading plugins dir: %s", err)
		return
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".go" {
			continue
		}

		path := filepath.Join(dir, e.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("[ERROR]: Reading %s: %d", path, err)
			continue
		}

		// Interpret the plugin source as package main
		_, err = m.interpreter.Eval(string(src))
		if err != nil {
			fmt.Printf("[ERROR]: Eval %s: %d", path, err)
			continue
		}

		// If plugin has Init, call it to self-register handlers
		if sym, err := m.interpreter.Eval("main.Init"); err == nil {
			initFn := sym.Interface().(func())
			initFn()
		}
	}
}

// Entry is hook point inside any function of the host app. Basically where to call plugins.
func (m *Manager) Entry(name string, payload interface{}) {
	if handlers, ok := m.hooks[name]; ok {
		for _, h := range handlers {
			h(payload)
		}
	}
}
