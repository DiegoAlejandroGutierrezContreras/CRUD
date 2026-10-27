// Package jsondb implementa un almacén genérico sobre archivos JSON con acceso concurrente seguro.
package jsondb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store es un almacén genérico que lee y escribe un arreglo de objetos JSON.
type Store[T any] struct {
	mu       sync.RWMutex
	filePath string
	nextID   int
	getID    func(*T) int
	setID    func(*T, int)
}

// New crea un nuevo Store. Si el archivo no existe, lo crea con un arreglo vacío [].
func New[T any](filePath string, startID int, getID func(*T) int, setID func(*T, int)) (*Store[T], error) {
	s := &Store[T]{
		filePath: filePath,
		nextID:   startID,
		getID:    getID,
		setID:    setID,
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio %s: %w", dir, err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := s.writeFile([]T{}); err != nil {
			return nil, fmt.Errorf("error creando archivo %s: %w", filePath, err)
		}
	} else {
		items, err := s.readFile()
		if err != nil {
			return nil, fmt.Errorf("error leyendo archivo %s: %w", filePath, err)
		}
		for _, item := range items {
			id := s.getID(&item)
			if id >= s.nextID {
				s.nextID = id + 1
			}
		}
	}

	return s, nil
}

// readFile lee el archivo JSON y retorna el arreglo de elementos.
func (s *Store[T]) readFile() ([]T, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo %s: %w", s.filePath, err)
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("error decodificando %s: %w", s.filePath, err)
	}

	return items, nil
}

// writeFile escribe el arreglo al archivo JSON de forma atómica.
func (s *Store[T]) writeFile(items []T) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("error codificando JSON: %w", err)
	}

	dir := filepath.Dir(s.filePath)
	tmpFile, err := os.CreateTemp(dir, "tmp-*.json")
	if err != nil {
		return fmt.Errorf("error creando archivo temporal: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("error escribiendo archivo temporal: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error cerrando archivo temporal: %w", err)
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error renombrando archivo: %w", err)
	}

	return nil
}

// GetAll retorna todos los elementos.
func (s *Store[T]) GetAll() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readFile()
}

// GetByID retorna un elemento por su ID. Retorna nil si no existe.
func (s *Store[T]) GetByID(id int) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, err := s.readFile()
	if err != nil {
		return nil, err
	}

	for i := range items {
		if s.getID(&items[i]) == id {
			return &items[i], nil
		}
	}

	return nil, nil
}

// Create agrega un nuevo elemento con ID autoincremental y lo retorna.
func (s *Store[T]) Create(item T) (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readFile()
	if err != nil {
		return item, err
	}

	s.setID(&item, s.nextID)
	s.nextID++

	items = append(items, item)

	if err := s.writeFile(items); err != nil {
		s.nextID--
		return item, err
	}

	return item, nil
}

// Update actualiza un elemento existente por su ID. Retorna false si no existe.
func (s *Store[T]) Update(id int, updater func(*T)) (T, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readFile()
	if err != nil {
		var zero T
		return zero, false, err
	}

	for i := range items {
		if s.getID(&items[i]) == id {
			updater(&items[i])
			if err := s.writeFile(items); err != nil {
				var zero T
				return zero, false, err
			}
			return items[i], true, nil
		}
	}

	var zero T
	return zero, false, nil
}

// Delete elimina un elemento por su ID. Retorna false si no existe.
func (s *Store[T]) Delete(id int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readFile()
	if err != nil {
		return false, err
	}

	idx := -1
	for i := range items {
		if s.getID(&items[i]) == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, nil
	}

	items = append(items[:idx], items[idx+1:]...)

	if err := s.writeFile(items); err != nil {
		return false, err
	}

	return true, nil
}

// FindOne busca el primer elemento que cumpla el predicado. Retorna nil si no lo encuentra.
func (s *Store[T]) FindOne(predicate func(*T) bool) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, err := s.readFile()
	if err != nil {
		return nil, err
	}

	for i := range items {
		if predicate(&items[i]) {
			return &items[i], nil
		}
	}

	return nil, nil
}
