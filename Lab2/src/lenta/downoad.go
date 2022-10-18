package main

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
)

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

type Item struct {
	Ref, Time, Title string
}

func readItem(item *html.Node) *Item {
	if a := item.FirstChild; isElem(a, "a") {
		if cs := getChildren(a); len(cs) == 2 && isElem(cs[0], "time") && isText(cs[1]) {
			return &Item{
				Ref:   getAttr(a, "href"),
				Time:  getAttr(cs[0], "title"),
				Title: cs[1].Data,
			}
		}
	}
	return nil
}

func search(node *html.Node) []*Item {
	if isDiv(node, "list-group") {
		var items []*Item
		for a := getChildren(node)[0]; a != nil; a = a.NextSibling {
			if getAttr(a, "class") == "list-group-item" {
				for b := getChildren(a)[0]; b != nil; b = b.NextSibling {
					if getAttr(b, "class") == "list-group-item-text" {
						for c := getChildren(b)[0]; c != nil; c = c.NextSibling {
							if isText(c) {
								items = append(items, &Item{
									Ref:   "",
									Title: c.Data,
								})
							}
						}
					}
				}
			}
		}
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c); items != nil {
			return items
		}
	}
	return nil
}

func downloadNews() []*Item {
	log.Info("sending request to elpol.ru/news.asp")
	if response, err := http.Get("http://elpol.ru/news.asp"); err != nil {
		log.Error("request to elpol.ru/news.asp failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from http://elpol.ru/news.asp", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from http://elpol.ru/news.asp", "error", err)
			} else {
				log.Info("HTML from http://elpol.ru/news.asp parsed successfully")
				return search(doc)
			}
		}
	}
	return nil
}
