package htmlparser

import (
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Xpath []path

type path struct {
	Name  string
	Index int
}

func NewXpath(xpath string) Xpath {
	split := strings.Split(xpath, "/")

	var p Xpath
	reg := regexp.MustCompile(`(.*)\[(-?\d+)\]`)
	regId := regexp.MustCompile(`\*\[@id="(.*)"]`)

	for _, s := range split {
		if s == "" {
			continue
		}
		e := path{}
		if r := reg.FindStringSubmatch(s); r != nil {
			i, err := strconv.Atoi(r[2])
			if err != nil {
				fmt.Printf("%v", err)
				return nil
			}
			e.Name = r[1]
			e.Index = i
		} else if r := regId.FindStringSubmatch(s); r != nil {
			e.Name = r[1]
			e.Index = 0
		} else {
			e.Name = s
			e.Index = 1
		}
		p = append(p, e)
	}
	return p
}

func (p Xpath) Print(w io.Writer, n *html.Node) error {
	node, err := p.Parse(n)
	if err != nil {
		return err
	}
	html.Render(w, node)
	return nil
}

// Parse the XPath value
func (p Xpath) Parse(n *html.Node) (*html.Node, error) {
	c := n
	for _, path := range p {
		if path.Index == 0 {
			c = getChildById(c, path.Name)
			fmt.Println("got by id", path.Name)
		} else {
			c = getChildByName(c, path.Name, path.Index)
			fmt.Println("got by name", path.Name)
		}
		if c == nil {
			return nil, errors.New("could not parse " + path.Name)
		}
	}
	//	for i := 0; i < len(p); i++ {
	//		if p[i].Index == 0 {
	//			c = getChildById(c, p[i].Name)
	//			fmt.Println("got by id", p[i].Name)
	//		} else {
	//			c = getChildByName(c, p[i].Name, p[i].Index)
	//			fmt.Println("got by name", p[i].Name)
	//		}
	//		if c == nil {
	//			return nil, errors.New("could not parse " + p[i].Name)
	//		}
	//	}
	return c, nil
}

// Parse the XPath value
func (p Xpath) ParseMulti(n *html.Node) ([]*html.Node, error) {
	nodes := []*html.Node{n}
	for _, path := range p {
		tmp := []*html.Node{}
		for _, node := range nodes {
			switch path.Index {
			case -1:
				tmp = append(tmp, AllByName(node, path.Name)...)
			case 0:
				if n := getChildById(node, path.Name); n != nil {
					tmp = append(tmp, n)
				} else {
					return nil, errors.New("could not get child by id: " + path.Name)
				}
			default:
				if n := getChildByName(node, path.Name, path.Index); n != nil {
					tmp = append(tmp, n)
				} else {
					return nil, errors.New("could not get child by name: " + path.Name)
				}
			}

			if len(tmp) == 0 {
				return nil, errors.New("could not parse " + path.Name)
			}
		}
		nodes = tmp
	}
	//	for i := 0; i < len(p); i++ {
	//		tmp := []*html.Node{}
	//		for _, node := range nodes {
	//			switch p[i].Index {
	//			case -1:
	//				tmp = append(tmp, AllByName(node, p[i].Name)...)
	//			case 0:
	//				if n := getChildById(node, p[i].Name); n != nil {
	//					tmp = append(tmp, n)
	//				}
	//			default:
	//				tmp = append(tmp, getChildByName(node, p[i].Name, p[i].Index))
	//			}
	//			if len(tmp) == 0 {
	//				return nil, errors.New("could not parse " + p[i].Name)
	//			}
	//		}
	//		nodes = tmp
	//	}
	return nodes, nil
}
