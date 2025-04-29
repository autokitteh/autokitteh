package pythonrt

import "fmt"

func ExamplePrinter() {
	printFn := func(line string) error {
		fmt.Println(line)
		return nil
	}

	printer := Printer{
		printFn: printFn,
	}

	data := []byte("a\nb\nc\nd")
	printer.Write(data)
	printer.Flush()

	// Output:
	// a
	// b
	// c
	// d
}
