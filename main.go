package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var etdCSVFilePath = flag.String("in", "etd-output.csv", "Path to etd CSV file.")
var prefix = flag.String("prefix", "", "DOI prefix.")
var depositorName = flag.String("depositor", "", "Name of the organization registering the DOIs. The name placed in this element should match the name under which a depositing organization has registered with CrossRef.")
var depositorEmail = flag.String("email", "", "Email address to which batch success and/or error messages are sent. It is recommended that this address be unique to a position within the organization submitting data (e.g. \"doi@...\") rather than unique to a person. In this way, the alias for delivery of this mail can be changed as responsibility for submission of DOI data within the organization changes from one person to another.")
var registrant = flag.String("registrant", "", "The organization that owns the information being registered.")
var starting = flag.Int("starting", 1, "The starting value for the incrementing integer section of the DOI pattern 'prefix/etd/year-intvalue'")

func main() {
	flag.Parse()

	if *depositorName == "" {
		log.Fatalln("depositor required")
	}
	if *depositorEmail == "" {
		log.Fatalln("email required")
	}
	if *registrant == "" {
		log.Fatalln("registrant required")
	}
	if *prefix == "" {
		log.Fatalln("prefix required")
	}

	dois := make(map[string]bool)

	// Open the ETD export from CURVE.
	etdCSVFile, err := os.Open(*etdCSVFilePath)
	if err != nil {
		log.Fatal(err)
	}

	etdCSVReader := csv.NewReader(etdCSVFile)

	dissertations := []*Dissertation{}

	lineNumber := 0

	for {
		lineNumber = lineNumber + 1

		record, err := etdCSVReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// If this record already has a DOI, skip it
		if strings.Contains(record[4], "https://doi.org") {
			continue
		}

		dissertation := new(Dissertation)

		dissertation.Title = strings.TrimSpace(record[0])
		if dissertation.Title == "" {
			log.Fatalf("On line %v: Empty title!", lineNumber)
		}

		mononymous := false
		splitName := strings.Split(record[1], ",")
		if len(splitName) < 2 {
			log.Printf("Found record with only one name: %v\n", record[1])
			if !askForConfirmation("Mononymous name?") {
				log.Fatalln("Exit!")
			} else {
				mononymous = true
			}
		}

		if mononymous {
			dissertation.Surname = strings.TrimSpace(splitName[0])
		} else {
			dissertation.Surname = strings.TrimSpace(splitName[0])
			if dissertation.Surname == "" {
				log.Fatalf("On line %v: Empty Surname!\n", lineNumber)
			}

			restOfName := strings.TrimSpace(splitName[1])
			dissertation.GivenName = strings.TrimSpace(strings.Split(restOfName, " ")[0])
			if dissertation.GivenName == "" {
				log.Fatalf("On line %v: Empty GivenName!\n", lineNumber)
			}
		}

		if record[2] == "" {
			log.Fatalf("On line %v: Empty Year!\n", lineNumber)
		}
		dissertation.Year = record[2][0:4]
		value, err := strconv.Atoi(dissertation.Year)
		if err != nil {
			log.Fatalf("On line %v: Couldn't convert Year to int value!\n", lineNumber)
		}
		if value < 1930 {
			log.Fatalf("On line %v: Likely an invalid year\n", lineNumber)
		}
		if value > 2999 {
			log.Fatalf("On line %v: Likely an invalid year\n", lineNumber)
		}

		dissertation.DegreeName = strings.TrimSpace(record[3])
		if dissertation.DegreeName == "" {
			log.Fatalf("On line %v: Empty DegreeName!\n", lineNumber)
		}

		findProquestIDRegexp := regexp.MustCompile(`pqdiss\: (\w+)\|http`)
		regexpResult := findProquestIDRegexp.FindStringSubmatch(record[4])
		if len(regexpResult) > 1 {
			dissertation.ProQuestID = regexpResult[1]
		}

		dissertation.UUID = strings.TrimSpace(record[5])
		if dissertation.UUID == "" {
			log.Fatalf("On line %v: Empty UUID!", lineNumber)
		}

		dissertation.URI = "https://curve.carleton.ca/" + dissertation.UUID

		dissertation.DOI = fmt.Sprintf("%v/etd/%v-%05v", *prefix, dissertation.Year, lineNumber+*starting-1)

		if _, ok := dois[dissertation.DOI]; ok {
			log.Fatalln("DOI collision!")
		} else {
			dois[dissertation.DOI] = true
		}

		dissertations = append(dissertations, dissertation)
	}

	fullBatches := len(dissertations) / 5000
	remainder := len(dissertations) % 5000

	batches := []*TemplateData{}

	for i := 0; i < fullBatches; i++ {
		templateData := new(TemplateData)
		runAt := time.Now().UTC()
		templateData.HeadData = HeadData{
			DOIBatch:       runAt.Unix(),
			Timestamp:      runAt.UnixNano(),
			DepositorName:  *depositorName,
			DepositorEmail: *depositorEmail,
			Registrant:     *registrant,
		}
		templateData.BodyData.Dissertations = dissertations[i*5000 : ((i + 1) * 5000)]
		batches = append(batches, templateData)
		time.Sleep(1 * time.Second)
	}

	if remainder != 0 {
		templateData := new(TemplateData)
		runAt := time.Now().UTC()
		templateData.HeadData = HeadData{
			DOIBatch:       runAt.Unix(),
			Timestamp:      runAt.UnixNano(),
			DepositorName:  *depositorName,
			DepositorEmail: *depositorEmail,
			Registrant:     *registrant,
		}
		templateData.BodyData.Dissertations = dissertations[fullBatches*5000:]
		batches = append(batches, templateData)
	}

	for i, templateData := range batches {
		output, err := os.Create(fmt.Sprintf("%v-crossref-%v.xml", time.Now().UTC().Format("2006-01-02"), i+1))
		if err != nil {
			log.Fatalln(err)
		}
		defer output.Close()

		t := template.Must(template.New("template").Parse(templateSkeleton))
		err = t.Execute(output, &templateData)
		if err != nil {
			log.Fatalln(err)
		}
	}

	report, err := os.Create(fmt.Sprintf("%v-report.csv", time.Now().UTC().Format("2006-01-02")))
	if err != nil {
		log.Fatalln(err)
	}
	defer report.Close()

	w := csv.NewWriter(report)

	err = w.Write([]string{"URI", "DOI"})
	if err != nil {
		log.Fatalln("Error writing to csv:", err)
	}

	for _, dissertation := range dissertations {
		err = w.Write([]string{dissertation.UUID, dissertation.DOI})
		if err != nil {
			log.Fatalln("Error writing to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatalln(err)
	}

}

// askForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
