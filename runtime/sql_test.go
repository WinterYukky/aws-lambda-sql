package runtime

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	crkit "github.com/WinterYukky/aws-lambda-custom-runtime-kit"
)

func pointer(value interface{}) *interface{} {
	return &value
}

var count = 0

func incrementedCount() int {
	count++
	return count
}

func TestAWSLambdaSQLRuntime_Invoke(t *testing.T) {
	type fields struct {
		reader  FileReader
		handler []byte
	}
	type args struct {
		event   []byte
		context *crkit.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Multiple queries responses last query result",
			fields: fields{
				handler: []byte("SELECT 'first' name; SELECT 'second' name;"),
			},
			args: args{
				event: []byte(`{"key1":"value1","key2":"value2"}`),
				context: &crkit.Context{
					RequestID: strconv.Itoa(incrementedCount()),
				},
			},
			want: map[string]interface{}{
				"name": pointer("second"),
			},
			wantErr: false,
		},
		{
			name: "Event UDF with empty key returns all event data",
			fields: fields{
				handler: []byte("SELECT event('') event;"),
			},
			args: args{
				event: []byte(`{"key1":"value1","key2":"value2"}`),
				context: &crkit.Context{
					RequestID: strconv.Itoa(incrementedCount()),
				},
			},
			want: map[string]interface{}{
				"event": pointer(`{"key1":"value1","key2":"value2"}`),
			},
			wantErr: false,
		},
		{
			name: "Event UDF with key returns event data at key",
			fields: fields{
				handler: []byte("SELECT event('key1') key1;"),
			},
			args: args{
				event: []byte(`{"key1":"value1","key2":"value2"}`),
				context: &crkit.Context{
					RequestID: strconv.Itoa(incrementedCount()),
				},
			},
			want: map[string]interface{}{
				"key1": pointer("value1"),
			},
			wantErr: false,
		},
		{
			name: "Print UDF doesn't not return error",
			fields: fields{
				handler: []byte("SELECT print(event('key1')) key1, print(1) number;"),
			},
			args: args{
				event: []byte(`{"key1":"value1","key2":"value2"}`),
				context: &crkit.Context{
					RequestID: strconv.Itoa(incrementedCount()),
				},
			},
			want: map[string]interface{}{
				"key1":   pointer("value1"),
				"number": pointer("1"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AWSLambdaSQLRuntime{
				reader:  tt.fields.reader,
				handler: tt.fields.handler,
			}
			got, err := a.Invoke(tt.args.event, tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("AWSLambdaSQLRuntime.Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AWSLambdaSQLRuntime.Invoke() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockFileReader struct {
	read func(filename string) ([]byte, error)
}

func (m *MockFileReader) Read(filename string) ([]byte, error) {
	return m.read(filename)
}

func TestAWSLambdaSQLRuntime_Setup(t *testing.T) {
	type fields struct {
		reader FileReader
	}
	type args struct {
		env *crkit.AWSLambdaRuntimeEnvironemnt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Should install handler",
			fields: fields{
				reader: &MockFileReader{
					read: func(filename string) ([]byte, error) {
						if filename != "/my/workspace/path/index.sql" {
							return nil, errors.New("path is not valid")
						}
						return []byte(`{"key1":"value1","key2":"value2"}`), nil
					},
				},
			},
			args: args{
				env: &crkit.AWSLambdaRuntimeEnvironemnt{
					LambdaTaskRoot: "/my/workspace/path",
					Handler:        "index",
				},
			},
			want: []byte(`{"key1":"value1","key2":"value2"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAWSLambdaSQLRuntime(tt.fields.reader)
			if err := a.Setup(tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("AWSLambdaSQLRuntime.Setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(a.handler, tt.want) {
				t.Errorf("AWSLambdaSQLRuntime.Setup() = %v, want %v", a.handler, tt.want)
			}
		})
	}
}
