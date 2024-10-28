package linking

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/oolong-sh/oolong/internal/config"
	"github.com/oolong-sh/oolong/internal/linking/lexer"
)

var cfg config.OolongConfig

func init() {
	var err error
	cfg, err = config.Setup("")
	if err != nil {
		panic(err)
	}
}

func TestReadDocumentSimple(t *testing.T) {
	// Basic test
	s := "Hello world!"
	reader := strings.NewReader(s)
	doc, err := readDocument(reader, "")
	if err != nil {
		t.Fatalf("Failed to read document: %v", err)
	}
	fmt.Println("Input:", s, "Output:", doc.tokens)
	if len(doc.tokens) != 2 {
		t.Fatalf("Incorrect Document.Content length. Expected %d, got %d\n", 2, len(doc.tokens))
	}
	expectedTokens := []lexer.Lexeme{
		{
			Value:    "hello",
			Location: [2]int{0, 0},
		}, {
			Value:    "world",
			Location: [2]int{0, 6},
		},
	}
	if !slices.Equal(doc.tokens, expectedTokens) {
		t.Fatalf("Incorrect order in output tokens slice. Expected %+v, got %+v\n", doc.tokens, expectedTokens)
	}

	// basic test with newlines (should contain `breakToken`)
	s = "Hello, \nworld!"
	reader = strings.NewReader(s)
	doc, err = readDocument(reader, "")
	if err != nil {
		t.Fatalf("Failed to read document: %v", err)
	}
	fmt.Println("Input:", s, " Output:", doc.tokens)
	if len(doc.tokens) != 3 {
		t.Fatalf("Incorrect Document.tokens length. Expected %d, got %d", 2, len(doc.tokens))
	}
	expectedTokens = []lexer.Lexeme{
		{
			Value:    "hello",
			Location: 0,
		},
		{
			Value:    lexer.BreakToken,
			Location: 0,
		},
		{
			Value:    "world",
			Location: 1,
		},
	}
	if !slices.Equal(doc.tokens, expectedTokens) {
		t.Fatalf("Incorrect order in output tokens slice. Expected %+v, got %+v\n", doc.tokens, expectedTokens)
	}

	// test with many newlines and multiple single-line lexemes
	s = "\nHello, \nworld! Foo-bar baz   \n\n foo"
	reader = strings.NewReader(s)
	doc, err = readDocument(reader, "")
	if err != nil {
		t.Fatalf("Failed to read document: %v", err)
	}
	fmt.Println("Input:", s, " Output:", doc.tokens)
	if len(doc.tokens) != 9 {
		t.Fatalf("Incorrect Document.Content length. Expected %d, got %d", 5, len(doc.tokens))
	}
}

// TODO: tests with hyphens and other specially handled cases
