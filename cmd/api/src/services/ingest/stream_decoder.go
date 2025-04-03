package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/specterops/bloodhound/src/model/ingest"
)

type StreamDecoder struct {
	dec *json.Decoder
}

func NewStreamDecoder(r io.Reader) *StreamDecoder {
	return &StreamDecoder{dec: json.NewDecoder(r)}
}

// EatOpeningBracket consumes the opening bracket '['
func (s *StreamDecoder) EatOpeningBracket() error {
	tok, err := s.dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return errors.New("expected opening bracket '['")
	}
	return nil
}

// EatClosingBracket consumes the closing bracket ']'
func (s *StreamDecoder) EatClosingBracket() error {
	tok, err := s.dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return errors.New("expected closing bracket ']'")
	}
	return nil
}

// EatOpeningCurlyBracket consumes the opening curly bracket '{'
func (s *StreamDecoder) EatOpeningCurlyBracket() error {
	tok, err := s.dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return errors.New("expected opening curly bracket '{'")
	}
	return nil
}

// EatClosingCurlyBracket consumes the closing curly bracket '}'
func (s *StreamDecoder) EatClosingCurlyBracket() error {
	tok, err := s.dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return errors.New("expected closing curly bracket '}'")
	}
	return nil
}

// DecodeNext decodes the next JSON value into the provided type
func (s *StreamDecoder) DecodeNext(v any) error {
	err := s.dec.Decode(v)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Println("End of JSON stream")
			return nil
		} else if syntaxErr, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("syntax error at byte %d: %v", syntaxErr.Offset, err)
		} else if _, ok := err.(*json.UnmarshalTypeError); ok {
			switch v.(type) {
			case *Node:
				return ingest.ErrNodeSchema
			case *Edge:
				return ingest.ErrEdgeSchema
			}
		} else {
			return fmt.Errorf("unexpected error: %v", err)
		}
	}

	return nil
}

// More checks if there are more tokens in the current array scope
func (s *StreamDecoder) More() bool {
	return s.dec.More()
}
