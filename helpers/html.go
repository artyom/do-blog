package helpers

import (
	"code.google.com/p/go.net/html"
)

// Extracts flattened content of first element of given type
func GetFirstElement(n *html.Node, element string) (found bool, t string) {
	if n.Type == html.ElementNode && n.Data == element {
		return true, Flatten(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found, t = GetFirstElement(c, element)
		if found {
			return
		}
	}
	return
}

// Flatten html node, get its text content
func Flatten(n *html.Node) (res string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res += Flatten(c)
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	return
}
