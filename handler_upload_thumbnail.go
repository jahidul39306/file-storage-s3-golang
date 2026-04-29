package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "not supported file type", err)
		return
	}

	videoInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find video", err)
		return
	}

	if videoInfo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	filePath, err := cfg.getAssetPath(mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create file name", err)
		return
	}
	newFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create a new file", err)
		return
	}

	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy the file", err)
		return
	}

	thumbnailURL := cfg.getAssetURL(filePath)
	fmt.Printf(thumbnailURL)
	videoInfo.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update the video metadata", err)
		return
	}
	respondWithJSON(w, http.StatusOK, videoInfo)
}
