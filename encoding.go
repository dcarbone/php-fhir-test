package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
)

/*
Lots of repeated code in here, could probably be cleaned up, might even do that one day.
*/

func encodeValueToString(v any) (string, error) {
	switch vt := v.(type) {
	case json.Number:
		return vt.String(), nil
	case string:
		return vt, nil
	case bool:
		return strconv.FormatBool(vt), nil
	case nil:
		return "", nil

	default:
		return "", fmt.Errorf("unexpected value type %[1]T (%[1]v)", v)
	}
}

func encodeFromDelimiter(stack *[]any, jd *json.Decoder, lastKey string, delim json.Delim) error {
	// indicates start of object
	if '{' == delim {
		el := &xml.StartElement{Name: xml.Name{Local: lastKey}}
		*stack = append(*stack, el)
		if err := encodeObjectToXML(stack, jd, el); err != nil {
			return fmt.Errorf("error encoding object at key %q: %w", lastKey, err)
		}
		*stack = append(*stack, el.End())
		return nil
	}

	// indicates start of an array
	if '[' == delim {
		if err := encodeArrayToXML(stack, jd, lastKey); err != nil {
			return fmt.Errorf("error encoding array at key %q: %w", lastKey, err)
		}
		return nil
	}

	// any / all other values, yell
	return fmt.Errorf("unexpected delimiter token seen: %q", delim)
}

func encodeArrayToXML(stack *[]any, jd *json.Decoder, lastKey string) error {
	var (
		tok json.Token
		err error
	)

	for jd.More() {
		tok, err = jd.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch tv := tok.(type) {
		case json.Delim:
			el := &xml.StartElement{
				Name: xml.Name{Local: lastKey},
			}
			*stack = append(*stack, el)

			if err = encodeFromDelimiter(stack, jd, lastKey, tv); err != nil {
				return err
			}
			*stack = append(*stack, el.End())

		case json.Number, string, bool, nil:
			strVal, err := encodeValueToString(tv)
			if err != nil {
				return err
			}
			el := xml.StartElement{
				Name: xml.Name{Local: lastKey},
			}
			*stack = append(*stack, el, strVal, el.End())

		default:
			return fmt.Errorf("unexpected token seen in array: %[1]T (%[1]v)", tok)
		}
	}

	if tok, err = jd.Token(); err != nil {
		return fmt.Errorf("error locating ending array token for %q: %w", lastKey, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return fmt.Errorf("expected closing array token for key %q, saw %[2]T (%[2]v)", lastKey, tok)
	}

	return nil
}

func encodeObjectToXML(stack *[]any, jd *json.Decoder, el *xml.StartElement) error {
	var (
		tok     json.Token
		lastKey string
		err     error
	)

	for jd.More() {
		tok, err = jd.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch tv := tok.(type) {

		// delimiter tokens
		case json.Delim:
			if err = encodeFromDelimiter(stack, jd, lastKey, tv); err != nil {
				return err
			}
			lastKey = ""

		// if a non-delimiter token is seen, this is _either_ a key or a value, with a key always coming first.
		case string, json.Number, bool, nil:
			strVal, err := encodeValueToString(tv)
			if err != nil {
				return err
			}
			if lastKey == "" {
				lastKey = strVal
				continue
			}
			// if we see a "resourceType" key, need to change the element's name to its value.
			if lastKey == "resourceType" {
				el.Name = xml.Name{Local: strVal, Space: el.Name.Space}
			} else {
				el.Attr = append(el.Attr, xml.Attr{
					Name:  xml.Name{Local: lastKey},
					Value: strVal,
				})
			}
			lastKey = ""

		default:
			return fmt.Errorf("unexpected token seen in object %q: %[2]T (%[2]v)", el.Name.Local, tok)
		}
	}

	if tok, err = jd.Token(); err != nil {
		return fmt.Errorf("error locating ending object token for %q: %w", lastKey, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expected closing object token for key %q, saw %[2]T (%[2]v)", lastKey, tok)
	}

	return nil
}

func buildXMLEncodingStackFromJSON(data []byte) ([]any, error) {
	var err error

	// initialize json decoder
	jd := json.NewDecoder(bytes.NewReader(data))
	jd.UseNumber()

	// first should always be "{"
	if tok, err := jd.Token(); err != nil {
		return nil, fmt.Errorf("error starting json decode: %w", err)
	} else if s, ok := tok.(json.Delim); !ok || s != '{' {
		return nil, fmt.Errorf(`expected first token to be "{", saw %[1]v (%[1]T)`, tok)
	}

	// initialize root element
	el := &xml.StartElement{
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns"},
				Value: "https://hl7.org/fhir",
			},
		},
	}

	// init xml stack
	stack := []any{el}

	// start encoding at root json object
	err = encodeObjectToXML(&stack, jd, el)
	if err != nil {
		return nil, err
	}

	// add end tag
	stack = append(stack, el.End())

	return stack, nil
}

func encodeXMLStack(w io.Writer, stack []any) error {
	var err error

	// write header
	if _, err = w.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("error encoding XML header: %w", err)
	}

	// init xml encoder
	xe := xml.NewEncoder(w)

	// process token stack, encoding tokens and values as we see them.
	for i, v := range stack {
		switch vt := v.(type) {
		case *xml.StartElement:
			if err = xe.EncodeToken(*vt); err != nil {
				return fmt.Errorf("error encoding token %d (%T) to XML: %w", i, v, err)
			}

		case xml.StartElement, xml.EndElement, xml.CharData, xml.Comment:
			if err = xe.EncodeToken(vt); err != nil {
				return fmt.Errorf("error encoding token %d (%T) to XML: %w", i, v, err)
			}

		default:
			if err = xe.Encode(vt); err != nil {
				return fmt.Errorf("error XML encoding value %[1]v (%[1]T): %w", vt, err)
			}
		}
	}

	if err = xe.Close(); err != nil {
		return fmt.Errorf("error finalizing XML: %w", err)
	}

	return nil
}
