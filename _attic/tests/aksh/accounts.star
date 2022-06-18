load('assert', 'assert')

store = accounts.open(getenv('DSN'))
id0 = store.create(data={'name':'0'})
id1 = store.create(data={'name':'1'})
id2 = store.create(data={'name':'2'})

a0 = store.get(id0)
assert.eq(a0.id, id0)
assert.eq(a0.data.name, '0')

a1 = store.get(id1)
assert.eq(a1.id, id1)
assert.eq(a1.data.name, '1')

a2 = store.get(id2)
assert.eq(a2.id, id2)
assert.eq(a2.data.name, '2')

a3 = store.get('A0')
assert.eq(a3, None)

aa = store.batch_get([id0, id1, id2, 'A0'])
assert.eq(aa, {id0: a0, id1: a1, id2: a2, 'A0': None})
