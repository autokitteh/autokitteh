define([], () => {
  return {
    modules: [
      {
        name: 'clear',
        aliases: ['cls'],
        help: `~Command Help
              ~~Command: clear
              ~Alias:   cls
              ~Details: resets the terminal's buffer and clear the screen.
              ~Usage:   clear~`,
        run: async function() {
          this.terminal.clear()
        }
      },

      {
        name: 'help',
        help: `~Command Help
              ~~Command: help
              ~Details: displays all registered commands on the terminal or display a specific command info.
              ~Usage:   help [command]~`,
        error_codes: {
          H01: {
            code: 'H01',
            details: 'cannot find command or command doesn\'t have a manual...'
          }
        },

        run: async function(args) {
          const terminal = this.terminal
          const print = (i) => terminal.print(i)
          const newLine = () => terminal.new_line()

          if (!args[0]) {
            newLine()
            print('This is the help page. Here are the registered commands:')
            newLine()
            terminal.get_installed_commands()
              .sort()
              .forEach(command => print(`${command}`))
            newLine()
            print(`Type 'help [command]' for more info`)
            newLine()
          } else {
            try {
              const help = await terminal.get_command_help(args[0])
              const lines = help
                .split('~')
                // get rid of extrenous nls/crs.
                .map(line => line.replace(/\r?\n|\r/, ''))
                .forEach(line => {
                  if (line) {
                    print(line)
                  } else {
                    newLine()
                  }
                })
            } catch (err) {
              throw (this.error_codes.H01)
            }
          }
        }
      },

    ]
  }
})
