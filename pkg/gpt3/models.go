package gpt3

import "fmt"

// APIError represents an error that occurred on an API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("[%d:%s] %s", e.StatusCode, e.Type, e.Message)
}

// APIErrorResponse is the full error respnose that has been returned by an API.
type APIErrorResponse struct {
	Error APIError `json:"error"`
}

// EngineObject contained in an engine response.
type EngineObject struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Owner  string `json:"owner"`
	Ready  bool   `json:"ready"`
}

// EnginesResponse is returned from the Engines API.
type EnginesResponse struct {
	Data   []EngineObject `json:"data"`
	Object string         `json:"object"`
}

// ChatCompletionRequestMessage is a message to use as the context for the chat completion API.
type ChatCompletionRequestMessage struct {
	// Role is the role is the role of the the message. Can be "system", "user", or "assistant"
	Role string `json:"role"`

	// Content is the content of the message
	Content string `json:"content"`
}

// ChatCompletionRequest is a request for the chat completion API.
type ChatCompletionRequest struct {
	// Model is the name of the model to use. If not specified, will default to gpt-3.5-turbo.
	Model string `json:"model"`

	// Messages is a list of messages to use as the context for the chat completion.
	Messages []ChatCompletionRequestMessage `json:"messages"`

	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic
	Temperature *float32 `json:"temperature,omitempty"`

	// An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP float32 `json:"top_p,omitempty"`

	// Number of responses to generate
	N int `json:"n,omitempty"`

	// Whether or not to stream responses back as they are generated
	Stream bool `json:"stream,omitempty"`

	// Up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`

	// MaxTokens is the maximum number of tokens to return.
	MaxTokens int `json:"max_tokens,omitempty"`

	// (-2, 2) Penalize tokens that haven't appeared yet in the history.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// (-2, 2) Penalize tokens that appear too frequently in the history.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// Modify the probability of specific tokens appearing in the completion.
	LogitBias map[string]float32 `json:"logit_bias,omitempty"`

	// Can be used to identify an end-user
	User string `json:"user,omitempty"`
}

// CompletionRequest is a request for the completions API.
type CompletionRequest struct {
	// A list of string prompts to use.
	// TODO there are other prompt types here for using token integers that we could add support for.
	Prompt []string `json:"prompt"`
	// How many tokens to complete up to. Max of 512
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Sampling temperature to use
	Temperature *float32 `json:"temperature,omitempty"`
	// Alternative to temperature for nucleus sampling
	TopP *float32 `json:"top_p,omitempty"`
	// How many choice to create for each prompt
	N *int `json:"n"`
	// Include the probabilities of most likely tokens
	LogProbs *int `json:"logprobs,omitempty"`
	// Echo back the prompt in addition to the completion
	Echo bool `json:"echo,omitempty"`
	// Up to 4 sequences where the API will stop generating tokens. Response will not contain the stop sequence.
	Stop []string `json:"stop,omitempty"`
	// PresencePenalty number between 0 and 1 that penalizes tokens that have already appeared in the text so far.
	PresencePenalty float32 `json:"presence_penalty"`
	// FrequencyPenalty number between 0 and 1 that penalizes tokens on existing frequency in the text so far.
	FrequencyPenalty float32 `json:"frequency_penalty"`

	// Whether to stream back results or not. Don't set this value in the request yourself
	// as it will be overridden depending on if you use CompletionStream or Completion methods.
	Stream bool `json:"stream,omitempty"`
}

// EditsRequest is a request for the edits API.
type EditsRequest struct {
	// ID of the model to use. You can use the List models API to see all of your available models, or see our Model overview for descriptions of them.
	Model string `json:"model"`
	// The input text to use as a starting point for the edit.
	Input string `json:"input"`
	// The instruction that tells the model how to edit the prompt.
	Instruction string `json:"instruction"`
	// Sampling temperature to use
	Temperature *float32 `json:"temperature,omitempty"`
	// Alternative to temperature for nucleus sampling
	TopP *float32 `json:"top_p,omitempty"`
	// How many edits to generate for the input and instruction. Defaults to 1
	N *int `json:"n"`
}

// EmbeddingsRequest is a request for the Embeddings API.
type EmbeddingsRequest struct {
	// Input text to get embeddings for, encoded as a string or array of tokens. To get embeddings
	// for multiple inputs in a single request, pass an array of strings or array of token arrays.
	// Each input must not exceed 2048 tokens in length.
	Input []string `json:"input"`
	// ID of the model to use
	Model string `json:"model"`
	// The request user is an optional parameter meant to be used to trace abusive requests
	// back to the originating user. OpenAI states:
	// "The [user] IDs should be a string that uniquely identifies each user. We recommend hashing
	// their username or email address, in order to avoid sending us any identifying information.
	// If you offer a preview of your product to non-logged in users, you can send a session ID
	// instead."
	User string `json:"user,omitempty"`
}

// ChatCompletionResponseMessage is a message returned in the response to the Chat Completions API.
type ChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponseChoice is one of the choices returned in the response to the Chat Completions API.
type ChatCompletionResponseChoice struct {
	Index        int                           `json:"index"`
	FinishReason string                        `json:"finish_reason"`
	Message      ChatCompletionResponseMessage `json:"message"`
}

// ChatCompletionsResponseUsage is the object that returns how many tokens the completion's request used.
type ChatCompletionsResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is the full response from a request to the Chat Completions API.
type ChatCompletionResponse struct {
	ID      string                         `json:"id"`
	Object  string                         `json:"object"`
	Created int                            `json:"created"`
	Model   string                         `json:"model"`
	Choices []ChatCompletionResponseChoice `json:"choices"`
	Usage   ChatCompletionsResponseUsage   `json:"usage"`
}

// LogprobResult represents logprob result of Choice.
type LogprobResult struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float32            `json:"token_logprobs"`
	TopLogprobs   []map[string]float32 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// CompletionResponseChoice is one of the choices returned in the response to the Completions API.
type CompletionResponseChoice struct {
	Text         string        `json:"text"`
	Index        int           `json:"index"`
	LogProbs     LogprobResult `json:"logprobs"`
	FinishReason string        `json:"finish_reason"`
}

// CompletionResponse is the full response from a request to the completions API.
type CompletionResponse struct {
	ID      string                     `json:"id"`
	Object  string                     `json:"object"`
	Created int                        `json:"created"`
	Model   string                     `json:"model"`
	Choices []CompletionResponseChoice `json:"choices"`
	Usage   CompletionResponseUsage    `json:"usage"`
}

// CompletionResponseUsage is the object that returns how many tokens the completion's request used.
type CompletionResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EditsResponse is the full response from a request to the edits API.
type EditsResponse struct {
	Object  string                `json:"object"`
	Created int                   `json:"created"`
	Choices []EditsResponseChoice `json:"choices"`
	Usage   EditsResponseUsage    `json:"usage"`
}

// The inner result of a create embeddings request, containing the embeddings for a single input.
type EmbeddingsResult struct {
	// The type of object returned (e.g., "list", "object")
	Object string `json:"object"`
	// The embedding data for the input
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// The usage stats for an embeddings response.
type EmbeddingsUsage struct {
	// The number of tokens used by the prompt
	PromptTokens int `json:"prompt_tokens"`
	// The total tokens used
	TotalTokens int `json:"total_tokens"`
}

// EmbeddingsResponse is the response from a create embeddings request.
//
// See: https://beta.openai.com/docs/api-reference/embeddings/create
type EmbeddingsResponse struct {
	Object string             `json:"object"`
	Data   []EmbeddingsResult `json:"data"`
	Usage  EmbeddingsUsage    `json:"usage"`
}

// EditsResponseChoice is one of the choices returned in the response to the Edits API.
type EditsResponseChoice struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// EditsResponseUsage is a structure used in the response from a request to the edits API.
type EditsResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// SearchRequest is a request for the document search API.
type SearchRequest struct {
	Documents []string `json:"documents"`
	Query     string   `json:"query"`
}

// SearchData is a single search result from the document search API.
type SearchData struct {
	Document int     `json:"document"`
	Object   string  `json:"object"`
	Score    float64 `json:"score"`
}

// SearchResponse is the full response from a request to the document search API.
type SearchResponse struct {
	Data   []SearchData `json:"data"`
	Object string       `json:"object"`
}
