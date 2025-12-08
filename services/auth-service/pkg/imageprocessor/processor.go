package imageprocessor

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// Config contiene la configuración para el procesamiento de imágenes
type Config struct {
	MaxWidth    int   // Ancho máximo en píxeles
	MaxHeight   int   // Alto máximo en píxeles
	Quality     int   // Calidad JPEG (1-100)
	MaxFileSize int64 // Tamaño máximo del archivo resultante en bytes
}

// DefaultAvatarConfig configuración por defecto para avatares
var DefaultAvatarConfig = Config{
	MaxWidth:    300,
	MaxHeight:   300,
	Quality:     85,
	MaxFileSize: 150 * 1024, // 150KB máximo
}

// ProcessUploadedImage procesa una imagen subida: redimensiona, comprime y guarda
func ProcessUploadedImage(src io.Reader, outputPath string, config Config) error {
	log.Printf("🔧 [ImageProcessor] Iniciando procesamiento de imagen: %s", outputPath)

	// Decodificar la imagen
	img, format, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	log.Printf("🔧 [ImageProcessor] Imagen original: formato=%s, dimensiones=%dx%d",
		format, img.Bounds().Dx(), img.Bounds().Dy())

	// Redimensionar la imagen manteniendo la proporción
	resized := imaging.Fit(img, config.MaxWidth, config.MaxHeight, imaging.Lanczos)

	log.Printf("🔧 [ImageProcessor] Imagen redimensionada a: %dx%d",
		resized.Bounds().Dx(), resized.Bounds().Dy())

	// Crear el directorio si no existe
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Determinar el formato de salida (siempre JPEG para avatares)
	outputPath = changeExtensionToJPEG(outputPath)

	// Guardar la imagen comprimida
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Convertir a JPEG con la calidad especificada
	err = jpeg.Encode(out, resized, &jpeg.Options{Quality: config.Quality})
	if err != nil {
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}

	// Verificar el tamaño del archivo resultante
	fileInfo, err := out.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()
	log.Printf("✅ [ImageProcessor] Imagen procesada exitosamente: %s (tamaño: %d bytes, %.1fKB)",
		outputPath, fileSize, float64(fileSize)/1024)

	// Verificar si excede el tamaño máximo
	if fileSize > config.MaxFileSize {
		log.Printf("⚠️ [ImageProcessor] Imagen excede tamaño máximo (%d bytes > %d bytes), recomprimiendo...",
			fileSize, config.MaxFileSize)

		// Recomprimir con menor calidad
		return recompressWithLowerQuality(resized, outputPath, config)
	}

	return nil
}

// changeExtensionToJPEG cambia la extensión del archivo a .jpeg
func changeExtensionToJPEG(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + ".jpeg"
}

// recompressWithLowerQuality recomprime la imagen con menor calidad si es muy grande
func recompressWithLowerQuality(img image.Image, outputPath string, config Config) error {
	qualities := []int{75, 65, 55, 45, 35} // Calidades progresivamente menores

	for _, quality := range qualities {
		log.Printf("🔧 [ImageProcessor] Intentando calidad %d%%", quality)

		out, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}

		err = jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
		out.Close()

		if err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}

		// Verificar el tamaño
		fileInfo, err := os.Stat(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		fileSize := fileInfo.Size()
		log.Printf("🔧 [ImageProcessor] Tamaño con calidad %d%%: %d bytes (%.1fKB)",
			quality, fileSize, float64(fileSize)/1024)

		if fileSize <= config.MaxFileSize {
			log.Printf("✅ [ImageProcessor] Compresión exitosa con calidad %d%% (tamaño final: %.1fKB)",
				quality, float64(fileSize)/1024)
			return nil
		}
	}

	return fmt.Errorf("unable to compress image below %d bytes", config.MaxFileSize)
}

// GetSupportedFormats retorna los formatos de imagen soportados
func GetSupportedFormats() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif"}
}

// IsValidImageFormat verifica si la extensión es un formato soportado
func IsValidImageFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := GetSupportedFormats()

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}
