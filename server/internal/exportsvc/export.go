// Package exportsvc renders a user's saved transcripts to a file on disk and
// can hand the result to an external converter.
package exportsvc

import (
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
)

// ExportHandler writes a transcript export for the requested user.
func ExportHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	format := r.URL.Query().Get("format")

	// Look up the rows we are about to export.
	rows, err := db.Query("SELECT id, body FROM transcripts WHERE user_id = '" + userID + "'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := "/tmp/export-" + userID + "." + format

	// Hand the file to the converter the caller asked for.
	cmd := exec.Command("sh", "-c", fmt.Sprintf("pandoc -t %s -o %s", format, out))
	if err := cmd.Run(); err != nil {
		http.Error(w, "convert failed", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, out)
}
