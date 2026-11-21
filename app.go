package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---------- Types ----------

type Collection struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ParentID  *string `json:"parentId"`
	Path      string  `json:"path"`
	HasNotes  bool    `json:"hasNotes"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   any    `json:"content"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Vault struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ---------- App ----------

type App struct {
	app *application.App
}

func NewApp() *App {
	return &App{}
}

func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.app = application.Get()
	return nil
}

// ---------- Helpers ----------

func (a *App) configFilePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, "luma")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.json"), nil
}

func (a *App) SaveVaultPath(path string) error {
	cfgPath, err := a.configFilePath()
	if err != nil {
		return err
	}
	data := map[string]string{"path": path}
	b, _ := json.Marshal(data)
	return os.WriteFile(cfgPath, b, 0644)
}

func (a *App) loadSavedVaultPath() string {
	cfgPath, err := a.configFilePath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return ""
	}
	var data map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return ""
	}
	return data["path"]
}

func safeName(name string) string {
	name = strings.TrimSpace(name)
	r := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	return r.Replace(name)
}

func readCollectionMeta(folderPath string, parentID *string) Collection {
	metaPath := filepath.Join(folderPath, "collection.json")
	if b, err := os.ReadFile(metaPath); err == nil {
		var c Collection
		if err := json.Unmarshal(b, &c); err == nil {
			if c.Path == "" {
				c.Path = folderPath
			}
			return c
		}
	}
	info, _ := os.Stat(folderPath)
	created := ""
	updated := ""
	if info != nil {
		created = info.ModTime().Format(time.RFC3339)
		updated = created
	}
	return Collection{
		ID:        filepath.Base(folderPath),
		Name:      filepath.Base(folderPath),
		ParentID:  parentID,
		Path:      folderPath,
		HasNotes:  false,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}

func writeCollectionMeta(c Collection) error {
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(filepath.Join(c.Path, "collection.json"), b, 0644)
}

// ---------- Vault ----------

func (a *App) GetVault() (*Vault, error) {
	path := a.loadSavedVaultPath()
	if path == "" {
		return nil, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil
	}
	if !info.IsDir() {
		return nil, nil
	}
	vault := &Vault{
		ID:        filepath.Base(path),
		Name:      filepath.Base(path),
		Path:      path,
		CreatedAt: info.ModTime().Format(time.RFC3339),
		UpdatedAt: info.ModTime().Format(time.RFC3339),
	}
	return vault, nil
}

func (a *App) ImportVault() (*Vault, error) {
	path, err := a.app.Dialog.OpenFile().
		SetTitle("Select Vault Folder").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("folder does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a folder")
	}
	if err := a.SaveVaultPath(path); err != nil {
		fmt.Printf("Warning: Could not save vault path: %v\n", err)
	}
	vault := &Vault{
		ID:        filepath.Base(path),
		Name:      filepath.Base(path),
		Path:      path,
		CreatedAt: info.ModTime().Format(time.RFC3339),
		UpdatedAt: info.ModTime().Format(time.RFC3339),
	}
	return vault, nil
}

// ---------- Collections ----------

func (a *App) GetCollections(vaultPath string) ([]Collection, error) {
	if vaultPath == "" {
		return []Collection{}, nil
	}
	var result []Collection
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return []Collection{}, nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		folder := filepath.Join(vaultPath, e.Name())
		c := readCollectionMeta(folder, nil)
		files, _ := os.ReadDir(folder)
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") && f.Name() != "collection.json" {
				c.HasNotes = true
				break
			}
		}
		result = append(result, c)
		subEntries, _ := os.ReadDir(folder)
		for _, se := range subEntries {
			if !se.IsDir() {
				continue
			}
			if strings.HasPrefix(se.Name(), ".") {
				continue
			}
			subFolder := filepath.Join(folder, se.Name())
			parentID := c.ID
			sc := readCollectionMeta(subFolder, &parentID)
			subFiles, _ := os.ReadDir(subFolder)
			for _, f := range subFiles {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") && f.Name() != "collection.json" {
					sc.HasNotes = true
					break
				}
			}
			result = append(result, sc)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	if result == nil {
		result = []Collection{}
	}
	return result, nil
}

func (a *App) CreateCollection(parentPath string, name string) (Collection, error) {
	name = safeName(name)
	if name == "" {
		return Collection{}, fmt.Errorf("name is empty")
	}
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		return Collection{}, err
	}
	folder := filepath.Join(parentPath, name)
	if err := os.MkdirAll(folder, 0755); err != nil {
		return Collection{}, err
	}
	now := time.Now().Format(time.RFC3339)
	// parentId nikaalo: agar parentPath ke andar collection.json hai to wo parent hai, warna root hai
	var parentId *string
	if b, err := os.ReadFile(filepath.Join(parentPath, "collection.json")); err == nil {
		var parent Collection
		if err := json.Unmarshal(b, &parent); err == nil {
			parentId = &parent.ID
		}
	}
	c := Collection{
		ID:        name,
		Name:      name,
		ParentID:  parentId,
		Path:      folder,
		HasNotes:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if parentId != nil {
		c.ID = *parentId + "/" + name
	}
	_ = writeCollectionMeta(c)
	return c, nil
}

func (a *App) DeleteCollection(collectionPath string) error {
	if _, err := os.Stat(collectionPath); err != nil {
		return fmt.Errorf("collection not found")
	}
	return os.RemoveAll(collectionPath)
}

func (a *App) RenameCollection(collectionPath string, newName string) (Collection, error) {
	newName = safeName(newName)
	if _, err := os.Stat(collectionPath); err != nil {
		return Collection{}, fmt.Errorf("collection not found")
	}
	// pehle purana meta padh lo
	old := readCollectionMeta(collectionPath, nil)
	newPath := filepath.Join(filepath.Dir(collectionPath), newName)
	if err := os.Rename(collectionPath, newPath); err != nil {
		return Collection{}, err
	}
	old.Name = newName
	old.Path = newPath
	old.UpdatedAt = time.Now().Format(time.RFC3339)
	if !strings.Contains(old.ID, "/") {
		old.ID = newName
	}
	_ = writeCollectionMeta(old)
	return old, nil
}

// ---------- Notes ----------

func (a *App) resolveCollectionFolder(vaultPath string, collectionPath string) string {
	if info, err := os.Stat(collectionPath); err == nil && info.IsDir() {
		return collectionPath
	}
	cols, _ := a.GetCollections(vaultPath)
	for _, c := range cols {
		if c.ID == collectionPath {
			return c.Path
		}
	}
	return filepath.Join(vaultPath, collectionPath)
}

func (a *App) resolveNotePath(notePath string) string {
	// agar ye pehle se poora path hai to wahi
	if _, err := os.Stat(notePath); err == nil {
		return notePath
	}
	vaultPath := a.loadSavedVaultPath()
	if vaultPath == "" {
		return notePath
	}
	// warna vault me dhoondo: file ka naam ID.json hai ya ID hai
	var found string
	_ = filepath.Walk(vaultPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		base := filepath.Base(p)
		if base == notePath || base == notePath+".json" {
			found = p
			return filepath.SkipAll
		}
		// andar ID bhi check karo
		if strings.HasSuffix(p, ".json") {
			b, err := os.ReadFile(p)
			if err == nil {
				var n Note
				if json.Unmarshal(b, &n) == nil && n.ID == notePath {
					found = p
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if found != "" {
		return found
	}
	return notePath
}
func (a *App) GetNotes(collectionPath string) ([]Note, error) {
	folder := collectionPath
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		vaultPath := a.loadSavedVaultPath()
		if vaultPath != "" {
			folder = a.resolveCollectionFolder(vaultPath, collectionPath)
		}
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return []Note{}, nil
	}
	var result []Note
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") || e.Name() == "collection.json" {
			continue
		}
		full := filepath.Join(folder, e.Name())
		b, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		var n Note
		if err := json.Unmarshal(b, &n); err != nil {
			continue
		}
		if n.Path == "" {
			n.Path = full
		}
		result = append(result, n)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})
	if result == nil {
		result = []Note{}
	}
	return result, nil
}

func (a *App) GetNote(notePath string) (*Note, error) {
	notePath = a.resolveNotePath(notePath)
	b, err := os.ReadFile(notePath)
	if err != nil {
		return nil, err
	}
	var n Note
	if err := json.Unmarshal(b, &n); err != nil {
		return nil, err
	}
	if n.Path == "" {
		n.Path = notePath
	}
	return &n, nil
}

// Editor me ReadNote naam se bhi call ho raha hai, isliye alias
func (a *App) ReadNote(notePath string) (*Note, error) {
	return a.GetNote(notePath)
}

func (a *App) CreateNote(collectionPath string, title string) (Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Untitled"
	}
	folder := collectionPath
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		vaultPath := a.loadSavedVaultPath()
		if vaultPath != "" {
			folder = a.resolveCollectionFolder(vaultPath, collectionPath)
		}
	}
	if err := os.MkdirAll(folder, 0755); err != nil {
		return Note{}, err
	}
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	now := time.Now().Format(time.RFC3339)
	filePath := filepath.Join(folder, id+".json")
	n := Note{
		ID:        id,
		Title:     title,
		Content:   map[string]any{"type": "doc", "content": []any{}},
		Path:      filePath,
		CreatedAt: now,
		UpdatedAt: now,
	}
	b, _ := json.MarshalIndent(n, "", "  ")
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		return Note{}, err
	}
	return n, nil
}

func (a *App) UpdateNote(notePath string, title string, content any) (Note, error) {
	notePath = a.resolveNotePath(notePath)
	n, err := a.GetNote(notePath)
	if err != nil {
		return Note{}, err
	}
	if title != "" {
		n.Title = title
	}
	if content != nil {
		n.Content = content
	}
	n.UpdatedAt = time.Now().Format(time.RFC3339)
	b, _ := json.MarshalIndent(n, "", "  ")
	if err := os.WriteFile(notePath, b, 0644); err != nil {
		return Note{}, err
	}
	return *n, nil
}

func (a *App) SaveNote(notePath string, title string, content any) (Note, error) {
	return a.UpdateNote(notePath, title, content)
}

func (a *App) DeleteNote(notePath string) error {
	notePath = a.resolveNotePath(notePath)
	if _, err := os.Stat(notePath); err != nil {
		vaultPath := a.loadSavedVaultPath()
		if vaultPath != "" {
			var found string
			_ = filepath.Walk(vaultPath, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && filepath.Base(p) == filepath.Base(notePath) {
					found = p
				}
				return nil
			})
			if found != "" {
				notePath = found
			}
		}
	}
	return os.Remove(notePath)
}

func (a *App) RenameNote(notePath string, newTitle string) (Note, error) {
	notePath = a.resolveNotePath(notePath)
	n, err := a.GetNote(notePath)
	if err != nil {
		return Note{}, err
	}
	n.Title = strings.TrimSpace(newTitle)
	n.UpdatedAt = time.Now().Format(time.RFC3339)
	b, _ := json.MarshalIndent(n, "", "  ")
	if err := os.WriteFile(notePath, b, 0644); err != nil {
		return Note{}, err
	}
	return *n, nil
}

func (a *App) SearchNotes(vaultPath string, query string) ([]Note, error) {
	if query == "" {
		return []Note{}, nil
	}
	q := strings.ToLower(query)
	var result []Note
	_ = filepath.Walk(vaultPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".json") || filepath.Base(p) == "collection.json" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		var n Note
		if err := json.Unmarshal(b, &n); err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(n.Title), q) {
			if n.Path == "" {
				n.Path = p
			}
			result = append(result, n)
		}
		return nil
	})
	if result == nil {
		result = []Note{}
	}
	return result, nil
}
