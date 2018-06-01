package main

import (
	"time"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"text/template"
	"io"
)

var etdCSVFilePath = flag.String("in", "etd-output.csv", "Path to etd CSV file.")
var reportFilePath = flag.String("report", "report.csv", "Path to which the report csv file will be written.")
var crossrefOutputFilePath = flag.String("out", "crossref.xml", "Path to which the output XML file will be written.")
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

	// Open the ETD report file.
	etdCSVFile, err := os.Open(*etdCSVFilePath) // For read access.
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

	for {
		record, err := etdCSVReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dissertation := new(Dissertation)

		dissertation.Title = record[0]
		dissertation.Surname = record[1]
		dissertation.GivenName = record[1]
		dissertation.Year = record[2]
		dissertation.DegreeName = record[3]
		dissertation.ProQuestID = record[4]
		dissertation.UUID = record[5]
		dissertation.URI = "https://curve.carleton.ca/" + dissertation.UUID
		dissertation.DOI = dissertation.Year + "-" + dissertation.UUID

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
