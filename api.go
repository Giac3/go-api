package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var imageSizes = map[string]struct{}{
	"256x256":   {},
	"512x512":   {},
	"1024x1024": {},
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletion struct {
	Messages          []Message `json:"messages"`
	Model             string    `json:"model"`
	Temperature       *float32  `json:"temperature,omitempty"`
	Top_p             *float32  `json:"top_p,omitempty"`
	Max_tokens        *int      `json:"max_tokens,omitempty"`
	Frequency_penalty float32   `json:"frequency_penalty,omitempty"`
	Presence_penalty  float32   `json:"presence_penalty,omitempty"`
}

type GenerateImage struct {
	Prompt          string `json:"prompt"`
	N               int    `json:"n,omitempty"`
	Size            string `json:"size,omitempty"`
	Response_format string `json:"response_format,omitempty"`
	User            string `json:"user,omitempty"`
}

type OpenAIChatResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Finish_reason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		Prompt_tokens     int `json:"prompt_tokens"`
		Completion_tokens int `json:"completion_tokens"`
		Total_tokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenAIImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		Url string `json:"url"`
	}
}

var API_TOKEN string = ""

func getChatCompletion(chatCompletion ChatCompletion) (*http.Response, error) {
	jsonBody, err := json.Marshal(chatCompletion)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+API_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getImageGeneration(generateImage GenerateImage) (*http.Response, error) {
	jsonBody, err := json.Marshal(generateImage)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+API_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func setDefaultValuesChat(chatCompletion *ChatCompletion) {
	if chatCompletion.Model == "" {
		chatCompletion.Model = "gpt-3.5-turbo"
	}

	if chatCompletion.Temperature == nil {
		temp := float32(0.1)
		chatCompletion.Temperature = &temp
	}
	if chatCompletion.Top_p == nil {
		temp := float32(1)
		chatCompletion.Top_p = &temp
	}
	if chatCompletion.Max_tokens == nil {
		temp := int(1000)
		chatCompletion.Max_tokens = &temp
	}
}
func setDefaultValuesImage(generateImage *GenerateImage) {
	if generateImage.Prompt == "" {
		generateImage.Prompt = "something random"
	}
	if generateImage.Response_format == "" {
		generateImage.Response_format = "url"
	}
	if generateImage.N == 0 {
		generateImage.N = 1
	}
	_, validSize := imageSizes[generateImage.Size]

	if generateImage.Size == "" || !validSize {
		generateImage.Size = "256x256"
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	API_TOKEN = os.Getenv("OPENAI_API_KEY")
	app := NewApp()

	app.GET("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello, go to /docs to check out the endpoints")
	})
	app.GET("/docs", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello docs are coming soon")
	})

	app.POST("/chatCompletion", func(res http.ResponseWriter, req *http.Request) {

		decoder := json.NewDecoder(req.Body)
		var chatCompletion ChatCompletion
		err := decoder.Decode(&chatCompletion)
		if err != nil {
			http.Error(res, "Error decoding request body", http.StatusBadRequest)
			return
		}

		setDefaultValuesChat(&chatCompletion)

		if len(chatCompletion.Messages) == 0 {
			http.Error(res, "Error: At least one message is required", http.StatusBadRequest)
			return
		}

		resp, err := getChatCompletion(chatCompletion)
		if err != nil {
			http.Error(res, "Error getting chat completion", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		var aiRes OpenAIChatResponse
		err = json.NewDecoder(resp.Body).Decode(&aiRes)
		if err != nil {
			http.Error(res, "Error reading OpenAI response", http.StatusInternalServerError)
			return
		}

		if len(aiRes.Choices) > 0 {
			res.Header().Set("Content-Type", "application/json")
			json.NewEncoder(res).Encode(aiRes.Choices[0].Message)
			return
		}

		http.Error(res, "No choices in OpenAI response", http.StatusInternalServerError)
	})

	app.POST("/generateImage", func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var generateImage GenerateImage
		decoder.Decode(&generateImage)
		setDefaultValuesImage(&generateImage)

		resp, err := getImageGeneration(generateImage)
		if err != nil {
			http.Error(res, "Error generating Image", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var aiRes OpenAIImageResponse
		err = json.NewDecoder(resp.Body).Decode(&aiRes)
		if err != nil {
			http.Error(res, "Error reading OpenAI response", http.StatusInternalServerError)
			return
		}
		if len(aiRes.Data) > 0 {
			res.Header().Set("Content-Type", "application/json")
			json.NewEncoder(res).Encode(aiRes.Data)
			return
		}

		http.Error(res, "No Images in OpenAI response", http.StatusInternalServerError)
	})

	fmt.Println("Server is live on localhost:3033")
	log.Fatal(http.ListenAndServe(":3033", app))
}
