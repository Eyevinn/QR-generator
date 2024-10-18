package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
)

var (
	defaultPort = "8080"
	defaultSize = 256
)

func fetchImageFromURL(url string) (image.Image, error) {
	// The default HTTP client does not have any timeout so it can hang forever.
	// It is therefore recommended to create a new client with a timeout.
	clientWithTimeout := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := clientWithTimeout.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "image/png":
		return png.Decode(resp.Body)
	case "image/jpeg", "image/jpg":
		return jpeg.Decode(resp.Body)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", contentType)
	}
}

type server struct {
	port     string
	text     string
	logoPath string
}

func (s *server) makeGenerateQRCodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		text := s.text
		logoPath := s.logoPath
		qrSize := defaultSize

		if textParam := r.URL.Query().Get("text"); textParam != "" {
			text = textParam
		}

		qrSizeParam := r.URL.Query().Get("size")
		size, err := strconv.Atoi(qrSizeParam)

		if err == nil {
			qrSize = size
		}

		qr, err := qrcode.New(text, qrcode.Medium)
		if err != nil {
			http.Error(w, "Failed to generate QR Code", http.StatusInternalServerError)
			slog.Error("Failed to generated QR Code", "error", err)
			return
		}

		var logo image.Image

		if logoPath != "" && strings.HasPrefix(logoPath, "http") {
			logo, err = fetchImageFromURL(logoPath)
			if err != nil {
				slog.Error("Failed to load logo", "url", logoPath, "error", err)
				http.Error(w, "Failed to load logo", http.StatusInternalServerError)
				return
			} else {
				qrImage := qr.Image(qrSize)
				logoSize := qrImage.Bounds().Dx() / 5
				scaledLogo := resize.Resize(uint(logoSize), uint(logoSize), logo, resize.Lanczos3)

				offsetX := (qrImage.Bounds().Dx() - scaledLogo.Bounds().Dx()) / 2
				offsetY := (qrImage.Bounds().Dy() - scaledLogo.Bounds().Dy()) / 2
				combined := image.NewRGBA(qrImage.Bounds())
				draw.Draw(combined, qrImage.Bounds(), qrImage, image.Point{}, draw.Src)
				draw.Draw(combined, scaledLogo.Bounds().Add(image.Pt(offsetX, offsetY)), scaledLogo, image.Point{}, draw.Over)

				w.Header().Set("Content-Type", "image/png")
				err := png.Encode(w, combined)
				if err != nil {
					slog.Error("Failed to encode image", "error", err)
				}
				return
			}
		}

		w.Header().Set("Content-Type", "image/png")
		err = png.Encode(w, qr.Image(qrSize))
		if err != nil {
			slog.Error("Failed to encode image", "error", err)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Error("failed to load existing .env file", "error", err)
		os.Exit(1)
	}
	srv := server{
		port:     os.Getenv("PORT"),
		text:     os.Getenv("TEXT"),
		logoPath: os.Getenv("LOGO_PATH"),
	}
	if srv.port == "" {
		srv.port = defaultPort
	}
	err := run(srv) // Use a special function to run the server so it is easier to test
	if err != nil {
		slog.Error("Failed to run", "error", err)
		os.Exit(1)
	}
}

func run(srv server) error {
	http.HandleFunc("/generate", srv.makeGenerateQRCodeHandler())
	slog.Info("server starting", "port", srv.port)
	return http.ListenAndServe(":"+srv.port, nil)
}
