package main

import (
	"io/ioutil"
	"log"

	"github.com/WinterYukky/aws-lambda-sql-runtime/runtime"

	crkit "github.com/WinterYukky/aws-lambda-custom-runtime-kit"
)

type DefaultFileReader struct{}

func (d DefaultFileReader) Read(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func main() {
	sqlRuntime := runtime.NewAWSLambdaSQLRuntime(DefaultFileReader{})
	customRuntime := crkit.NewAWSLambdaCustomRuntime(sqlRuntime)
	if err := customRuntime.Invoke(); err != nil {
		log.Fatalf("Failed to invoke lambda: %v", err)
	}
}
