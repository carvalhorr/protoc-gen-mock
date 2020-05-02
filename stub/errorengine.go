package stub

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/carvalhorr/protoc-gen-mock/util"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"plugin"
	"strings"
)

func NewCustomErrorEngine(path string) (CustomErrorEngine, error) {
	err := util.CreateDir(path)
	return &customErrorEngine{
		BasePath:       path,
		errorTypeCache: make(map[string]customError, 0),
	}, err
}

type CustomErrorEngine interface {
	GetNewInstance(spec *ErrorDetailsSpec) (interface{}, error)
}

type customError struct {
	Hash      string
	ErrorType interface{}
}

type customErrorEngine struct {
	BasePath       string
	errorTypeCache map[string]customError
}

func (e *customErrorEngine) GetNewInstance(spec *ErrorDetailsSpec) (interface{}, error) {
	if !e.exists(spec) {
		errorType, err := e.createErrorType(spec)
		if err != nil {
			return nil, err
		}
		e.errorTypeCache[getKey(spec)] = *errorType
	}
	return e.errorTypeCache[getKey(spec)].ErrorType, nil
}

func (e *customErrorEngine) exists(spec *ErrorDetailsSpec) bool {
	_, exists := e.errorTypeCache[getKey(spec)]
	return exists
}

func (e *customErrorEngine) createErrorType(spec *ErrorDetailsSpec) (customErr *customError, err error) {
	hash := generateHash(spec)
	path := e.BasePath + hash
	genPluginErr := generatePlugin(path, spec)
	if genPluginErr != nil {
		fmt.Println(genPluginErr.Error())
	}
	compilePlugin(path)
	errorType, err := loadType(path)
	if err != nil {
		return nil, err
	}
	return &customError{
		Hash:      hash,
		ErrorType: errorType,
	}, nil
}

func generateHash(spec *ErrorDetailsSpec) string {
	str := spec.Import + spec.Type
	h := sha1.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func loadType(path string) (interface{}, error) {
	p, r := plugin.Open(path + "/plugin.so")
	if r != nil {
		return nil, r
	}

	loader, err := p.Lookup("GetType")
	if err != nil {
		return nil, err
	}
	instance := loader.(func() interface{})()
	return instance, nil
}

func generatePlugin(path string, spec *ErrorDetailsSpec) error {
	err := util.CreateDir(path)
	if err != nil {
		return err
	}
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

func compilePlugin(path string) error {
	cmd := exec.Command("go", "build", "-trimpath", "-buildmode=plugin", "-o", path+"/plugin.so", path+"/plugin.go")
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	cmd.Run()
	cmd.Process.Release()
	if string(stderrBuf.Bytes()) != "" {
		return fmt.Errorf(string(stderrBuf.Bytes()))
	}
	return nil
}

func getKey(spec *ErrorDetailsSpec) string {
	return spec.Import + "-" + spec.Type
}
