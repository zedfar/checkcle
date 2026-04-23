/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_1738231158")

  // add notified_start field
  collection.fields.add(new Field({
    "hidden": false,
    "id": "bool_notified_start",
    "name": "notified_start",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "bool"
  }))

  // add notified_end field
  collection.fields.add(new Field({
    "hidden": false,
    "id": "bool_notified_end",
    "name": "notified_end",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "bool"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1738231158")

  collection.fields.removeById("bool_notified_start")
  collection.fields.removeById("bool_notified_end")

  return app.save(collection)
})
