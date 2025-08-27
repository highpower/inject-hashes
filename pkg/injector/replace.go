package injector

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func replace(file string, rules []Rule) error {
	temp, err := replaceIn(file, rules)
	if err != nil {
		return err
	}
	if err := os.Rename(file, fmt.Sprintf("%s.bak", file)); err != nil {
		return err
	}
	if err := os.Rename(temp, file); err != nil {
		return err
	}
	return nil
}

func replaceIn(file string, rules []Rule) (string, error) {
	source, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = source.Close() }()
	base, ext := filepath.Base(file), filepath.Ext(file)
	pattern := fmt.Sprintf("%s-*%s", base[:len(base)-len(ext)], ext)
	target, err := os.CreateTemp(filepath.Dir(file), pattern)
	if err != nil {
		return "", err
	}
	defer func() { _ = target.Close() }()
	if err := copyReplace(source, target, rules); err != nil {
		return "", err
	}
	return target.Name(), nil
}

func copyReplace(reader io.Reader, writer io.Writer, rules []Rule) error {
	tokenizer := html.NewTokenizer(reader)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			tryApplyRules(&token, rules)
			if _, err := fmt.Fprintf(writer, "%s", token.String()); err != nil {
				return err
			}
		case html.TextToken:
			token := tokenizer.Token()
			if _, err := fmt.Fprintf(writer, "%s", token.Data); err != nil {
				return err
			}
		default:
			token := tokenizer.Token()
			if _, err := fmt.Fprintf(writer, "%s", token.String()); err != nil {
				return err
			}
		}
	}
}

func tryApplyRules(token *html.Token, rules []Rule) {
	switch {
	case token.DataAtom == atom.Script:
		log.Printf("applying rules to <script>")
		replaceSrc(token, rules)
	case token.DataAtom == atom.Link && isStyleSheet(token):
		log.Printf("applying rules to <link>")
		replaceHref(token, rules)
	}
}

func replaceSrc(token *html.Token, rules []Rule) {
	for i := 0; i < len(token.Attr); i++ {
		if strings.EqualFold(token.Attr[i].Key, "src") {
			tryReplaceValue(&token.Attr[i], rules)
			return
		}
	}
}

func replaceHref(token *html.Token, rules []Rule) {
	for i := 0; i < len(token.Attr); i++ {
		if strings.EqualFold(token.Attr[i].Key, "href") {
			tryReplaceValue(&token.Attr[i], rules)
			return
		}
	}
}

func tryReplaceValue(attr *html.Attribute, rules []Rule) {
	for _, rule := range rules {
		matched := attr.Val == rule.Ref
		log.Printf("trying to compare '%s' to '%s': %t", attr.Val, rule.Ref, matched)
		if matched {
			log.Printf("replacing '%s' to '%s'", attr.Val, rule.NewRef)
			attr.Val = rule.NewRef
			return
		}
	}
}

func isStyleSheet(token *html.Token) bool {
	for _, attr := range token.Attr {
		if strings.EqualFold(attr.Key, "rel") && strings.EqualFold(attr.Val, "stylesheet") {
			return true
		}
	}
	return false
}
