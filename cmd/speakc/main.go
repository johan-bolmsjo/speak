package main

import (
	"fmt"
)

var text = `
package paint

enum Color
    1: Red
    2: Green
    3: Blue
end

type XyCoordinate [2]float32

message PaintRequest
    1: id           msg.Id        // Use type 'Id' in package 'msg'.
    2: color        Color
    3: brushSize    float32       // Brush size in millimetres.
    4: xyCoordinate XyCoordinate
end
`

func main() {
	lex := lex("test", text)
	for {
		item := lex.nextItem()
		if item.kind == itemError {
			fmt.Printf("error:%d: %v\n", lex.lineNumber(), item)
		} else {
			fmt.Printf("%v\n", item)
		}
		if item.kind == itemEof || item.kind == itemError {
			break
		}
	}
}
