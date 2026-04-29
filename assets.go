package main

import (
	"crypto/rand"
	"encoding/base64"
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
