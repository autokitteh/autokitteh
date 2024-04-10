load("env", "TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN", "TWILIO_MSG_SERVICE_SID")
load("libs/integrations/twilio.star", "twilio")

client = twilio.connect(http, TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN)

def on_http_send(data):
    print(client.create_message(
        to="+19173329447",
        messaging_service_sid=TWILIO_MSG_SERVICE_SID,
        body="meow, world!",
    ))

def on_twilio_webhook(data):
    print(data)
    print(data.body.text())
    print(data.body.form())
