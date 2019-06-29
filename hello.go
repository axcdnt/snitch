package main

// SayHello a salute function
func SayHello(name string) string {
	if name == "" {
		return "Hello, world!"
	}

	return "Hello, " + name + "!"
}
