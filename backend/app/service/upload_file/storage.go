package upload_file

import (
	"context"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/utils/file_storage"
	"github.com/tencentyun/cos-go-sdk-v5"
	"mime"
	"path"
	"path/filepath"
	"strings"
)

func useCOSStorage() bool {
	return file_storage.UseCOSStorage()
}

func buildPublicFileURL(relativeDir, filename string) string {
	if useCOSStorage() {
		key := buildCOSObjectKey(relativeDir, filename)
		return buildCOSPublicURL(key)
	}

	prefix := strings.TrimRight(strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.SourceUrlPrefix")), "/")
	cleanDir := "/" + strings.Trim(strings.ReplaceAll(relativeDir, "\\", "/"), "/")
	return prefix + cleanDir + "/" + filename
}

func buildCOSObjectKey(relativeDir, filename string) string {
	cleanDir := strings.Trim(strings.ReplaceAll(relativeDir, "\\", "/"), "/")
	pathPrefix := strings.Trim(strings.ReplaceAll(variable.ConfigYml.GetString("FileUploadSetting.Cos.PathPrefix"), "\\", "/"), "/")
	if pathPrefix != "" {
		return path.Join(pathPrefix, cleanDir, filename)
	}
	return path.Join(cleanDir, filename)
}

func buildCOSPublicURL(objectKey string) string {
	baseURL := strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.BaseURL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(variable.ConfigYml.GetString("FileUploadSetting.Cos.BucketURL"))
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

func uploadLocalFileToCOS(localPath, relativeDir, filename, contentType string) (string, error) {
	client, err := file_storage.NewCOSClient()
	if err != nil {
		return "", err
	}

	key := buildCOSObjectKey(relativeDir, filename)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: detectContentType(filename, contentType),
		},
	}

	if _, err = client.Object.PutFromFile(context.Background(), key, localPath, opt); err != nil {
		return "", err
	}

	return buildCOSPublicURL(key), nil
}

func detectContentType(filename, headerValue string) string {
	if strings.TrimSpace(headerValue) != "" {
		return headerValue
	}
	if extType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); extType != "" {
		return extType
	}
	return "application/octet-stream"
}
