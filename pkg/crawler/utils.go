package crawler

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func logError(logger *logrus.Logger, err error, requestID string) {
	errFormat := "[error] in %s[%s:%d]"
	pc, fn, line, _ := runtime.Caller(1)
	logger.WithFields(logrus.Fields{
		"caller":    fmt.Sprintf(errFormat, runtime.FuncForPC(pc).Name(), fn, line),
		"requestID": requestID,
	}).Error(err.Error())
}

// All takes a reader object and returns links from `href` elements from the DOM.
// It does not close the reader.
func All(httpBody io.Reader) []string {
	links := []string{}
	col := []string{}
	page := html.NewTokenizer(httpBody)

	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					tl := trimAnchor(attr.Val)
					col = append(col, tl)
					resolv(&links, col)
				}
			}
		}
	}
}

func trimAnchor(link string) string {
	if strings.Contains(link, "#") {
		var index int
		for n, str := range link {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return link[:index]
	}
	return link
}

func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

func resolv(sl *[]string, ml []string) {
	for _, str := range ml {
		if check(*sl, str) == false {
			*sl = append(*sl, str)
		}
	}
}
