package xpz

import (
	"strings"

	"github.com/antchfx/xmlquery"
)

// GetText returns the text content of the first node matching the XPath
func GetText(node *xmlquery.Node, xpath string) string {
	if node == nil {
		return ""
	}
	found := xmlquery.FindOne(node, xpath)
	if found == nil {
		return ""
	}
	return strings.TrimSpace(found.InnerText())
}

// GetAttr returns the attribute value from the first node matching the XPath
func GetAttr(node *xmlquery.Node, xpath, attr string) string {
	if node == nil {
		return ""
	}
	found := xmlquery.FindOne(node, xpath)
	if found == nil {
		return ""
	}
	for _, a := range found.Attr {
		if a.Name.Local == attr {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}

// GetAttrDirect returns an attribute directly from the provided node
func GetAttrDirect(node *xmlquery.Node, attr string) string {
	if node == nil {
		return ""
	}
	for _, a := range node.Attr {
		if a.Name.Local == attr {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}

// FindAll returns all nodes matching the XPath
func FindAll(node *xmlquery.Node, xpath string) []*xmlquery.Node {
	if node == nil {
		return nil
	}
	return xmlquery.Find(node, xpath)
}
