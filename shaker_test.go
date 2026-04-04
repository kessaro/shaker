package shaker

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maxatome/go-testdeep/helpers/tdhttp"
	"github.com/maxatome/go-testdeep/td"
)

type methodType int

const (
	get    methodType = iota
	post   methodType = 1 * iota
	put    methodType = 2 * iota
	delete methodType = 3 * iota
)

var ErrSample = errors.New("sample error")

type emptyStructType struct{}

type endpointCall struct {
	name            string
	endpoint        string
	payload         string
	expectedStatus  int
	expectedPayload string
}

func goodHandler(ctx *Context, in *emptyStructType) error {
	return nil
}

func wrongHandlerInputCount(ctx *Context, in *emptyStructType, inn *emptyStructType) error {
	return nil
}

func TestRegisteringEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		regFunc       func(*Shaker) error
		expectedError error
	}{
		{
			name: "Good signature",
			regFunc: func(s *Shaker) error {
				return s.Get("/test", goodHandler, http.StatusOK)
			},
			expectedError: nil,
		},
		{
			name: "Wrong input arguments",
			regFunc: func(s *Shaker) error {
				return s.Get("/test2", wrongHandlerInputCount, http.StatusOK)
			},
			expectedError: ErrInvalidHandlerSignature,
		},
	}

	shakerr := NewShaker(nil)

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			expected, got := test.expectedError, test.regFunc(&shakerr)
			if !errors.Is(expected, got) {
				tt.Fatalf("errors does not match : %v != %v", got, expected)
			}
		})

	}
}

func TestCallEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		method    methodType
		beforeFct func(*Shaker)
		calls     []endpointCall
	}{
		{
			name:   "No input with output",
			method: get,
			beforeFct: func(skr *Shaker) {
				type Out struct {
					Var string `json:"var"`
				}

				skr.Get("/test", func(ctx *gin.Context) (Out, error) {
					return Out{Var: "test"}, nil
				}, http.StatusOK)
			},
			calls: []endpointCall{
				{
					endpoint:       "/test",
					payload:        "test",
					expectedStatus: http.StatusOK,
					expectedPayload: `{
			"var": "test"
		}`,
				},
			},
		},
		{
			name:   "Both input and output",
			method: get,
			beforeFct: func(skr *Shaker) {
				type In struct {
					Var string `uri:"var"`
					Opt string `form:"option"`
				}

				type Out struct {
					Var    string `json:"var"`
					Option string `json:"option"`
				}

				skr.Get("/test/:var", func(ctx *Context, input *In) (Out, error) {
					return Out{Var: input.Var, Option: input.Opt}, nil
				}, http.StatusOK)
			},
			calls: []endpointCall{
				{
					endpoint:       "/test/a?option=aa",
					payload:        "test",
					expectedStatus: http.StatusOK,
					expectedPayload: `
			{
				"var": "a",
				"option": "aa"
			}`,
				},
				{
					endpoint:       "/test/b?option=bb",
					payload:        "test",
					expectedStatus: http.StatusOK,
					expectedPayload: `
			{
				"var": "b",
				"option": "bb"
			}`,
				},
				{
					endpoint:       "/test/b",
					payload:        "test",
					expectedStatus: http.StatusOK,
					expectedPayload: `
				{
					"var": "b",
					"option": ""
				}`,
				},
			},
		},
		{
			name:   "Post method",
			method: post,
			beforeFct: func(skr *Shaker) {
				type In struct {
					Var string `json:"var"`
				}

				type Out struct {
					Var string `json:"var"`
				}

				skr.Post("/test", func(ctx *gin.Context, input *In) (Out, error) {
					return Out{Var: input.Var}, nil
				}, http.StatusCreated)
			},
			calls: []endpointCall{
				{
					endpoint: "/test",
					payload: `{
			"var": "abc"
		}`,
					expectedStatus: http.StatusCreated,
					expectedPayload: `
			{
				"var": "abc"
			}`,
				},
			},
		},
		{
			name:   "Validation error",
			method: post,
			beforeFct: func(skr *Shaker) {
				type In struct {
					Var string `json:"var" binding:"required"`
				}

				type Out struct {
					Var string `json:"var"`
				}

				skr.Post("/validate", func(ctx *gin.Context, input *In) (Out, error) {
					return Out{Var: input.Var}, nil
				}, http.StatusOK)
			},
			calls: []endpointCall{
				{
					endpoint:       "/validate",
					payload:        `{}`,
					expectedStatus: http.StatusBadRequest,
					expectedPayload: `{
					"error": "Key: 'In.Var' Error:Field validation for 'Var' failed on the 'required' tag"
				}`,
				},
			},
		},
		{
			name:   "Custom error",
			method: get,
			beforeFct: func(skr *Shaker) {

				skr.Get("/sample", func(ctx *gin.Context) error {
					return ErrSample
				}, http.StatusCreated)
			},
			calls: []endpointCall{
				{
					endpoint:       "/sample",
					payload:        "test",
					expectedStatus: http.StatusExpectationFailed,
					expectedPayload: `
			{
				"error": "sample error"
			}`,
				},
			},
		},
		{
			name:   "Default error",
			method: get,
			beforeFct: func(skr *Shaker) {

				skr.Get("/defaultError", func(ctx *gin.Context) error {
					return errors.New("fancy error")
				}, http.StatusCreated)
			},
			calls: []endpointCall{
				{
					endpoint:       "/defaultError",
					payload:        "test",
					expectedStatus: http.StatusInternalServerError,
					expectedPayload: `
			{
				"error": "internal server error"
			}`,
				},
			},
		},
		{
			name:   "Validator constraints",
			method: post,
			beforeFct: func(skr *Shaker) {
				type In struct {
					Var string `json:"var" binding:"required"`
				}

				skr.Post("/validate", func(ctx *Context, in *In) error {
					return nil
				}, http.StatusOK)
			},
			calls: []endpointCall{
				{
					name:           "all is good",
					endpoint:       "/validate",
					payload:        `{"var":"ok"}`,
					expectedStatus: http.StatusOK,
				},
				{
					name:           "missing field",
					endpoint:       "/validate",
					payload:        `{}`,
					expectedStatus: http.StatusBadRequest,
				},
			},
		},
	}

	for _, test := range tests {
		shaker := NewShaker(&MappedErrors{
			ErrSample: http.StatusExpectationFailed,
		})
		testAPI := tdhttp.NewTestAPI(t, shaker.engine)
		test.beforeFct(&shaker)

		for _, call := range test.calls {
			tt := testAPI.Name(fmt.Sprintf("%s: %s", test.name, call.name))

			switch test.method {
			case get:
				tt = tt.Get(call.endpoint)
			case post:
				tt = tt.Post(call.endpoint, strings.NewReader(call.payload), "Content-Type", "application/json")
			case put:
				tt = tt.Put(call.endpoint, strings.NewReader(call.payload), "Content-Type", "application/json")
			case delete:
				tt = tt.Delete(call.endpoint, strings.NewReader(call.payload), "Content-Type", "application/json")
			}

			tt.CmpStatus(call.expectedStatus)

			if call.expectedPayload != "" {
				tt.CmpJSONBody(td.JSON(call.expectedPayload))
			}
		}
	}
}
