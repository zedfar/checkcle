/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const email    = $os.getenv("PB_ADMIN_EMAIL");
  const password = $os.getenv("PB_ADMIN_PASSWORD");

  if (!email || !password) {
    return;
  }

  const fullName = $os.getenv("PB_ADMIN_FULL_NAME") || "Administrator";
  const username = $os.getenv("PB_ADMIN_USERNAME")  || "admin";

  let record;
  let exists = false;

  try {
    record = app.findAuthRecordByEmail("_superusers", email);
    exists = true;
  } catch(e) {
    exists = false;
  }

  if (!exists) {
    const col = app.findCollectionByNameOrId("_superusers");
    record = new Record(col);
    record.set("email", email);
    record.setPassword(password);
  }

  // Always ensure these fields are set (upsert behaviour)
  record.set("full_name", fullName);
  record.set("username", username);
  record.set("status", "active");

  app.save(record);

}, (app) => {
  const email = $os.getenv("PB_ADMIN_EMAIL");
  if (!email) return;
  try {
    const record = app.findAuthRecordByEmail("_superusers", email);
    app.delete(record);
  } catch(e) {}
});
