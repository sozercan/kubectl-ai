package gpt3_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/PullRequestInc/go-gpt3"
	fakes "github.com/PullRequestInc/go-gpt3/go-gpt3fakes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 net/http.RoundTripper

func TestInitNewClient(t *testing.T) {
	client := gpt3.NewClient("test-key")
	assert.NotNil(t, client)
}

func fakeHTTPClient() (*fakes.FakeRoundTripper, *http.Client) {
	rt := &fakes.FakeRoundTripper{}
	return rt, &http.Client{
		Transport: rt,
	}
}

func TestRequestCreationFails(t *testing.T) {
	ctx := context.Background()
	rt, httpClient := fakeHTTPClient()
	client := gpt3.NewClient("test-key", gpt3.WithHTTPClient(httpClient))
	rt.RoundTripReturns(nil, errors.New("request error"))

	type testCase struct {
		name        string
		apiCall     func() (interface{}, error)
		errorString string
	}

	testCases := []testCase{
		{
			"Engines",
			func() (interface{}, error) {
				return client.Engines(ctx)
			},
			"Get \"https://api.openai.com/v1/engines\": request error",
		},
		{
			"Engine",
			func() (interface{}, error) {
				return client.Engine(ctx, gpt3.DefaultEngine)
			},
			"Get \"https://api.openai.com/v1/engines/davinci\": request error",
		},
		{
			"Completion",
			func() (interface{}, error) {
				return client.Completion(ctx, gpt3.CompletionRequest{})
			},
			"Post \"https://api.openai.com/v1/engines/davinci/completions\": request error",
		},
		{
			"CompletionStream",
			func() (interface{}, error) {
				var rsp *gpt3.CompletionResponse
				onData := func(data *gpt3.CompletionResponse) {
					rsp = data
				}
				return rsp, client.CompletionStream(ctx, gpt3.CompletionRequest{}, onData)
			},
			"Post \"https://api.openai.com/v1/engines/davinci/completions\": request error",
		},
		{
			"CompletionWithEngine",
			func() (interface{}, error) {
				return client.CompletionWithEngine(ctx, gpt3.AdaEngine, gpt3.CompletionRequest{})
			},
			"Post \"https://api.openai.com/v1/engines/ada/completions\": request error",
		},
		{
			"CompletionStreamWithEngine",
			func() (interface{}, error) {
				var rsp *gpt3.CompletionResponse
				onData := func(data *gpt3.CompletionResponse) {
					rsp = data
				}
				return rsp, client.CompletionStreamWithEngine(ctx, gpt3.AdaEngine, gpt3.CompletionRequest{}, onData)
			},
			"Post \"https://api.openai.com/v1/engines/ada/completions\": request error",
		},
		{
			"Edits",
			func() (interface{}, error) {
				return client.Edits(ctx, gpt3.EditsRequest{})
			},
			"Post \"https://api.openai.com/v1/edits\": request error",
		},
		{
			"Search",
			func() (interface{}, error) {
				return client.Search(ctx, gpt3.SearchRequest{})
			},
			"Post \"https://api.openai.com/v1/engines/davinci/search\": request error",
		},
		{
			"SearchWithEngine",
			func() (interface{}, error) {
				return client.SearchWithEngine(ctx, gpt3.AdaEngine, gpt3.SearchRequest{})
			},
			"Post \"https://api.openai.com/v1/engines/ada/search\": request error",
		},
		{
			"Embeddings",
			func() (interface{}, error) {
				return client.Embeddings(ctx, gpt3.EmbeddingsRequest{})
			},
			"Post \"https://api.openai.com/v1/embeddings\": request error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rsp, err := tc.apiCall()
			assert.EqualError(t, err, tc.errorString)
			assert.Nil(t, rsp)
		})
	}
}

type errReader int

func (errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestResponses(t *testing.T) {
	ctx := context.Background()
	rt, httpClient := fakeHTTPClient()
	client := gpt3.NewClient("test-key", gpt3.WithHTTPClient(httpClient))

	type testCase struct {
		name           string
		apiCall        func() (interface{}, error)
		responseObject interface{}
	}

	testCases := []testCase{
		{
			"Engines",
			func() (interface{}, error) {
				return client.Engines(ctx)
			},
			&gpt3.EnginesResponse{
				Data: []gpt3.EngineObject{
					{
						ID:     "123",
						Object: "list",
						Owner:  "owner",
						Ready:  true,
					},
				},
			},
		},
		{
			"Engine",
			func() (interface{}, error) {
				return client.Engine(ctx, gpt3.DefaultEngine)
			},
			&gpt3.EngineObject{
				ID:     "123",
				Object: "list",
				Owner:  "owner",
				Ready:  true,
			},
		},
		{
			"Completion",
			func() (interface{}, error) {
				return client.Completion(ctx, gpt3.CompletionRequest{})
			},
			&gpt3.CompletionResponse{
				ID:      "123",
				Object:  "list",
				Created: 123456789,
				Model:   "davinci-12",
				Choices: []gpt3.CompletionResponseChoice{
					{
						Text:         "output",
						FinishReason: "stop",
					},
				},
			},
		},
		{
			"CompletionStream",
			func() (interface{}, error) {
				var rsp *gpt3.CompletionResponse
				onData := func(data *gpt3.CompletionResponse) {
					rsp = data
				}
				return rsp, client.CompletionStream(ctx, gpt3.CompletionRequest{}, onData)
			},
			nil, // streaming responses are tested separately
		},
		{
			"CompletionWithEngine",
			func() (interface{}, error) {
				return client.CompletionWithEngine(ctx, gpt3.AdaEngine, gpt3.CompletionRequest{})
			},
			&gpt3.CompletionResponse{
				ID:      "123",
				Object:  "list",
				Created: 123456789,
				Model:   "davinci-12",
				Choices: []gpt3.CompletionResponseChoice{
					{
						Text:         "output",
						FinishReason: "stop",
					},
				},
			},
		},
		{
			"CompletionStreamWithEngine",
			func() (interface{}, error) {
				var rsp *gpt3.CompletionResponse
				onData := func(data *gpt3.CompletionResponse) {
					rsp = data
				}
				return rsp, client.CompletionStreamWithEngine(ctx, gpt3.AdaEngine, gpt3.CompletionRequest{}, onData)
			},
			nil, // streaming responses are tested separately
		},
		{
			"Search",
			func() (interface{}, error) {
				return client.Search(ctx, gpt3.SearchRequest{})
			},
			&gpt3.SearchResponse{
				Data: []gpt3.SearchData{
					{
						Document: 1,
						Object:   "search_result",
						Score:    40.312,
					},
				},
			},
		},
		{
			"SearchWithEngine",
			func() (interface{}, error) {
				return client.SearchWithEngine(ctx, gpt3.AdaEngine, gpt3.SearchRequest{})
			},
			&gpt3.SearchResponse{
				Data: []gpt3.SearchData{
					{
						Document: 1,
						Object:   "search_result",
						Score:    40.312,
					},
				},
			},
		},
		{
			"Embeddings",
			func() (interface{}, error) {
				return client.Embeddings(ctx, gpt3.EmbeddingsRequest{})
			},
			&gpt3.EmbeddingsResponse{
				Object: "list",
				Data: []gpt3.EmbeddingsResult{{
					Object:    "object",
					Embedding: []float64{0.1, 0.2, 0.3},
					Index:     0,
				}},
				Usage: gpt3.EmbeddingsUsage{
					PromptTokens: 1,
					TotalTokens:  2,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("bad status codes", func(t *testing.T) {
				for _, code := range []int{400, 401, 404, 422, 500} {
					// first mock with error with body failure
					mockResponse := &http.Response{
						StatusCode: code,
						Body:       io.NopCloser(errReader(0)),
					}

					rt.RoundTripReturns(mockResponse, nil)
					rsp, err := tc.apiCall()
					assert.Nil(t, rsp)
					assert.EqualError(t, err, "failed to read from body: read error")

					// then mock with an unknown error string
					mockResponse = &http.Response{
						StatusCode: code,
						Body:       io.NopCloser(bytes.NewBufferString("unknown error")),
					}

					rt.RoundTripReturns(mockResponse, nil)
					rsp, err = tc.apiCall()
					assert.Nil(t, rsp)
					assert.EqualError(t, err, fmt.Sprintf("[%d:Unexpected] unknown error", code))

					// then mock with an json APIErrorResponse
					apiErrorResponse := &gpt3.APIErrorResponse{
						Error: gpt3.APIError{
							Type:    "test_type",
							Message: "test message",
						},
					}

					data, err := json.Marshal(apiErrorResponse)
					assert.NoError(t, err)

					mockResponse = &http.Response{
						StatusCode: code,
						Body:       io.NopCloser(bytes.NewBuffer(data)),
					}

					rt.RoundTripReturns(mockResponse, nil)
					rsp, err = tc.apiCall()
					assert.Nil(t, rsp)
					assert.EqualError(t, err, fmt.Sprintf("[%d:test_type] test message", code))
					apiErrorResponse.Error.StatusCode = code
					assert.Equal(t, apiErrorResponse.Error, err)
				}
			})
			t.Run("success code json decode failure", func(t *testing.T) {
				mockResponse := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
				}

				rt.RoundTripReturns(mockResponse, nil)

				rsp, err := tc.apiCall()
				assert.Error(t, err, "invalid json response: invalid character 'i' looking for beginning of value")
				assert.Nil(t, rsp)
			})
			// skip streaming/nil response objects here as those will be tested separately
			if tc.responseObject != nil {
				t.Run("successful response", func(t *testing.T) {
					data, err := json.Marshal(tc.responseObject)
					assert.NoError(t, err)

					mockResponse := &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBuffer(data)),
					}

					rt.RoundTripReturns(mockResponse, nil)

					rsp, err := tc.apiCall()
					assert.NoError(t, err)
					assert.Equal(t, tc.responseObject, rsp)
				})
			}
		})
	}
}

// TODO: add streaming response tests
