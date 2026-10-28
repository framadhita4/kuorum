package retrieval

import (
	"errors"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	ErrEmptyFilename  = errors.New("filename cannot be empty")
	ErrInvalidPaperID = errors.New("filename contains no supported characters for paper ID")
)

var windowsReservedNames = map[string]bool{
	"con":  true,
	"prn":  true,
	"aux":  true,
	"nul":  true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	"lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// FilenameParser encapsulates logic for transforming raw uploaded filenames into safe, filesystem- and URL-compliant paper IDs.
type FilenameParser struct {
	maxLen int
}

// NewFilenameParser creates a default FilenameParser with a standard length limit (120 chars).
func NewFilenameParser() *FilenameParser {
	return &FilenameParser{
		maxLen: 120,
	}
}

// Parse converts a raw filename into a sanitized paper ID.
// Returns an error if the filename is empty or contains no supported characters.
func (p *FilenameParser) Parse(filename string) (string, error) {
	sanitized := p.Sanitize(filename)
	if sanitized == "" {
		if strings.TrimSpace(filename) == "" {
			return "", ErrEmptyFilename
		}
		return "", ErrInvalidPaperID
	}
	return sanitized, nil
}

// Sanitize cleans any unsupported characters, path components, and extensions from a filename.
// If the input produces no valid identifier, an empty string is returned.
func (p *FilenameParser) Sanitize(filename string) string {
	raw := strings.TrimSpace(filename)
	if raw == "" {
		return ""
	}

	// 1. Normalize directory separators (\ and /) and strip path traversal components
	normalized := strings.ReplaceAll(raw, "\\", "/")
	base := path.Base(normalized)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == "/" {
		return ""
	}

	// 2. Strip extension (e.g. .pdf or other extension)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// 3. Lowercase and normalize Unicode (NFD decomposes diacritics like é -> e + ´)
	name = strings.ToLower(name)
	decomp := norm.NFD.String(name)

	var b strings.Builder
	b.Grow(len(decomp))

	prevHyphen := false
	prevUnderscore := false
	prevDot := false

	for _, r := range decomp {
		// Skip non-spacing diacritical marks
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
			prevUnderscore = false
			prevDot = false

		case r == '_':
			if !prevUnderscore && !prevHyphen && b.Len() > 0 {
				b.WriteByte('_')
				prevUnderscore = true
				prevHyphen = false
				prevDot = false
			}

		case r == '.':
			if !prevDot && !prevHyphen && !prevUnderscore && b.Len() > 0 {
				b.WriteByte('.')
				prevDot = true
				prevHyphen = false
				prevUnderscore = false
			}

		default:
			// Any unsupported or separator character becomes a hyphen
			if !prevHyphen && !prevUnderscore && !prevDot && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	res := b.String()

	// 4. Trim leading and trailing hyphens, underscores, dots
	res = strings.Trim(res, "-_.")

	// 5. If empty, return empty
	if res == "" {
		return ""
	}

	// 6. Enforce length limit
	if p.maxLen > 0 && len(res) > p.maxLen {
		res = res[:p.maxLen]
		res = strings.TrimRight(res, "-_.")
	}

	// 7. Check for Windows reserved names (e.g. "con", "prn", "aux")
	if windowsReservedNames[res] {
		res = "paper-" + res
	}

	return res
}

var defaultFilenameParser = NewFilenameParser()

// ParsePaperID converts an arbitrary filename (e.g. header.Filename) into a clean, safe paper ID.
func ParsePaperID(filename string) (string, error) {
	return defaultFilenameParser.Parse(filename)
}

// SanitizePaperID cleans an arbitrary filename into a safe paper ID string (returning "" if invalid).
func SanitizePaperID(filename string) string {
	return defaultFilenameParser.Sanitize(filename)
}

