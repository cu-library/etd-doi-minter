package main

// TemplateData contains the data to use when creating the template
type TemplateData struct {
	HeadData
	BodyData
}

// HeadData contains the data to use in the header of the template
type HeadData struct {
	DOIBatch       int64
	Timestamp      int64
	DepositorName  string
	DepositorEmail string
	Registrant     string
}

// BodyData contains the data to use in the body of the template
type BodyData struct {
	Dissertations []*Dissertation
}

// Dissertation contains the data for each dissertation
type Dissertation struct {
	Title      string
	GivenName  string
	Surname    string
	Year       string
	DegreeName string
	ProQuestID string
	DOI        string
	URI        string
	UUID       string
}

const templateSkeleton string = `<?xml version="1.0" encoding="UTF-8"?>
<doi_batch version="4.4.1" 
           xmlns="http://www.crossref.org/schema/4.4.1"
           xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
           xsi:schemaLocation="http://www.crossref.org/schema/4.4.1 http://www.crossref.org/schemas/crossref4.4.1.xsd">
	<head>
		{{- with .HeadData}}
		<doi_batch_id>{{.DOIBatch}}</doi_batch_id>
		<timestamp>{{.Timestamp}}</timestamp>
		<depositor>
			<depositor_name>{{.DepositorName}}</depositor_name>
			<email_address>{{.DepositorEmail}}</email_address>
		</depositor>
		<registrant>{{.Registrant}}</registrant>
		{{- end}}
	</head>
	<body>
		{{- range .Dissertations }}
		<dissertation>
			<person_name sequence="first" contributor_role="author">
{{- if .GivenName}}{{"\n"}}				<given_name>{{.GivenName}}</given_name>{{end}}
				<surname>{{.Surname}}</surname>
			</person_name>
			<titles>
				<title>{{.Title}}</title>
			</titles>
			<approval_date media_type="electronic">
				<year>{{.Year}}</year>
			</approval_date>
			<institution>
				<institution_name>Carleton University</institution>
				<institution_place>Ottawa, Ontario</institution_place>
			</institution>
			<degree>{{.DegreeName}}</degree>
{{- if .ProQuestID}}
			<publisher_item>
				<identifier id_type="dai">{{.ProQuestID}}</identifier>
			<publisher_item>{{end}}
			<doi_data>
				<doi>{{.DOI}}</doi>
				<resource>{{.URI}}</resource>
			</doi_data>
		</dissertation>
		{{- end}}
	</body>
</doi_batch>
`
