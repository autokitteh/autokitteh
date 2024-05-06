# Life of an Event

1. For example, an external provider sends some data to an integration.
2. The integration processes this data.
3. The integration needs to find out which connection IDs to send the data to. For this, it can query the **Vars** service for connection IDs that has a specific variable set with an optional specific value. This is done by the `FindConnectionIDs` method.
4. The connection contacts the `Dispatcher` service to disptach the event.
5. The dispatcher persists the event in the Database, and triggers a workflow to process the event. It also returns synchronously the event ID generated for the event.
6. The dispatcher workflow uses the `Connections` and `Triggers` services to find out which project and environment to launch a session in.
7. The dispatcher creates a session using the `Sessions` service.
