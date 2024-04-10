// @ts-nocheck

const chatHeader = document.querySelector('.chat-header')
const chatMessages = document.querySelector('.chat-messages')
const chatInputForm = document.querySelector('.chat-input-form')
const chatInputBody = document.querySelector('.chat-input-body')
const chatInputPath = document.getElementById('chat-input-path-actual')
const clearChatBtn = document.querySelector('.clear-chat-button')
const inputPrefix = document.getElementById('path-prefix')
const inputForm = document.querySelector('.chat-input-form')

const addr = document.location.pathname.split('/').slice(-2)[0]
const host = document.location.host
const pathPrefixURL = `http://${host}/http/${addr}/`

const createChatMessageElement = (message) => {
  const t = (message.created_at ? new Date(message.created_at) : new Date()).toLocaleString({ hour: 'numeric', minute: 'numeric', hour12: true })

  const color = message.addr.includes('You ') ? 'gray-bg' : 'blue-bg'

  return `<div class="message ${color}">
      <div class="message-sender">${message.addr} <span class="message-id">${message.id ? message.id : ''}</span></div>
      <div class="message-text">${message.text}</div>
      <div class="message-timestamp">${t}</div>
    </div>`
}

const get = async () => {
  if (window.location.protocol === 'file:') {
    return []
  }

  try {
    return fetch('msgs').then(resp => resp.json())
  } catch (error) {
    console.error(error)
  }
}

window.onload = async () => {
  inputPrefix.innerText = pathPrefixURL

  last_msg_timestamp = null

  const poll = async () => {
    const msgs = await get()
    for (msg of msgs) {
      const t = new Date(msg.created_at)
      if (last_msg_timestamp && t <= last_msg_timestamp) {
        continue
      }

      last_msg_timestamp = t

      newMessage(msg)
    }
  }

  poll()

  setInterval(poll, 1000)
}

const newMessage = (msg) => {
  /* Add message to DOM */
  chatMessages.innerHTML += createChatMessageElement(msg)

  /* Clear input field */
  chatInputForm.reset()

  /*  Scroll to bottom of chat messages */
  chatMessages.scrollTop = chatMessages.scrollHeight
}

const sendMessage = async (e) => {
  e.preventDefault()

  if (window.location.protocol === 'file:') {
    console.log('Cannot send messages in file protocol')
    return
  }

  const body = chatInputBody.value

  const url = pathPrefixURL + chatInputPath.value

  try {
    const resp = await fetch(url, {
      method: 'POST',
      body: body,
    })

    const rtext = await resp.text()

    newMessage({
      addr: `You → <span class="like-pre">${url}</span>`,
      text: `Sent: <span class="like-pre">${body}</span><br/>Got <span class="like-pre">${resp.statusText}</span>`,
      created_at: new Date(),
      id: '',
    })
  } catch (error) {
    console.error(error)
  }
}

chatInputForm.addEventListener('submit', sendMessage)

clearChatBtn.addEventListener('click', async () => {
  try {
    console.log(await fetch('msgs', {method: 'DELETE'}))

    chatMessages.innerHTML = ''
  } catch (error) {
    console.error(error)
  }
})
