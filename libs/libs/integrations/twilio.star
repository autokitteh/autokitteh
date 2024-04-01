_BASE_URL = "https://api.twilio.com/2010-04-01"

def connect(conn, account_sid, auth_token=None):
    """Connect to the Twilio API"""

    def _url(path):
        return _BASE_URL + path

    def _post(path, body):
        resp = conn.post(
            _url(path), 
            form_body=body, 
            basic_auth=auth_token and (account_sid, auth_token) or None,
        )
        
        json = resp.body.json()

        if resp.status_code < 200 or resp.status_code >= 400:
            fail(json)
            
        return json

    def create_message(to, from_number=None, messaging_service_sid=None, body=None, media_url=None, content_sid=None):
        """Create a new message.
        
        https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource
        """

        if not from_number and not messaging_service_sid:
            fail("either from_number or messaging_service_sid is required")
        
        if not body and not media_url and not content_sid:
            fail("either body, media_url, or content_sid is required")

        req_body = {
            "To": to,
        }

        if from_number:
            req_body["From"] = from_number

        if messaging_service_sid:
            req_body["MessagingServiceSid"] = messaging_service_sid

        if body:
            req_body["Body"] = body

        if media_url:
            req_body["MediaUrl"] = media_url

        if content_sid:
            req_body["ContentSid"] = content_sid

        return _post(
            "/Accounts/{}/Messages.json".format(account_sid),
            req_body,
        )

    return struct(
        create_message=create_message,
    )
        
twilio = struct(
    connect=connect,
)
