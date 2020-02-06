package config

import (
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
	"gopkg.in/alecthomas/kingpin.v2"

	_ "github.com/lib/pq" // Postgres driver
)

// Version Application version
const Version string = "0.2.0"

// NumberWithSign represents arg having sign
type NumberWithSign struct {
	Value    int
	Explicit bool // True if + or - was passed
}

var (
	createIndexCommand = kingpin.Command("create-index", "Create ES indexes")
	exportCommand      = kingpin.Command("export", "Run export")
	ingestCommand      = kingpin.Command("ingest", "Start real time ingestion")
	statsCommand       = kingpin.Command("stats", "Print database ledger statistics")
	esStatsCommand     = kingpin.Command("es-stats", "Print ES ranges stats")
	fillGapsCommand    = kingpin.Command("fill-gaps", "Fill gaps")

	databaseUrl = kingpin.
			Flag("database-url", "Stellar Core database URL").
			Default("postgres://localhost/core?sslmode=disable").
			OverrideDefaultFromEnvar("DATABASE_URL").
			URL()

	EsUrl = kingpin.
		Flag("es-url", "ElasticSearch URL").
		Default("http://localhost:9200").
		OverrideDefaultFromEnvar("ES_URL").
		URL()

	// Concurrency How many tasks and goroutines to produce (all at once for now)
	Concurrency = kingpin.
			Flag("concurrency", "Concurrency for indexing").
			Short('c').
			Default("5").
			OverrideDefaultFromEnvar("CONCURRENCY").
			Int()

	// BatchSize Batch size for bulk export
	BatchSize = exportCommand.
			Flag("batch", "Ledger batch size").
			Short('b').
			Default("50").
			Int()

	// Number of retries
	Retries = exportCommand.
		Flag("retries", "Retries count").
		Default("25").
		Int()

	// Start ledger to start with
	Start = NumberWithSign{0, false}

	start = exportCommand.Arg("start", "Ledger to start indexing, +100 means offset 100 from the first").Default("0").String()

	// Count ledgers
	Count = exportCommand.Arg("count", "Count of ledgers to ingest, should be aliquout batch size").Default("0").Int()

	// StartIngest ledger to start with ingesting
	StartIngest = ingestCommand.Arg("start", "Ledger to start ingesting").Int()

	// Verbose print data
	Verbose = exportCommand.Flag("verbose", "Print indexed data").Bool()

	// ExportDryRun do not index data
	ExportDryRun = exportCommand.Flag("dry-run", "Do not send actual data to Elastic").Bool()

	// ForceRecreateIndexes Allows indexes to be deleted before creation
	ForceRecreateIndexes = createIndexCommand.Flag("force", "Delete indexes before creation").Bool()

	FillGapsDryRun = fillGapsCommand.Flag("dry-run", "Do not ingest anything, just print missing ledgers").Bool()

	// DB Instance of sqlx.DB
	DB *sqlx.DB

	// Command KingPin command
	Command string
)

func initDB() {
	databaseDriver := (*databaseUrl).Scheme

	db, err := sqlx.Connect(databaseDriver, (*databaseUrl).String())
	if err != nil {
		log.Fatal(err)
	}

	DB = db
}

func parseNumberWithSign(value string) (r NumberWithSign, err error) {
	v, err := strconv.Atoi(value)

	if err != nil {
		return r, err
	}

	r.Value = v
	if value[0] == '-' || value[0] == '+' {
		r.Explicit = true
	}

	return r, nil
}

func parseStart() {
	if *start == "" {
		return
	}

	s, err := parseNumberWithSign(*start)
	if err != nil {
		log.Fatal("Error parsing start value", err)
	}
	Start = s
}

func init() {
	kingpin.Version(Version)
	Command = kingpin.Parse()

	initDB()
	parseStart()
}
