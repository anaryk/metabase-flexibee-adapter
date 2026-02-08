package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	r := New()
	r.Register(Evidence{
		Slug:       "prodejka",
		Table:      "flexibee_prodejka",
		PrimaryKey: "id",
	})

	ev, ok := r.Get("prodejka")
	require.True(t, ok)
	assert.Equal(t, "flexibee_prodejka", ev.Table)
	assert.Equal(t, "id", ev.PrimaryKey)
}

func TestRegistry_GetNotFound(t *testing.T) {
	t.Parallel()

	r := New()
	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_All_PreservesOrder(t *testing.T) {
	t.Parallel()

	r := New()
	r.Register(Evidence{Slug: "b", Table: "t_b", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "a", Table: "t_a", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "c", Table: "t_c", PrimaryKey: "id"})

	all := r.All()
	require.Len(t, all, 3)
	assert.Equal(t, "b", all[0].Slug)
	assert.Equal(t, "a", all[1].Slug)
	assert.Equal(t, "c", all[2].Slug)
}

func TestRegistry_Len(t *testing.T) {
	t.Parallel()

	r := New()
	assert.Equal(t, 0, r.Len())

	r.Register(Evidence{Slug: "a", Table: "t_a", PrimaryKey: "id"})
	assert.Equal(t, 1, r.Len())

	r.Register(Evidence{Slug: "b", Table: "t_b", PrimaryKey: "id"})
	assert.Equal(t, 2, r.Len())
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	t.Parallel()

	r := New()
	r.Register(Evidence{Slug: "a", Table: "t_a", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "a", Table: "t_a_updated", PrimaryKey: "id"})

	assert.Equal(t, 1, r.Len())
	ev, ok := r.Get("a")
	require.True(t, ok)
	assert.Equal(t, "t_a_updated", ev.Table)
}

func TestNewDefault_Completeness(t *testing.T) {
	t.Parallel()

	r := NewDefault()

	expectedSlugs := []string{
		// Sales & Invoicing
		"prodejka", "faktura-vydana", "faktura-prijata", "pohledavka", "zavazek",
		// Orders
		"objednavka-prijata", "objednavka-vydana", "nabidka-vydana",
		"nabidka-prijata", "poptavka-vydana", "poptavka-prijata",
		// Inventory
		"sklad", "skladovy-pohyb", "skladova-karta",
		// Contacts
		"adresar", "kontakt",
		// Cash & Banking
		"banka", "pokladni-pohyb", "bankovni-ucet", "pokladna",
		// Products
		"cenik", "skupina-zbozi", "merna-jednotka",
		// Accounting
		"stredisko", "zakazka", "cinnost", "ucet", "sazba-dph", "kurz",
		// Contracts
		"smlouva", "dodavatelska-smlouva",
		// Assets
		"majetek",
	}

	assert.Equal(t, len(expectedSlugs), r.Len(), "registry should have %d evidence types", len(expectedSlugs))

	for _, slug := range expectedSlugs {
		ev, ok := r.Get(slug)
		if !assert.True(t, ok, "missing evidence: %s", slug) {
			continue
		}
		assert.NotEmpty(t, ev.Table, "table should be set for %s", slug)
		assert.Equal(t, "id", ev.PrimaryKey, "primary key should be 'id' for %s", slug)
	}
}

func TestNewDefault_MasterDataFlags(t *testing.T) {
	t.Parallel()

	r := NewDefault()

	masterData := []string{
		"sklad", "skladova-karta", "adresar", "kontakt",
		"bankovni-ucet", "pokladna", "cenik", "skupina-zbozi",
		"merna-jednotka", "stredisko", "zakazka", "cinnost",
		"ucet", "sazba-dph", "majetek",
	}

	transactional := []string{
		"prodejka", "faktura-vydana", "faktura-prijata",
		"banka", "pokladni-pohyb", "skladovy-pohyb", "kurz",
		"smlouva", "dodavatelska-smlouva",
	}

	for _, slug := range masterData {
		ev, _ := r.Get(slug)
		assert.True(t, ev.IsMasterData, "%s should be master data", slug)
	}

	for _, slug := range transactional {
		ev, _ := r.Get(slug)
		assert.False(t, ev.IsMasterData, "%s should be transactional", slug)
	}
}
