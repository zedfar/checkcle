/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("users")

  // Add isActive bool field
  collection.fields.add(new Field({
    "hidden": false,
    "id": "bool_is_active",
    "name": "isActive",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "bool"
  }))

  app.save(collection)

  // Set isActive = true on all existing users so nobody gets locked out
  const existing = app.findRecordsByFilter("users", "id != ''", "-created", 500, 0)
  existing.forEach((record) => {
    record.set("isActive", true)
    app.save(record)
  })

}, (app) => {
  const collection = app.findCollectionByNameOrId("users")
  collection.fields.removeById("bool_is_active")
  return app.save(collection)
})
