package assets

import (
	"os"
	"path/filepath"
	"testing"
)

// assetsRoot resolves to the repo's real /assets directory. Tests run with
// cwd set to this package's directory (internal/assets), so ../../assets is
// the repo root's assets folder.
const assetsRoot = "../../assets"

func TestCatalogEntriesHaveUniqueNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, m := range All() {
		if seen[m.Name] {
			t.Fatalf("duplicate catalog name %q", m.Name)
		}
		seen[m.Name] = true
	}
	if len(seen) == 0 {
		t.Fatalf("expected a non-empty catalog")
	}
}

func TestCatalogEntriesHavePositiveDimensions(t *testing.T) {
	for _, m := range All() {
		if m.Width <= 0 || m.Height <= 0 || m.Depth <= 0 {
			t.Errorf("%s: expected positive Width/Height/Depth, got %+v", m.Name, m)
		}
	}
}

func TestCatalogFilesExistOnDisk(t *testing.T) {
	for _, m := range All() {
		objPath := filepath.Join(assetsRoot, m.RelPath)
		if _, err := os.Stat(objPath); err != nil {
			t.Errorf("%s: obj file missing at %s: %v", m.Name, objPath, err)
		}

		mtlPath := objPath[:len(objPath)-len(filepath.Ext(objPath))] + ".mtl"
		if _, err := os.Stat(mtlPath); err != nil {
			t.Errorf("%s: mtl file missing at %s: %v", m.Name, mtlPath, err)
		}
	}
}

func TestCatalogRelPathMatchesCategoryFolder(t *testing.T) {
	for _, m := range All() {
		want := "models/" + string(m.Category) + "/" + m.Name + ".obj"
		if m.RelPath != want {
			t.Errorf("%s: RelPath %q doesn't match expected %q", m.Name, m.RelPath, want)
		}
	}
}

func TestOnlyAnimalsHaveYawOffset(t *testing.T) {
	for _, m := range All() {
		hasOffset := m.YawOffsetRadians != 0
		isAnimal := m.Category == CategoryAnimals
		if hasOffset != isAnimal {
			t.Errorf("%s: expected YawOffsetRadians != 0 iff category is animals (category=%s, offset=%v)", m.Name, m.Category, m.YawOffsetRadians)
		}
	}
}

func TestByNameAndByCategory(t *testing.T) {
	fox, ok := ByName("fox")
	if !ok || fox.Category != CategoryAnimals {
		t.Fatalf("expected to find fox in animals category, got %+v (ok=%v)", fox, ok)
	}

	if _, ok := ByName("does-not-exist"); ok {
		t.Fatalf("expected lookup of unknown name to fail")
	}

	trees := ByCategory(CategoryTrees)
	if len(trees) != 4 {
		t.Fatalf("expected 4 trees, got %d: %+v", len(trees), trees)
	}
}

func TestFootprintUsesWidestHorizontalExtent(t *testing.T) {
	m := Model{Width: 2, Depth: 1}
	if got := m.Footprint(); got != 1 {
		t.Fatalf("expected footprint 1 (half of the wider extent), got %v", got)
	}
	m.Depth = 4
	if got := m.Footprint(); got != 2 {
		t.Fatalf("expected footprint 2 after widening depth, got %v", got)
	}
}
