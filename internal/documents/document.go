package documents

import (
	"fmt"
	"io"
	"os"

	"github.com/oolong-sh/oolong/internal/config"
	"github.com/oolong-sh/oolong/internal/linking/lexer"
	"github.com/oolong-sh/oolong/internal/linking/ngrams"
)

// DOC:
type Document struct {
	path string
	// NOTE: this may need to store more information than just weights
	// - alternatively, the 'Keywords' function could be generated only when needed
	ngwgts map[string]float32

	ngrams map[string]*ngrams.NGram
	tokens []lexer.Lexeme
}

// Document implementation of Note interface
func (d *Document) Path() string                       { return d.path }
func (d *Document) KeywordWeights() map[string]float32 { return d.ngwgts }

// DOC:
func ReadDocument(documentPath string) (*Document, error) {
	f, err := os.Open(documentPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d, err := readDocument(f, documentPath)
	if err != nil {
		return nil, err
	}
	d.path = documentPath

	return d, nil
}

// DOC:
func readDocument(r io.Reader, documentPath string) (*Document, error) {
	initStage := 3

	l := lexer.New()
	fmt.Printf("Running lexer on %s...\n", documentPath)
	l.Lex(r, initStage)

	doc := &Document{
		path:   documentPath,
		tokens: l.Output,
	}

	fmt.Printf("Generating NGrams for %s...\n", documentPath)
	doc.ngrams = ngrams.Generate(doc.tokens, config.NGramRange(), doc.path)
	// TODO: multi-pass ngram calculation?
	// - allow ngram analyzer to handle uppercase chars and symbols

	doc.setWeightsMap()

	return doc, nil
}

// DOC:
func (d *Document) setWeightsMap() {
	wgts := make(map[string]float32)
	for k, v := range d.ngrams {
		wgts[k] = v.Weight()
	}
	d.ngwgts = wgts
}
