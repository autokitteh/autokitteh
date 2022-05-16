package manifest

#Memo: [string]: string

#Account: {
	name:      string
	disabled?: bool
	memo:      #Memo
}

#Accounts: [...#Account]

#EventSource: {
	id:        string
	disabled?: bool
	types?: [...string]
}

#EventSources: [...#EventSource]

#Plugin: {
	id:        string
	address?:  string
	port?:     number
	disabled?: bool
	exec?: {
		name: string
	}
}

#Plugins: [...#Plugin]

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
	id:           string
	name:         string
	account_name: string
	main_path:    string
	disabled?:    bool
	memo:         #Memo
	plugins:      #ProjectPlugins
	src_bindings: #ProjectSourceBindings
	Predecls: [string]: string
}

#Projects: [...#Project]

#Manifest: {
	accounts:  #Accounts
	projects:  #Projects
	eventsrcs: #EventSources
	plugins:   #Plugins
}
