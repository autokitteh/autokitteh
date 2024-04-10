define(
  ['utilities/htmlDom', 'utilities/styles'],
  function(dom, styles) {
    function Terminal() {}

    Terminal.prototype = (function(dom, styles) {
      let config = {}
      let process_stack = []

      let states = {
        curr_buffer: 0,
        curr_history: 0,
        history: [],
      }

      /**
       * The HTML elements used by the terminal.
       */
      let components = {
        main_stdin_cont: dom.getById('input-cont'),
        main_stdin: dom.getById('main-stdin'),
        curr_stdin: null,
        prompt_msg: dom.getById('main-stdin-msg'),
        prompt_symbol: dom.getById('main-stdin-symbol'),
        bottom_offset: dom.getById('bottom-offset'),
        stdout: dom.getById('stdout'),
      }

      /**
       * Contains the all the commands included in the terminal.
       */
      let commands = {
        installed: {},
        aliases: [],

        valid: function(command) {
          if (this.installed[command]) {
            return true
          } else {
            return this.aliases.filter(alias => alias.split('=')[0] === command).length !== 0
          }
        },

        help: function(command) {
          return this.installed[command].help
        },

        get: function(command) {
          let search = this.installed[command]

          if (!search) {
            search = this.aliases.filter(alias => alias.split('=')[0] === command)[0].split('=')[1]
            search = this.installed[search]
          }

          return search
        }
      }

      /**
       * Sets the value of a key in the configuration. Updates the CSS variables and reload the main
       * STDIN.
       * @param {String} key - Configuration key to be changed.
       * @param {String|Number} value - New value of the key specified.
       */
      async function set_config(key, value) {
        config[key] = value
        styles.setVar(`--config-${key}`, value)
        reload_main_stdin()
      }

      /**
       * Get the names of the installed commands in the terminal.
       */
      const get_installed_commands = () => Object.keys(commands.installed)

      /**
       * Gets the help/manual of the specified command.
       * @param {String} command - Name of the command
       */
      const get_command_help = (command) => commands.help(command) || Promise.reject()

      /**
       * Gets the terminal's configuration
       */
      const get_config = () => config

      /**
       * Must be run first before doing anything with the terminal. Retrieves the configuration in 
       * Local Storage, if there is any, and setups the main STDIN.
       */
      async function setup() {
        show_header()

        config = {
          'prompt-symbol-color': styles.getVar('--config-prompt-symbol-color'),
          'error-source-color': styles.getVar('--config-error-source-color'),
          'info-source-color': styles.getVar('--config-info-source-color'),
          'prompt--color': styles.getVar('--config-prompt-msg-color'),
          'error-code-color': styles.getVar('--config-error-code-color'),
          'margin-sides': styles.getVar('--config-margin-sides'),
          'font-weight': styles.getVar('--config-font-weight'),
          'background': styles.getVar('--config-background'),
          'foreground': styles.getVar('--config-foreground'),
          'font-size': styles.getVar('--config-font-size'),
          'secondary-color': '#608460',
          'prompt-symbol': '$',
          'prompt-msg': '',
          'max-history': 10,
          'max-buffer': 50,
          'tab-size': 2
        }

        await run_intro()

        show_main_stdin()
      }

      /**
       * Changes the current used STDIN and any click events will focus it. If the STDIN is the main
       * STDIN, it will add the correct keyboard events.
       * @param {Input} stdin - Input element that the terminal will focus on.
       */
      function change_curr_stdin(stdin) {
        if (components.curr_stdin) {
          components.curr_stdin.disabled = true
        }

        components.curr_stdin = stdin

        if (stdin) {
          document.onclick = () => {
            components.curr_stdin.focus()
            components.curr_stdin.click()
          }

          if (stdin === components.main_stdin) {
            components.main_stdin.disabled = false
            document.onkeydown = (e) => {
              if (e.key === 'l' && e.ctrlKey) {
                e.preventDefault()
                clear()
              } else if (e.key === 'Enter' && !e.shiftKey && !e.ctrlKey) {
                run_command(components.curr_stdin.value, false).then(() => show_main_stdin())
              } else if (e.key === 'Tab') {
                e.preventDefault()
                e.target.value = `${e.target.value}${' '.repeat(config['tab-size'])}`
              } else if (e.key === 'ArrowUp') {
                e.preventDefault()
                e.target.value = history_up() || e.target.value
              } else if (e.key === 'ArrowDown') {
                e.preventDefault()
                e.target.value = history_down() || e.target.value
              }
            }
          } else {
            document.onkeydown = () => {}
          }

          components.curr_stdin.focus()
        } else {
          document.onclick = () => {}
          document.onkeydown = () => {}
        }
      }

      /**
       * Checks first if the command is valid. If valid, runs a Promise Race with the run method of
       * the specified command and the cancel Promise, which will resolve if 'CTRL+C' is pressed. If
       * the command throws an error or reject, display it in the terminal.
       * @param {String} input - Input from the main STDIN.
       * @param {Boolean} quiet - If true, do not display a copy of the command line input.
       */
      async function run_command(input, quiet) {
        if (!quiet) {
          display_command(input)
        }

        add_to_history(input)

        input = input.trim().split(' ')
        const command = input.shift()

        const cancel_command = () => new Promise(resolve => {
          const key_handler = (e) => {
            if (e.ctrlKey && (e.key === 'c' || e.key === 'C')) {
              document.removeEventListener('keydown', key_handler)

              return resolve()
            }
          }

          document.addEventListener('keydown', key_handler, false)
        })

        const remove_cancel_listener = () => (
          document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c', ctrlKey: true }))
        )

        if (commands.valid(command)) {
          hide_main_stdin()

          await Promise.race([commands.get(command).run(input), cancel_command()])
            .catch((err) => {
              console.log(err)
              error(command, err.code, err.details)
            })
            .finally(() => {
              remove_cancel_listener()
              show_main_stdin()
            })
        } else if (command === '') {
          return
        } else {
          error('terminal', null, 'unrecognized command')
        }
      }

      /**
       * Add the input to the terminal's history state.
       * @param {String} input - Input from the main stdin.
       */
      function add_to_history(input) {
        states.history.unshift(input)
        states.curr_history = -1
        if (states.history.length > config['max-history']) {
          states.history.pop()
        }
      }

      /**
       * Cycles down in the terminal's history
       */
      function history_down() {
        if (states.curr_history - 1 >= 0) {
          states.curr_history -= 1
          return states.history[states.curr_history]
        } else if (states.curr_history === -1) {
          states.curr_history = 1
          return states.history[0]
        } else {
          return states.history[0]
        }
      }

      /**
       * Cycles up in the termina's history
       */
      function history_up() {
        if (states.curr_history + 1 < states.history.length) {
          states.curr_history += 1
          return states.history[states.curr_history]
        } else {
          return states.history[states.history.length - 1]
        }
      }

      /**
       * Installs the package and its command modules. If a module contains aliases add it to the
       * aliases directory.
       * @param {Terminal} terminal - Terminal where to install the module
       * @param {Object} package - Command package that contains one or more command.
       */
      function install(terminal, package) {
        package.modules.forEach(module => {
          module.terminal = terminal

          if (module.aliases) {
            module.aliases.forEach(alias => {
              commands.aliases.push(`${alias}=${module.name}`)
            })
          }

          commands.installed = { ...commands.installed, [`${module.name}`]: module }
        })
      }

      /**
       * Clears the STDOUT by removing all children node.
       */
      function clear() {
        const stdout = components.stdout

        while (stdout.firstChild)
          stdout.removeChild(stdout.firstChild)

        states.curr_buffer = 0
      }

      /**
       * Prints the given text to the STDOUT.
       * @params {String} text - The text to be printed.
       */
      function print(text) {
        const line = dom.create('pre')
        line.classList.add('line')
        line.appendChild(dom.text(text))
        append_stdout(line)
        components.bottom_offset.scrollIntoView()
      }

      /**
       * Adds a empty line in the STDOUT
       */
      function new_line() {
        const line = dom.create('pre')
        line.appendChild(dom.text(' '))
        line.classList.add('line')
        append_stdout(line)
        components.bottom_offset.scrollIntoView()
      }

      /**
       * delete the number of lines specified.
       */
      function delete_line(lines = 1) {
        if (typeof lines !== 'number') {
          throw ({ code: null, details: '[delete_line] argument provided is not of type "Number"' })
        }

        try {
          lines += 1

          while (lines-- > 0) {
            components.stdout.removeChild(components.stdout.lastChild)
          }
        } catch (ignore) {
          // This catch block will be used if STDOUT doesn't have any child nodes.
          return
        }
      }

      /**
       * Displays terminal header on startup.
       */
      function show_header() {
        new_line()
      }

      /**
       * Ask for the prompt's name and preferred prompt symbol.
       */
      async function run_intro() {
        await set_config('prompt-symbol', '=^.^=')

        new_line()
        print(`Type 'help' to see a list of available commands`)
        new_line()
      }

      /**
       * Creates an input element and wait for prompt input. Can have a label if specified.
       * @params {String} label_next - Label for the input element.
       */
      function read_line(label_text) {
        hide_main_stdin()
        return new Promise((resolve, reject) => {
          const line = dom.create('div')
          const input = dom.create('input')
          let label

          if (label_text) {
            label = dom.create('pre')
            label.appendChild(dom.text(label_text))
            label.classList.add('label')
            line.appendChild(label)
          }

          input.classList.add('stdin')
          line.classList.add('command')
          line.appendChild(input)

          function submit(e) {
            if (e.key === 'Enter') {
              e.preventDefault()
              this.removeEventListener('keydown', submit)
              return resolve(this.value)
            }
          }

          input.addEventListener('keydown', submit)

          append_stdout(line)
          change_curr_stdin(input)
        })
      }

      /**
       * Reads the next prompt input. Only reads Alphanumeric characters.
       * @params {String} label_text - Label for the input element.
       */
      function read_char(label_text) {
        hide_main_stdin()
        return new Promise((resolve, reject) => {
          const line = dom.create('div')
          const input = dom.create('input')
          let label

          if (label_text) {
            label = dom.create('pre')
            label.appendChild(dom.text(label_text))
            label.classList.add('label')
            line.appendChild(label)
          }

          input.classList.add('stdin')
          line.classList.add('command')
          line.appendChild(input)

          function submit(e) {
            if (/^[a-zA-Z0-9]$/.test(e.key)) {
              return resolve(e.key)
            } else {
              e.preventDefault()
            }
          }

          input.addEventListener('keydown', submit)

          append_stdout(line)
          change_curr_stdin(input)
        })
      }

      /**
       * Open a file chooser for uploading files.
       * @param {Boolean} multiple - If true, allow prompt to select multiple files.
       * @param {Array} file_types - Array of file types the prompt can select.
       */
      function read_files(multiple, file_types) {
        return new Promise((resolve, reject) => {
          const input = dom.create('input')
          input.value = null
          input.type = 'file'
          input.multiple = multiple || false
          input.accept = file_types ? file_types.join(',') : ''

          input.onclick = () => {
            function check_files() {
              setTimeout(() => {
                if (input.files.length === 0) {
                  return resolve(null)
                } else {
                  return resolve(input.files)
                }
              }, 250)
            }

            window.addEventListener('focus', check_files, { once: true })
          }

          input.click()
          change_curr_stdin(input)
        })
      }

      /**
       * Reads prompt input from multiple lines using HTML Textarea. When 'Shift + Enter' return input.
       * @param {String} text - Initial text of the textarea.
       */
      function read_textarea(text) {
        hide_main_stdin()
        return new Promise((resolve, reject) => {
          const textarea = dom.create('textarea')

          textarea.setAttribute('spellcheck', false)

          if (text) {
            textarea.value = text
          }

          resize = (target) => setTimeout(() => {
            target.style.height = 'auto'
            target.style.height = `${target.scrollHeight}px`
            components.bottom_offset.scrollIntoView()
          }, 0)

          function autosize(e) {
            if (e.key === 'Enter' && e.shiftKey) {
              e.preventDefault()
              return resolve(this.value)
            } else if (e.key === 'Tab') {
              e.preventDefault()
              this.value = `${this.value}${' '.repeat(config['tab-size'])}`
            } else {
              resize(e.target)
            }
          }

          textarea.addEventListener('keydown', autosize)
          print('Press Shift+Enter to submit...')
          new_line()
          append_stdout(textarea)
          resize(textarea)
          change_curr_stdin(textarea)
        })
      }

      /**
       * Display the error source, error code (If specified), and the error details in the terminal.
       * @param {String} source - Command that invoked an error.
       * @param {String} code - Error code from the source.
       * @param {details} details - Info about the error.
       */
      function error(source, code, details) {
        const err_cont = dom.create('div')
        const err_source = dom.create('pre')
        const err_details = dom.create('pre')

        err_cont.appendChild(err_source)
        err_source.appendChild(dom.text(`[${source}]`))
        err_source.classList.add('error-source')
        err_source.classList.add('line')
        err_details.classList.add('line')

        if (code) {
          const err_code = dom.create('pre')
          err_code.classList.add('error-code')
          err_code.classList.add('line')
          err_code.appendChild(dom.text(`[Error Code: ${code}]`))
          err_cont.appendChild(err_code)
          err_cont.classList.add('error-with-code')
        } else {
          err_cont.classList.add('error')
        }

        err_cont.appendChild(err_details)
        err_details.appendChild(dom.text(details))

        append_stdout(err_cont)
      }

      function info(source, details) {
        const cont = dom.create('div')
        const info_source = dom.create('pre')
        const info_details = dom.create('pre')

        cont.appendChild(info_source)
        info_source.appendChild(dom.text(`[${source}]`))
        info_source.classList.add('info-source')
        info_source.classList.add('line')
        info_details.classList.add('line')

        cont.classList.add('info')

        cont.appendChild(info_details)
        info_details.appendChild(dom.text(details))

        append_stdout(cont)
      }

      /**
       * Display the recent input from the main STDIN.
       * @param {String} command - Input from the command line.
       */
      function display_command(command) {
        const display = dom.create('div')
        const input = dom.create('pre')
        const prompt = dom.create('div')
        const prompt_msg = dom.create('span')
        const symbol = dom.create('span')

        prompt.classList.add('prompt')
        prompt.appendChild(prompt_msg)
        prompt.appendChild(symbol)
        prompt_msg.classList.add('prompt-msg')

        const msg = config['prompt-msg'];
        if (msg) {
          prompt_msg.appendChild(dom.text(`[${msg}]`))
        }

        symbol.classList.add('prompt-symbol')
        symbol.appendChild(dom.text(config['prompt-symbol']))
        input.appendChild(dom.text(command))
        input.classList.add('line')
        display.classList.add('command')
        display.appendChild(prompt)
        display.appendChild(input)
        append_stdout(display)
      }

      /**
       * Append the given HTML node to the STDOUT
       * @param {DOM Element} node - Element to be appended.
       */
      function append_stdout(node) {
        components.stdout.appendChild(node)
        if (++states.curr_buffer > config['max-buffer'])
          components.stdout.removeChild(stdout.firstChild)
      }

      /**
       * Hides the main STDIN and remove all keyboard events attached to it.
       */
      function hide_main_stdin() {
        components.main_stdin_cont.classList.add('hide-stdin')
        change_curr_stdin(null)
      }

      /**
       * Shows the main STDIN and re-enables all keyboard events attached to it.
       */
      function show_main_stdin() {
        components.main_stdin_cont.classList.remove('hide-stdin')
        components.main_stdin.value = ''
        components.bottom_offset.scrollIntoView()
        setTimeout(() => change_curr_stdin(components.main_stdin))
      }

      /**
       * Checks if there were any changes made to the configuration (prompt_msg & prompt_symbol).
       */
      function reload_main_stdin() {
        const msg = config['prompt-msg'];
        if (msg) {
          components.prompt_msg.innerText = `[${config['prompt-msg']}]`;
        }

        components.prompt_symbol.innerText = config['prompt-symbol']
      }

      const add_process = (process) => process_stack.push({ ...process, PID: process_stack.length + 1 })

      const end_process = (name) => {
        const _process = get_process(name)
        process_stack = process_stack.filter(p => p !== _process)
      }

      const get_process = (name) => process_stack.find(p => p.name === name)

      const get_all_process = () => [...process_stack]

      return {
        constructor: Terminal,

        /* stdout methods */
        clear: clear,
        print: print,
        new_line: new_line,
        delete_line: delete_line,

        info: info,
        error: error,

        /* stdin methods */
        read_textarea: read_textarea,
        read_line: read_line,
        read_char: read_char,
        read_files: read_files,

        /* terminal methods */
        setup: setup,
        install: function(module) { install(this, module) },
        run_command: run_command,

        /* commands */
        get_installed_commands: get_installed_commands,
        get_command_help: get_command_help,

        /* config */
        get_config: get_config,
        set_config: set_config,

        /* process */
        add_process: add_process,
        end_process: end_process,
        get_process: get_process,
        get_all_process: get_all_process,

        test: commands
      }
    })(dom, styles)

    return Terminal
  }
)
