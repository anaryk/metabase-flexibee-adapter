//go:build integration

package flexibee

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	demoURL     = "https://demo.flexibee.eu"
	demoCompany = "demo"
	demoUser    = "winstrom"
	demoPass    = "winstrom"
)

var integrationLogger = slog.New(slog.DiscardHandler)

func demoClient() *Client {
	return NewClient(demoURL, demoCompany, demoUser, demoPass, integrationLogger)
}

func ctxWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// --- Evidence fetch tests ---

func TestIntegration_FetchProdejka(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "prodejka", FetchOptions{
		Limit:       5,
		Detail:      "full",
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount, "rowCount should be present")
	assert.Greater(t, *resp.Winstrom.RowCount, 0, "demo should have prodejka records")
	assert.NotEmpty(t, resp.Winstrom.Records)
	assert.LessOrEqual(t, len(resp.Winstrom.Records), 5)

	// Verify expected fields exist
	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	assert.Contains(t, rec, "sumCelkem")
	assert.Contains(t, rec, "lastUpdate")
	t.Logf("prodejka: %d total, first kod=%v", *resp.Winstrom.RowCount, rec["kod"])
}

func TestIntegration_FetchCenik(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "cenik", FetchOptions{
		Limit:       3,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0, "demo should have cenik records")
	assert.NotEmpty(t, resp.Winstrom.Records)

	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	t.Logf("cenik: %d total, first kod=%v", *resp.Winstrom.RowCount, rec["kod"])
}

func TestIntegration_FetchSklad(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "sklad", FetchOptions{AddRowCount: true})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)

	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	assert.Contains(t, rec, "nazev")
	t.Logf("sklad: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchSkladovyPohyb(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "skladovy-pohyb", FetchOptions{
		Limit:       2,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)
	t.Logf("skladovy-pohyb: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchFakturaVydana(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "faktura-vydana", FetchOptions{
		Limit:       3,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)

	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	assert.Contains(t, rec, "sumCelkem")
	t.Logf("faktura-vydana: %d total, first kod=%v", *resp.Winstrom.RowCount, rec["kod"])
}

func TestIntegration_FetchPokladna(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "pokladna", FetchOptions{AddRowCount: true})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)

	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	t.Logf("pokladna: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchPokladniPohyb(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "pokladni-pohyb", FetchOptions{
		Limit:       2,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)
	t.Logf("pokladni-pohyb: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchStredisko(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "stredisko", FetchOptions{AddRowCount: true})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)

	rec := resp.Winstrom.Records[0]
	assert.Contains(t, rec, "id")
	assert.Contains(t, rec, "kod")
	assert.Contains(t, rec, "nazev")
	t.Logf("stredisko: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchAdresar(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "adresar", FetchOptions{
		Limit:       3,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)
	t.Logf("adresar: %d total", *resp.Winstrom.RowCount)
}

func TestIntegration_FetchBanka(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "banka", FetchOptions{
		Limit:       2,
		AddRowCount: true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount)
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
	assert.NotEmpty(t, resp.Winstrom.Records)
	t.Logf("banka: %d total", *resp.Winstrom.RowCount)
}

// --- rowCount parsing (string vs number) ---

func TestIntegration_RowCountIsString(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	resp, err := c.FetchEvidence(ctx, "stredisko", FetchOptions{AddRowCount: true})
	require.NoError(t, err)
	require.NotNil(t, resp.Winstrom.RowCount, "@rowCount must be parsed even though Flexibee sends it as a string")
	assert.Greater(t, *resp.Winstrom.RowCount, 0)
}

// --- Properties endpoint ---

func TestIntegration_FetchProperties_Prodejka(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	props, err := c.FetchEvidenceProperties(ctx, "prodejka")
	require.NoError(t, err)
	assert.NotEmpty(t, props)

	// Check some expected properties exist
	propNames := make(map[string]string)
	for _, p := range props {
		propNames[p.Name] = p.Type
	}
	assert.Contains(t, propNames, "id")
	assert.Contains(t, propNames, "kod")
	assert.Contains(t, propNames, "sumCelkem")
	assert.Contains(t, propNames, "lastUpdate")
	assert.Contains(t, propNames, "datVyst")

	t.Logf("prodejka properties: %d fields", len(props))
	t.Logf("  id type=%s, kod type=%s, sumCelkem type=%s, datVyst type=%s",
		propNames["id"], propNames["kod"], propNames["sumCelkem"], propNames["datVyst"])
}

func TestIntegration_FetchProperties_Cenik(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	props, err := c.FetchEvidenceProperties(ctx, "cenik")
	require.NoError(t, err)
	assert.NotEmpty(t, props)

	propNames := make(map[string]string)
	for _, p := range props {
		propNames[p.Name] = p.Type
	}
	assert.Contains(t, propNames, "id")
	assert.Contains(t, propNames, "kod")
	assert.Contains(t, propNames, "nazev")
	assert.Contains(t, propNames, "cenaZakl")

	t.Logf("cenik properties: %d fields", len(props))
}

// --- Pagination ---

func TestIntegration_Pagination_Cenik(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	// Cenik has thousands of records, paginate with small pages
	it := c.IterateEvidence(ctx, "cenik", FetchOptions{Limit: 20})

	var totalRecords int
	pages := 0
	for {
		records, err := it.Next(ctx)
		require.NoError(t, err)
		if records == nil {
			break
		}
		totalRecords += len(records)
		pages++

		// Don't fetch everything, stop after a few pages
		if pages >= 3 {
			break
		}
	}

	assert.Equal(t, 3, pages)
	assert.Equal(t, 60, totalRecords, "3 pages * 20 records")
	t.Logf("cenik pagination: fetched %d records in %d pages", totalRecords, pages)
}

func TestIntegration_Pagination_SmallEvidence(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	// Stredisko has ~40 records, fetch with page size 100 → should be 1 page
	it := c.IterateEvidence(ctx, "stredisko", FetchOptions{Limit: 100})

	var totalRecords int
	pages := 0
	for {
		records, err := it.Next(ctx)
		require.NoError(t, err)
		if records == nil {
			break
		}
		totalRecords += len(records)
		pages++
	}

	assert.Equal(t, 1, pages, "stredisko should fit in a single page")
	assert.Greater(t, totalRecords, 0)
	t.Logf("stredisko: fetched all %d records in %d page", totalRecords, pages)
}

// --- Incremental sync filter ---

func TestIntegration_FilterParameter(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	// Verify filter parameter is sent correctly in the request.
	// Note: the demo instance ignores filter parameters and always returns
	// all records. This test only verifies the API accepts the filter
	// without errors — actual filtering works on paid instances.
	since := time.Now().AddDate(-1, 0, 0)
	resp, err := c.FetchEvidence(ctx, "stredisko", FetchOptions{
		Limit:       5,
		AddRowCount: true,
		Filter:      "lastUpdate > '" + since.Format(time.RFC3339) + "'",
	})
	require.NoError(t, err, "filter parameter should not cause an error")
	require.NotNil(t, resp.Winstrom.RowCount)
	t.Logf("stredisko with filter: %d records (demo ignores filters)", *resp.Winstrom.RowCount)
}

// --- All registered evidence types ---

func TestIntegration_AllRegisteredEvidences(t *testing.T) {
	c := demoClient()
	ctx := ctxWithTimeout(t)

	evidences := []string{
		"prodejka", "faktura-vydana", "faktura-prijata",
		"pohledavka", "zavazek",
		"objednavka-prijata", "objednavka-vydana",
		"nabidka-vydana", "nabidka-prijata",
		"poptavka-vydana", "poptavka-prijata",
		"sklad", "skladovy-pohyb", "skladova-karta",
		"adresar", "kontakt",
		"banka", "pokladni-pohyb", "bankovni-ucet", "pokladna",
		"cenik", "skupina-zbozi", "merna-jednotka",
		"stredisko", "zakazka", "cinnost", "ucet", "sazba-dph", "kurz",
		"smlouva", "dodavatelska-smlouva",
		"majetek",
	}

	for _, ev := range evidences {
		t.Run(ev, func(t *testing.T) {
			resp, err := c.FetchEvidence(ctx, ev, FetchOptions{
				Limit:       1,
				AddRowCount: true,
			})
			require.NoError(t, err, "evidence %s should be fetchable", ev)
			require.NotNil(t, resp.Winstrom.RowCount, "evidence %s should return @rowCount", ev)
			t.Logf("%s: %d records", ev, *resp.Winstrom.RowCount)
		})
	}
}
