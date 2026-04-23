/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("alert_configurations");

  const field = new Field({
    "id": "text_telegram_thread_id",
    "name": "telegram_thread_id",
    "type": "text",
    "required": false,
    "presentable": false,
    "hidden": false,
    "max": 20,
    "min": 0,
    "pattern": "^[0-9]*$",
    "autogeneratePattern": ""
  });

  collection.fields.add(field);
  return app.save(collection);
}, (app) => {
  const collection = app.findCollectionByNameOrId("alert_configurations");
  const field = collection.fields.getByName("telegram_thread_id");
  collection.fields.remove(field.id);
  return app.save(collection);
});
