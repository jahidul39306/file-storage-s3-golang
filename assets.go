package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg apiConfig) getAssetPath(videoIDString string, mediaType string) string {
	mediaTypeParts := strings.Split(mediaType, "/")
	fileName := fmt.Sprint(videoIDString + "." + mediaTypeParts[1])
	filePath := filepath.Join(cfg.assetsRoot, fileName)
	return filePath
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("/assets/" + assetPath)
}
