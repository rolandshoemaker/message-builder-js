package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"

	"golang.org/x/crypto/cryptobyte"
)

type tree []value

type value struct {
	Tag string

	Type string // "concrete" or "prefix"

	// if type == concrete
	ValueType string // uint8, uint16, string, bytes
	Value     string

	// if type == prefix
	PrefixSize int // 1 or 2 (uint8 or uint16)
	Children   []value
}

func main() {
	js.Global().Set("build", js.FuncOf(build))
	<-make(chan struct{})
}

func build(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return "Invalid no of arguments passed"
	}
	inputJSON := args[0].String()

	var t tree
	if err := json.Unmarshal([]byte(inputJSON), &t); err != nil {
		return err.Error()
	}

	b := cryptobyte.NewBuilder(nil)

	for _, v := range t {
		dfsBuilder(v, b)
	}
	fmt.Printf("%x\n", b.BytesOrPanic())

	ret := map[string]any{
		"hex": hex.EncodeToString(b.BytesOrPanic()),
	}

	return js.ValueOf(ret)
}

func dfsBuilder(v value, b *cryptobyte.Builder) []byte {
	initialOffset := len(b.BytesOrPanic())

	if v.Type == "concrete" {
		switch v.ValueType {
		case "uint8":
			val, err := strconv.ParseUint(v.Value, 0, 8)
			if err != nil {
				panic(err)
			}
			b.AddUint8(uint8(val))
		case "uint16":
			val, err := strconv.ParseUint(v.Value, 0, 16)
			if err != nil {
				panic(err)
			}
			b.AddUint16(uint16(val))
		case "string":
			b.AddBytes([]byte(v.Value))
		case "hex bytes":
			h, err := hex.DecodeString(v.Value)
			if err != nil {
				panic(err)
			}
			b.AddBytes(h)
		}
		postBytes := b.BytesOrPanic()
		postOffset := len(postBytes)
		fmt.Printf("%s %x %d:%d\n", v.Tag, postBytes[initialOffset:postOffset], initialOffset, postOffset)
		return b.BytesOrPanic()
	}
	var prefixFunc func(cryptobyte.BuilderContinuation)
	switch v.PrefixSize {
	case 1:
		prefixFunc = b.AddUint8LengthPrefixed
	case 2:
		prefixFunc = b.AddUint16LengthPrefixed
	}
	prefixFunc(func(b *cryptobyte.Builder) {
		for _, c := range v.Children {
			dfsBuilder(c, b)
		}
	})
	postBytes := b.BytesOrPanic()
	postOffset := len(postBytes)
	fmt.Printf("%s %x %d:%d\n", v.Tag, postBytes[initialOffset:postOffset], initialOffset, postOffset)
	return b.BytesOrPanic()
}

const jsonExample = `[
    {
        "Tag": "a",
        "Type": "concrete",
        "ValueType": "uint8",
        "Value": "32"
    },
    {
        "Tag": "b",
        "Type": "prefix",
        "PrefixSize": 2,
        "Children": [
            {
                "Tag": "c",
                "Type": "concrete",
                "ValueType": "uint16",
                "Value": "65295"
            },
            {
                "Tag": "d",
                "Type": "prefix",
                "PrefixSize": 1,
                "Children": [
                    {
                        "Tag": "e",
                        "Type": "concrete",
                        "ValueType": "uint16",
                        "Value": "48879"
                    }
                ]
            },
            {
                "Tag": "f",
                "Type": "concrete",
                "ValueType": "uint8",
                "Value": "128"
            }
        ]
    },
    {
        "Tag": "g",
        "Type": "concrete",
        "ValueType": "uint16",
        "Value": "254"
    }
]`
