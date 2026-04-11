package file_storage

import (
	"context"
	"douyin-backend/app/global/variable"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func UseCOSStorage() bool {
	return strings.EqualFold(strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.StorageDriver")), "cos")
}

func DeletePublicResource(resourceURL string) error {
	resourceURL = strings.TrimSpace(resourceURL)
	if resourceURL == "" {
		return nil
	}

	if UseCOSStorage() {
		return deleteCOSResource(resourceURL)
	}
	return deleteLocalResource(resourceURL)
}

func deleteCOSResource(resourceURL string) error {
	objectKey, ok := ExtractCOSObjectKey(resourceURL)
	if !ok || objectKey == "" {
		return nil
	}

	client, err := NewCOSClient()
	if err != nil {
		return err
	}

	_, err = client.Object.Delete(context.Background(), objectKey)
	return err
}

func deleteLocalResource(resourceURL string) error {
	sourcePrefix := strings.TrimRight(strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.SourceUrlPrefix")), "/")
	if !strings.HasPrefix(resourceURL, sourcePrefix+"/") {
		return nil
	}

	relativePath := strings.TrimLeft(strings.TrimPrefix(resourceURL, sourcePrefix), "/")
	if relativePath == "" {
		return nil
	}

	uploadRootPath := strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.UploadRootPath"))
	var basePath string
	if filepath.IsAbs(uploadRootPath) {
		basePath = uploadRootPath
	} else {
		cleanRoot := strings.TrimPrefix(strings.TrimPrefix(uploadRootPath, "./"), ".\\")
		basePath = filepath.Join(variable.BasePath, cleanRoot)
	}

	fullPath := filepath.Join(basePath, filepath.FromSlash(relativePath))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ExtractCOSObjectKey(resourceURL string) (string, bool) {
	candidates := []string{
		strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.BaseURL")),
		strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.BucketURL")),
	}

	for _, candidate := range candidates {
		candidate = strings.TrimRight(candidate, "/")
		if candidate == "" {
			continue
		}
		if strings.HasPrefix(resourceURL, candidate+"/") {
			return strings.TrimLeft(strings.TrimPrefix(resourceURL, candidate), "/"), true
		}
	}

	return "", false
}

func NewCOSClient() (*cos.Client, error) {
	bucketURL := strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.BucketURL"))
	secretID := strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.SecretID"))
	secretKey := strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.SecretKey"))

	if bucketURL == "" || secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("cos config is incomplete")
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("invalid cos bucket url: %w", err)
	}

	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return client, nil
}
