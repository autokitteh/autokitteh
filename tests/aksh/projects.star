load("assert", "assert")

store = projects.open(getenv("DSN"))

astore = accounts.open(getenv('DSN'))
aid = astore.create(data={'name':'test'})

id0 = store.create(account_id=aid, data={"name":"0", "main_path": projects.path("file:/tmp"), "enabled": True})
id1 = store.create(account_id=aid, data={"name":"1", "main_path": projects.path("file:/tmp"), "enabled": True})
id2 = store.create(account_id=aid, data={"name":"2", "main_path": projects.path("file:/tmp"), "enabled": True})

p0 = store.get(id0)
assert.eq(p0.data.name, "0")
assert.eq(p0.data.main_path, projects.path("file:/tmp"))
assert.true(p0.data.enabled)

p1 = store.get(id1)
assert.eq(p1.data.name, "1")
assert.eq(p1.data.main_path, projects.path("file:/tmp"))
assert.true(p1.data.enabled)

p2 = store.get(id2)
assert.eq(p2.data.name, "2")
assert.eq(p2.data.main_path, projects.path("file:/tmp"))
assert.true(p2.data.enabled)

store.update(id0, data={"enabled": False})

p0 = store.get(id0)
assert.eq(p0.data.name, "0")
assert.eq(p0.data.main_path, projects.path("file:/tmp"))
assert.true(not p0.data.enabled)

assert.eq(store.get("P0"), None)

pp = store.batch_get([id0, id1, id2, "P0"])
assert.eq(pp, {id0: p0, id1: p1, id2: p2, "P0": None})
