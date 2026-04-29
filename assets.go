package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cfg *apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg *apiConfig) getAssetPath(mediaType string) (string, error) {
	mediaTypeParts := strings.Split(mediaType, "/")
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(b)
	fileName := fmt.Sprint(encoded + "." + mediaTypeParts[1])
	filePath := filepath.Join(cfg.assetsRoot, fileName)
	return filePath, nil
}

func (cfg *apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, assetPath)
}

func (cfg *apiConfig) generateS3Key(filename string) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoded := hex.EncodeToString(b)
	ext := filepath.Ext(filename)

	return fmt.Sprintf("%s%s", encoded, ext), nil
}
