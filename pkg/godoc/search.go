package godoc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/yikakia/cachalot"
	"github.com/yikakia/cachalot/core/cache"
	"github.com/yikakia/cachalot/core/codec"
	"go.uber.org/multierr"
	"golang.org/x/net/html"
)

type SearchResult struct {
	Packages []*SearchPackageInfo
}

type SearchPackageInfo struct {
	Name        string
	Path        string
	Synopsis    string
	GoDocUrl    string
	ImportedBy  int
	SubPackages []string `json:"sub_packages,omitempty"`
}

var searchCache = sync.OnceValue(func() cache.Cache[[]byte] {
	b, err := cachalot.NewBuilder[[]byte]("search", store())
	if err != nil {
		panic(err)
	}

	b.WithCacheMissLoader(searchLoader).WithLogicExpireLoader(searchLoader).WithLogicExpireBytesAdapter(true).WithCompression(codec.GzipCompressionCodec{})

	build, err := b.Build()
	if err != nil {
		panic(err)
	}

	return build
})

func searchLoader(ctx context.Context, q string) ([]byte, error) {
	if !strings.HasPrefix(q, "search") {
		return nil, fmt.Errorf("search query must start with 'search'")
	}

	q = strings.TrimPrefix(q, "search")

	resp, err := client().R().
		SetQueryParams(map[string]string{
			"q": q,
			"m": "package",
		}).
		Get(baseURL() + "/search")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Body(), nil
}

func doSearch(q string) ([]byte, error) {
	get, err := searchCache().Get(context.Background(), "search"+q)
	if err != nil {
		return nil, err
	}
	return get, nil
}

func Search(query string) (*SearchResult, error) {
	cacheGet, err := doSearch(query)
	if err != nil {
		return nil, err
	}

	return extractSearchResult(string(cacheGet))
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
	imptBy, err := extractImportedBy(selection)
	if err != nil {
		return nil, err
	}

	otherPackages, err := extractOtherPackages(selection)
	if err != nil {
		return nil, err
	}

	return &SearchPackageInfo{
		Name:        name,
		Path:        path,
		Synopsis:    synopsis,
		GoDocUrl:    baseURL() + url,
		SubPackages: otherPackages,
		ImportedBy:  imptBy,
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

func extractImportedBy(selection *goquery.Selection) (int, error) {

	im := selection.
		Find("div.SearchSnippet-infoLabel").
		Find("a[aria-label='Go to Imported By']").
		Find("strong").Text()

	atoi, _ := strconv.Atoi(im)
	return atoi, nil
}

func extractPackageGoDocUrl(selection *goquery.Selection) (string, error) {
	var goDocUrl string

	goDocUrl, _ = selection.
		Find("a[data-test-id='snippet-title']").
		Attr("href")

	goDocUrl = strings.TrimSpace(goDocUrl)
	return goDocUrl, nil
}

func extractOtherPackages(s *goquery.Selection) ([]string, error) {
	var otherPackages []string
	var moduleName string
	tmp := s.Find("div.SearchSnippet-sub.go-textSubtle").
		Find("strong").Text()
	tmp = strings.TrimPrefix(tmp, "Other packages in module")
	tmp = strings.Trim(tmp, ":")
	moduleName = strings.TrimSpace(tmp)
	if moduleName == "" {
		return nil, nil
	}

	s.Find("a.go-Chip.go-Chip--subtle").
		Each(func(i int, s *goquery.Selection) {
			p := s.Text()
			p = strings.TrimSpace(p)
			if p == "" {
				return
			}
			otherPackages = append(otherPackages, moduleName+"/"+p)
		})

	return otherPackages, nil
}

func getDoc(query string) (*goquery.Document, error) {
	p, e := html.Parse(strings.NewReader(query))
	if e != nil {
		return nil, errors.WithStack(e)
	}

	return goquery.NewDocumentFromNode(p), nil
}
