define([], () => {
  last_msg_timestamp = null

  return {
    modules: [
      {
        name: "http",
        help: `~Command Help
              ~~Command: http
              ~Details: send an http request
              ~Usage:   http <method> <url> [body]~`,
        run: async function(args) {
          const terminal = this.terminal

          if (window.location.protocol === 'file:') {
            // when testing locally (ie, just a local file in browser), fetch is not allowed.
            terminal.error('connection', '', 'fetch not allowed in file protocol')
            return
          }

          const [method, url] = args.slice(0, 2)
          const rest = args.slice(2)
          const body = rest.length ? rest.join(' ') : undefined

          if (!method || !url) {
            terminal.error('http', '', 'missing method or url')
            return
          }

          try {
            const resp = await fetch(url, { method, body })
            terminal.print(resp.statusText)
          } catch (error) {
            terminal.error('http', '', `${error}`)
          }
        },
      },

      {
        name: 'messages',
        aliases: ['ms', 'msgs'],
        help: `~Command Help
              ~~Command: messages
              ~Alias:   ms, msgs
              ~Details: displays all messages.
              ~Usage:   messages [list|delete {<id>|all}]~`,
        run: async function(args) {
          const terminal = this.terminal

          if (window.location.protocol === 'file:') {
            // when testing locally (ie, just a local file in browser), fetch is not allowed.
            terminal.error('connection', '', 'fetch not allowed in file protocol')
            return
          }

          const get = async () => {
            return fetch('msgs').then(resp => resp.json())
          }

          const emit = (msg, breaking) => {
            const t = new Date(msg['created_at'])
            if (last_msg_timestamp && t <= last_msg_timestamp) {
              return
            }

            last_msg_timestamp = t
            const txt = `${msg['id']} [${t.toLocaleString({ hour: 'numeric', minute: 'numeric', hour12: true })}] ${msg['text']}`
            if (breaking) {
              terminal.info('message', txt)
            } else {
              terminal.print(txt)
            }
          }

          switch (args[0] ? args[0] : 'list') {
            case 'wait': { // intentionally not listed in help. this is ran at init.
              err = false

              while (true) {
                try {
                  const msgs = await get()
                  if (err) {
                    err = false
                    terminal.info('connection', 'connection restored')
                  }

                  for (const msg of msgs) emit(msg, true)
                } catch (error) {
                  if (!err) {
                    err = true
                    terminal.error('connection', '', `${error}`)
                  }
                }

                await new Promise(r => setTimeout(r, 1000))
              }
            }

            case 'list': {
              last_msg_timestamp = null
              try {
                const msgs = await get()
                for (const msg of msgs) emit(msg, false)
              } catch (error) {
                terminal.error('connection', '', `${error}`)
                return
              }
              break
            }

            case 'delete': {
              const id = args[1]
              if (!id) { 
                terminal.print('missing message id')
                return
              }

              try {
                const resp = await fetch(`msgs/${id}`, {method: 'DELETE'})        
                terminal.print(resp.statusText)        
              } catch (error) {
                terminal.error('connection', '', `${error}`)
                return
              }

              break
            }

            default:
              terminal.error('msgs', '', 'invalid command')
          }
        }
      }
    ]
  }
})
