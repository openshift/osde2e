package main

import "sort"

type DocsData struct {
	Title    string
	Sections []Section
}

func (d *DocsData) Populate(opts map[string]Options) {
	for sectName, sectOpts := range opts {
		sort.Sort(sectOpts)

		// put options into a section defaulting to other
		sectID := d.SectionID(sectName)
		if sectID < 0 {
			if sectID = d.SectionID("other"); sectID < 0 {
				sectID = len(d.Sections)
				d.Sections = append(d.Sections, Section{
					Name:        "other",
					Description: "Various additional options for configuring osde2e.",
				})
			}
		}

		d.Sections[sectID].Options = append(d.Sections[sectID].Options, sectOpts...)
	}
}

func (d *DocsData) SectionID(name string) int {
	for i, sect := range d.Sections {
		if sect.Name == name {
			return i
		}
	}
	return -1
}

type Section struct {
	Name        string
	Description string
	Options
}

type Options []Option

func (o Options) Len() int {
	return len(o)
}

func (o Options) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Options) Less(i, j int) bool {
	return o[i].Variable < o[j].Variable
}

type Option struct {
	Variable    string
	Description string
	Type        string
	DefaultValue string
}
