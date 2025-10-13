package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &strings.Builder{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("fileToUpload", file.Name())
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "https://catbox.moe/user/api.php", strings.NewReader(body.String()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(respData), nil
}
