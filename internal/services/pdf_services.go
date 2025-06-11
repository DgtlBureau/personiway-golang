package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var laravel_api string = os.Getenv("URL_LARAVEL")

func Convert(file multipart.File) {
	tempDir := os.TempDir()
	log.Printf("Using temp directory: %s", tempDir)

	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		log.Printf("failed to create temp directory: %v", err)
		return
	}

	tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("input_%d_*.pdf", time.Now().UnixNano()))
	if err != nil {
		log.Printf("failed to create temp file: %v", err)
		return
	}
	tempPDF := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPDF)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("failed to read file: %v", err)
		return
	}

	err = os.WriteFile(tempPDF, fileBytes, 0644)
	if err != nil {
		log.Printf("failed to write temp PDF: %v", err)
		return
	}

	err = convertPDFToPNG(tempPDF)
	if err != nil {
		log.Printf("failed to convert PDF: %v", err)
		return
	}
}

func convertPDFToPNG(pdfPath string) error {
	baseName := fmt.Sprintf("page_%d", time.Now().UnixNano())

	cmd := exec.Command("pdftoppm",
		"-png",
		"-r", "300",
		pdfPath,
		baseName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("pdftoppm error: %v, output: %s", err, string(output))
		return err
	}

	log.Printf("Successfully converted PDF to images")

	return processGeneratedImages(baseName)
}

func processGeneratedImages(baseName string) error {
	pattern := fmt.Sprintf("%s-*.png", baseName)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error finding generated images: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no images were generated")
	}

	log.Printf("Found %d generated images", len(files))

	for i, file := range files {
		log.Printf("Processing image: %s", file)

		imageData, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading %s: %v", file, err)
			continue
		}

		go func(data []byte, pageNum int, filename string) {
			defer os.Remove(filename)

			result := sendToAntropic(data)

			go func() {
				var requestBody bytes.Buffer
				writer := multipart.NewWriter(&requestBody)

				fileWriter, _ := writer.CreateFormFile("file", "page.png")
				fileWriter.Write(data)

				writer.WriteField("json", result)

				placeIdsJSON, _ := json.Marshal([]int{2})
				writer.WriteField("place_ids", string(placeIdsJSON))

				writer.Close()

				req, _ := http.NewRequest("POST", laravel_api+"admin/save-pdf-schemas", &requestBody)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				client := &http.Client{}
				client.Do(req)
			}()

		}(imageData, i, file)
	}

	return nil
}

func sendToAntropic(imageData []byte) string {
	apiKey := os.Getenv("ANTROPIC_API_KEY")
	if apiKey == "" {
		log.Printf("ANTHROPIC_API_KEY not set")
		return ""
	}

	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	request := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 4000,
		"system":     "Please return only the JSON object in the response, without introduction, conclusion, or explanation. The response must contain only valid JSON for php json_decode() function.",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": user_prompt,
					},
					{
						"type": "image",
						"source": map[string]string{
							"type":       "base64",
							"media_type": "image/png",
							"data":       imageBase64,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error marshaling: %v", err)
		return ""
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("API Error %d: %s", resp.StatusCode, string(body))
		return ""
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return ""
	}

	if len(response.Content) > 0 {
		return response.Content[0].Text
	}

	return ""
}
