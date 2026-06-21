package air

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// BuildAIR assembles the final AIR struct from source context and normalized data.
func BuildAIR(src SourceContext, normPortfolio *AIRPortfolio, normRules []Rule, decisionHierarchy DecisionHierarchy, execConfig ExecutionConfig, prov *Provenance) *AIR {
	air := &AIR{
		Metadata: AIRMetadata{
			StrategyName:    src.Manifest.Name,
			StrategyID:      src.Manifest.StrategyID,
			Description:     src.Manifest.Description,
			Author:          src.Manifest.Author,
			Version:         src.Manifest.Version,
			SpecVersion:     src.Manifest.SpecVersion,
			CompilerVersion: "v0.1.0",
			GeneratedAt:     src.GeneratedAt,
			Tags:            src.Manifest.Tags,
		},
		Universe:          buildUniverse(src),
		Signals:           src.Signals,
		SignalInterests:   src.SignalInterests,
		Relations:         src.Relations,
		Regimes:           src.Regimes,
		Portfolio:         *normPortfolio,
		DecisionHierarchy: decisionHierarchy,
		Rules:             normRules,
		Execution:         execConfig,
		Provenance:        prov,
	}

	return air
}

// knownAssets maps symbols to their canonical names and sectors.
var knownAssets = map[string]struct {
	Name   string
	Sector string
}{
	"QQQ":    {Name: "Invesco QQQ Trust", Sector: "technology"},
	"NVDA":   {Name: "NVIDIA Corporation", Sector: "technology"},
	"AMD":    {Name: "Advanced Micro Devices, Inc.", Sector: "technology"},
	"TLT":    {Name: "iShares 20+ Year Treasury Bond ETF", Sector: "duration"},
	"USO":    {Name: "United States Oil Fund", Sector: "energy"},
	"SPY":    {Name: "SPDR S&P 500 ETF Trust", Sector: "broad_market"},
	"MSFT":   {Name: "Microsoft Corporation", Sector: "technology"},
	"AAPL":   {Name: "Apple Inc.", Sector: "technology"},
	"AVGO":   {Name: "Broadcom Inc.", Sector: "technology"},
	"SMH":    {Name: "VanEck Semiconductor ETF", Sector: "technology"},
	"SHY":    {Name: "iShares 1-3 Year Treasury Bond ETF", Sector: "duration"},
	"IEF":    {Name: "iShares 7-10 Year Treasury Bond ETF", Sector: "duration"},
	"XLE":    {Name: "Energy Select Sector SPDR Fund", Sector: "energy"},
	"XOM":    {Name: "Exxon Mobil Corporation", Sector: "energy"},
	"CVX":    {Name: "Chevron Corporation", Sector: "energy"},
	"XLU":    {Name: "Utilities Select Sector SPDR Fund", Sector: "utilities"},
	"XLP":    {Name: "Consumer Staples Select Sector SPDR Fund", Sector: "consumer_staples"},
	"XLV":    {Name: "Health Care Select Sector SPDR Fund", Sector: "healthcare"},
	"JNJ":    {Name: "Johnson & Johnson", Sector: "healthcare"},
	"PG":     {Name: "Procter & Gamble Co.", Sector: "consumer_staples"},
	"KO":     {Name: "Coca-Cola Company", Sector: "consumer_staples"},
	"IWM":    {Name: "iShares Russell 2000 ETF", Sector: "broad_market"},
	"AGG":    {Name: "iShares Core U.S. Aggregate Bond ETF", Sector: "duration"},
	"DBC":    {Name: "Invesco DB Commodity Index Tracking Fund", Sector: "commodities"},
	"GLD":    {Name: "SPDR Gold Shares", Sector: "commodities"},
	"SLV":    {Name: "iShares Silver Trust", Sector: "commodities"},
	"DIA":    {Name: "SPDR Dow Jones Industrial Average ETF Trust", Sector: "broad_market"},
	"VTI":    {Name: "Vanguard Total Stock Market ETF", Sector: "broad_market"},
	"LQD":    {Name: "iShares iBoxx $ Investment Grade Corporate Bond ETF", Sector: "duration"},
	"XLK":    {Name: "Technology Select Sector SPDR Fund", Sector: "technology"},
	"XLF":    {Name: "Financial Select Sector SPDR Fund", Sector: "financials"},
	"CASH":   {Name: "Cash / risk-free baseline", Sector: ""},
	"XAUUSD": {Name: "Gold Spot Price USD", Sector: "commodities"},
}

func buildUniverse(src SourceContext) AIRUniverse {
	assetMap := make(map[string]*Asset)
	sectorSet := make(map[string]bool)

	// Collect symbols from universe config
	symbols := make([]string, 0)
	seen := make(map[string]bool)

	addSymbol := func(sym string) {
		if seen[sym] {
			return
		}
		seen[sym] = true
		symbols = append(symbols, sym)
	}

	for _, s := range src.Manifest.Universe.Symbols {
		addSymbol(s)
	}
	for _, b := range src.Manifest.Portfolio.CandidateBaskets {
		for _, s := range b.Symbols {
			addSymbol(s)
		}
	}
	if src.Manifest.Portfolio.InitialAllocation != nil {
		for _, p := range src.Manifest.Portfolio.InitialAllocation.Positions {
			addSymbol(p.Symbol)
		}
	}
	// Also collect from backtest benchmarks
	for _, b := range src.Manifest.Backtest.Benchmarks {
		addSymbol(b.Symbol)
	}

	// Build assets with known metadata
	var assetList []Asset
	for _, sym := range symbols {
		// Skip cash from asset list - it's a portfolio allocation target, not a tradable asset.
		if sym == "cash" || sym == "CASH" {
			continue
		}

		known := knownAssets[sym]
		sector := known.Sector
		if sector != "" {
			sectorSet[sector] = true
		}

		asset := Asset{
			Symbol:     sym,
			Name:       known.Name,
			AssetClass: inferAssetClass(sym),
			Sector:     sector,
			Currency:   "USD",
			Tradable:   sym != "cash" && sym != "CASH",
		}
		assetMap[sym] = &asset
		assetList = append(assetList, asset)
	}

	// Sort for determinism
	sort.Slice(assetList, func(i, j int) bool {
		return assetList[i].Symbol < assetList[j].Symbol
	})

	// Build themes from candidate baskets
	themes := buildThemes(src.Manifest.Portfolio.CandidateBaskets)

	// Collect asset classes
	classSet := make(map[string]bool)
	for _, a := range assetList {
		classSet[a.AssetClass] = true
	}
	classSet["cash"] = true

	var classes, sectors []string
	for c := range classSet {
		classes = append(classes, c)
	}
	for s := range sectorSet {
		sectors = append(sectors, s)
	}
	sort.Strings(classes)
	sort.Strings(sectors)

	// Determine benchmark
	benchmark := ""
	for _, b := range src.Manifest.Backtest.Benchmarks {
		if b.Type == "fund" || b.Type == "" {
			benchmark = b.Symbol
			break
		}
	}
	if benchmark == "" {
		benchmark = "SPY"
	}

	return AIRUniverse{
		Assets:       assetList,
		AssetClasses: classes,
		Sectors:      sectors,
		Themes:       themes,
		Benchmark:    benchmark,
	}
}

// buildThemes creates Theme entries from candidate basket definitions.
// Baskets with an identifiable role and clear symbol list become themes.
// Only baskets whose symbol sets overlap with the universe's core basket
// members are considered for inclusion.
func buildThemes(baskets []CandidateBasket) []Theme {
	// Priority baskets for theme generation
	type themeDef struct {
		basketID    string
		themeID     string
		name        string
		description string
		members     []string
	}

	candidates := []themeDef{
		{basketID: "growth_technology", themeID: "growth_technology", name: "Growth Technology", description: "Long-duration growth technology exposure."},
		{basketID: "defensive_equities", themeID: "defensive_equities", name: "Defensive Equities", description: "Defensive equity candidates for risk-off rotation."},
		{basketID: "commodities_energy", themeID: "commodities_energy", name: "Commodities / Energy", description: "Energy and commodity exposure."},
		{basketID: "duration", themeID: "duration", name: "Duration", description: "Treasury duration instruments."},
	}

	// Build lookup of basket IDs to their symbols
	basketSymbols := make(map[string][]string)
	for _, b := range baskets {
		basketSymbols[b.BasketID] = b.Symbols
	}

	var themes []Theme
	for _, c := range candidates {
		syms, ok := basketSymbols[c.basketID]
		if !ok || len(syms) == 0 {
			continue
		}
		// Only include first 3 members for the theme
		members := syms
		if len(members) > 3 {
			members = members[:3]
		}
		themes = append(themes, Theme{
			ThemeID:     c.themeID,
			Name:        c.name,
			Members:     members,
			Description: c.description,
		})
	}

	return themes
}

func inferAssetClass(symbol string) string {
	switch symbol {
	case "SPY", "QQQ", "DIA", "IWM", "VTI":
		return "equities"
	case "TLT", "IEF", "SHY", "AGG", "BND", "LQD":
		return "bonds"
	case "USO", "GLD", "SLV", "DBC", "XAUUSD":
		return "commodities"
	case "XLE", "XLU", "XLP", "XLV", "XLK", "XLF":
		return "equities"
	}
	return "equities"
}

// DefaultDecisionHierarchy returns the standard layer definitions.
func DefaultDecisionHierarchy() DecisionHierarchy {
	return DecisionHierarchy{
		Layers: []Layer{
			{Name: "portfolio_safety", Priority: 100},
			{Name: "risk_management", Priority: 90},
			{Name: "regime", Priority: 80},
			{Name: "cross_asset_relations", Priority: 70},
			{Name: "strategy", Priority: 60},
			{Name: "tactical", Priority: 50},
			{Name: "experimental", Priority: 25},
		},
		ConflictResolution: []string{
			"layer_priority",
			"rule_priority",
			"confidence",
			"tie_breaker",
		},
		TieBreaker: "rule_order",
	}
}

// DefaultExecutionConfig returns standard execution defaults.
func DefaultExecutionConfig() ExecutionConfig {
	return ExecutionConfig{
		RebalanceFrequency:    "daily",
		OrderTiming:           "next_open",
		TransactionCostBps:    1,
		SlippageBps:           2,
		AllowFractionalShares: true,
		DividendHandling:      "cash",
		CorporateActions:      "adjusted_prices",
		ValuationFrequency:    "daily",
	}
}

// EnrichExecutionConfig overlays manifest backtest config onto execution defaults.
func EnrichExecutionConfig(exec ExecutionConfig, bt BacktestConfig) ExecutionConfig {
	if bt.DecisionSampling != nil {
		exec.DecisionSampling = bt.DecisionSampling
	}
	if bt.ValuationFrequency != "" {
		exec.ValuationFrequency = bt.ValuationFrequency
	}
	if len(bt.Benchmarks) > 0 {
		exec.Benchmarks = bt.Benchmarks
	}
	return exec
}

// SourceContext holds all data needed to build AIR.
type SourceContext struct {
	Manifest        *Manifest
	StrategyMD      string
	Rules           []Rule
	Signals         []Signal
	SignalInterests []SignalInterest
	Relations       []Relation
	Regimes         []Regime
	GeneratedAt     string

	// Raw bytes for hashing
	ManifestRaw        []byte
	StrategyRaw        []byte
	RulesRaw           []byte
	SignalsRaw         []byte
	SignalInterestsRaw []byte
}

// CanonicalJSON serializes a value to deterministic JSON with sorted keys.
func CanonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}

	return canonicalMarshal(tree)
}

func canonicalMarshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		result := []byte{'{'}
		for i, k := range keys {
			if i > 0 {
				result = append(result, ',')
			}
			keyBytes, _ := json.Marshal(k)
			result = append(result, keyBytes...)
			result = append(result, ':')
			inner, err := canonicalMarshal(val[k])
			if err != nil {
				return nil, err
			}
			result = append(result, inner...)
		}
		result = append(result, '}')
		return result, nil

	case []any:
		result := []byte{'['}
		for i, item := range val {
			if i > 0 {
				result = append(result, ',')
			}
			inner, err := canonicalMarshal(item)
			if err != nil {
				return nil, err
			}
			result = append(result, inner...)
		}
		result = append(result, ']')
		return result, nil

	default:
		return json.Marshal(val)
	}
}

// HashBytes computes the SHA-256 hash of data, returning a hex string prefixed with "sha256:".
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", h)
}

// HashSource computes SHA-256 hashes for source file bytes.
func HashSource(src *SourceContext) map[string]string {
	hashes := make(map[string]string)
	if src.ManifestRaw != nil {
		hashes["manifest.json"] = HashBytes(src.ManifestRaw)
	}
	if src.StrategyRaw != nil {
		hashes["strategy.md"] = HashBytes(src.StrategyRaw)
	}
	if src.RulesRaw != nil {
		hashes["rules.json"] = HashBytes(src.RulesRaw)
	}
	if src.SignalsRaw != nil {
		hashes["signals.json"] = HashBytes(src.SignalsRaw)
	}
	if src.SignalInterestsRaw != nil {
		hashes["signal_interests.json"] = HashBytes(src.SignalInterestsRaw)
	}
	return hashes
}
