package handler

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "image/gif"  // GIF デコード用（そのまま保存）
	_ "image/jpeg" // JPEG デコード用
	_ "image/png"  // PNG デコード用

	"github.com/TenshoOHASHI/knowhub/services/gateway/swagger"
)

// swagger type reference
var _ swagger.UploadResponse

const maxUploadSize = 5 << 20 // 5 MB

const jpegQuality = 80 // JPEG再エンコード品質（1-100）。80はWeb標準的な値

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// compressedTypes はデコード→再エンコードで圧縮する対象
// GIFはアニメーションがあるため、WebPはすでに効率的なためそのまま保存
var compressedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

// Upload 画像アップロード
// @Summary      画像アップロード
// @Description  画像ファイルをアップロードし、URLを返す（要認証）
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "画像ファイル (JPEG/PNG/GIF/WebP, 最大5MB)"
// @Success      200  {object}  swagger.UploadResponse
// @Failure      400  {string}  string  "bad request"
// @Failure      413  {string}  string  "file too large"
// @Failure      415  {string}  string  "unsupported media type"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/upload [post]
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Size limit
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Detect MIME from content (first 512 bytes)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(buf[:n])
	if !allowedTypes[contentType] {
		http.Error(w, fmt.Sprintf("unsupported media type: %s", contentType), http.StatusUnsupportedMediaType)
		return
	}

	// Seek back to start after reading header
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to process file", http.StatusInternalServerError)
		return
	}

	// Generate filename
	exts, _ := mime.ExtensionsByType(contentType)
	ext := ".bin"
	if len(exts) > 0 {
		ext = exts[0]
	}
	sanitized := sanitizeFilename(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), sanitized, ext)

	// Ensure directory exists
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		slog.Error("failed to create upload directory", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Create destination file
	dstPath := filepath.Join(h.uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		slog.Error("failed to create file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 画像タイプに応じて保存方法を切り替え
	if compressedTypes[contentType] {
		// JPEG/PNG: デコード → 再エンコードで圧縮
		if err := compressAndSave(dst, file, contentType); err != nil {
			slog.Error("failed to compress image", "error", err, "type", contentType)
			// 圧縮に失敗したら生ファイルを保存（フォールバック）
			dst.Seek(0, io.SeekStart)  // 書き込み位置を先頭に配置
			dst.Truncate(0)            // 不完全の中身を削除
			file.Seek(0, io.SeekStart) // 読み書き位置を前途に配置
			if _, copyErr := io.Copy(dst, file); copyErr != nil {
				slog.Error("fallback copy also failed", "error", copyErr)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// GIF/WebP: そのまま保存
		if _, err := io.Copy(dst, file); err != nil {
			slog.Error("failed to save file", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// 保存先のパスを用意（ブラウザー側でsrcパスから、サーバにGETで画像を取得する）
	url := "/uploads/" + filename
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// compressAndSave 画像をデコードして再エンコード（圧縮）して保存
func compressAndSave(dst *os.File, src io.Reader, contentType string) error {
	// デコード（image.Decode は登録されたデコーダを自動選択）
	img, _, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	switch contentType {
	case "image/jpeg":
		// quality 80 で再エンコード（目に見える劣化はほぼなし）
		return jpeg.Encode(dst, img, &jpeg.Options{Quality: jpegQuality})
	case "image/png":
		// PNGは不可逆圧縮しない。標準エンコーダで最適化して保存
		return png.Encode(dst, img)
	default:
		return fmt.Errorf("unsupported type for compression: %s", contentType)
	}
}

// sanitizeFilename strips path components and replaces non-alphanumeric chars
func sanitizeFilename(name string) string {
	// Remove directory components(_../../etc/pic.png -> pic.png)
	name = filepath.Base(name)
	// Remove extension (we'll add our own), pic.png -> pic
	name = strings.TrimSuffix(name, filepath.Ext(name))
	// Replace non-alphanumeric (including unicode) with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	name = reg.ReplaceAllString(name, "_")
	// Truncate to 64 chars
	if len(name) > 64 {
		name = name[:64]
	}
	if name == "" {
		name = "upload"
	}
	return name
}
