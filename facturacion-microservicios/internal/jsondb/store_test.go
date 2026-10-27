package jsondb

import (
	"os"
	"path/filepath"
	"testing"
)

// Tipo de prueba
type Item struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

func tempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func newTestStore(t *testing.T, startID int) *Store[Item] {
	t.Helper()
	dir := tempDir(t)
	fp := filepath.Join(dir, "test.json")
	store, err := New[Item](fp, startID, func(i *Item) int { return i.ID }, func(i *Item, id int) { i.ID = id })
	if err != nil {
		t.Fatalf("Error creando store: %v", err)
	}
	return store
}

func TestNuevoArchivoCreadoVacio(t *testing.T) {
	store := newTestStore(t, 1)
	items, err := store.GetAll()
	if err != nil {
		t.Fatalf("Error en GetAll: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Se esperaban 0 items, se obtuvieron %d", len(items))
	}
}

func TestCreateYGetByID(t *testing.T) {
	store := newTestStore(t, 10)

	item, err := store.Create(Item{Nombre: "Prueba 1"})
	if err != nil {
		t.Fatalf("Error en Create: %v", err)
	}
	if item.ID != 10 {
		t.Errorf("Se esperaba ID 10, se obtuvo %d", item.ID)
	}

	item2, err := store.Create(Item{Nombre: "Prueba 2"})
	if err != nil {
		t.Fatalf("Error en Create: %v", err)
	}
	if item2.ID != 11 {
		t.Errorf("Se esperaba ID 11, se obtuvo %d", item2.ID)
	}

	found, err := store.GetByID(10)
	if err != nil {
		t.Fatalf("Error en GetByID: %v", err)
	}
	if found == nil {
		t.Fatal("No se encontró el item con ID 10")
	}
	if found.Nombre != "Prueba 1" {
		t.Errorf("Se esperaba 'Prueba 1', se obtuvo '%s'", found.Nombre)
	}
}

func TestGetByIDNoExiste(t *testing.T) {
	store := newTestStore(t, 1)
	found, err := store.GetByID(999)
	if err != nil {
		t.Fatalf("Error en GetByID: %v", err)
	}
	if found != nil {
		t.Error("Se esperaba nil para un ID inexistente")
	}
}

func TestUpdate(t *testing.T) {
	store := newTestStore(t, 1)

	_, err := store.Create(Item{Nombre: "Original"})
	if err != nil {
		t.Fatalf("Error en Create: %v", err)
	}

	updated, ok, err := store.Update(1, func(i *Item) {
		i.Nombre = "Modificado"
	})
	if err != nil {
		t.Fatalf("Error en Update: %v", err)
	}
	if !ok {
		t.Fatal("Update retornó false, se esperaba true")
	}
	if updated.Nombre != "Modificado" {
		t.Errorf("Se esperaba 'Modificado', se obtuvo '%s'", updated.Nombre)
	}

	// Verificar que se persistió
	found, err := store.GetByID(1)
	if err != nil {
		t.Fatalf("Error en GetByID: %v", err)
	}
	if found.Nombre != "Modificado" {
		t.Errorf("El cambio no se persistió: '%s'", found.Nombre)
	}
}

func TestUpdateNoExiste(t *testing.T) {
	store := newTestStore(t, 1)
	_, ok, err := store.Update(999, func(i *Item) { i.Nombre = "X" })
	if err != nil {
		t.Fatalf("Error en Update: %v", err)
	}
	if ok {
		t.Error("Update retornó true para un ID inexistente")
	}
}

func TestDelete(t *testing.T) {
	store := newTestStore(t, 1)

	store.Create(Item{Nombre: "A borrar"})

	ok, err := store.Delete(1)
	if err != nil {
		t.Fatalf("Error en Delete: %v", err)
	}
	if !ok {
		t.Fatal("Delete retornó false, se esperaba true")
	}

	found, _ := store.GetByID(1)
	if found != nil {
		t.Error("El item no se eliminó correctamente")
	}
}

func TestDeleteNoExiste(t *testing.T) {
	store := newTestStore(t, 1)
	ok, err := store.Delete(999)
	if err != nil {
		t.Fatalf("Error en Delete: %v", err)
	}
	if ok {
		t.Error("Delete retornó true para un ID inexistente")
	}
}

func TestFindOne(t *testing.T) {
	store := newTestStore(t, 1)
	store.Create(Item{Nombre: "Alpha"})
	store.Create(Item{Nombre: "Beta"})

	found, err := store.FindOne(func(i *Item) bool { return i.Nombre == "Beta" })
	if err != nil {
		t.Fatalf("Error en FindOne: %v", err)
	}
	if found == nil {
		t.Fatal("No se encontró el item")
	}
	if found.ID != 2 {
		t.Errorf("Se esperaba ID 2, se obtuvo %d", found.ID)
	}
}

func TestFilter(t *testing.T) {
	store := newTestStore(t, 1)
	store.Create(Item{Nombre: "Laptop"})
	store.Create(Item{Nombre: "Mouse"})
	store.Create(Item{Nombre: "Laptop Pro"})

	results, err := store.Filter(func(i *Item) bool { return len(i.Nombre) > 5 })
	if err != nil {
		t.Fatalf("Error en Filter: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Se esperaban 2 resultados, se obtuvieron %d", len(results))
	}
}

func TestEscrituraAtomica(t *testing.T) {
	store := newTestStore(t, 1)
	store.Create(Item{Nombre: "Test"})

	// Verificar que no quedan archivos temporales
	dir := filepath.Dir(store.filePath)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" && e.Name() != filepath.Base(store.filePath) {
			t.Errorf("Archivo temporal no limpiado: %s", e.Name())
		}
	}
}

func TestGetAllRetornaArregloVacio(t *testing.T) {
	store := newTestStore(t, 1)
	items, err := store.GetAll()
	if err != nil {
		t.Fatalf("Error en GetAll: %v", err)
	}
	if items == nil {
		t.Error("GetAll retornó nil en lugar de arreglo vacío")
	}
}

func TestWithLock(t *testing.T) {
	store := newTestStore(t, 1)
	store.Create(Item{Nombre: "A"})
	store.Create(Item{Nombre: "B"})

	err := store.WithLock(func(items []Item) ([]Item, error) {
		for i := range items {
			items[i].Nombre = items[i].Nombre + "_modificado"
		}
		return items, nil
	})
	if err != nil {
		t.Fatalf("Error en WithLock: %v", err)
	}

	items, _ := store.GetAll()
	if items[0].Nombre != "A_modificado" || items[1].Nombre != "B_modificado" {
		t.Error("WithLock no aplicó los cambios correctamente")
	}
}

func TestIDAutoincrementalDesdeStartID(t *testing.T) {
	store := newTestStore(t, 100)
	item, _ := store.Create(Item{Nombre: "Primero"})
	if item.ID != 100 {
		t.Errorf("Se esperaba ID 100, se obtuvo %d", item.ID)
	}
	item2, _ := store.Create(Item{Nombre: "Segundo"})
	if item2.ID != 101 {
		t.Errorf("Se esperaba ID 101, se obtuvo %d", item2.ID)
	}
}

func TestIDContinuaDespueDatosSemilla(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.json")
	// Escribir datos semilla con IDs existentes
	os.WriteFile(fp, []byte(`[{"id":101,"nombre":"A"},{"id":103,"nombre":"B"}]`), 0644)

	store, err := New[Item](fp, 104, func(i *Item) int { return i.ID }, func(i *Item, id int) { i.ID = id })
	if err != nil {
		t.Fatalf("Error creando store: %v", err)
	}

	item, _ := store.Create(Item{Nombre: "Nuevo"})
	if item.ID != 104 {
		t.Errorf("Se esperaba ID 104, se obtuvo %d (debió continuar después de los datos semilla)", item.ID)
	}
}
