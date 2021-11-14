package plugins

type Plugin interface {
	Triggers() []string
	Execute() error
}

var Plugins = make(map[string]Plugin)

func Register(p Plugin) {
	for _, t := range p.Triggers() {
		Plugins[t] = p
	}
}

func ProcessCommands(cmd string) error {
	for trigger, plugin := range Plugins {
		if cmd == trigger {
			err := plugin.Execute()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
