package godoc

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type PackageDocument struct {
	Overview  string
	Consts    []ConstBlock
	Variables []VariableBlock
	Functions []FunctionBlock
	Types     []TypeBlock
}

type ConstBlock struct {
	SourceURL  string
	Definition string
	Comment    string
}

type VariableBlock struct {
	SourceURL  string
	Definition string
	Comment    string
}

type FunctionBlock struct {
	SourceURL  string
	Definition string
	Comment    string
	// TODO 支持 Examples 可能需要再拿一个结构体
	//	Examples   []string
}

type TypeBlock struct {
	SourceURL   string
	Definition  string
	Comment     string
	TypeMethods []TypeMethod
}

type TypeMethod struct {
	SourceURL  string
	Definition string
	Comment    string
}

func GetPackageDocument(pkgName string) (*PackageDocument, error) {
	resp, err := client().
		R().
		Get(baseURL() + "/" + pkgName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result, err := extractDocResult(resp.String())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func extractDocResult(html string) (*PackageDocument, error) {
	doc, err := getDoc(html)
	if err != nil {
		return nil, err
	}
	overview, err := extractDocOverview(doc)
	if err != nil {
		return nil, err
	}
	consts, err := extractDocConsts(doc)
	if err != nil {
		return nil, err
	}
	variables, err := extractDocVariables(doc)
	if err != nil {
		return nil, err
	}
	fns, err := extractDocFunctions(doc)
	if err != nil {
		return nil, err
	}
	types, err := extractDocTypes(doc)
	if err != nil {
		return nil, err
	}

	return &PackageDocument{
		Overview:  overview,
		Consts:    consts,
		Variables: variables,
		Functions: fns,
		Types:     types,
	}, nil
}

func extractDocOverview(doc *goquery.Document) (string, error) {
	var overview string

	overview = doc.Find("section.Documentation-overview p").Text()
	return overview, nil
}

func extractDocConsts(doc *goquery.Document) ([]ConstBlock, error) {
	var consts []ConstBlock
	doc.
		Find("section.Documentation-constants").
		Children().
		Each(func(i int, s *goquery.Selection) {
			// 如果现在是 div 标签，则是常量定义
			if s.Is("div") {
				cb := ConstBlock{}
				// 常量定义里有超链接和定义
				url, _ := s.Find("a.Documentation-source").Attr("href")
				cb.SourceURL = url
				// 找到 pre 标签，里面是定义
				var lines []string
				s.Find("pre").Each(func(i int, s *goquery.Selection) {
					lines = append(lines, s.Text())
				})
				cb.Definition = strings.Join(lines, "\n")
				consts = append(consts, cb)
				return
			}
			// 获取p标签，放到最后一个元素里
			// 可能没有注释 此时 class = Documentation-empty
			if s.Is("p") && s.AttrOr("class", "") == "" {
				consts[len(consts)-1].Comment = s.Text()
				return
			}
			// 其他标签，忽略
		})
	return consts, nil
}

func extractDocVariables(doc *goquery.Document) ([]VariableBlock, error) {
	var vars []VariableBlock
	doc.
		Find("section.Documentation-variables").
		Children().
		Each(func(i int, s *goquery.Selection) {
			// 如果现在是 div 标签，则是常量定义
			if s.Is("div") && s.AttrOr("class", "") == "Documentation-declaration" {
				vb := VariableBlock{}
				// 常量定义里有超链接和定义
				url, _ := s.Find("a.Documentation-source").Attr("href")
				vb.SourceURL = url
				// 找到 pre 标签，里面是定义
				var lines []string
				s.Find("pre").Each(func(i int, s *goquery.Selection) {
					lines = append(lines, s.Text())
				})
				vb.Definition = strings.Join(lines, "\n")
				vars = append(vars, vb)
				return
			}
			// 获取p标签，放到最后一个元素里
			// 可能没有注释 此时 class = Documentation-empty
			if s.Is("p") && s.AttrOr("class", "") == "" {
				vars[len(vars)-1].Comment = s.Text()
				return
			}
			// 其他标签，忽略
		})
	return vars, nil
}

func extractDocFunctions(doc *goquery.Document) ([]FunctionBlock, error) {
	var fns []FunctionBlock
	doc.
		Find("section.Documentation-functions").
		// 和前面的不一样，函数都被 div.Documentation-function 包裹了
		Find("div.Documentation-function").
		Each(func(i int, s *goquery.Selection) {
			fnb := FunctionBlock{}
			// Documentation-source 超链接到定义
			fnb.SourceURL = s.Find("a.Documentation-source").AttrOr("href", "")

			//Documentation-declaration 是函数定义
			s.Find("div.Documentation-declaration").
				Each(func(i int, s *goquery.Selection) {
					// 找到 pre 标签，里面是定义
					var lines []string
					s.Find("pre").Each(func(i int, s *goquery.Selection) {
						lines = append(lines, s.Text())
					})
					fnb.Definition = strings.Join(lines, "\n")
				})
			// 找到p标签 是注释
			fnb.Comment = s.Find("p").Text()
			fns = append(fns, fnb)
		})

	return fns, nil
}

func extractDocTypes(doc *goquery.Document) ([]TypeBlock, error) {
	var types []TypeBlock
	var err error
	doc.
		Find("section.Documentation-types").
		// type 被 div.Documentation-type 包裹了
		Find("div.Documentation-type").
		Each(func(i int, s *goquery.Selection) {
			tpb := TypeBlock{}
			// 找到 h4 标签 Documentation-typeHeader
			// 找到 a 标签 Documentation-source
			// 这是类型的链接
			tpb.SourceURL = s.
				Find("h4.Documentation-typeHeader").
				Find("a.Documentation-source").
				AttrOr("href", "")

			// 找到 div.Documentation-declaration 是定义
			s.Find("div.Documentation-declaration").
				Each(func(i int, s *goquery.Selection) {
					// 找到 pre 标签，里面是定义
					var lines []string
					s.Find("pre").Each(func(i int, s *goquery.Selection) {
						lines = append(lines, s.Text())
					})
					tpb.Definition = strings.Join(lines, "\n")
				})
			// 找到p标签 是注释
			tpb.Comment = s.Find("p").Text()
			tpMethods, _err := extractDocTypeMethods(s)
			if _err != nil {
				err = multierr.Append(err, _err)
				return
			}
			tpb.TypeMethods = tpMethods
			types = append(types, tpb)
		})
	if err != nil {
		return nil, err
	}
	return types, nil
}

// 需要传入 extractDocTypes 中的 Documentation-type 节点
func extractDocTypeMethods(s *goquery.Selection) ([]TypeMethod, error) {
	var methods []TypeMethod
	s.
		Find("div.Documentation-typeMethod").
		Each(func(i int, s *goquery.Selection) {
			method := TypeMethod{}
			// url
			method.SourceURL = s.
				Find("h4.Documentation-typeMethodHeader").
				Find("a.Documentation-source").
				AttrOr("href", "")
			// 定义
			s.Find("div.Documentation-declaration").
				Each(func(i int, s *goquery.Selection) {
					// 找到 pre 标签，里面是定义
					var lines []string
					s.Find("pre").Each(func(i int, s *goquery.Selection) {
						lines = append(lines, s.Text())
					})
					method.Definition = strings.Join(lines, "\n")
				})
			// p 标签是注释
			method.Comment = s.Find("p").Text()
			methods = append(methods, method)
		})
	return methods, nil
}
