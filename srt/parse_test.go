package srt

import (
	"fmt"
	"strings"
	"testing"
)

const srtSource string = `
1
00:00:10,500 --> 00:00:13,000
Elephant's Dream

2
00:00:15,000 --> 00:00:18,000
At the left we can see...
`

func TestParse(t *testing.T) {
	strips := make([]Strip, 0, 8)
	input := strings.NewReader(srtSource)
	res := Parse(input)

	for r := range res {
		if r.Error != nil {
			t.Error(r.Error)
			return
		}
		strips = append(strips, *r.Strip)
	}

	for _, s := range strips {
		fmt.Print("strip: ", s)
	}
}
