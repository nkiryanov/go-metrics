package shadow

func fake() {
	os := struct {
		Exit func(int)
	}{Exit: func(int) {}}
	os.Exit(1) // not the real os.Exit — must not be reported
}
