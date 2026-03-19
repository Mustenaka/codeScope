package filebrowser

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"codescope/server/internal/session"
)

const maxPreviewSize int64 = 256 * 1024

var (
	ErrPathOutsideWorkspace = errors.New("path is outside workspace_root")
	ErrInvalidPath          = errors.New("invalid file path")
)

var ignoredDirectories = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	".codescope":   {},
}

var previewExtensions = map[string]struct{}{
	".bat":     {},
	".c":       {},
	".cc":      {},
	".conf":    {},
	".cpp":     {},
	".css":     {},
	".csv":     {},
	".env":     {},
	".go":      {},
	".graphql": {},
	".h":       {},
	".hpp":     {},
	".html":    {},
	".ini":     {},
	".java":    {},
	".js":      {},
	".json":    {},
	".jsx":     {},
	".kt":      {},
	".kts":     {},
	".log":     {},
	".lua":     {},
	".md":      {},
	".mjs":     {},
	".php":     {},
	".proto":   {},
	".ps1":     {},
	".py":      {},
	".rb":      {},
	".rs":      {},
	".scss":    {},
	".sh":      {},
	".sql":     {},
	".svg":     {},
	".swift":   {},
	".toml":    {},
	".ts":      {},
	".tsx":     {},
	".txt":     {},
	".xml":     {},
	".yaml":    {},
	".yml":     {},
}

var previewBaseNames = map[string]struct{}{
	"Dockerfile": {},
	"LICENSE":    {},
	"Makefile":   {},
	"README":     {},
}

type SessionReader interface {
	Get(id string) (session.Session, error)
}

type Node struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int64  `json:"size,omitempty"`
	Previewable bool   `json:"previewable,omitempty"`
	Children    []Node `json:"children,omitempty"`
}

type Content struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Previewable bool   `json:"previewable"`
	Reason      string `json:"reason,omitempty"`
	Content     string `json:"content,omitempty"`
	Language    string `json:"language,omitempty"`
}

type Service struct {
	sessions SessionReader
}

func NewService(sessions SessionReader) *Service {
	return &Service{sessions: sessions}
}

func (s *Service) ListTree(sessionID string) ([]Node, error) {
	record, err := s.sessions.Get(sessionID)
	if err != nil {
		return nil, fmt.Errorf("load session %s: %w", sessionID, err)
	}
	return s.ListTreeByWorkspace(record.WorkspaceRoot)
}

func (s *Service) ListTreeByWorkspace(workspaceRoot string) ([]Node, error) {
	root, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read workspace root: %w", err)
	}

	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		if shouldIgnore(entry.Name()) {
			continue
		}
		node, err := s.buildNode(root, filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, compareNodes)
	return nodes, nil
}

func (s *Service) ReadContent(sessionID, requestedPath string) (Content, error) {
	record, err := s.sessions.Get(sessionID)
	if err != nil {
		return Content{}, fmt.Errorf("load session %s: %w", sessionID, err)
	}
	return s.ReadContentByWorkspace(record.WorkspaceRoot, requestedPath)
}

func (s *Service) ReadContentByWorkspace(workspaceRoot, requestedPath string) (Content, error) {
	resolvedPath, relativePath, err := resolveWorkspacePath(workspaceRoot, requestedPath)
	if err != nil {
		return Content{}, err
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return Content{}, fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return Content{}, fmt.Errorf("%w: directory content is not previewable", ErrInvalidPath)
	}

	result := Content{
		Path:        relativePath,
		Size:        info.Size(),
		Previewable: false,
		Language:    detectLanguage(relativePath),
	}
	if !isPreviewableExtension(relativePath) {
		result.Reason = "extension_not_previewable"
		return result, nil
	}
	if info.Size() > maxPreviewSize {
		result.Reason = "file_too_large"
		return result, nil
	}

	file, err := os.Open(resolvedPath)
	if err != nil {
		return Content{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Content{}, fmt.Errorf("read file: %w", err)
	}
	if !utf8.Valid(data) || bytesContainNUL(data) {
		result.Reason = "binary_not_previewable"
		return result, nil
	}

	result.Previewable = true
	result.Content = string(data)
	return result, nil
}

func (s *Service) buildNode(root, absolutePath string) (Node, error) {
	info, err := os.Stat(absolutePath)
	if err != nil {
		return Node{}, fmt.Errorf("stat path %s: %w", absolutePath, err)
	}

	relativePath, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return Node{}, fmt.Errorf("build relative path: %w", err)
	}
	relativePath = filepath.ToSlash(relativePath)

	node := Node{
		Name: info.Name(),
		Path: relativePath,
	}
	if info.IsDir() {
		node.Type = "directory"
		entries, err := os.ReadDir(absolutePath)
		if err != nil {
			return Node{}, fmt.Errorf("read directory %s: %w", absolutePath, err)
		}
		children := make([]Node, 0, len(entries))
		for _, entry := range entries {
			if shouldIgnore(entry.Name()) {
				continue
			}
			child, err := s.buildNode(root, filepath.Join(absolutePath, entry.Name()))
			if err != nil {
				return Node{}, err
			}
			children = append(children, child)
		}
		slices.SortFunc(children, compareNodes)
		node.Children = children
		return node, nil
	}

	node.Type = "file"
	node.Size = info.Size()
	node.Previewable = isPreviewableExtension(relativePath)
	return node, nil
}

func resolveWorkspacePath(workspaceRoot, requestedPath string) (string, string, error) {
	if strings.TrimSpace(requestedPath) == "" {
		return "", "", ErrInvalidPath
	}

	root, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve workspace root: %w", err)
	}

	cleanRelative := filepath.Clean(requestedPath)
	if cleanRelative == "." || cleanRelative == string(filepath.Separator) {
		return "", "", ErrInvalidPath
	}

	resolved := filepath.Join(root, cleanRelative)
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", "", fmt.Errorf("resolve file path: %w", err)
	}

	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return "", "", fmt.Errorf("resolve relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", ErrPathOutsideWorkspace
	}

	return resolved, filepath.ToSlash(rel), nil
}

func shouldIgnore(name string) bool {
	_, ignored := ignoredDirectories[name]
	return ignored
}

func isPreviewableExtension(path string) bool {
	base := filepath.Base(path)
	if _, ok := previewBaseNames[base]; ok {
		return true
	}
	ext := strings.ToLower(filepath.Ext(base))
	_, ok := previewExtensions[ext]
	return ok
}

func detectLanguage(path string) string {
	base := filepath.Base(path)
	if _, ok := previewBaseNames[base]; ok {
		return strings.ToLower(base)
	}
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(base)), ".")
}

func bytesContainNUL(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func compareNodes(left, right Node) int {
	if left.Type != right.Type {
		if left.Type == "directory" {
			return -1
		}
		return 1
	}
	return strings.Compare(left.Name, right.Name)
}
