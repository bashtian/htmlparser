package htmlparser

import (
	"exp/html"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Path struct {
	Name  string
	Index int
}

// Parse writes the content for the given XPath from the URL to a writer
func Parse(w io.Writer, url, xpath string) {
	r, err := http.Get(url)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		log.Fatal(err)
		return
	}

	c := parseXpath(doc, xpath)

	html.Render(w, c)
}

func parseXpath(n *html.Node, xpath string) *html.Node {
	p := getPath(xpath)
	c := n
	for i := 0; i < len(p); i++ {
		if p[i].Index == 0 {
			c = getChildById(c, p[i].Name)
		} else {
			c = getChildByName(c, p[i].Name, p[i].Index)
		}
		if c == nil {
			return nil
		}
	}
	return c
}

func getChildById(n *html.Node, idName string) *html.Node {
	var f func(*html.Node) *html.Node
	f = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == idName {
					return n
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if node := f(c); node != nil {
				return node
			}
		}
		return nil
	}
	return f(n)
}

func getChildByName(n *html.Node, name string, index int) *html.Node {
	//fmt.Println(n, name, index)
	childIndex := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		//fmt.Println(c, name, index)
		if c.Type == html.ElementNode && c.Data == name {
			//fmt.Printf("[Found] %v\n", c.Data)
			childIndex++
			if childIndex == index {
				return c
			}
		}
	}
	return nil
}

func getPath(xpath string) []Path {
	split := strings.Split(xpath, "/")

	var p []Path
	reg := regexp.MustCompile(`(.*)\[(\d+)\]`)
	regId := regexp.MustCompile(`\*\[@id="(.*)"]`)

	for _, s := range split {
		if s == "" {
			continue
		}
		e := Path{}
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
