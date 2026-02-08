package registry

// Evidence describes a Flexibee evidence type and its mapping to PostgreSQL.
type Evidence struct {
	Slug         string // Flexibee evidence slug (e.g. "prodejka")
	Table        string // PostgreSQL table name (e.g. "flexibee_prodejka")
	PrimaryKey   string // Primary key field (always "id")
	IsMasterData bool   // Master/reference data - never cleaned up
}

// Registry holds all registered evidence types.
type Registry struct {
	evidences map[string]Evidence
	order     []string // preserves registration order
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{
		evidences: make(map[string]Evidence),
	}
}

// Register adds an evidence type to the registry.
func (r *Registry) Register(ev Evidence) {
	if _, exists := r.evidences[ev.Slug]; !exists {
		r.order = append(r.order, ev.Slug)
	}
	r.evidences[ev.Slug] = ev
}

// Get returns an evidence type by slug.
func (r *Registry) Get(slug string) (Evidence, bool) {
	ev, ok := r.evidences[slug]
	return ev, ok
}

// All returns all registered evidence types in registration order.
func (r *Registry) All() []Evidence {
	result := make([]Evidence, 0, len(r.order))
	for _, slug := range r.order {
		result = append(result, r.evidences[slug])
	}
	return result
}

// Len returns the number of registered evidence types.
func (r *Registry) Len() int {
	return len(r.evidences)
}

// NewDefault creates a registry with all known Flexibee evidence types.
func NewDefault() *Registry {
	r := New()

	// Sales & Invoicing (transactional)
	r.Register(Evidence{Slug: "prodejka", Table: "flexibee_prodejka", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "faktura-vydana", Table: "flexibee_faktura_vydana", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "faktura-prijata", Table: "flexibee_faktura_prijata", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "pohledavka", Table: "flexibee_pohledavka", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "zavazek", Table: "flexibee_zavazek", PrimaryKey: "id"})

	// Orders (transactional)
	r.Register(Evidence{Slug: "objednavka-prijata", Table: "flexibee_objednavka_prijata", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "objednavka-vydana", Table: "flexibee_objednavka_vydana", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "nabidka-vydana", Table: "flexibee_nabidka_vydana", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "nabidka-prijata", Table: "flexibee_nabidka_prijata", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "poptavka-vydana", Table: "flexibee_poptavka_vydana", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "poptavka-prijata", Table: "flexibee_poptavka_prijata", PrimaryKey: "id"})

	// Inventory
	r.Register(Evidence{Slug: "sklad", Table: "flexibee_sklad", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "skladovy-pohyb", Table: "flexibee_skladovy_pohyb", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "skladova-karta", Table: "flexibee_skladova_karta", PrimaryKey: "id", IsMasterData: true})

	// Contacts (master data)
	r.Register(Evidence{Slug: "adresar", Table: "flexibee_adresar", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "kontakt", Table: "flexibee_kontakt", PrimaryKey: "id", IsMasterData: true})

	// Cash & Banking
	r.Register(Evidence{Slug: "banka", Table: "flexibee_banka", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "pokladni-pohyb", Table: "flexibee_pokladni_pohyb", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "bankovni-ucet", Table: "flexibee_bankovni_ucet", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "pokladna", Table: "flexibee_pokladna", PrimaryKey: "id", IsMasterData: true})

	// Products (master data)
	r.Register(Evidence{Slug: "cenik", Table: "flexibee_cenik", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "skupina-zbozi", Table: "flexibee_skupina_zbozi", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "merna-jednotka", Table: "flexibee_merna_jednotka", PrimaryKey: "id", IsMasterData: true})

	// Accounting (master data)
	r.Register(Evidence{Slug: "stredisko", Table: "flexibee_stredisko", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "zakazka", Table: "flexibee_zakazka", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "cinnost", Table: "flexibee_cinnost", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "ucet", Table: "flexibee_ucet", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "sazba-dph", Table: "flexibee_sazba_dph", PrimaryKey: "id", IsMasterData: true})
	r.Register(Evidence{Slug: "kurz", Table: "flexibee_kurz", PrimaryKey: "id"})

	// Contracts (transactional)
	r.Register(Evidence{Slug: "smlouva", Table: "flexibee_smlouva", PrimaryKey: "id"})
	r.Register(Evidence{Slug: "dodavatelska-smlouva", Table: "flexibee_dodavatelska_smlouva", PrimaryKey: "id"})

	// Assets (master data)
	r.Register(Evidence{Slug: "majetek", Table: "flexibee_majetek", PrimaryKey: "id", IsMasterData: true})

	return r
}
