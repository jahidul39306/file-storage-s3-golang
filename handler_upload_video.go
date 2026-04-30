package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid videoID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
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

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get the video", err)
		return
	}

	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse the content type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Not supported file type", err)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely_temp.mp4")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create temp file in local disk", err)
		return
	}

	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to copy file", err)
		return
	}

	tempFile.Seek(0, io.SeekStart)

	ratio, err := getVideAspectRation(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get video ratio", err)
		return
	}

	prefix := "other"

	if ratio == "16:9" {
		prefix = "landscape"
	} else if ratio == "9:16" {
		prefix = "portrait"
	}

	fileKey, err := cfg.generateS3Key(header.Filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate s3 key", err)
		return
	}

	newFileKey := fmt.Sprintf("%s/%s", prefix, fileKey)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(newFileKey),
		Body:        tempFile,
		ContentType: aws.String(mediaType),
	})

	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, newFileKey)
	videoInfo.VideoURL = &videoURL
	fmt.Println("Updating video URL to:", videoURL)
	err = cfg.db.UpdateVideo(videoInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update the video metadata", err)
		return
	}
	respondWithJSON(w, http.StatusOK, videoInfo)
}

func getVideAspectRation(filePath string) (string, error) {
	type VideoMetaData struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	var buf bytes.Buffer
	var data VideoMetaData
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		return "", err
	}

	if len(data.Streams) > 0 {
		w := float64(data.Streams[0].Width)
		h := float64(data.Streams[0].Height)

		ratio := w / h

		if ratio > 1 {
			return "16:9", nil
		} else if ratio < 1 {
			return "9:16", nil
		} else {
			return "other", nil
		}
	}
	return "", nil
}
