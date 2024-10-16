package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
)

func fetchImageFromURL(url string) (image.Image, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf(("could not fetch image from URL: %w"), err)
    }
    defer resp.Body.Close()

    contentType := resp.Header.Get("Content-Type")
    switch contentType {
    case "image/png":
        img, err := png.Decode(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("could not decode PNG image: %w", err)
        }
        return img, nil
    case "image/jpeg", "image/jpg":
        img, err := jpeg.Decode(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("could not decode JPEG image: %w", err)
        }
        return img, nil
    default:
        return nil, fmt.Errorf("unsupported image format: %s", contentType)
    }
}

func generateQRCodeHandler(w http.ResponseWriter, r *http.Request) {
    text := os.Getenv("TEXT")
    if textParam := r.URL.Query().Get("text"); textParam != "" {
        text = textParam
    }

    qr, err := qrcode.New(text, qrcode.Medium)
    if err != nil {
        log.Printf("Failed to generating QR Code: %v", err)
        http.Error(w, "Failed to generate QR Code", http.StatusInternalServerError)
        return
    }

    logoPath := os.Getenv("LOGO_PATH")
    if logoPathParam := r.URL.Query().Get("logo"); logoPathParam != "" {
        logoPath = logoPathParam
    }

    var logo image.Image

    if logoPath != "" && strings.HasPrefix(logoPath, "http") {
        logo, err = fetchImageFromURL(logoPath)
        if err != nil {
            log.Printf("Error loading logo from URL %s: %v", logoPath, err)
        }
    }

    qrImage := qr.Image(256)
    if logo != nil {
        logoSize := qrImage.Bounds().Dx() / 5
        scaledLogo := resize.Resize(uint(logoSize), uint(logoSize), logo, resize.Lanczos3)

        offsetX := (qrImage.Bounds().Dx() - scaledLogo.Bounds().Dx()) / 2
        offsetY := (qrImage.Bounds().Dy() - scaledLogo.Bounds().Dy()) / 2
        combined := image.NewRGBA(qrImage.Bounds())
        draw.Draw(combined, qrImage.Bounds(), qrImage, image.Point{}, draw.Src)
        draw.Draw(combined, scaledLogo.Bounds().Add(image.Pt(offsetX, offsetY)), scaledLogo, image.Point{}, draw.Over)

        w.Header().Set("Content-Type", "image/png")
        if err := png.Encode(w, combined); err != nil {
            log.Printf("Failed to encode combined image: %v", err)
            http.Error(w, "Internal Server Error: Failed to encode image", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "image/png")
    if err := png.Encode(w, qrImage); err != nil {
        log.Printf("Failed to encode QR Code image: %v", err)
        http.Error(w, "Internal Server Error: Failed to encode image", http.StatusInternalServerError)
    }
}

func main() {
    if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
        log.Fatal("Error loading .env file")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    http.HandleFunc("/generate", generateQRCodeHandler)
    fmt.Println("Server starting on port", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
