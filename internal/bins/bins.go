// Package bins locates and downloads the external tools the app drives.
// MediaMTX is kept inside the app data dir so the app never relies on a
// system-wide install.
package bins

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const fallbackMediaMtx = "v1.21.4"

func client() *http.Client { return &http.Client{Timeout: 5 * time.Minute} }

// LookPath returns the absolute path of a binary or "" when it is missing.
func LookPath(name string) string {
	if name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		if st, err := os.Stat(name); err == nil && !st.IsDir() {
			return name
		}
		return ""
	}
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

func MediaMtxName() string {
	if runtime.GOOS == "windows" {
		return "mediamtx.exe"
	}
	return "mediamtx"
}

// EnsureMediaMtx downloads the latest MediaMTX release into binDir when the
// binary is not there yet and returns its path.
func EnsureMediaMtx(ctx context.Context, binDir string, logf func(string)) (string, error) {
	dest := filepath.Join(binDir, MediaMtxName())
	if LookPath(dest) != "" {
		return dest, nil
	}
	logf("MediaMTX absent, téléchargement dans " + binDir)

	version := latestMediaMtxVersion(ctx)
	asset, err := assetName(version)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://github.com/bluenviron/mediamtx/releases/download/%s/%s", version, asset)
	logf("→ " + url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client().Do(req)
	if err != nil {
		return "", fmt.Errorf("téléchargement de MediaMTX : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("téléchargement de MediaMTX : HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(binDir, "mediamtx-*"+filepath.Ext(asset))
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return "", fmt.Errorf("téléchargement de MediaMTX : %w", err)
	}
	tmp.Close()

	if err := extractBinary(tmpName, MediaMtxName(), dest); err != nil {
		return "", err
	}
	if err := os.Chmod(dest, 0o755); err != nil {
		return "", err
	}
	logf("MediaMTX " + version + " installé")
	return dest, nil
}

func latestMediaMtxVersion(ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/bluenviron/mediamtx/releases/latest", nil)
	if err != nil {
		return fallbackMediaMtx
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client().Do(req)
	if err != nil {
		return fallbackMediaMtx
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fallbackMediaMtx
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.TagName == "" {
		return fallbackMediaMtx
	}
	return payload.TagName
}

func assetName(version string) (string, error) {
	arch := runtime.GOARCH
	switch runtime.GOOS {
	case "linux", "darwin":
		return fmt.Sprintf("mediamtx_%s_%s_%s.tar.gz", version, runtime.GOOS, arch), nil
	case "windows":
		return fmt.Sprintf("mediamtx_%s_windows_%s.zip", version, arch), nil
	}
	return "", fmt.Errorf("plateforme non supportée pour MediaMTX : %s/%s", runtime.GOOS, runtime.GOARCH)
}

func extractBinary(archive, wanted, dest string) error {
	if strings.HasSuffix(archive, ".zip") {
		return extractZip(archive, wanted, dest)
	}
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != wanted {
			continue
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("binaire %s introuvable dans l'archive", wanted)
}

func extractZip(archive, wanted, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if filepath.Base(f.Name) != wanted {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("binaire %s introuvable dans l'archive", wanted)
}
