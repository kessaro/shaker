package shaker

import (
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

func TestShaker(t *testing.T) {
	tests := []struct {
		name               string
		method             methodType
		beforeFct          func(skr *shaker)
		endpointToCall     string
		payloadToSend      string
		expectedStatusCode int
		expectedPayload    string
	}{
		{
			name:   "No input with output",
			method: get,
			beforeFct: func(skr *shaker) {
				type Out struct {
					Var string `json:"var"`
				}

				skr.Get("/test", func(ctx *gin.Context) (Out, error) {
					return Out{Var: "test"}, nil
				}, http.StatusOK)
			},
			endpointToCall:     "/test",
			expectedStatusCode: http.StatusOK,
			expectedPayload: `{
			"var": "test"
		}`,
		},
		{
			name:   "Both input and output",
			method: get,
			beforeFct: func(skr *shaker) {
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
			endpointToCall:     "/test/a?option=aa",
			expectedStatusCode: http.StatusOK,
			expectedPayload: `
			{
				"var": "a",
				"option": "aa"
			}`,
		},
		{
			name:   "Post method",
			method: post,
			beforeFct: func(skr *shaker) {
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
			endpointToCall: "/test",
			payloadToSend: `
			{
			"var": "abc"
			}`,
			expectedStatusCode: http.StatusCreated,
			expectedPayload: `
			{
				"var": "abc"
			}`,
		},
	}

	for _, test := range tests {
		shaker := NewShaker()
		testAPI := tdhttp.NewTestAPI(t, shaker.engine)
		test.beforeFct(&shaker)

		tt := testAPI.
			Name(test.name)

		switch test.method {
		case get:
			tt = tt.Get(test.endpointToCall)
		case post:
			tt = tt.Post(test.endpointToCall, strings.NewReader(test.payloadToSend), "Content-Type", "application/json")
		case put:
			tt = tt.Put(test.endpointToCall, strings.NewReader(test.payloadToSend), "Content-Type", "application/json")
		case delete:
			tt = tt.Delete(test.endpointToCall, strings.NewReader(test.payloadToSend), "Content-Type", "application/json")
		}

		tt.CmpStatus(test.expectedStatusCode)

		if test.expectedPayload != "" {
			tt.CmpJSONBody(td.JSON(test.expectedPayload))
		}
	}
}
