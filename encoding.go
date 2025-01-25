package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func buildXHTMLStack(xhtml string) ([]any, error) {
	var (
		tok   xml.Token
		stack []any
		err   error
	)
	xd := xml.NewDecoder(strings.NewReader(xhtml))
	for {
		tok, err = xd.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return stack, nil
			}
			return nil, fmt.Errorf("error decoding xhtml: %w", err)
		}

		switch tt := tok.(type) {
		case xml.StartElement:
			el := tt.Copy()
			el.Name = xml.Name{Local: el.Name.Local}
			stack = append(stack, el)

		case xml.EndElement:
			el := xml.EndElement{
				Name: xml.Name{Local: tt.Name.Local},
			}
			stack = append(stack, el)

		case xml.CharData:
			stack = append(stack, tt.Copy())

		case xml.Comment:
			stack = append(stack, tt.Copy())

		case xml.ProcInst:
			stack = append(stack, tt.Copy())

		case xml.Directive:
			stack = append(stack, tt.Copy())
		}
	}
}

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

func buildArrayXMLStack(jd *json.Decoder, elName string) ([]any, error) {
	var (
		tok   json.Token
		stack []any
		err   error
	)

	for jd.More() {
		tok, err = jd.Token()
		if err != nil {
			return nil, err
		}

		switch tv := tok.(type) {
		case json.Delim:
			if tv == '{' {
				el := &xml.StartElement{
					Name: xml.Name{Local: elName},
				}
				subStack, err := buildObjectXMLStack(jd, el)
				if err != nil {
					return nil, err
				}
				stack = append(stack, el)
				stack = append(stack, subStack...)
				stack = append(stack, el.End())
			} else {
				return nil, fmt.Errorf("unexpected token seen in array %q: %[2]T (%[2]v", elName, tv)
			}

		case json.Number, string, bool, nil:
			strVal, err := encodeValueToString(tv)
			if err != nil {
				return nil, err
			}
			el := xml.StartElement{
				Name: xml.Name{Local: elName},
			}
			stack = append(stack, el, xml.CharData(strVal), el.End())

		default:
			return nil, fmt.Errorf("unexpected token seen in array: %[1]T (%[1]v)", tok)
		}
	}

	if tok, err = jd.Token(); err != nil {
		return nil, fmt.Errorf("error locating ending array token for %q: %w", elName, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return nil, fmt.Errorf("expected closing array token for key %q, saw %[2]T (%[2]v)", elName, tok)
	}

	return stack, nil
}

func buildObjectXMLStack(jd *json.Decoder, el *xml.StartElement) ([]any, error) {
	var (
		tok     json.Token
		lastKey string
		stack   []any
		err     error
	)

	for jd.More() {
		tok, err = jd.Token()
		if err != nil {
			return nil, err
		}

		switch tv := tok.(type) {

		// delimiter tokens
		case json.Delim:
			if tv == '{' {
				child := &xml.StartElement{Name: xml.Name{Local: lastKey}}
				subStack, err := buildObjectXMLStack(jd, child)
				if err != nil {
					return nil, err
				}
				stack = append(stack, child)
				stack = append(stack, subStack...)
				stack = append(stack, child.End())
			} else if tv == '[' {
				subStack, err := buildArrayXMLStack(jd, lastKey)
				if err != nil {
					return nil, err
				}
				stack = append(stack, subStack...)
			} else {
				return nil, fmt.Errorf("unexpected delimiter token seen: %q", tv)
			}
			lastKey = ""

		// if a non-delimiter token is seen, this is _either_ a key or a value, with a key always coming first.
		case string, json.Number, bool, nil:
			strVal, err := encodeValueToString(tv)
			if err != nil {
				return nil, err
			}
			if lastKey == "" {
				lastKey = strVal
				continue
			}

			switch lastKey {

			case "resourceType":
				// skip

			case "div":
				subStack, err := buildXHTMLStack(strVal)
				if err != nil {
					return nil, err
				}
				stack = append(stack, subStack...)

			default:
				el := xml.StartElement{
					Name: xml.Name{Local: lastKey},
					Attr: []xml.Attr{
						{
							Name:  xml.Name{Local: "value"},
							Value: strVal,
						},
					},
				}
				stack = append(stack, el, el.End())
			}
			lastKey = ""

		default:
			return nil, fmt.Errorf("unexpected token seen in object %q: %[2]T (%[2]v)", el.Name.Local, tok)
		}
	}

	if tok, err = jd.Token(); err != nil {
		return nil, fmt.Errorf("error locating ending object token for %q: %w", el.Name.Local, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, fmt.Errorf("expected closing object token for key %q, saw %[2]T (%[2]v)", lastKey, tok)
	}

	return stack, nil
}

func encodeXMLStack(xe *xml.Encoder, stack []any) error {
	var err error

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

	return nil
}
