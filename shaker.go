package shaker

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func (s *Shaker) Handler() http.Handler {
	return s.engine.Handler()
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

	var inputType reflect.Type
	if inputCount == 2 {
		inputType = functType.In(1).Elem()
	}

	return func(ctx *gin.Context) {
		var inputStruct reflect.Value
		if inputCount == 2 {
			inputStruct = reflect.New(inputType)
			bindingDest := inputStruct.Interface()
			// use ShouldBind variants so gin does not write a response on error; we
			// want to delegate error handling to handleErr.
			if hasUriTags(bindingDest) {
				if err := ctx.ShouldBindUri(bindingDest); err != nil {
					handleErr(ctx, &sf, err, nil)
					return
				}
			}

			if hasQueryTags(bindingDest) {
				if err := ctx.ShouldBindQuery(bindingDest); err != nil {
					handleErr(ctx, &sf, err, nil)
					return
				}
			}

			if hasHeaderTags(bindingDest) {
				if err := ctx.ShouldBindHeader(bindingDest); err != nil {
					handleErr(ctx, &sf, err, nil)
					return
				}
			}

			if err := ctx.ShouldBind(bindingDest); err != nil {
				handleErr(ctx, &sf, err, nil)
				return
			}
		}

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

	// validator errors should always result in a bad request response. They are
	// produced by gin's binding mechanism when the input doesn't satisfy tags
	// like `binding:"required"` or `min=...`.
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ctx.JSON(http.StatusBadRequest, errorBody{Err: err.Error()})
		return
	}

	// look through the mapped errors using errors.Is instead of indexing the map
	// directly with `err`. The latter can panic if `err` holds an unhashable
	// concrete type (e.g. validator.ValidationErrors).
	for mappedErr, status := range sf.mappedErrors {
		if errors.Is(err, mappedErr) {
			ctx.JSON(status, errorBody{Err: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusInternalServerError, errorBody{Err: "internal server error"})
}

// hasUriTags returns true when at least one field in the value pointed to by
// v has a `uri:"..."` struct tag. We expect v to be a pointer to struct.
func hasTag(v interface{}, tagNames ...string) bool {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return false
	}
	typ := val.Elem().Type()
	for i := 0; i < typ.NumField(); i++ {
		for _, tagName := range tagNames {
			if tag := typ.Field(i).Tag.Get(tagName); tag != "" {
				return true
			}
		}
	}
	return false
}

func hasUriTags(v interface{}) bool {
	return hasTag(v, "uri")
}

func hasQueryTags(v interface{}) bool {
	return hasTag(v, "form", "query")
}

func hasHeaderTags(v interface{}) bool {
	return hasTag(v, "header")
}

func (s *Shaker) Shake() error {
	return s.engine.Run()
}
