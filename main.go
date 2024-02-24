package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

// Page represents a single page of the website.
type Page struct {
	Title string
	Body  template.HTML
}

// renderMarkdown converts markdown content to HTML using Goldmark.
func renderMarkdown(content []byte) template.HTML {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)
	if err := md.Convert(content, &buf); err != nil {
		log.Printf("Error rendering markdown: %v\n", err)
		return ""
	}
	return template.HTML(buf.String())
}

// generateHTML generates an HTML file for a page.
func generateHTML(page *Page, templatePath, outputPath string) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, page); err != nil {
		return err
	}

	html := buf.Bytes()

	if err := ioutil.WriteFile(outputPath, html, 0644); err != nil {
		return err
	}

	return nil
}

// generateSite generates HTML files for all markdown files in the input directory.
func generateSite(inputDir, outputDir, templatePath string) error {
	// Read all markdown files from input directory
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
			continue
		}

		// Read markdown content
		mdPath := filepath.Join(inputDir, file.Name())
		mdContent, err := ioutil.ReadFile(mdPath)
		if err != nil {
			return err
		}

		// Convert markdown to HTML
		htmlContent := renderMarkdown(mdContent)

		// Create a Page instance
		page := &Page{
			Title: file.Name(),
			Body:  htmlContent,
		}

		// Generate HTML file for the page
		htmlPath := filepath.Join(outputDir, file.Name()+".html")
		if err := generateHTML(page, templatePath, htmlPath); err != nil {
			return err
		}
	}

	return nil
}

func Create() {
	inputDir := "content"
	outputDir := "output"
	templatePath := "templates/template.html"

	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.Mkdir(outputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	// Generate the site
	if err := generateSite(inputDir, outputDir, templatePath); err != nil {
		log.Fatalf("Error generating site: %v", err)
	}

	log.Println("Site generated successfully!")
}
func servePage(w http.ResponseWriter, r *http.Request) {
	pageName := filepath.Base(r.URL.Path)
	if pageName == "" || pageName == "/" {
		pageName = "index.html"
	}
	http.ServeFile(w, r, filepath.Join("output", pageName))
}
func main() {
	_, err := os.MkdirAll("./content", 0777), os.MkdirAll("./templates", 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("DONE")
	Create()
	http.HandleFunc("/", servePage)

	// Start HTTP server
	port := ":8080"
	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
