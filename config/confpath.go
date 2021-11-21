package config

import (
	_ "embed"
	"log"
	"os"
	"path"
)

//go:embed example.yaml
var exampleConfig []byte

func configPaths() []string {
	var configChoices []string
	// from most desirable to least desirable config path
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configChoices = append(configChoices, path.Join(xdgConfigHome, "paprika.yml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		configChoices = append(configChoices, path.Join(home, ".config", "paprika.yml"))
	}
	systemdConfigDir := os.Getenv("CONFIGURATION_DIRECTORY")
	if systemdConfigDir != "" {
		configChoices = append(configChoices, path.Join(systemdConfigDir, "paprika.yml"))
	} else {
		configChoices = append(configChoices, path.Join("/etc", "paprika.yml"))
	}
	return configChoices
}

func createConfig(userPath string) {
	var file *os.File
	fpath := userPath
	var err error
	var errs []error

	if userPath != "" {
		file, err = os.Create(userPath)
		if err != nil {
			log.Fatalf("Error: Failed to create config: %s", err)
		}
	} else {
		configChoices := configPaths()
		for _, v := range configChoices {
			file, err = os.Create(v)
			if err == nil {
				fpath = v
				break
			} else {
				errs = append(errs, err)
			}
		}
		if file == nil {
			log.Print("Error: Failed to create config file.")
			for i := range configChoices {
				log.Printf("- reason: %s", errs[i])
			}
			os.Exit(1)
		}
	}
	defer file.Close()

	if _, err = file.Write(exampleConfig); err != nil {
		log.Fatalf("Error: Failed to create config: %s", err)
	} else {
		log.Printf("- %s", fpath)
		os.Exit(0)
	}
}

func findConfig() *os.File {
	configChoices := configPaths()

	var errs []error
	for _, v := range configChoices {
		file, err := os.Open(v)
		if err == nil {
			return file
		} else {
			errs = append(errs, err)
		}
	}

	log.Print("Error: Could not open any config paths.")
	for i := range configChoices {
		log.Printf("- %s", errs[i])
	}
	os.Exit(1)
	return nil
}
