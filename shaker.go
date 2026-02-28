package shaker

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MappedErrors map[error]int

type Shaker struct {
	engine     *gin.Engine
	mappedErrs MappedErrors
}

type Context = gin.Context

func NewShaker(customErrors *MappedErrors) Shaker {
	errs := MappedErrors{
		ErrRessourceNotFound:       http.StatusNotFound,
		ErrInvalidHandlerSignature: http.StatusTeapot,
	}

	if customErrors != nil {
		for e, v := range *customErrors {
			errs[e] = v
		}
	}

	return Shaker{
		engine:     gin.New(),
		mappedErrs: errs,
	}
}

func (s *Shaker) Get(path string, handler interface{}, defaultStatusCode int) error {
	ginHandler, err := shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
		mappedErrors:      s.mappedErrs,
	}.ginize()

	if err != nil {
		return err
	}

	s.engine.GET(path, ginHandler)
	return nil
}
func (s *Shaker) Post(path string, handler interface{}, defaultStatusCode int) error {
	ginHandler, err := shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
		mappedErrors:      s.mappedErrs,
	}.ginize()

	if err != nil {
		return err
	}

	s.engine.POST(path, ginHandler)
	return nil
}

func (s *Shaker) Put(path string, handler interface{}, defaultStatusCode int) error {
	ginHandler, err := shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
		mappedErrors:      s.mappedErrs,
	}.ginize()

	if err != nil {
		return err
	}

	s.engine.PUT(path, ginHandler)
	return nil
}
func (s *Shaker) Delete(path string, handler interface{}, defaultStatusCode int) error {
	ginHandler, err := shakerFunc{
		callback:          handler,
		defaultStatusCode: defaultStatusCode,
		mappedErrors:      s.mappedErrs,
	}.ginize()

	if err != nil {
		return err
	}

	s.engine.DELETE(path, ginHandler)
	return nil
}

type shakerFunc struct {
	callback          interface{}
	defaultStatusCode int
	mappedErrors      MappedErrors
}

func (sf shakerFunc) ginize() (gin.HandlerFunc, error) {
	cbValue := reflect.ValueOf(sf.callback)
	functType := cbValue.Type()

	inputCount, outputCount := functType.NumIn(), functType.NumOut()

	// Check input and output parameters
	if inputCount > 2 || outputCount > 2 {
		logrus.Error("invalid handler signature")
		return nil, ErrInvalidHandlerSignature
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
	}, nil
}

type errorBody struct {
	Err string `json:"error"`
}

func handleErr(ctx *gin.Context, sf *shakerFunc, err error, out any) {
	if err == nil {
		ctx.JSON(sf.defaultStatusCode, out)
		return
	}

	if errorFromMapping, found := sf.mappedErrors[err]; found {
		ctx.JSON(errorFromMapping, errorBody{Err: err.Error()})
	} else {
		ctx.JSON(http.StatusInternalServerError, errorBody{Err: "internal server error"})
	}
}

func (s *Shaker) Shake() error {
	return s.engine.Run()
}
