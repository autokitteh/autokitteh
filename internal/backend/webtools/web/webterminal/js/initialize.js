require(['terminal', 'commands/internal', 'commands'], (Terminal, internal, commands) => {
  const terminal = new Terminal()

  terminal.setup()
  terminal.install(internal)
  terminal.install(commands)
  terminal.run_command('ms wait', true)
})
