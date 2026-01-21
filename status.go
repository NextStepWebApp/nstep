package main

import "fmt"

func statusNextStep(plj *packageLocalJson) {
	fmt.Printf(":: Current version: nextstep/%s\n", plj.getVersion())
}
