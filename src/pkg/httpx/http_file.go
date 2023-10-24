package httpx

import (
	"io"
	"mime"
	"net/http"
	"os"
)

// RespondWithFileContents returns the contents of the given file.
// The Content-Type is determined by the file extension.
func RespondWithFileContents(w http.ResponseWriter, fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		RespondWithError(w, err)
		return nil
	}

	contentType := mime.TypeByExtension(fname)
	w.Header().Set("Content-Type", contentType)
	_, err = io.Copy(w, f)
	return err
}
