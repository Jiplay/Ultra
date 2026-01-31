package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FoodRow represents a parsed food item from CSV
type FoodRow struct {
	Name     string
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
}

// FileStats tracks statistics for a single CSV file
type FileStats struct {
	Filename   string
	Total      int
	Success    int
	Skipped    int
	Duplicates int
	OutputFile string
}

// ProcessingStats tracks overall statistics across all files
type ProcessingStats struct {
	Files       []FileStats
	TotalFiles  int
	TotalFoods  int
	UniqueFoods int
	Duplicates  int
}

// Merge adds a FileStats to overall stats
func (ps *ProcessingStats) Merge(fs FileStats) {
	ps.Files = append(ps.Files, fs)
	ps.TotalFiles++
	ps.TotalFoods += fs.Total
	ps.UniqueFoods += fs.Success
	ps.Duplicates += fs.Duplicates
}

func main() {
	dataDir := "../data"

	fmt.Println("CSV Food Data Cleaner")
	fmt.Println("=====================")
	fmt.Printf("Processing directory: %s\n\n", dataDir)

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Fatalf("Data directory does not exist: %s", dataDir)
	}

	// Find all CSV files
	csvFiles, err := findCSVFiles(dataDir)
	if err != nil {
		log.Fatalf("Error scanning directory: %v", err)
	}

	if len(csvFiles) == 0 {
		log.Println("No CSV files found in data directory")
		return
	}

	// Track duplicates across all files
	seenNames := make(map[string]bool)

	// Process each CSV file
	var overallStats ProcessingStats
	for _, csvFile := range csvFiles {
		stats := processCSVFile(csvFile, seenNames)
		overallStats.Merge(stats)
	}

	// Print summary
	printSummary(overallStats)
}

// findCSVFiles scans a directory for CSV files
func findCSVFiles(dir string) ([]string, error) {
	var csvFiles []string

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check for .csv extension (case-insensitive)
		if strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
			csvFiles = append(csvFiles, filepath.Join(dir, file.Name()))
		}
	}

	sort.Strings(csvFiles) // Process in alphabetical order
	return csvFiles, nil
}

// processCSVFile processes a single CSV file
func processCSVFile(csvPath string, seenNames map[string]bool) FileStats {
	stats := FileStats{
		Filename: filepath.Base(csvPath),
	}

	// Parse CSV
	file, err := os.Open(csvPath)
	if err != nil {
		log.Printf("Error opening file %s: %v", csvPath, err)
		return stats
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV %s: %v", csvPath, err)
		return stats
	}

	var foods []FoodRow
	stats.Total = len(records) - 1 // Exclude header

	// Skip first row (header)
	for i := 1; i < len(records); i++ {
		record := records[i]

		// Skip if not enough fields
		if len(record) < 5 {
			stats.Skipped++
			continue
		}

		// Parse food item
		name := simplifyName(strings.TrimSpace(record[0]))
		calories := cleanNumeric(record[1])
		protein := cleanNumeric(record[2])
		carbs := cleanNumeric(record[3])
		fat := cleanNumeric(record[4])

		// Skip if name is empty
		if name == "" {
			stats.Skipped++
			continue
		}

		// Check for duplicate (case-insensitive)
		nameKey := strings.ToLower(name)
		if seenNames[nameKey] {
			stats.Duplicates++
			continue
		}

		// Mark as seen
		seenNames[nameKey] = true

		foods = append(foods, FoodRow{
			Name:     name,
			Calories: calories,
			Protein:  protein,
			Carbs:    carbs,
			Fat:      fat,
		})
		stats.Success++
	}

	// Generate SQL output
	outputFile := generateOutputFilename(csvPath)
	err = generateSQL(foods, outputFile, filepath.Base(csvPath))
	if err != nil {
		log.Printf("Error generating SQL for %s: %v", csvPath, err)
	}
	stats.OutputFile = outputFile

	return stats
}

// generateOutputFilename creates SQL filename from CSV filename
func generateOutputFilename(csvPath string) string {
	// Extract base filename without extension
	base := filepath.Base(csvPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Generate SQL filename in scripts directory (current directory)
	return filepath.Join(".", name+".sql")
}

// cleanNumeric converts French decimal format and handles missing values
func cleanNumeric(value string) float64 {
	value = strings.TrimSpace(value)

	// Handle missing/trace values
	if value == "" || value == "-" || strings.ToLower(value) == "traces" || strings.HasPrefix(value, "<") {
		return 0
	}

	// Remove quotes
	value = strings.Trim(value, "\"")

	// Replace French decimal separator (comma) with period
	value = strings.Replace(value, ",", ".", -1)

	// Parse to float
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	// Round to nearest integer
	return math.Round(parsed)
}

// simplifyName simplifies long food names
func simplifyName(name string) string {
	// Remove quotes
	name = strings.Trim(name, "\"")

	// Remove "(aliment moyen)" suffix
	name = regexp.MustCompile(`\s*\(aliment moyen\)$`).ReplaceAllString(name, "")

	// Replace "ou X" variations (keep first name)
	name = regexp.MustCompile(`\s+ou\s+[^,]+`).ReplaceAllString(name, "")

	// Remove detailed prep instructions
	removePatterns := []string{
		`, chair et peau`,
		`, chair sans peau`,
		`, sans peau`,
		`, sans noyau`,
		`, sans pépins`,
		`, pelée?`,
		`, épluchée?`,
		`, pelé`,
		`, dénoyauté`,
		`, égoutté`,
		`, égouttée`,
		`, non égouttée?`,
		`, préemballée?`,
		`, à tartiner`,
		`, à cuire`,
		`, à finir de cuire`,
		`, précuit`,
		`, précuite`,
		`, préfrite?`,
		`, tranches`,
		`, entière?`,
		`, tout type`,
		`, type`,
		`, sans précision`,
		`, sans sel ajouté`,
		`, sans sucres ajoutés`,
		`, sans croûte`,
		`, source ou riche en fibres`,
	}

	for _, pattern := range removePatterns {
		name = regexp.MustCompile(pattern).ReplaceAllString(name, "")
	}

	// Simplify cooking methods with slashes
	name = regexp.MustCompile(`(bouilli|rôti|sauté|grillé)/[^,]+`).ReplaceAllString(name, "$1")

	// Simplify "X, Y, Z" to just "X Y" for cleaner names
	// But keep important qualifiers like "rouge", "blanc", "vert", "complet"
	name = cleanCommas(name)

	// Limit length to 80 characters
	if len(name) > 80 {
		name = name[:77] + "..."
	}

	return strings.TrimSpace(name)
}

// cleanCommas simplifies comma-separated qualifiers
func cleanCommas(name string) string {
	// Keep important qualifiers like color/type before comma
	importantQualifiers := regexp.MustCompile(`\b(rouge|blanc|vert|jaune|noir|complet|entier|frais|sec|cru|cuit|bouilli|grillé|rôti|sauté|appertisé|surgelé)\b`)

	parts := strings.Split(name, ",")
	if len(parts) <= 1 {
		return name
	}

	// Keep first part (main name)
	result := parts[0]

	// Add important qualifiers from remaining parts
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if importantQualifiers.MatchString(part) {
			result += " " + part
		}
	}

	return result
}

// generateSQL creates SQL INSERT statements
func generateSQL(foods []FoodRow, filename string, sourceCSV string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "-- Generated from data/%s\n", sourceCSV)
	fmt.Fprintf(file, "-- Total foods: %d\n", len(foods))
	fmt.Fprintf(file, "-- Generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Handle empty food list
	if len(foods) == 0 {
		fmt.Fprintf(file, "-- No unique foods found (all entries were duplicates or invalid)\n")
		fmt.Fprintf(file, "-- No INSERT statements generated\n")
		return nil
	}

	// Start INSERT statement
	fmt.Fprintf(file, "INSERT INTO general_foods (name, description, calories, protein, carbs, fat, fiber, tag, created_at, updated_at)\n")
	fmt.Fprintf(file, "VALUES\n")

	// Write each food item
	for i, food := range foods {
		// Escape single quotes in name
		name := strings.ReplaceAll(food.Name, "'", "''")

		// Write VALUES row
		fmt.Fprintf(file, "  ('%s', '', %.0f, %.0f, %.0f, %.0f, 0, 'general', NOW(), NOW())",
			name,
			food.Calories,
			food.Protein,
			food.Carbs,
			food.Fat,
		)

		// Add comma or semicolon
		if i < len(foods)-1 {
			fmt.Fprintf(file, ",\n")
		} else {
			fmt.Fprintf(file, ";\n")
		}
	}

	return nil
}

// printSummary prints comprehensive statistics
func printSummary(stats ProcessingStats) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Processed %d CSV file(s)\n\n", stats.TotalFiles)

	for i, file := range stats.Files {
		fmt.Printf("%d. %s\n", i+1, file.Filename)
		fmt.Printf("   - Total rows:          %d\n", file.Total)
		fmt.Printf("   - Successfully parsed: %d\n", file.Success)
		fmt.Printf("   - Skipped (invalid):   %d\n", file.Skipped)
		fmt.Printf("   - Duplicates:          %d\n", file.Duplicates)
		fmt.Printf("   - Output:              %s\n\n", file.OutputFile)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Overall Summary:")
	fmt.Printf("  Files processed:  %d\n", stats.TotalFiles)
	fmt.Printf("  Total foods:      %d\n", stats.TotalFoods)
	fmt.Printf("  Unique foods:     %d\n", stats.UniqueFoods)
	fmt.Printf("  Duplicates:       %d\n", stats.Duplicates)
	fmt.Println(strings.Repeat("=", 60))
}
