package main

// TemplateData contains the data to use when creating the template
type TemplateData struct {
	XMLName           string   `xml:"doi_batch"`
	Version           string   `xml:"version,attr"`
	Xmlns             string   `xml:"xmlns,attr"`
	Xmlnsxsi          string   `xml:"xmlns:xsi,attr"`
	Xsischemalocation string   `xml:"xsi:schemaLocation,attr"`
	Head              HeadData `xml:"head"`
	Body              BodyData `xml:"body"`
}

func NewTemplateData() *TemplateData {
	return &TemplateData{
		Version:           "4.4.1",
		Xmlns:             "http://www.crossref.org/schema/4.4.1",
		Xmlnsxsi:          "http://www.w3.org/2001/XMLSchema-instance",
		Xsischemalocation: "http://www.crossref.org/schema/4.4.1 http://www.crossref.org/schemas/crossref4.4.1.xsd",
	}
}

// HeadData contains the data to use in the header of the template
type HeadData struct {
	DOIBatch       int64  `xml:"doi_batch_id"`
	Timestamp      int64  `xml:"timestamp"`
	DepositorName  string `xml:"depositor>depositor_name"`
	DepositorEmail string `xml:"depositor>email_address"`
	Registrant     string `xml:"registrant"`
}

// BodyData contains the data to use in the body of the template
type BodyData struct {
	Dissertations []*Dissertation `xml:"dissertation"`
}

// Dissertation contains the data for each dissertation
type Dissertation struct {
	Person struct {
		Sequence        string `xml:"sequence,attr"`
		ContributorRole string `xml:"contributor_role,attr"`
		GivenName       string `xml:"given_name"`
		Surname         string `xml:"surname"`
	} `xml:"person_name"`
	Title        string `xml:"titles>title"`
	ApprovalDate struct {
		Year      string `xml:"year"`
		MediaType string `xml:"media_type,attr"`
	} `xml:"approval_date"`
	InstitutionName  string `xml:"institution>institution_name"`
	InstitutionPlace string `xml:"institution>institution_place"`
	DegreeName       string `xml:"degree"`
	Identifier       struct {
		IdType string `xml:"id_type,attr"`
		Value  string `xml:",chardata"`
	} `xml:"publisher_item>identifier,omitempty"`
	DOI  string `xml:"doi_data>doi"`
	URI  string `xml:"doi_data>resource"`
	UUID string `xml:"-"`
}

func NewDissertation() *Dissertation {
	d := &Dissertation{
		InstitutionName:  "Carleton University",
		InstitutionPlace: "Ottawa, Ontario",
	}
	d.Person.Sequence = "first"
	d.Person.ContributorRole = "author"
	d.ApprovalDate.MediaType = "online"
	return d
}
