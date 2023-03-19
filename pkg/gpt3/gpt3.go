package gpt3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultApiVersion     = "2022-12-01"
	defaultUserAgent      = "openai-slack-example"
	defaultTimeoutSeconds = 30
)

// A Client is an API client to communicate with the OpenAI gpt-3 APIs
type Client interface {
	// Completion creates a completion with the default engine. This is the main endpoint of the API
	// which auto-completes based on the given prompt.
	Completion(ctx context.Context, request CompletionRequest) (*CompletionResponse, error)

	// CompletionStream creates a completion with the default engine and streams the results through
	// multiple calls to onData.
	CompletionStream(ctx context.Context, request CompletionRequest, onData func(*CompletionResponse)) error

	// Given a prompt and an instruction, the model will return an edited version of the prompt.
	Edits(ctx context.Context, request EditsRequest) (*EditsResponse, error)

	// Search performs a semantic search over a list of documents with the default engine.
	Search(ctx context.Context, request SearchRequest) (*SearchResponse, error)

	// Returns an embedding using the provided request.
	Embeddings(ctx context.Context, request EmbeddingsRequest) (*EmbeddingsResponse, error)
}

type client struct {
	endpoint       string
	apiKey         string
	deploymentName string
	apiVersion     string
	userAgent      string
	httpClient     *http.Client
}

// NewClient returns a new OpenAI GPT-3 API client. An apiKey is required to use the client
func NewClient(endpoint string, apiKey string, deploymentName string, options ...ClientOption) Client {
	httpClient := &http.Client{
		Timeout: time.Duration(defaultTimeoutSeconds * time.Second),
	}

	c := &client{
		endpoint:       endpoint,
		apiKey:         apiKey,
		deploymentName: deploymentName,
		apiVersion:     defaultApiVersion,
		userAgent:      defaultUserAgent,
		httpClient:     httpClient,
	}
	for _, o := range options {
		o(c)
	}
	return c
}

func (c *client) Completion(ctx context.Context, request CompletionRequest) (*CompletionResponse, error) {
	request.Stream = false
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/openai/deployments/%s/completions", c.deploymentName), request)
	if err != nil {
		return nil, err
	}
	resp, err := c.performRequest(req)
	if err != nil {
		return nil, err
	}

	output := new(CompletionResponse)
	if err := getResponseObject(resp, output); err != nil {
		return nil, err
	}
	return output, nil
}

var dataPrefix = []byte("data: ")
var doneSequence = []byte("[DONE]")

func (c *client) CompletionStream(ctx context.Context, request CompletionRequest, onData func(*CompletionResponse)) error {
	request.Stream = true
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/openai/deployments/%s/completions", c.deploymentName), request)
	if err != nil {
		return err
	}
	resp, err := c.performRequest(req)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		// make sure there isn't any extra whitespace before or after
		line = bytes.TrimSpace(line)
		// the completion API only returns data events
		if !bytes.HasPrefix(line, dataPrefix) {
			continue
		}
		line = bytes.TrimPrefix(line, dataPrefix)

		// the stream is completed when terminated by [DONE]
		if bytes.HasPrefix(line, doneSequence) {
			break
		}
		output := new(CompletionResponse)
		if err := json.Unmarshal(line, output); err != nil {
			return fmt.Errorf("invalid json stream data: %v", err)
		}
		onData(output)
	}

	return nil
}

func (c *client) Edits(ctx context.Context, request EditsRequest) (*EditsResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/edits", request)
	if err != nil {
		return nil, err
	}
	resp, err := c.performRequest(req)
	if err != nil {
		return nil, err
	}

	output := new(EditsResponse)
	if err := getResponseObject(resp, output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *client) Search(ctx context.Context, request SearchRequest) (*SearchResponse, error) {
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/openai/deployments/%s/search", c.deploymentName), request)
	if err != nil {
		return nil, err
	}
	resp, err := c.performRequest(req)
	if err != nil {
		return nil, err
	}
	output := new(SearchResponse)
	if err := getResponseObject(resp, output); err != nil {
		return nil, err
	}
	return output, nil
}

// Embeddings creates text embeddings for a supplied slice of inputs with a provided model.
//
// See: https://beta.openai.com/docs/api-reference/embeddings
func (c *client) Embeddings(ctx context.Context, request EmbeddingsRequest) (*EmbeddingsResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/embeddings", request)
	if err != nil {
		return nil, err
	}
	resp, err := c.performRequest(req)
	if err != nil {
		return nil, err
	}

	output := EmbeddingsResponse{}
	if err := getResponseObject(resp, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (c *client) performRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := checkForSuccess(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// returns an error if this response includes an error.
func checkForSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read from body: %w", err)
	}
	var result APIErrorResponse
	if err := json.Unmarshal(data, &result); err != nil {
		// if we can't decode the json error then create an unexpected error
		apiError := APIError{
			StatusCode: resp.StatusCode,
			Type:       "Unexpected",
			Message:    string(data),
		}
		return apiError
	}
	result.Error.StatusCode = resp.StatusCode
	return result.Error
}

func getResponseObject(rsp *http.Response, v interface{}) error {
	defer rsp.Body.Close()
	if err := json.NewDecoder(rsp.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid json response: %w", err)
	}
	return nil
}

func jsonBodyReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return bytes.NewBuffer(nil), nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed encoding json: %w", err)
	}
	return bytes.NewBuffer(raw), nil
}

func (c *client) newRequest(ctx context.Context, method, path string, payload interface{}) (*http.Request, error) {
	bodyReader, err := jsonBodyReader(payload)
	if err != nil {
		return nil, err
	}
	reqUrl := fmt.Sprintf("%s%s?api-version=%s", c.endpoint, path, c.apiVersion)
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("api-key", c.apiKey)
	return req, nil
}
