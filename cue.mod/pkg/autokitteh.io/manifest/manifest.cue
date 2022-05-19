package manifest

#Memo: [string]: string

#Account: {
	name?:     string
	disabled?: bool
	memo:      #Memo
}

#Accounts: [string]: #Account

#EventSource: {
	id?:       string
	disabled?: bool
	types?:    [...string]
}

#EventSources: [string]: #EventSource

#Plugin: {
	id?:       string
	address?:  string
	port?:     number
	disabled?: bool
	exec?: {
		name: string
	}
}

#Plugins: [string]: #Plugin

#ProjectPlugin: {
	disabled?: bool
}

#ProjectPlugins: [string]: #ProjectPlugin

#ProjectSourceBinding: {
	disabled?:   bool
	src_id:      string
	assoc?:      string
	src_config?: string
}

#ProjectSourceBindings: [string]: #ProjectSourceBinding

#Project: {
	id?:           string
	name?:         string
	account_name?: string
	main_path:     string
	disabled?:     bool
	memo:          #Memo
	plugins:       #ProjectPlugins
	src_bindings:  #ProjectSourceBindings
	Predecls: [string]: string
}

#Projects: [string]: #Project

#Manifest: {
	accounts:  #Accounts
	projects:  #Projects
	eventsrcs: #EventSources
	plugins:   #Plugins
}
