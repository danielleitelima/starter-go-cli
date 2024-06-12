package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ResponsePayload struct {
	Response string `json:"response"`
}

type TranslationPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type TranslationResponse struct {
	Translation string `json:"response"`
}

type ResultItem struct {
	Source      string `json:"source"`
	Translation string `json:"translation"`
}

var analiseCmd = &cobra.Command{
	Use:   "analise [text]",
	Short: "Analyze and output the words in JSON format",
	Long: `The "analise" command takes a string of text as an argument, sends it to an Ollama instance for processing, using the llama3 model, and outputs the result in JSON format.
Optionally, you can specify the Ollama instance URL and the translation language locale.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]

		llmHost, err := cmd.Flags().GetString("llm-host")
		if err != nil {
			fmt.Println("Error retrieving llm-host flag:", err)
			os.Exit(1)
		}
		if llmHost == "" {
			llmHost = os.Getenv("STARTER_GO_CLI_LLM_HOST")
			if llmHost == "" {
				llmHost = "http://localhost:11434/api/generate"
				fmt.Println("Using default Ollama host:", llmHost)
			}
		}

		translationLanguage, err := cmd.Flags().GetString("translation-language")
		if err != nil {
			fmt.Println("Error retrieving language flag:", err)
			os.Exit(1)
		}
		if translationLanguage == "" {
			translationLanguage = os.Getenv("STARTER_GO_CLI_TRANSLATION_LANGUAGE")
			if translationLanguage == "" {
				translationLanguage = "en-US"
				fmt.Println("Using default translation language:", translationLanguage)
			}
		}

		prompt := fmt.Sprintf("Divide the text below into small sections, each representing a particular thought or idea. Use grammar as a basis and avoid creating a section with a single word. You can break a phrase into subject and predicate.\n\nExample text:\n\nHey, kannst du mir den heutigen Mittagsmenü schicken? Ich bin gerade total eingebunden bei der Arbeit und schaffe es nicht reinzukommen.\n\nExample output:\n\n[\n    \"Hey\",\n    \"kannst du mir\",\n    \"den heutigen Mittagsmenü schicken?\",\n    \"Ich bin gerade\",\n    \"total eingebunden\",\n    \"bei der Arbeit\",\n    \"und\",\n    \"schaffe es nicht reinzukommen.\"\n]\n\nActual text:\n\n%s\n\nActual output:\n\nProvide only the JSON array as the output without any additional text or explanation.", text)

		payload := RequestPayload{
			Model:  "llama3",
			Prompt: prompt,
			Stream: false,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Error marshalling request payload:", err)
			os.Exit(1)
		}

		resp, err := http.Post(llmHost, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: received status code %d\n", resp.StatusCode)
			os.Exit(1)
		}

		var responsePayload ResponsePayload
		err = json.Unmarshal(body, &responsePayload)
		if err != nil {
			fmt.Println("Error parsing JSON response:", err)
			os.Exit(1)
		}

		var sections []string
		err = json.Unmarshal([]byte(responsePayload.Response), &sections)
		if err != nil {
			fmt.Println("Error parsing response array:", err)
			os.Exit(1)
		}

		var results []ResultItem

		for _, section := range sections {
			translationPrompt := fmt.Sprintf("Translate the following text to %s:\n\n%s\n\nProvide only the translation without any additional text or explanation.", translationLanguage, section)
			translationPayload := TranslationPayload{
				Model:  "llama3",
				Prompt: translationPrompt,
				Stream: false,
			}

			translationPayloadBytes, err := json.Marshal(translationPayload)
			if err != nil {
				fmt.Println("Error marshalling translation request payload:", err)
				os.Exit(1)
			}

			translationResp, err := http.Post(llmHost, "application/json", bytes.NewBuffer(translationPayloadBytes))
			if err != nil {
				fmt.Println("Error making HTTP request for translation:", err)
				os.Exit(1)
			}
			defer translationResp.Body.Close()

			translationBody, err := ioutil.ReadAll(translationResp.Body)
			if err != nil {
				fmt.Println("Error reading translation response body:", err)
				os.Exit(1)
			}

			if translationResp.StatusCode != http.StatusOK {
				fmt.Printf("Error: received status code %d for translation\n", translationResp.StatusCode)
				os.Exit(1)
			}

			var translationResponse TranslationResponse
			err = json.Unmarshal(translationBody, &translationResponse)
			if err != nil {
				fmt.Println("Error parsing translation JSON response:", err)
				os.Exit(1)
			}

			result := ResultItem{
				Source:      section,
				Translation: translationResponse.Translation,
			}
			results = append(results, result)
		}

		resultsJSON, err := json.MarshalIndent(results, "", "    ")
		if err != nil {
			fmt.Println("Error marshalling final results to JSON:", err)
			os.Exit(1)
		}

		fmt.Println(string(resultsJSON))
	},
}

func init() {
	analiseCmd.Flags().StringP("llm-host", "l", "", "The Ollama host URL for the LLM service (default is 'http://localhost:11434/api/generate')")
	analiseCmd.Flags().StringP("translation-language", "t", "", "The language for translation in locale format (default is 'en-US')")

	rootCmd.AddCommand(analiseCmd)
}
