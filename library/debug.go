package library

import "fmt"

func DEBUG(val interface{}) {

	fmt.Println("=====================================================================STAR=========================================================================")
	fmt.Println("=============")
	fmt.Println("=============", val)
	fmt.Println("=============")
	fmt.Println("==================================================================END=================================================================================")

}