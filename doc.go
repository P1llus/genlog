// Package genlog provides high-performance fake log generation for testing and benchmarking.
//
// # Overview
//
// genlog generates realistic log data using customizable templates and pre-generated random values.
// It achieves millions of logs per second throughput through:
//   - Pre-compiled templates
//   - Value pools for fast lookups
//   - Producer-consumer pipeline architecture
//   - Parallel generation and batched writes
//
// # Quick Start
//
// The simplest way to use genlog is with a configuration file:
//
//	gen, err := genlog.NewFromFile("config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	gen.Start()
//	gen.Wait()
//	gen.Stop()
//
// # Programmatic Configuration
//
// You can also create configurations programmatically:
//
//	cfg := &genlog.Config{
//		Templates: []genlog.Template{
//			{
//				Template: "{{timestamp}} [{{level}}] {{message}}",
//				Weight:   10,
//			},
//		},
//		CustomTypes: map[string][]string{
//			"level":   {"INFO", "WARN", "ERROR"},
//			"message": {"User logged in", "Request processed"},
//		},
//		Outputs: []genlog.OutputConfig{
//			{
//				Type: "file",
//				Config: map[string]interface{}{
//					"path": "output.log",
//				},
//			},
//		},
//		Limits: genlog.Limits{
//			MaxCount: 1000000, // Generate 1M logs
//		},
//	}
//
//	gen, err := genlog.New(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	gen.Start()
//	gen.Wait()
//	gen.Stop()
//
// # Template Syntax
//
// Templates use {{placeholder}} syntax for dynamic values:
//
//   - Built-in placeholders: {{ipv4}}, {{username}}, {{email}}, {{timestamp}}
//   - Custom types: Define in CustomTypes map and reference by name
//   - Functions: {{FormattedDate "2006-01-02"}}, {{Number 1 100}}
//
// # Performance Tuning
//
// Control generation performance through the Performance configuration:
//
//	cfg.Performance = genlog.Performance{
//		ProducerWorkers: 8,      // CPU-bound generation workers
//		ConsumerWorkers: 2,      // I/O-bound write workers
//		ChannelBuffer:   100000, // Pipeline buffer size
//		WriteBatchSize:  1000,   // Logs per write batch
//		ValuePoolSize:   10000,  // Pre-generated values per type
//	}
//
// # Generation Limits
//
// Control when generation stops:
//
//	cfg.Limits = genlog.Limits{
//		MaxCount:    1000000,    // Stop after 1M logs
//		MaxDuration: time.Hour,  // Or after 1 hour
//		MaxSize:     1 << 30,    // Or after 1GB
//	}
//
// # Statistics
//
// Monitor generation in real-time:
//
//	stats := gen.Stats().Snapshot()
//	fmt.Printf("Generated: %d logs at %.0f logs/sec\n",
//		stats.Generated, stats.LogsPerSecond)
//
// # Output Types
//
// Built-in output types:
//
//   - file: Write to local files
//   - udp: Send via UDP socket
//
// Additional output types can be added by implementing the output.Output interface.
package genlog
