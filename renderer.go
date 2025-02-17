package admonitions

import (
	"fmt"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	Writer    html.Writer
	HardWraps bool
	XHTML     bool
	Unsafe    bool
}

// HeadingAttributeFilter defines attribute names which heading elements can have
var AdmonitionAttributeFilter = html.GlobalAttributeFilter

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as (X)HTML.
type Renderer struct {
	Config
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAdmonition, r.renderAdmon)
}

// Define BlockQuoteType enum
type BlockQuoteType int

const (
	Info BlockQuoteType = iota
	Note
	Warn
	Tip
	None
)

func (t BlockQuoteType) String() string {
	return []string{"info", "note", "warning", "tip", "none"}[t]
}

type BlockQuoteLevelMap map[ast.Node]int

func (m BlockQuoteLevelMap) Level(node ast.Node) int {
	return m[node]
}

type BlockQuoteClassifier struct {
	patternMap map[string]*regexp.Regexp
}

func LegacyBlockQuoteClassifier() BlockQuoteClassifier {
	return BlockQuoteClassifier{
		patternMap: map[string]*regexp.Regexp{
			"info": regexp.MustCompile(`(?i)info`),
			"note": regexp.MustCompile(`(?i)note`),
			"warn": regexp.MustCompile(`(?i)warn`),
			"tip":  regexp.MustCompile(`(?i)tip`),
		},
	}
}

func GHAlertsBlockQuoteClassifier() BlockQuoteClassifier {
	return BlockQuoteClassifier{
		patternMap: map[string]*regexp.Regexp{
			"info": regexp.MustCompile(`(?i)^\!(note|important)`),
			"note": regexp.MustCompile(`(?i)^\!warning`),
			"warn": regexp.MustCompile(`(?i)^\!caution`),
			"tip":  regexp.MustCompile(`(?i)^\!tip`),
		},
	}
}

// ClassifyingBlockQuote compares a string against a set of patterns and returns a BlockQuoteType
func (classifier BlockQuoteClassifier) ClassifyingBlockQuote(literal string) BlockQuoteType {

	var t = None
	switch {
	case classifier.patternMap["info"].MatchString(literal):
		t = Info
	case classifier.patternMap["note"].MatchString(literal):
		t = Note
	case classifier.patternMap["warn"].MatchString(literal):
		t = Warn
	case classifier.patternMap["tip"].MatchString(literal):
		t = Tip
	}
	return t
}

// renderBlockQuote will render a BlockQuote
func (r *Renderer) renderAdmon(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Initialize BlockQuote level map
	// if r.LevelMap == nil {
	// 	r.LevelMap = GenerateBlockQuoteLevel(node)
	// }

	// quoteType := ParseBlockQuoteType(node, source)
	// quoteLevel := r.LevelMap.Level(node)
	quoteType := Warn
	quoteLevel := 2

	if quoteLevel == 0 && entering && quoteType != None {
		prefix := fmt.Sprintf("<ac:structured-macro ac:name=\"%s\"><ac:parameter ac:name=\"icon\">true</ac:parameter><ac:rich-text-body>\n", quoteType)
		if _, err := writer.Write([]byte(prefix)); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}
	if quoteLevel == 0 && !entering && quoteType != None {
		suffix := "</ac:rich-text-body></ac:structured-macro>\n"
		if _, err := writer.Write([]byte(suffix)); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}
	return r.renderAdmonition(writer, source, node, entering)
}

func (r *Renderer) renderAdmonition(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*Admonition)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<blockquote")
			html.RenderAttributes(w, n, AdmonitionAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<blockquote>\n")
		}
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}
