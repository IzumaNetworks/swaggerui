package swaggerui

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:generate go run generate.go

//go:embed embed
var swagfs embed.FS

func byteHandler(b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Write(b)
	}
}

func byteHandlerHTML(b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
	}
}

// Handler returns a handler that will serve a self-hosted Swagger UI with your spec embedded
func Handler(spec []byte) http.Handler {
	// render the index template with the proper spec name inserted
	static, _ := fs.Sub(swagfs, "embed")
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger_spec", byteHandler(spec))
	mux.Handle("/", http.FileServer(http.FS(static)))
	return mux
}

func HandlerWithSubpath(spec []byte, subpath string) http.Handler {
	static, _ := fs.Sub(swagfs, "embed")
	mux := http.NewServeMux()
	//	mux.HandleFunc(path.Join(subpath, "/swagger_spec"), byteHandler(spec))
	mux.HandleFunc("/swagger_spec", byteHandler(spec))
	mux.Handle("/", http.FileServer(http.FS(static)))
	return mux
}

func GenerateHTMLIndexOfEmbeddedSpecs(swaggerfiles embed.FS) (ret string, err error) {
	var buf bytes.Buffer
	err = fs.WalkDir(swaggerfiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process files (skip directories)
		if !d.IsDir() {
			// Read the file's content
			// data, err := swaggerfiles.ReadFile(path)
			// if err != nil {
			// 	log.Printf("Error reading %s: %v", path, err)
			// 	return err
			// }

			urlpath := strings.TrimPrefix(path, "embed/swagger/")
			// remove the .json extension
			urlpath = strings.TrimSuffix(urlpath, ".swagger.json")
			buf.WriteString(fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, urlpath, urlpath))
		}
		return nil
	})

	if err != nil {
		return
	}
	ret = fmt.Sprintf(`<ul>%s</ul>`, buf.String())
	return
}

// HandlerFromFS returns a handler that will serve a self-hosted Swagger UI with your spec embedded
func HandlerFromEmbedFS(swaggerfiles embed.FS) (ret http.Handler, err error) {

	// render the index template with the proper spec name inserted
	//	static, _ := fs.Sub(swagfs, "embed")
	mux := http.NewServeMux()

	err = fs.WalkDir(swaggerfiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process files (skip directories)
		if !d.IsDir() {
			// Read the file's content
			data, err := swaggerfiles.ReadFile(path)
			if err != nil {
				log.Printf("Error reading %s: %v", path, err)
				return err
			}

			urlpath := strings.TrimPrefix(path, "embed/")
			// remove the .json extension
			urlpath = strings.TrimSuffix(urlpath, ".swagger.json")
			log.Printf("urlpath: %s", urlpath)
			mux.Handle("/"+urlpath+"/", http.StripPrefix("/"+urlpath, HandlerWithSubpath(data, "/"+urlpath)))

			// // Print the filename and content
			// fmt.Printf("Filename: %s\n", path)
			// fmt.Printf("Content:\n%s\n\n", string(data))
		}
		return nil
	})

	//	mux.HandleFunc("/swagger_spec", byteHandler(spec))
	//	mux.Handle("/", http.FileServer(http.FS(static)))
	html, err := GenerateHTMLIndexOfEmbeddedSpecs(swaggerfiles)
	if err != nil {
		return
	}
	mux.HandleFunc("/", byteHandlerHTML([]byte(html)))

	ret = mux
	return

	// static, _ := fs.Sub(fsys, "embed")
	// mux := http.NewServeMux()
	// mux.Handle("/", http.FileServer(http.FS(static)))
	// return mux
}
