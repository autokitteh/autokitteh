from flask import Flask, request, redirect
from os import getenv

from stringcase import snakecase # type: ignore

from autokitteh.api import Client, ProjectID
from autokitteh.eventsrc import EventSource, init as init_eventsrc


_src = init_eventsrc('twilio')

app = Flask(__name__)


@app.route('/twiliosrc/message', methods=['POST'])
def on_message() -> str:
    # TODO: validate request as authentic from twilio.

    account_sid = request.form['AccountSid']
    message_sid = request.form['MessageSid']

    data = {snakecase(k): v for k, v in request.form.items()}

    event_id = _src.send(
        assoc=account_sid,
        type_='message',
        data=data,
        orig_id=f'{account_sid}/{message_sid}',
        memo={},
    )

    app.logger.debug(f'received message {account_sid}/{message_sid}, sent event {event_id}')

    # TODO: might configure some bindings to wait for response from script and
    # then send it back to user.

    return ''


@app.route('/twiliosrc/bind', methods=['POST'])
def on_bind() -> tuple[str, int]:
    project_id = request.args.get('project_id')
    account_sid = request.args.get('account_sid')
    approved = request.args.get('approved') == '1'

    if not project_id:
        return 'missing project_id', 400

    if not account_sid:
        return 'mising account_sid', 400

    _src.bind(
        name='twilio',
        project_id=ProjectID(project_id),
        assoc=account_sid,
        approved=approved,
    )

    return 'bind request sent', 200
