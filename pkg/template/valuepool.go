// Package template provides template compilation and pre-generated random value pools.
//
// This package is responsible for compiling log templates into efficient execution plans
// and maintaining pools of pre-generated random values for fast log generation.
package template

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// ValuePool holds pre-generated random values for extremely fast lookup during log generation.
//
// Each producer worker gets its own ValuePool instance to avoid any lock contention.
// Values are pre-generated at initialization time using gofakeit, and subsequent
// log generation simply indexes into these pre-populated arrays.
type ValuePool struct {
	// Pre-generated values for common types
	ipv4Addresses []string
	ipv6Addresses []string
	usernames     []string
	emails        []string
	domains       []string
	urls          []string
	userAgents    []string
	httpMethods   []string
	httpStatuses  []string
	logLevels     []string
	timestamps    []string

	// Custom types from configuration
	customTypes map[string][]string

	// Per-goroutine random source (no locking needed)
	rng *rand.Rand
}

// NewValuePool creates a new value pool with pre-generated random values.
func NewValuePool(seed int64, poolSize int, customTypes map[string][]string) *ValuePool {
	// Create per-goroutine random source
	rng := rand.New(rand.NewSource(seed))

	// Create a gofakeit instance with the seed
	faker := gofakeit.New(uint64(seed))

	vp := &ValuePool{
		ipv4Addresses: make([]string, poolSize),
		ipv6Addresses: make([]string, poolSize),
		usernames:     make([]string, poolSize),
		emails:        make([]string, poolSize),
		domains:       make([]string, poolSize),
		urls:          make([]string, poolSize),
		userAgents:    make([]string, poolSize),
		httpMethods:   make([]string, poolSize),
		httpStatuses:  make([]string, poolSize),
		logLevels:     make([]string, poolSize),
		timestamps:    make([]string, poolSize),
		customTypes:   customTypes,
		rng:           rng,
	}

	// Pre-generate all common values
	for i := 0; i < poolSize; i++ {
		vp.ipv4Addresses[i] = faker.IPv4Address()
		vp.ipv6Addresses[i] = faker.IPv6Address()
		vp.usernames[i] = faker.Username()
		vp.emails[i] = faker.Email()
		vp.domains[i] = faker.DomainName()
		vp.urls[i] = faker.URL()
		vp.userAgents[i] = faker.UserAgent()
		vp.httpMethods[i] = faker.HTTPMethod()
		vp.httpStatuses[i] = strconv.Itoa(faker.HTTPStatusCode())
		vp.logLevels[i] = faker.LogLevel("general")

		// Generate timestamps with realistic variety
		minDate := time.Now().Add(-365 * 24 * time.Hour) // 1 year ago
		maxDate := time.Now()
		vp.timestamps[i] = faker.DateRange(minDate, maxDate).Format(time.RFC3339)
	}

	return vp
}

// Get returns a random value for the given placeholder name.
func (vp *ValuePool) Get(name string) string {
	switch name {
	// Network
	case "ipv4", "IPv4", "IPv4Address":
		return vp.ipv4Addresses[vp.rng.Intn(len(vp.ipv4Addresses))]
	case "ipv6", "IPv6", "IPv6Address":
		return vp.ipv6Addresses[vp.rng.Intn(len(vp.ipv6Addresses))]
	case "domain", "Domain", "DomainName":
		return vp.domains[vp.rng.Intn(len(vp.domains))]
	case "url", "URL":
		return vp.urls[vp.rng.Intn(len(vp.urls))]

	// User info
	case "username", "Username":
		return vp.usernames[vp.rng.Intn(len(vp.usernames))]
	case "email", "Email":
		return vp.emails[vp.rng.Intn(len(vp.emails))]

	// HTTP
	case "useragent", "UserAgent":
		return vp.userAgents[vp.rng.Intn(len(vp.userAgents))]
	case "httpmethod", "HTTPMethod":
		return vp.httpMethods[vp.rng.Intn(len(vp.httpMethods))]
	case "httpstatus", "HTTPStatus", "HTTPStatusCode":
		return vp.httpStatuses[vp.rng.Intn(len(vp.httpStatuses))]

	// Logging
	case "loglevel", "LogLevel", "level":
		return vp.logLevels[vp.rng.Intn(len(vp.logLevels))]

	// Time
	case "timestamp", "Timestamp":
		return vp.timestamps[vp.rng.Intn(len(vp.timestamps))]

	// Numbers
	case "number", "Number":
		return strconv.Itoa(vp.rng.Intn(100000))
	case "port", "Port":
		return strconv.Itoa(vp.rng.Intn(65535) + 1)

	// Check custom types
	default:
		if values, ok := vp.customTypes[name]; ok && len(values) > 0 {
			return values[vp.rng.Intn(len(values))]
		}

		// Return a placeholder for unknown types
		return fmt.Sprintf("{{%s}}", name)
	}
}

// GetFormattedDate returns a formatted timestamp.
func (vp *ValuePool) GetFormattedDate(format string) string {
	// Pick a random pre-generated timestamp and reformat it
	ts := vp.timestamps[vp.rng.Intn(len(vp.timestamps))]
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Fallback to current time if parsing fails
		t = time.Now()
	}
	return t.Format(format)
}

// GetNumberRange returns a random number in the specified range.
func (vp *ValuePool) GetNumberRange(min, max int) string {
	if min >= max {
		return strconv.Itoa(min)
	}
	return strconv.Itoa(vp.rng.Intn(max-min+1) + min)
}

// RNG returns the underlying random number generator for this value pool.
func (vp *ValuePool) RNG() *rand.Rand {
	return vp.rng
}
