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
var reportFilePath = flag.String("report", "report.csv", "Path to which the report csv file will be written.")
var crossrefOutputFilePath = flag.String("out", "crossref.xml", "Path to which the output XML file will be written.")
var prefix = flag.String("prefix", "", "DOI prefix.")
var depositorName = flag.String("depositor", "", "Name of the organization registering the DOIs. The name placed in this element should match the name under which a depositing organization has registered with CrossRef.")
var depositorEmail = flag.String("email", "", "Email address to which batch success and/or error messages are sent. It is recommended that this address be unique to a position within the organization submitting data (e.g. \"doi@...\") rather than unique to a person. In this way, the alias for delivery of this mail can be changed as responsibility for submission of DOI data within the organization changes from one person to another.")
var registrant = flag.String("registrant", "", "The organization that owns the information being registered.")

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

	// Open the ETD export from CURVE.
	etdCSVFile, err := os.Open(*etdCSVFilePath)
	if err != nil {
		log.Fatal(err)
	}
	etdCSVReader := csv.NewReader(etdCSVFile)

	templateData := new(TemplateData)

	templateData.HeadData = HeadData{
		DOIBatch:       time.Now().UTC().Unix(),
		Timestamp:      time.Now().UTC().UnixNano(),
		DepositorName:  *depositorName,
		DepositorEmail: *depositorEmail,
		Registrant:     *registrant,
	}

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

		log.Println(lineNumber)

		dissertation := new(Dissertation)

		dissertation.Title = strings.TrimSpace(record[0])
		if dissertation.Title == "" {
			log.Fatalln("Empty title!")
		}

		mononymous := false
		splitName := strings.Split(record[1], ",")
		if len(splitName) < 2 {
			log.Println(record[1])
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
				log.Fatalln("Empty Surname!")
			}

			restOfName := strings.TrimSpace(splitName[1])
			dissertation.GivenName = strings.TrimSpace(strings.Split(restOfName, " ")[0])
			if dissertation.GivenName == "" {
				log.Fatalln("Empty GivenName!")
			}
		}

		if record[2] == "" {
			log.Fatalln("Empty Year!")
		}
		dissertation.Year = record[2][0:4]
		value, err := strconv.Atoi(dissertation.Year)
		if err != nil {
			log.Fatalln("Couldn't convert Year to int value!")
		}
		if value < 1930 {
			log.Fatalln("Likely an invalid year")
		}
		if value > 2099 {
			log.Fatalln("Likely an invalid year")
		}

		dissertation.DegreeName = strings.TrimSpace(record[3])
		if dissertation.DegreeName == "" {
			log.Fatalln("Empty DegreeName!")
		}

		findProquestIDRegexp := regexp.MustCompile(`pqdiss\: (\w+)\|http`)
		regexpResult := findProquestIDRegexp.FindStringSubmatch(record[4])
		if len(regexpResult) > 1 {
			dissertation.ProQuestID = regexpResult[1]
		}

		dissertation.UUID = strings.TrimSpace(record[5])
		if dissertation.UUID == "" {
			log.Fatalln("Empty UUID!")
		}

		dissertation.URI = "https://curve.carleton.ca/" + dissertation.UUID
		splitUUID := strings.Split(dissertation.UUID, "-")
		dissertation.DOI = *prefix + "-" + dissertation.Year + "-" + splitUUID[len(splitUUID)-1]

		templateData.BodyData.Dissertations = append(templateData.BodyData.Dissertations, dissertation)
	}

	output, err := os.Create(*crossrefOutputFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer output.Close()

	report, err := os.Create(*reportFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer report.Close()

	t := template.Must(template.New("template").Parse(templateSkeleton))
	err = t.Execute(output, &templateData)
	if err != nil {
		log.Fatalln(err)
	}

	w := csv.NewWriter(report)

	err = w.Write([]string{"URI", "DOI"})
	if err != nil {
		log.Fatalln("Error writing to csv:", err)
	}

	for _, dissertation := range templateData.BodyData.Dissertations {
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
