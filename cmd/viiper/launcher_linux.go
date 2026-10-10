package main

func commands() ([]command, error) {
	return []command{
		{
			label: "Run VIIPER",
			args:  []string{"server"},
		},
		showHelp,
		configInit,
		proxy,
		{
			label: "Install/Update VIIPER",
			args:  []string{"install"},
		},
		{
			label:   "Uninstall VIIPER",
			args:    []string{"uninstall"},
			confirm: "Really uninstall VIIPER?",
		},
	}, nil
}
