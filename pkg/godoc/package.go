package godoc

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type PackageDocument struct {
	Overview    string
	Consts      []ConstBlock
	Variables   []VariableBlock
	Functions   []FunctionBlock
	Types       []TypeBlock
	SubPackages []*SubPackage
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

type SubPackage struct {
	Name    string
	Comment string
}

type GetPackageRequest struct {
	PackageName string
	NeedURL     bool
}

func GetPackageDocument(req GetPackageRequest) (*PackageDocument, error) {
	body, err := getWithFn(req.PackageName, func() ([]byte, error) {
		resp, err := client().
			R().
			Get(baseURL() + "/" + req.PackageName)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return resp.Body(), nil
	})
	if err != nil {
		return nil, err
	}

	result, err := extractDocResult(string(body), req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func extractDocResult(html string, req GetPackageRequest) (*PackageDocument, error) {
	doc, err := getDoc(html)
	if err != nil {
		return nil, err
	}
	overview, err := extractDocOverview(doc, req)
	if err != nil {
		return nil, err
	}
	consts, err := extractDocConsts(doc, req)
	if err != nil {
		return nil, err
	}
	variables, err := extractDocVariables(doc, req)
	if err != nil {
		return nil, err
	}
	fns, err := extractDocFunctions(doc, req)
	if err != nil {
		return nil, err
	}
	types, err := extractDocTypes(doc, req)
	if err != nil {
		return nil, err
	}
	subPackages, err := extractSubPackages(doc, req)
	if err != nil {
		return nil, err
	}

	return &PackageDocument{
		Overview:    overview,
		Consts:      consts,
		Variables:   variables,
		Functions:   fns,
		Types:       types,
		SubPackages: subPackages,
	}, nil
}

func extractDocOverview(doc *goquery.Document, req GetPackageRequest) (string, error) {
	var overview string

	overview = doc.Find("section.Documentation-overview p").Text()
	return overview, nil
}

func extractDocConsts(doc *goquery.Document, req GetPackageRequest) ([]ConstBlock, error) {
	var consts []ConstBlock
	doc.
		Find("section.Documentation-constants").
		Children().
		Each(func(i int, s *goquery.Selection) {
			// 如果现在是 div 标签，则是常量定义
			if s.Is("div") {
				cb := ConstBlock{}
				if req.NeedURL {
					// 常量定义里有超链接和定义
					url, _ := s.Find("a.Documentation-source").Attr("href")
					cb.SourceURL = url
				}

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

func extractDocVariables(doc *goquery.Document, req GetPackageRequest) ([]VariableBlock, error) {
	var vars []VariableBlock
	doc.
		Find("section.Documentation-variables").
		Children().
		Each(func(i int, s *goquery.Selection) {
			if s.Is("div") && s.AttrOr("class", "") == "Documentation-declaration" {
				vb := VariableBlock{}
				if req.NeedURL {
					// 定义里有超链接和定义
					url, _ := s.Find("a.Documentation-source").Attr("href")
					vb.SourceURL = url
				}

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

func extractDocFunctions(doc *goquery.Document, req GetPackageRequest) ([]FunctionBlock, error) {
	var fns []FunctionBlock
	doc.
		Find("section.Documentation-functions").
		// 和前面的不一样，函数都被 div.Documentation-function 包裹了
		Find("div.Documentation-function").
		Each(func(i int, s *goquery.Selection) {
			fnb := FunctionBlock{}
			// Documentation-source 超链接到定义
			if req.NeedURL {
				fnb.SourceURL = s.Find("a.Documentation-source").AttrOr("href", "")
			}

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

func extractDocTypes(doc *goquery.Document, req GetPackageRequest) ([]TypeBlock, error) {
	var types []TypeBlock
	var err error
	doc.
		Find("section.Documentation-types").
		// type 被 div.Documentation-type 包裹了
		Find("div.Documentation-type").
		Each(func(i int, s *goquery.Selection) {
			tpb := TypeBlock{}
			if req.NeedURL {
				// 找到 h4 标签 Documentation-typeHeader
				// 找到 a 标签 Documentation-source
				// 这是类型的链接
				tpb.SourceURL = s.
					Find("h4.Documentation-typeHeader").
					Find("a.Documentation-source").
					AttrOr("href", "")
			}

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
			tpMethods, _err := extractDocTypeMethods(s, req)
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
func extractDocTypeMethods(s *goquery.Selection, req GetPackageRequest) ([]TypeMethod, error) {
	var methods []TypeMethod
	s.
		Find("div.Documentation-typeMethod").
		Each(func(i int, s *goquery.Selection) {
			method := TypeMethod{}
			if req.NeedURL {
				// url
				method.SourceURL = s.
					Find("h4.Documentation-typeMethodHeader").
					Find("a.Documentation-source").
					AttrOr("href", "")
			}

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

func extractSubPackages(doc *goquery.Document, req GetPackageRequest) ([]*SubPackage, error) {
	var subPackages []*SubPackage
	var err error

	doc.Find("table[data-test-id='UnitDirectories-table']").
		Find("tbody").
		Children().
		Each(func(i int, s *goquery.Selection) {

			v, hasSubPackage := s.Attr("data-aria-controls")
			fmt.Println(v, hasSubPackage, goquery.NodeName(s))

			if hasSubPackage {
				// 目录要处理自己和子包
				dir, _err := extractSubPackageAsDir(s, req)
				if _err != nil {
					err = multierr.Append(err, _err)
					return
				}
				if dir != nil {
					subPackages = append(subPackages, dir)
				}

			} else {
				subPackage, _err := extractSubPackage(s)
				if _err != nil {
					err = multierr.Append(err, _err)
					return
				}
				if subPackage != nil {
					subPackages = append(subPackages, subPackage)
				}
			}
		})
	if err != nil {
		return nil, err
	}

	return subPackages, nil
}

func extractSubPackage(s *goquery.Selection) (*SubPackage, error) {
	// Name
	// 有 id 优先用id
	// 没有 id 用 a 标签的文本
	name := s.AttrOr("data-id", "")
	if name == "" {
		name = s.Find("a").Text()
	} else {
		name = strings.TrimSpace(name)
		name = strings.ReplaceAll(name, "-", "/")
	}
	if name == "" {
		return nil, nil
	}
	// comment
	comment := s.Find("td.UnitDirectories-desktopSynopsis").Text()
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, nil
	}

	return &SubPackage{
		Name:    name,
		Comment: comment,
	}, nil
}

func extractSubPackageAsDir(s *goquery.Selection, req GetPackageRequest) (*SubPackage, error) {

	// pathCell name
	name := s.Find("div.UnitDirectories-pathCell").Find("span").Text()
	name = strings.TrimSpace(name)
	if name == "" {
		name = s.Find("a").Text()
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, nil
		}
	}
	// comment
	comment := s.Find("td.UnitDirectories-desktopSynopsis").Text()
	comment = strings.TrimSpace(comment)

	return &SubPackage{
		Name:    name,
		Comment: comment,
	}, nil
}
