from unittest.mock import AsyncMock
from msgraph.generated.models.o_data_errors.o_data_error import ODataError

teams = AsyncMock()
teams.me.joined_teams.get.side_effect = ODataError(message="oops (1)")


async def on_event(_):
    try:
        res = await teams.me.joined_teams.get()
        for t in res.value:
            print(t.display_name, t.id)
    except Exception as err:
        print("ERROR:", err)
        raise (err)
