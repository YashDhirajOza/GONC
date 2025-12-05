package main

import (
	"fmt"
	"log"
	"floatchat-gopy/gonc"
	"time"
)

func main() {
	start := time.Now()
	nc, err := gonc.Open("argo_2019_01/nodc_D1900975_339.nc")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	if nc.Format == gonc.ClassicFormat {
		fmt.Println("Format: Classic (CDF-1)")
	} else {
		fmt.Println("Format: 64-bit Offset (CDF-2)")
	}

	fmt.Println("NumRecs:", nc.NumRecs)

	fmt.Println("Dimensions:")
	for _, d := range nc.Dims {
		fmt.Printf("  %s = %d\n", d.Name, d.Length)
	}

	fmt.Println("Total execution time:", time.Since(start))
}
