package repo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"clarus/internal/config"
	"clarus/internal/utils"
)

// Service defines repository exploration operations
type Service interface {
	ListRepositories(ctx context.Context) ([]*Repository, error)
	GetRepository(ctx context.Context, name string) (*Repository, error)
	GetDirectoryTree(ctx context.Context, repo string, opts TreeOptions) (*DirectoryTree, error)
	ListFiles(ctx context.Context, repo string, path string) ([]FileInfo, error)
	ReadFile(ctx context.Context, repo string, path string) (string, error)
	SearchCode(ctx context.Context, query string, opts CodeSearchOptions) ([]CodeSearchResult, error)
	FindFiles(ctx context.Context, pattern string, repo string) ([]FileInfo, error)
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	cfg *config.ReposConfig
}

// NewService creates a new repository service
func NewService(cfg *config.ReposConfig) Service {
	return &serviceImpl{cfg: cfg}
}

// ListRepositories returns all repositories under REPOS_ROOT
func (s *serviceImpl) ListRepositories(ctx context.Context) ([]*Repository, error) {
	entries, err := os.ReadDir(s.cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to read repos directory: %w", err)
	}

	var repos []*Repository
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Check if it's excluded
		if utils.IsExcluded(name, s.cfg.ExcludePatterns) {
			continue
		}

		repoPath := filepath.Join(s.cfg.Root, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		repo := &Repository{
			Name:         name,
			Path:         repoPath,
			LastModified: info.ModTime(),
		}

		// Detect primary language
		repo.PrimaryLanguage = s.detectPrimaryLanguage(repoPath)

		// Count files
		repo.FileCount = s.countFiles(repoPath)

		// Try to read description from README
		repo.Description = s.readDescription(repoPath)

		repos = append(repos, repo)
	}

	// Sort by name
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

func (s *serviceImpl) detectPrimaryLanguage(repoPath string) string {
	langCounts := make(map[string]int)

	filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Skip excluded patterns
		if utils.IsExcluded(d.Name(), s.cfg.ExcludePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		lang := utils.DetectLanguage(d.Name())
		if lang != "unknown" {
			langCounts[lang]++
		}
		return nil
	})

	var maxLang string
	var maxCount int
	for lang, count := range langCounts {
		if count > maxCount {
			maxLang = lang
			maxCount = count
		}
	}

	return maxLang
}

func (s *serviceImpl) countFiles(repoPath string) int {
	count := 0
	filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if utils.IsExcluded(d.Name(), s.cfg.ExcludePatterns) {
			return nil
		}
		count++
		return nil
	})
	return count
}

func (s *serviceImpl) readDescription(repoPath string) string {
	readmeFiles := []string{"README.md", "README", "readme.md", "README.txt"}

	for _, readme := range readmeFiles {
		path := filepath.Join(repoPath, readme)
		content, err := utils.ReadFileContent(path, 4096)
		if err != nil {
			continue
		}

		// Extract first paragraph
		lines := strings.Split(content, "\n")
		var desc strings.Builder
		foundContent := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip title
			if strings.HasPrefix(trimmed, "#") {
				foundContent = true
				continue
			}
			// Skip empty lines before content
			if !foundContent && trimmed == "" {
				continue
			}
			// Found content
			if trimmed != "" {
				foundContent = true
				desc.WriteString(trimmed)
				desc.WriteString(" ")
			} else if foundContent {
				// End of first paragraph
				break
			}
		}

		result := strings.TrimSpace(desc.String())
		if len(result) > 200 {
			return result[:200] + "..."
		}
		return result
	}

	return ""
}

// GetRepository returns information about a specific repository
func (s *serviceImpl) GetRepository(ctx context.Context, name string) (*Repository, error) {
	if err := utils.ValidateRepoName(name); err != nil {
		return nil, err
	}

	repoPath := filepath.Join(s.cfg.Root, name)
	info, err := os.Stat(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("repository not found: %s", name)
		}
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", name)
	}

	return &Repository{
		Name:            name,
		Path:            repoPath,
		LastModified:    info.ModTime(),
		PrimaryLanguage: s.detectPrimaryLanguage(repoPath),
		FileCount:       s.countFiles(repoPath),
		Description:     s.readDescription(repoPath),
	}, nil
}

// GetDirectoryTree returns the directory structure of a repository
func (s *serviceImpl) GetDirectoryTree(ctx context.Context, repoName string, opts TreeOptions) (*DirectoryTree, error) {
	if err := utils.ValidateRepoName(repoName); err != nil {
		return nil, err
	}

	repoPath := filepath.Join(s.cfg.Root, repoName)
	if _, err := os.Stat(repoPath); err != nil {
		return nil, fmt.Errorf("repository not found: %s", repoName)
	}

	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 3
	}

	excludePatterns := opts.ExcludePatterns
	if len(excludePatterns) == 0 {
		excludePatterns = s.cfg.ExcludePatterns
	}

	root := &TreeNode{
		Name:     repoName,
		Path:     "",
		IsDir:    true,
		Children: make([]*TreeNode, 0),
	}

	s.buildTree(root, repoPath, "", 0, opts.MaxDepth, opts.IncludeFiles, excludePatterns)

	return &DirectoryTree{
		Root:     root,
		MaxDepth: opts.MaxDepth,
	}, nil
}

func (s *serviceImpl) buildTree(node *TreeNode, basePath, relPath string, depth, maxDepth int, includeFiles bool, excludePatterns []string) {
	if depth >= maxDepth {
		return
	}

	fullPath := filepath.Join(basePath, relPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()

		// Check exclusions
		if utils.IsExcluded(name, excludePatterns) {
			continue
		}

		childPath := filepath.Join(relPath, name)
		if relPath == "" {
			childPath = name
		}

		if entry.IsDir() {
			child := &TreeNode{
				Name:     name,
				Path:     childPath,
				IsDir:    true,
				Children: make([]*TreeNode, 0),
			}
			node.Children = append(node.Children, child)
			s.buildTree(child, basePath, childPath, depth+1, maxDepth, includeFiles, excludePatterns)
		} else if includeFiles {
			node.Children = append(node.Children, &TreeNode{
				Name:  name,
				Path:  childPath,
				IsDir: false,
			})
		}
	}
}

// ListFiles returns files in a specific directory of a repository
func (s *serviceImpl) ListFiles(ctx context.Context, repoName string, path string) ([]FileInfo, error) {
	if err := utils.ValidateRepoName(repoName); err != nil {
		return nil, err
	}

	repoPath := filepath.Join(s.cfg.Root, repoName)
	fullPath, err := utils.SanitizePath(repoPath, path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		name := entry.Name()
		if utils.IsExcluded(name, s.cfg.ExcludePatterns) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath := filepath.Join(path, name)
		files = append(files, FileInfo{
			Path:         relPath,
			Name:         name,
			IsDirectory:  entry.IsDir(),
			Size:         info.Size(),
			LastModified: info.ModTime(),
			Language:     utils.DetectLanguage(name),
		})
	}

	return files, nil
}

// ReadFile reads a file from a repository
func (s *serviceImpl) ReadFile(ctx context.Context, repoName string, path string) (string, error) {
	if err := utils.ValidateRepoName(repoName); err != nil {
		return "", err
	}

	repoPath := filepath.Join(s.cfg.Root, repoName)
	fullPath, err := utils.SanitizePath(repoPath, path)
	if err != nil {
		return "", err
	}

	// Check if it's a sensitive file
	if utils.IsSensitiveFile(path) {
		return "", fmt.Errorf("cannot read potentially sensitive file: %s", path)
	}

	return utils.ReadFileContent(fullPath, s.cfg.MaxFileSize)
}

// SearchCode searches for patterns in repository code
func (s *serviceImpl) SearchCode(ctx context.Context, query string, opts CodeSearchOptions) ([]CodeSearchResult, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 50
	}
	if opts.Context <= 0 {
		opts.Context = 2
	}

	var repos []string
	if opts.Repository != "" {
		if err := utils.ValidateRepoName(opts.Repository); err != nil {
			return nil, err
		}
		repos = []string{opts.Repository}
	} else {
		repoList, err := s.ListRepositories(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range repoList {
			repos = append(repos, r.Name)
		}
	}

	regex, err := regexp.Compile(query)
	if err != nil {
		// Fall back to literal search
		regex = regexp.MustCompile(regexp.QuoteMeta(query))
	}

	var results []CodeSearchResult

	for _, repoName := range repos {
		if len(results) >= opts.MaxResults {
			break
		}

		repoPath := filepath.Join(s.cfg.Root, repoName)
		filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			// Check exclusions
			if utils.IsExcluded(d.Name(), s.cfg.ExcludePatterns) {
				return nil
			}

			// Check file pattern
			if opts.FilePattern != "" {
				matched, _ := filepath.Match(opts.FilePattern, d.Name())
				if !matched {
					return nil
				}
			}

			// Skip binary files (simple heuristic)
			info, err := d.Info()
			if err != nil || info.Size() > s.cfg.MaxFileSize {
				return nil
			}

			relPath, _ := filepath.Rel(repoPath, path)

			// Search file
			matches := s.searchFile(path, relPath, repoName, regex, opts.Context)
			for _, match := range matches {
				if len(results) >= opts.MaxResults {
					return filepath.SkipAll
				}
				results = append(results, match)
			}

			return nil
		})
	}

	return results, nil
}

func (s *serviceImpl) searchFile(fullPath, relPath, repoName string, regex *regexp.Regexp, contextLines int) []CodeSearchResult {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []CodeSearchResult
	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for i, line := range lines {
		if regex.MatchString(line) {
			// Build context
			var context []string
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			end := i + contextLines + 1
			if end > len(lines) {
				end = len(lines)
			}

			for j := start; j < end; j++ {
				if j != i {
					context = append(context, lines[j])
				}
			}

			results = append(results, CodeSearchResult{
				Repository: repoName,
				FilePath:   relPath,
				Line:       i + 1,
				Content:    line,
				Context:    context,
			})
		}
	}

	return results
}

// FindFiles finds files matching a pattern in a repository
func (s *serviceImpl) FindFiles(ctx context.Context, pattern string, repoName string) ([]FileInfo, error) {
	var repos []string
	if repoName != "" {
		if err := utils.ValidateRepoName(repoName); err != nil {
			return nil, err
		}
		repos = []string{repoName}
	} else {
		repoList, err := s.ListRepositories(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range repoList {
			repos = append(repos, r.Name)
		}
	}

	var files []FileInfo

	for _, repo := range repos {
		repoPath := filepath.Join(s.cfg.Root, repo)

		filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			if utils.IsExcluded(d.Name(), s.cfg.ExcludePatterns) {
				return nil
			}

			matched, _ := filepath.Match(pattern, d.Name())
			if !matched {
				return nil
			}

			relPath, _ := filepath.Rel(repoPath, path)
			info, _ := d.Info()

			files = append(files, FileInfo{
				Path:         filepath.Join(repo, relPath),
				Name:         d.Name(),
				IsDirectory:  false,
				Size:         info.Size(),
				LastModified: info.ModTime(),
				Language:     utils.DetectLanguage(d.Name()),
			})

			return nil
		})
	}

	return files, nil
}
