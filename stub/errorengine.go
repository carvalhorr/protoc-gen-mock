package stub

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"plugin"
	"reflect"
	"strings"
)

func NewCustomErrorEngine(path string) CustomErrorEngine {
	return &customErrorEngine{
		BasePath: path,
	}
}

type CustomErrorEngine interface {
	GetNewInstance(spec ErrorDetailsSpec) (interface{}, error)
}

type customError struct {
	Hash      string
	ErrorType reflect.Type
}

type customErrorEngine struct {
	BasePath       string
	errorTypeCache map[string]customError
}

func (e *customErrorEngine) GetNewInstance(spec ErrorDetailsSpec) (interface{}, error) {
	if !e.exists(spec) {
		errorType, err := e.createErrorType(spec)
		if err != nil {
			return nil, err
		}
		e.errorTypeCache[getKey(spec)] = *errorType
	}
	return reflect.New(e.errorTypeCache[getKey(spec)].ErrorType).Interface(), nil
}

func (e *customErrorEngine) exists(spec ErrorDetailsSpec) bool {
	_, exists := e.errorTypeCache[getKey(spec)]
	return exists
}

func (e *customErrorEngine) createErrorType(spec ErrorDetailsSpec) (customErr *customError, err error) {
	hash := generateHash(spec)
	path := e.BasePath + hash
	generatePlugin(path, spec)
	compilePlugin(path, spec)
	errorType, err := loadType(path)
	if err != nil {
		return nil, err
	}
	return &customError{
		Hash:      hash,
		ErrorType: errorType,
	}, nil
}

func generateHash(spec ErrorDetailsSpec) string {
	str := spec.Import + spec.Type
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

type CustomErrorPlugin interface {
	GetType() interface{}
}

func loadType(path string) (reflect.Type, error) {
	p, r := plugin.Open(path + "/plugin.so")
	if r != nil {
		return nil, r
	}

	loader, err := p.Lookup("GetType")
	if err != nil {
		return nil, err
	}
	instance := loader.(func() interface{})()
	return reflect.TypeOf(instance), nil
}

func generatePlugin(path string, spec ErrorDetailsSpec) error {
	template := `package main

import (
	custom "$import"
)

func GetType() interface{} {
	return new(custom.$type)
}
`
	generatedCode := strings.ReplaceAll(template, "$import", spec.Import)
	generatedCode = strings.ReplaceAll(generatedCode, "$type", spec.Type)
	return ioutil.WriteFile(path+"/plugin.go", []byte(generatedCode), 0644)

}

func compilePlugin(path string, spec ErrorDetailsSpec) error {
	cmd := exec.Command("go", "build", "-trimpath", "-buildmode=plugin", "-o", path+"/plugin.so", path+"/plugin.go")
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Run()
	if string(stderrBuf.Bytes()) != "" {
		return fmt.Errorf(string(stderrBuf.Bytes()))
	}
	return nil
}

func getKey(spec ErrorDetailsSpec) string {
	return spec.Import + "-" + spec.Type
}
