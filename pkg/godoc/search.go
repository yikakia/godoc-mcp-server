package godoc

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/net/html"
)

type SearchResult struct {
	Packages []*SearchPackageInfo
}

type SearchPackageInfo struct {
	Name     string
	Path     string
	Synopsis string
	GoDocUrl string
}

func Search(query string) (*SearchResult, error) {
	resp, err := client().R().
		SetQueryParams(map[string]string{
			"q": query,
			"m": "package",
		}).
		Get(baseURL() + "/search")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return extractSearchResult(resp.String())
}

func extractSearchResult(query string) (*SearchResult, error) {
	doc, err := getDoc(query)
	if err != nil {
		return nil, err
	}

	var infos []*SearchPackageInfo

	doc.Find(".SearchSnippet").Each(func(i int, selection *goquery.Selection) {
		info, _err := extractPackageInfo(selection)
		if _err != nil {
			err = multierr.Append(err, _err)
			return
		}
		infos = append(infos, info)
	})
	if err != nil {
		return nil, err
	}

	return &SearchResult{Packages: infos}, nil
}

func extractPackageInfo(selection *goquery.Selection) (*SearchPackageInfo, error) {
	name, err := extractPackageName(selection)
	if err != nil {
		return nil, err
	}
	path, err := extractPackagePath(selection)
	if err != nil {
		return nil, err
	}
	synopsis, err := extractPackageSynopsis(selection)
	if err != nil {
		return nil, err
	}
	url, err := extractPackageGoDocUrl(selection)
	if err != nil {
		return nil, err
	}

	return &SearchPackageInfo{
		Name:     name,
		Path:     path,
		Synopsis: synopsis,
		GoDocUrl: baseURL() + url,
	}, nil
}

func extractPackageName(selection *goquery.Selection) (string, error) {
	var name string
	name = selection.
		Find("a[data-test-id='snippet-title']").
		Contents().Not("span").Text()

	name = strings.TrimSpace(name)
	return name, nil
}

func extractPackagePath(selection *goquery.Selection) (string, error) {
	var path string
	path = selection.
		Find("a[data-test-id='snippet-title']").
		Find(".SearchSnippet-header-path").
		Text()

	path = strings.TrimSpace(path)
	path = strings.Trim(path, "()")
	return path, nil
}

func extractPackageSynopsis(selection *goquery.Selection) (string, error) {
	var synopsis string
	synopsis = selection.Find("p[data-test-id='snippet-synopsis']").Text()

	synopsis = strings.TrimSpace(synopsis)
	return synopsis, nil
}

func extractPackageGoDocUrl(selection *goquery.Selection) (string, error) {
	var goDocUrl string

	goDocUrl, _ = selection.
		Find("a[data-test-id='snippet-title']").
		Attr("href")

	goDocUrl = strings.TrimSpace(goDocUrl)
	return goDocUrl, nil
}

func getDoc(query string) (*goquery.Document, error) {
	p, e := html.Parse(strings.NewReader(query))
	if e != nil {
		return nil, errors.WithStack(e)
	}

	return goquery.NewDocumentFromNode(p), nil
}
