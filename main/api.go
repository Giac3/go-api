package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"modules/api/pacey"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

type CreateEmbedding struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	User  string   `json:"user,omitempty"`
}
type TextFromURL struct {
	URL string `json:"url"`
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
		Url string `json:"url,omitempty"`
		B64 string `json:"b64_json,omitempty"`
	} `json:"data"`
}
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		Prompt_tokens int `json:"prompt_tokens"`
		Total_tokens  int `json:"total_tokens"`
	} `json:"usage"`
}

var API_TOKEN string = ""
var PORT int = 3033

func getData[T CreateEmbedding | GenerateImage | ChatCompletion](requestData T, endpoint string) (*http.Response, error) {
	jsonBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
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
func setDefaultValuesEmbedding(createEmbedding *CreateEmbedding) {
	if createEmbedding.Model == "" {
		createEmbedding.Model = "text-embedding-ada-002"
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	API_TOKEN = os.Getenv("OPENAI_API_KEY")
	if API_TOKEN == "" {
		log.Fatal("Unable to get API KEY, check your env")
	}
	port := os.Getenv("PORT")
	Int, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("No port in .env using 3033 as fallback")
	} else {
		PORT = Int
	}
	app := pacey.NewApp()

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

		resp, err := getData(chatCompletion, "https://api.openai.com/v1/chat/completions")
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

		resp, err := getData(generateImage, "https://api.openai.com/v1/images/generations")
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
	app.POST("/createEmbedding", func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var createEmbedding CreateEmbedding
		err = decoder.Decode(&createEmbedding)
		if err != nil {
			if err.Error() == "json: cannot unmarshal string into Go struct field CreateEmbedding.input of type []string" {
				http.Error(res, "Invalid Input: please provide a valid array of string(s) as input eg. \n{\n\t'input': ['Hello there']\n}", http.StatusBadRequest)
				return
			}
			http.Error(res, "Error decoding request body: ", http.StatusBadRequest)
			return
		}
		if len(createEmbedding.Input) == 0 || (len(createEmbedding.Input) == 1 && createEmbedding.Input[0] == "") {
			http.Error(res, "Invalid Input: please provide a valid array of string(s) as input eg. \n{\n\t'input': ['Hello there']\n}", http.StatusBadRequest)
			return
		}
		setDefaultValuesEmbedding(&createEmbedding)

		resp, err := getData(createEmbedding, "https://api.openai.com/v1/embeddings")
		if err != nil {
			http.Error(res, "Error Creating Embeddings", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		var aiRes OpenAIEmbeddingResponse
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

		http.Error(res, "No Embeddings in OpenAI response", http.StatusInternalServerError)
	})

	app.POST("/getTextFromURL", func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var textFromUrl TextFromURL
		err = decoder.Decode(&textFromUrl)
		if err != nil {
			http.Error(res, "Invalid request params", http.StatusBadRequest)
			return
		}
		if textFromUrl.URL == "" {
			http.Error(res, "Must pass a valid", http.StatusBadRequest)
			return
		}
		cmd := exec.Command("lynx", "--dump", textFromUrl.URL)
		text, _ := cmd.CombinedOutput()
		var response = struct {
			Text string `json:"text"`
		}{string(text)}

		if strings.Contains(string(text), "\nCan't Access `file") {
			http.Error(res, "Nope", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(response)
	})

	app.GoLive(PORT)
}
