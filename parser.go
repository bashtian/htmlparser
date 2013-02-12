package htmlparser

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.google.com/p/go.net/html"
)

type path struct {
	Name  string
	Index int
}

type Config struct {
	Client *http.Client
	Writer io.Writer
}

var DefaultConfig = &Config{
	Client: &http.Client{},
	Writer: os.Stdout,
}

func Parse(w io.Writer, url string, xpath string) error {
	return DefaultConfig.ParseMulti(w, url, []string{xpath})
}

func ParseMulti(w io.Writer, url string, xpath []string) error {
	return DefaultConfig.ParseMulti(w, url, xpath)
}

func NewParseMulti(url string, xpath []string) error {
	return DefaultConfig.NewParseMulti(url, xpath)
}

// Parse writes the content for the given XPath from the URL to a writer
func (conf *Config) Parse(w io.Writer, url string, xpath string) error {
	return conf.ParseMulti(w, url, []string{xpath})
}

// Parse writes the content for the given XPath from the URL to a writer
func (conf *Config) ParseMulti(w io.Writer, url string, xpath []string) error {
	r, err := conf.Client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	for _, p := range xpath {
		c, err := ParseXpath(doc, p)
		if err != nil {
			return err
		}
		html.Render(w, c)
	}
	return nil
}

// Parse writes the content for the given XPath from the URL to a writer
func (conf *Config) NewParseMulti(url string, xpath []string) error {
	doc, _, err := conf.FetchDocumentNode(url)
	if err != nil {
		return err
	}

	err = writeXpaths(conf.Writer, doc, xpath)
	if err != nil {
		return err
	}

	return nil
}

func (conf *Config) ParseUrl(url string, do func(string) []string) (*html.Node, error) {
	doc, loc, err := conf.FetchDocumentNode(url)
	if err != nil {
		return nil, err
	}

	xpath := do(loc)

	n, err := ParseXpath(doc, xpath[0])
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Parse writes the content for the given XPath from the URL to a writer
func (conf *Config) NewParseMultiFunc(url string, do func(string) []string) error {
	doc, loc, err := conf.FetchDocumentNode(url)
	if err != nil {
		return err
	}

	xpath := do(loc)

	err = writeXpaths(conf.Writer, doc, xpath)
	if err != nil {
		return err
	}

	return nil
}

func (conf *Config) FetchDocumentNode(url string) (*html.Node, string, error) {
	r, err := conf.Client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, "", err
	}

	loc := r.Request.URL.String()
	return doc, loc, nil
}

func writeXpaths(w io.Writer, doc *html.Node, xpath []string) error {
	for _, p := range xpath {
		c, err := ParseXpath(doc, p)
		if err != nil {
			return err
		}
		html.Render(w, c)
	}
	return nil
}

func (c *Config) Render(n *html.Node) {
	html.Render(c.Writer, n)
}

func ParseXpath(n *html.Node, xpath string) (*html.Node, error) {
	p := getPath(xpath)
	c := n
	for i := 0; i < len(p); i++ {
		if p[i].Index == 0 {
			c = getChildById(c, p[i].Name)
		} else {
			c = getChildByName(c, p[i].Name, p[i].Index)
		}
		if c == nil {
			return nil, errors.New("could not parse " + xpath)
		}
	}
	return c, nil
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

func AttrValue(n *html.Node, keyName string) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println(a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
}

func getChildByName(n *html.Node, name string, index int) *html.Node {
	//fmt.Println(n, name, index)
	childIndex := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		//fmt.Println(c, name, index)
		if c.Type == html.ElementNode && c.Data == name {
			//fmt.Printf("[Found] %v [Node] %v\n", c.Data, c.Attr)
			childIndex++
			if childIndex == index {
				return c
			}
		}
	}
	return nil
}

func getPath(xpath string) []path {
	split := strings.Split(xpath, "/")

	var p []path
	reg := regexp.MustCompile(`(.*)\[(\d+)\]`)
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
