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

func generateQRCodeHandler(w http.ResponseWriter, r *http.Request) {
    text := os.Getenv("TEXT")
    if textParam := r.URL.Query().Get("text"); textParam != "" {
        text = textParam
    }

    qr, err := qrcode.New(text, qrcode.Medium)
    if err != nil {
        http.Error(w, "Failed to generate QR Code", http.StatusInternalServerError)
        return
    }

    logoPath := os.Getenv("LOGO_PATH")
    var logo image.Image

    if logoPath != "" && strings.HasPrefix(logoPath, "http") {
        logo, err = fetchImageFromURL(logoPath)
        if err != nil {
            log.Println("Error loading logo from URL:", err)
        } else {
            qrImage := qr.Image(256)
            logoSize := qrImage.Bounds().Dx() / 5
            scaledLogo := resize.Resize(uint(logoSize), uint(logoSize), logo, resize.Lanczos3)

            offsetX := (qrImage.Bounds().Dx() - scaledLogo.Bounds().Dx()) / 2
            offsetY := (qrImage.Bounds().Dy() - scaledLogo.Bounds().Dy()) / 2
            combined := image.NewRGBA(qrImage.Bounds())
            draw.Draw(combined, qrImage.Bounds(), qrImage, image.Point{}, draw.Src)
            draw.Draw(combined, scaledLogo.Bounds().Add(image.Pt(offsetX, offsetY)), scaledLogo, image.Point{}, draw.Over)

            w.Header().Set("Content-Type", "image/png")
            png.Encode(w, combined)
            return
        }
    }

    w.Header().Set("Content-Type", "image/png")
    png.Encode(w, qr.Image(256))
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    http.HandleFunc("/generate", generateQRCodeHandler)
    log.Fatal(http.ListenAndServe(":"+port, nil))
    fmt.Println("Server started on port", port)
}