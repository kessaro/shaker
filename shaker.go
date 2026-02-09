package shaker

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type shaker struct {
	engine *gin.Engine
}

type Context = gin.Context

func NewShaker() shaker {
	return shaker{
		engine: gin.New(),
	}
}

func (s *shaker) Get(path string, handler interface{}, defaultStatusCode int) {
	s.engine.GET(path, shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
	}.ginize())
}
func (s *shaker) Post(path string, handler interface{}, defaultStatusCode int) {
	s.engine.POST(path, shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
	}.ginize())
}

func (s *shaker) Put(path string, handler interface{}, defaultStatusCode int) {
	s.engine.PUT(path, shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
	}.ginize())
}
func (s *shaker) Delete(path string, handler interface{}, defaultStatusCode int) {
	s.engine.DELETE(path, shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
	}.ginize())
}

type shakerFunc struct {
	callback          interface{}
	defaultStatusCode int
}

func (sf shakerFunc) ginize() gin.HandlerFunc {
	cbValue := reflect.ValueOf(sf.callback)
	functType := cbValue.Type()

	inputCount, outputCount := functType.NumIn(), functType.NumOut()

	// Check input and output parameters
	if inputCount > 2 || outputCount > 2 {
		logrus.Fatal("invalid handler signature")
	}

	// TODO: check in/out types

	var inputStruct reflect.Value

	if inputCount == 2 {
		t := functType.In(1).Elem()
		inputStruct = reflect.New(t)
	}

	return func(ctx *gin.Context) {
		if inputCount == 2 {
			bindingDest := inputStruct.Interface()
			if err := ctx.BindUri(bindingDest); err != nil {
				handleErr(ctx, &sf, err, nil)
				return
			}

			if err := ctx.BindQuery(bindingDest); err != nil {
				handleErr(ctx, &sf, err, nil)
				return
			}

			if err := ctx.BindHeader(bindingDest); err != nil {
				handleErr(ctx, &sf, err, nil)
				return
			}

			if err := ctx.Bind(bindingDest); err != nil {
				handleErr(ctx, &sf, err, nil)
				return
			}
		}

		fmt.Println(reflect.ValueOf(ctx).Kind())

		inputs := []reflect.Value{
			reflect.ValueOf(ctx),
		}

		if inputCount == 2 {
			inputs = append(inputs, inputStruct)
		}

		out := cbValue.Call(inputs)

		var outputStruct any = nil
		if len(out) == 2 {
			outputStruct = out[0].Interface()
		}

		if errI := out[len(out)-1].Interface(); errI == nil {
			handleErr(ctx, &sf, nil, outputStruct)
		} else {
			handleErr(ctx, &sf, errI.(error), nil)
		}
	}
}

type errorBody struct {
	Err error `json:"error"`
}

func handleErr(ctx *gin.Context, sf *shakerFunc, err error, out any) {
	switch err.(type) {
	case errNotFound:
		ctx.JSON(http.StatusNotFound, errorBody{Err: err})
	default:
		ctx.JSON(sf.defaultStatusCode, out)
	}
}

func (s *shaker) Shake() error {
	return s.engine.Run()
}
