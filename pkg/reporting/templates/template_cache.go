package templates

import (
	"fmt"
	"sync"
	"text/template"

	"github.com/openshift/osde2e/pkg/common/templates"
)

type templateCacheKey struct {
	reportName string
	reportType string
}

type templateCache struct {
	cache map[templateCacheKey]*template.Template
	mutex *sync.Mutex
}

var cache = &templateCache{
	cache: map[templateCacheKey]*template.Template{},
	mutex: &sync.Mutex{},
}

func (t *templateCache) getReportTemplate(reportName string, reportType string) (*template.Template, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	key := templateCacheKey{reportName: reportName, reportType: reportType}
	if template, ok := t.cache[key]; ok {
		if template == nil {
			return nil, fmt.Errorf("unable to find report %s:%s", reportName, reportType)
		}

		return template, nil
	}

	template, err := templates.LoadTemplate(fmt.Sprintf("reports/%s/%s.template", reportName, reportType))

	t.cache[key] = template
	if err != nil {
		return nil, fmt.Errorf("error loading template: %v", err)
	}

	return template, nil
}
