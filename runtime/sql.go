package runtime

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	crkit "github.com/WinterYukky/aws-lambda-custom-runtime-kit"
	"github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type FileReader interface {
	Read(filename string) ([]byte, error)
}

func NewAWSLambdaSQLRuntime(fileReader FileReader) *AWSLambdaSQLRuntime {
	return &AWSLambdaSQLRuntime{
		reader: fileReader,
	}
}

type AWSLambdaSQLRuntime struct {
	reader  FileReader
	handler []byte
}

func (a *AWSLambdaSQLRuntime) Setup(env *crkit.AWSLambdaRuntimeEnvironemnt) error {
	handler, err := a.reader.Read(fmt.Sprintf("%v/%v.sql", env.LambdaTaskRoot, env.Handler))
	a.handler = handler
	return err
}

func (a *AWSLambdaSQLRuntime) Invoke(event []byte, context *crkit.Context) (interface{}, error) {
	db, err := a.open(event, context)
	if err != nil {
		return nil, err
	}
	return a.invoke(db)
}
func (a *AWSLambdaSQLRuntime) Cleanup(env *crkit.AWSLambdaRuntimeEnvironemnt) {}

func (a *AWSLambdaSQLRuntime) open(event []byte, context *crkit.Context) (*gorm.DB, error) {
	eventUDF := func(path string) string {
		if path == "" {
			return string(event)
		}
		value := gjson.Get(string(event), path)
		return value.String()
	}
	driverName := "sqlite3_" + context.RequestID
	sql.Register(driverName, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("event", eventUDF, false); err != nil {
				return err
			}
			if err := conn.RegisterFunc("print", printUDF, true); err != nil {
				return err
			}
			return nil
		},
	})

	db, err := gorm.Open(&sqlite.Dialector{DSN: ":memory:", DriverName: driverName}, &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			Colorful: false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return db, nil
}

func printUDF(value interface{}) string {
	switch v := value.(type) {
	case string:
		println(v)
		return v
	default:
		b, _ := json.Marshal(v)
		println(string(b))
		return string(b)
	}
}
func (a *AWSLambdaSQLRuntime) invoke(db *gorm.DB) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := db.Raw(string(a.handler)).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	return result, nil
}
