package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Attr map[string]string

// references:
// stream-parsing example: https://eli.thegreenplace.net/2019/faster-xml-stream-processing-in-go/
// having maps in structs: https://stackoverflow.com/a/34972468/605846
type EADNode struct {
	Name     string
	Attr     Attr
	Value    string
	Children []*EADNode
}

type Stack struct {
	S []*EADNode
}

// TODO: could generalize by using "any" for the param and return type
func (s *Stack) Peek() *EADNode {
	idx := len(s.S) - 1
	if idx < 0 {
		return nil
	}
	return s.S[idx]
}

func (s *Stack) Push(val *EADNode) {
	s.S = append(s.S, val)
	return
}

func (s *Stack) Pop() *EADNode {
	idx := len(s.S) - 1
	if idx < 0 {
		return nil
	}
	retval := s.S[idx]
	s.S = s.S[:idx]
	return retval
}

func (s *Stack) Len() int {
	return len(s.S)
}

type EADState struct {
	Stack  Stack
	Tree   *EADNode
	Errors []error
}

func NewEADNode(el xml.StartElement) *EADNode {
	// create new node
	var en *EADNode
	en = new(EADNode)
	en.Name = el.Name.Local
	// TODO: add size to make? len(el.Attr)
	en.Attr = make(Attr)

	for _, attr := range el.Attr {
		en.Attr[attr.Name.Local] = attr.Value
	}

	return en
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := xml.NewDecoder(f)

	eadState := new(EADState)

	indent := 0
	for {
		token, err := d.Token()
		if token == nil || err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error decoding token: %s", err)
		}

		// A Token is an interface holding one of the token types:
		//   StartElement, EndElement, CharData, Comment, ProcInst, or Directive.
		// https://stackoverflow.com/a/33139049/605846
		// https://www.socketloop.com/tutorials/golang-read-xml-elements-data-with-xml-chardata-example
		// https://code-maven.com/slides/golang/parse-html-extract-tags-and-attributes
		switch el := token.(type) {
		case xml.StartElement:
			fmt.Printf("%sStartElement --> %s\n", strings.Repeat(" ", indent), el.Name.Local)

			en := NewEADNode(el)

			indent += 4

			// sort map keys to make display order deterministic
			// https://go.dev/blog/maps
			// https://pkg.go.dev/sort
			var keys []string
			for k := range en.Attr {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// output attribute values
			for _, k := range keys {
				fmt.Printf("%s@%s = %s\n", strings.Repeat(" ", indent), k, en.Attr[k])
			}

			eadState.Stack.Push(en)

		case xml.CharData:
			str := strings.TrimSpace(string([]byte(el)))

			// see if there is actually any data...
			if len(str) != 0 {
				en := eadState.Stack.Peek()
				if en == nil {
					log.Fatalf("In CharData: Stack should not be empty! %s, %s", el, str)
				}
				en.Value = str

				fmt.Printf("%sCharData --> %s\n", strings.Repeat(" ", indent), en.Value)
			}

		case xml.EndElement:
			indent -= 4
			fmt.Printf("%sEndElement --> %s\n", strings.Repeat(" ", indent), el.Name.Local)

			en := eadState.Stack.Pop()
			if en == nil {
				log.Fatalf("In EndElement: Stack should not be empty! %s", el)
			}

			// get the parent node
			// if the parent node is nil, it means we're processing the root element, and
			//   so we can assign the root element to the EADState.Tree member.
			// if the parent node is NOT nil, then append the just-popped EADNode to the
			//   parent node's Children slice
			parent := eadState.Stack.Peek()
			if parent == nil {
				eadState.Tree = en
			} else {
				parent.Children = append(parent.Children, en)
			}
			// fmt.Printf("depth: %d\n", eadState.Stack.Len())
		}

	}

}

// Golang stack implementation
// https://yourbasic.org/golang/implement-stack/
// https://go.dev/play/p/uiYfmQHR1b9
// https://go.dev/play/p/VkWkOFadSYh

// 	package main

// import (
// 	"encoding/json"
// 	"encoding/xml"
// 	"fmt"
// 	"io/ioutil"
// )

// type EAD struct {
// 	Head     string  `xml:"head" json:"head"`
// 	Contents []Mixed `xml:",any" json:"contents"`
// }

// type Mixed struct {
// 	Type  string
// 	Value interface{}
// }

// func main() {
// 	bytes, err := ioutil.ReadFile("example.xml")
// 	if err != nil {
// 		panic(err)
// 	}

// 	var doc EAD
// 	if err := xml.Unmarshal([]byte(bytes), &doc); err != nil {
// 		panic(err)
// 	}

// 	jdoc, err := json.MarshalIndent(doc, "", "  ")
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(string(jdoc))
// }

// func (m *Mixed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
// 	switch start.Name.Local {
// 	case "head", "p", "list":
// 		var e string
// 		if err := d.DecodeElement(&e, &start); err != nil {
// 			return err
// 		}
// 		m.Value = e
// 		m.Type = start.Name.Local
// 	default:
// 		return fmt.Errorf("unknown element: %s", start)
// 	}
// 	return nil
// }
