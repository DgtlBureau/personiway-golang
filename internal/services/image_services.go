package services

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func svgToPNG(svgData []byte) ([]byte, error) {
	// Парсим SVG
	svg, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return nil, err
	}

	// Создаем изображение
	w, h := int(svg.ViewBox.W), int(svg.ViewBox.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Рендерим SVG в изображение
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	svg.Draw(raster, 1.0)

	// Конвертируем в PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func svgToJPEG(svgData []byte, quality int) ([]byte, error) {
	// Сначала в RGBA
	svg, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return nil, err
	}

	w, h := int(svg.ViewBox.W), int(svg.ViewBox.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	svg.Draw(raster, 1.0)

	// Конвертируем в JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
