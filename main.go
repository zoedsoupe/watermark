package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {
	http.Handle("GET /", templ.Handler(upload()))
	http.HandleFunc("POST /", uploadHandler)
	fmt.Println("Servidor iniciado em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Limite de 10MB para upload

	pdfFile, pdfHeader, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Erro ao enviar o PDF", http.StatusBadRequest)
		return
	}
	defer pdfFile.Close()

	logoFile, logoHeader, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "Erro ao enviar o Logotipo", http.StatusBadRequest)
		return
	}
	defer logoFile.Close()

	// Salvar arquivos temporariamente
	pdfPath := filepath.Join(os.TempDir(), pdfHeader.Filename)
	pdfOut, err := os.Create(pdfPath)
	if err != nil {
		http.Error(w, "Erro ao salvar o PDF", http.StatusInternalServerError)
		return
	}
	defer pdfOut.Close()
	io.Copy(pdfOut, pdfFile)

	logoPath := filepath.Join(os.TempDir(), logoHeader.Filename)
	logoOut, err := os.Create(logoPath)
	if err != nil {
		http.Error(w, "Erro ao salvar o Logotipo", http.StatusInternalServerError)
		return
	}
	defer logoOut.Close()
	io.Copy(logoOut, logoFile)

	// Processar o PDF
	outputPath := filepath.Join(os.TempDir(), "modificado_"+pdfHeader.Filename)
	err = addLogoToPDF(pdfPath, logoPath, outputPath)
	if err != nil {
		http.Error(w, "Erro ao processar o PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Enviar o PDF modificado para download
	w.Header().Set("Content-Disposition", "attachment; filename=modificado_"+pdfHeader.Filename)
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, outputPath)

	// Limpar arquivos temporários
	os.Remove(pdfPath)
	os.Remove(logoPath)
	os.Remove(outputPath)
}

func addLogoToPDF(pdfPath, logoPath, outputPath string) error {
	pos := "pos:tl, scale:0.1, offset: 0 0, rotation: 0"

	// Aplicar o watermark em cada página
	err := pdf.AddImageWatermarksFile(pdfPath, outputPath, []string{}, true, logoPath, pos, nil)

	if err != nil {
		return err
	}

	return nil
}
