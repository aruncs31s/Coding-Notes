---
id: file_uploads_go
aliases:
  - File Uploads in Go
  - Multipart Form Uploads
tags:
  - go
  - http
  - file-upload
  - web
  - security
dg-publish: true
---

# File Uploads in Go: `net/http` & `Gin`

File uploads in web applications use the `multipart/form-data` encoding. Handling uploads securely requires:
1. **Memory Bounds**: Limiting maximum payload size to protect against out-of-memory DoS attacks.
2. **MIME Type Sniffing**: Inspecting real file magic bytes rather than trusting client-provided file extensions.
3. **Path Traversal Prevention**: Sanitizing filenames before saving to disk.

---

## 1. Single File Upload (`net/http`)

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	MaxUploadSize = 10 << 20 // 10 MB limit
	UploadDir     = "./uploads"
)

func uploadSingleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Enforce strict max body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// 2. Parse multipart form with in-memory buffer limit (32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "File exceeds max allowed size (10MB)", http.StatusBadRequest)
		return
	}

	// 3. Retrieve file from form key "file"
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file key", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Validate MIME type by reading first 512 bytes (magic numbers)
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		http.Error(w, "Cannot read file header", http.StatusInternalServerError)
		return
	}

	detectedType := http.DetectContentType(buff)
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedTypes[detectedType] {
		http.Error(w, fmt.Sprintf("Unsupported file type: %s", detectedType), http.StatusBadRequest)
		return
	}

	// Rewind file reader back to start after sniffing
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to rewind file reader", http.StatusInternalServerError)
		return
	}

	// 5. Sanitize filename to prevent Directory Traversal attacks (e.g. "../../etc/passwd")
	cleanFilename := filepath.Base(fileHeader.Filename)
	destinationPath := filepath.Join(UploadDir, cleanFilename)

	dst, err := os.Create(destinationPath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 6. Stream content to destination
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to write file to disk", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' successfully uploaded (%d bytes, %s)!\n", cleanFilename, fileHeader.Size, detectedType)
}
```

---

## 2. Multi-File Upload (`net/http`)

When uploading multiple files under a single form field (e.g. `<input type="file" name="files" multiple>`):

```go
func uploadMultipleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB total

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Files exceed maximum size", http.StatusBadRequest)
		return
	}

	// Access files slice from multipart form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	savedFiles := 0
	for _, fileHeader := range files {
		src, err := fileHeader.Open()
		if err != nil {
			continue
		}

		cleanName := filepath.Base(fileHeader.Filename)
		dstPath := filepath.Join(UploadDir, cleanName)

		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}

		io.Copy(dst, src)
		src.Close()
		dst.Close()
		savedFiles++
	}

	fmt.Fprintf(w, "Successfully uploaded %d / %d files.\n", savedFiles, len(files))
}
```

---

## 3. File Uploads in `Gin` Framework

`Gin` simplifies file handling with helper methods `c.FormFile()` and `c.SaveUploadedFile()`:

```go
package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Set lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Single file
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
			return
		}

		filename := filepath.Base(file.Filename)
		target := filepath.Join("./uploads", filename)

		// Gin built-in safe upload helper
		if err := c.SaveUploadedFile(file, target); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Uploaded successfully",
			"filename": filename,
			"size":     file.Size,
		})
	})

	// Multiple files
	r.POST("/upload-multiple", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		files := form.File["files"]
		for _, file := range files {
			filename := filepath.Base(file.Filename)
			c.SaveUploadedFile(file, filepath.Join("./uploads", filename))
		}

		c.JSON(http.StatusOK, gin.H{"uploaded_count": len(files)})
	})

	r.Run(":8080")
}
```

---

## 🔒 Security Checklist for Go File Uploads

| Vulnerability | Mitigation in Go |
|---|---|
| **Memory Exhaustion / OOM** | Use `http.MaxBytesReader(w, r.Body, max)` and pass sensible buffer to `r.ParseMultipartForm()`. |
| **Path Traversal (`../../etc/passwd`)** | Always wrap filenames with `filepath.Base(fileHeader.Filename)`. |
| **Executable/Malicious File Extension** | Sniff magic bytes via `http.DetectContentType(512Bytes)` and whitelist MIME types. |
| **Filename Collisions** | Prefix or replace filename with a unique UUID (`uuid.New().String() + ext`). |
